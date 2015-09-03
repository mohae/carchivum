// Package carchivum works with compressed archives.
//
// Go's `archive` package, supports `tar` and `zip`
// Go's `compress` package supports: bzip2, flate, gzip, lzw, zlib
//
// Carchivum supports zip and tar. For tar, archiver also supports
// the following compression:
//
// When using archiver, compression is not optional.
//
package carchivum

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/mohae/magicnum"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

var unsetTime time.Time
var CreateDir bool
var defaultFormat = magicnum.Gzip

// default max random number for random number generation.
var MaxRand = 10000

// we assume this count isn't going to change during runtime
var cpu = runtime.NumCPU()

// Arbitrarily set the multiplier to some default value.
var CPUMultiplier = 4

// Car is a Compressed Archive. The struct holds information about Cars and
// their processing.
type Car struct {
	sync.Mutex
	// Name of the archive, this includes path information, if any.
	Name       string
	UseLongExt bool
	OutDir     string
	// Create operation modifiers
	Owner int
	Group int
	os.FileMode
	// Extract operation modifiers
	UseFullpath bool
	// Local file selection
	// List of files to delete if applicable.
	deleteList     []string
	DeleteArchived bool
	// Exclude file processing
	Exclude         string
	ExcludeExt      []string
	ExcludeExtCount int
	ExcludeAnchored string
	// Include file processing
	Include         string
	IncludeExt      []string
	IncludeExtCount int
	IncludeAnchored string
	// File time format handling
	Newer      string
	NewerMTime time.Time
	NewerFile  string
	//	TimeFormats []string

	// Output format for time
	outputNameTimeFormat string
	// Processing queue
	FileCh chan *os.File
	// Other Counters
	files           int32
	dirs            int32
	bytes           int64
	compressedBytes int64
	// timer
	t0 time.Time
	𝛥t float64
}

func (c *Car) setDelta() {
	c.𝛥t = float64(time.Since(c.t0)) / 1e9
}

// Delta returns the time delta between operation start and end.
func (c *Car) Delta() float64 {
	return c.𝛥t
}

// Message provides summary information about the operation performed.
func (c *Car) Message() string {
	return fmt.Sprintf("%q created in %4f seconds\n%d files totalling %d bytes were processed", c.Name, c.𝛥t, c.files, c.bytes)
}

// AddFile reads a file and pipes it to the zipper goroutine.
func (c *Car) AddFile(root, p string, fi os.FileInfo, err error) error {
	// Check fileInfo to see if this should be added to archive
	process, err := c.filterFileInfo(fi)
	if err != nil {
		return err
	}
	if !process {
		return nil
	}
	// Check path information to see if this should be added to archive
	process, err = c.filterPath(root, p)
	if err != nil {
		return err
	}
	if !process {
		return nil
	}
	var relPath string
	relPath, err = filepath.Rel(root, p)
	if err != nil {
		log.Print(err)
		return err
	}
	if relPath == "," {
		return nil
	}
	fullpath := p
	if !c.UseFullpath {
		p = filepath.Join(filepath.Base(root), relPath)
	}
	f, err := os.Open(p)
	if err != nil {
		log.Print(err)
		return err
	}
	c.Mutex.Lock()
	c.files++
	c.bytes += fi.Size()
	if c.DeleteArchived {
		c.deleteList = append(c.deleteList, fullpath)
	}
	c.FileCh <- f
	c.Mutex.Unlock()
	return nil
}

func (c *Car) filterFileInfo(fi os.FileInfo) (bool, error) {
	// Don't add symlinks, otherwise would have to code some cycle
	// detection amongst other stuff.
	if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		return false, nil
	}
	if c.NewerMTime != unsetTime {
		if !fi.ModTime().After(c.NewerMTime) {
			return false, nil
		}
	}
	return true, nil
}

func (c *Car) filterPath(root, p string) (bool, error) {
	if strings.HasSuffix(root, p) {
		return false, nil
	}
	b, err := c.includeFile(root, p)
	if err != nil {
		return false, err
	}
	if !b {
		return false, nil
	}
	b, err = c.excludeFile(root, p)
	if err != nil {
		return false, err
	}
	if b {
		return false, nil
	}
	return true, nil
}

func (c *Car) includeFile(root, p string) (bool, error) {
	if c.IncludeAnchored != "" {
		if strings.HasPrefix(filepath.Base(c.IncludeAnchored), p) {
			return true, nil
		}
	}
	// since we are just evaluating a file, we use match and look at the
	// fullpath
	if c.Include != "" {
		matches, err := filepath.Match(c.Include, filepath.Join(root, p))
		if err != nil {
			return false, err
		}

		if matches {
			return true, nil
		}
	}
	if c.IncludeExtCount > 0 {
		for _, ext := range c.IncludeExt {
			if strings.HasSuffix(filepath.Base(p), "."+ext) {
				return true, nil
			}
		}
		return false, nil
	}
	return true, nil
}

func (c *Car) excludeFile(root, p string) (bool, error) {
	if c.ExcludeAnchored != "" {
		if strings.HasPrefix(filepath.Base(p), c.ExcludeAnchored) {
			return true, nil
		}
	}
	// since we are just evaluating a file, we use match and look at the
	// fullpath
	if c.Exclude != "" {
		matches, err := filepath.Match(c.Exclude, filepath.Join(root, p))
		if err != nil {
			return false, err
		}

		if matches {
			return true, nil
		}
	}
	if c.ExcludeExtCount != 0 {
		for _, ext := range c.ExcludeExt {
			if strings.HasSuffix(filepath.Base(p), "."+ext) {
				return true, nil
			}
		}
	}
	return false, nil
}

// Extract can handle the processing and extraction of a source file. Dst is
// the destination directory of the output, if a location other than the CWD
// is desired. The source file can be a zip, tar, or compressed tar.
func Extract(dst, src string) error {
	// determine the type of archive
	f, err := os.Open(src)
	if err != nil {
		log.Print(err)
		return err
	}
	// find its format
	format, err := magicnum.GetFormat(f)
	if err != nil {
		log.Print(err)
		return err
	}
	if !isSupported(format) {
		err := fmt.Errorf("%s: %s is not a supported format", src, format)
		log.Print(err)
		return err
	}
	if format == magicnum.Zip {
		// close the file, the zip reader will open it
		f.Close()
		zip := NewZip(src)
		zip.OutDir = dst
		err := zip.Extract()
		if err != nil {
			log.Print(err)
		}
		return err
	}
	defer f.Close()
	tar := NewTar(src)
	tar.OutDir = dst
	tar.Format = format
	err = tar.ExtractArchive(f)
	if err != nil {
		log.Print(err)
	}
	return err
}

//func formattedNow() string {
//	return time.Now().Local().Format()
//}

func getFileParts(s string) (dir, filename, ext string, err error) {
	// see if there is path involved, if there is, get the last part of it
	dir, fname := filepath.Split(s)
	parts := strings.Split(fname, ".")
	l := len(parts)
	switch l {
	case 2:
		filename := parts[0]
		ext := parts[1]
		return dir, filename, ext, nil
	case 1:
		filename := parts[0]
		return dir, filename, ext, nil
	case 0:
		err := fmt.Errorf("no destination filename found in %s", s)
		log.Print(err)
		return dir, filename, ext, err
	default:
		// join all but the last parts together with a "."
		filename := strings.Join(parts[0:l-1], ".")
		ext := parts[l-1]
		return dir, filename, ext, nil
	}
	err = fmt.Errorf("unable to determine destination filename and extension")
	log.Print(err)
	return dir, filename, ext, err
}

func isSupported(format magicnum.Format) bool {
	switch format {
	case magicnum.Zip, magicnum.LZ4, magicnum.Tar, magicnum.LZW, magicnum.Gzip, magicnum.Bzip2:
		return true
	}
	return false
}
