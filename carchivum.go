// creates compressed archives.
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
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	FmtUnsupported Format = iota
	FmtGzip
	FmtTar
	FmtTar1
	FmtTar2
	FmtZip
	FmtZipEmpty
	FmtZipSpanned
	FmtBzip2
	FmtLZH
	FmtLZW
	FmtRAR
	FmtRAROld
)


// Header information for common archive/compression formats.
// Zip includes: zip, jar, odt, ods, odp, docx, xlsx, pptx, apx, odf, ooxml
var (
	headerGzip []byte = []byte{0x1f, 0x8b}
	headerTar1 []byte = []byte{0x75, 0x73, 0x74, 0x61, 0x72, 0x00, 0x30, 0x30} // offset: 257
	headerTar2 []byte = []byte{0x75, 0x73, 0x74, 0x61, 0x72, 0x00, 0x20, 0x00} // offset: 257
	headerZip []byte = []byte{0x50, 0x4b, 0x03, 0x04}
	headerZipEmpty []byte = []byte{0x50, 0x4b, 0x05, 0x06}
	headerZipSpanned []byte = []byte{0x50, 0x4b, 0x07, 0x08}
	headerBzip2 []byte = []byte{0x42, 0x5a, 0x68}
	headerLZH []byte = []byte{0x1F, 0xa0}
	headerLZW []byte = []byte{0x1F, 0x9d}
	headerRAR []byte = []byte{0x52, 0x61, 0x72, 0x21, 0x1a, 0x07, 0x01, 0x00}
	headerRAROld []byte = []byte{0x52, 0x61, 0x72, 0x21, 0x1a, 0x07, 0x00}
)

func GetFileFormat(r io.ReaderAt) (Format, error) {
	h := make([]byte, 8, 8)
	
	r.ReadAt(h, 0)

	if bytes.Equal(headerGzip, h[0:2]) {
		return FmtGzip, nil
	}

	if bytes.Equal(headerZip, h[0:4]) {
		return FmtZip, nil
	}

	r.ReadAt(h, 257)
	if bytes.Equal(headerTar1, h) || bytes.Equal(headerTar2, h) {
		return FmtTar, nil
	}

	// unsupported
	if bytes.Equal(headerRAROld, h) {
		return FmtUnsupported, FmtRAROld.NotSupportedError()
	}

	if bytes.Equal(headerRAR, h) {
		return FmtUnsupported, FmtRAR.NotSupportedError()
	}

	r.ReadAt(h, 0)
	if bytes.Equal(headerZipEmpty, h[0:4]) {
		return FmtUnsupported, FmtZipEmpty.NotSupportedError()
	}

	if bytes.Equal(headerZipSpanned, h[0:4]) {
		return FmtUnsupported, FmtZipSpanned.NotSupportedError()
	}

	if bytes.Equal(headerBzip2, h[0:3]) {
		return FmtUnsupported, FmtBzip2.NotSupportedError()
	}	

	if bytes.Equal(headerLZW, h[0:2]) {
		return FmtUnsupported, FmtLZW.NotSupportedError()
	}	

	if bytes.Equal(headerLZH, h[0:2]) {
		return FmtUnsupported, FmtLZH.NotSupportedError()
	}	

	return FmtUnsupported, FmtUnsupported.NotSupportedError()
}

type (
	Format int
)

func (f Format) ToString() string {
	switch f {
	case FmtGzip:
		return "gzip"
	case FmtTar1, FmtTar2:
		return "tar"
	case FmtZip:
		return "zip"
	case FmtZipEmpty:
		return "empty zip archive"
	case FmtZipSpanned:
		return "spanned zip archive"
	case FmtBzip2:
		return "bzip2"
	case FmtLZH:
		return "LZH"
	case FmtLZW:
		return "LZW"
	case FmtRAR:
		return "RAR post 5.0"
	case FmtRAROld:
		return  "RAR pre 1.5"
	}
	return "unsupported"
}
	
func (f Format) NotSupportedError() error {
	switch f {
	case FmtZipEmpty:
		return fmt.Errorf("empty zip archive is not supported")
	case FmtZipSpanned:
		return fmt.Errorf("spanned zip archive is not supported")
	case FmtBzip2 :
		return fmt.Errorf("bzip2 is not supported")
	case FmtLZH:
		return fmt.Errorf("LZH is not supported")
	case FmtLZW:
		return fmt.Errorf("LZW is not supported")
	case FmtRAR:
		return fmt.Errorf("RAR post 5.0 is not supported")
	case FmtRAROld:
		return fmt.Errorf("RAR pre 1.5 is not supported")
	}
	return fmt.Errorf("unsupported format error, more specific information unavailable")
}
	

var defaultFormat Format = FmtGzip

// Options
var AppendDate bool
var TimeFormat string = time.RFC3339
var UseNano bool
var Separator string = "-"
var MakeUnique bool = false
// default max random number for random number generation.
var MaxRand = 10000

// we assume this count isn't going to change during runtime
var cpu int = runtime.NumCPU()

// Arbitrarily set the multiplier to some default value.
var CPUMultiplier int = 4

// Car is a Compressed Archive. The struct holds information about Cars and
// their processing.
type Car struct {
	// Name of the archive, this includes path information, if any.
	Name string
	UseLongExt bool
	UseFullpath bool

	// Create operation modifiers
	Owner int
	Group int
	Mode os.FileMode

	// Extract operation modifiers


	// Local file selection
	DeleteFiles bool

	Exclude string
	ExcludeExt []string
	ExcludeExtCount int
	ExcludeAnchored string

	Include string
	IncludeExt []string
	IncludeExtCount int
	IncludeAnchored string

	Newer string
	NewerMTime string
	NewerFile string

	// Processing queue
	FileCh	chan *os.File

	// Other Counters
	counterLock sync.Mutex
	files int32
	bytes int64
	compressedBytes int64
	
	// timer
	t0 time.Time
	𝛥t float64
}

func (c *Car) setDelta() {
	c.𝛥t = float64(time.Since(c.t0)) / 1e9
}

func (c *Car) Delta() float64 {
	return c.𝛥t
}

func (c *Car) Message() string {
	return fmt.Sprintf("%q created in %4f seconds\n%d files totalling %d bytes were processed", c.Name, c.𝛥t, c.files, c.bytes)
}

// addFile  reads a file and pipes it to the zipper goroutine.
func (c *Car) AddFile(root, p string, fi os.FileInfo, err error) error {
	logger.Debugf("root: %s, p: %s, fi.Name: %s", root, p, fi.Name())
	// Don't add symlinks, otherwise would have to code some cycle
	// detection amongst other stuff.
	if fi.Mode() & os.ModeSymlink == os.ModeSymlink {
		logger.Debugf("don't follow symlinks: %q", p)
		return nil
	}

	add, err := c.addFile(root, p)
	if err != nil {
		return err
	}
	if !add {
		return nil
	}

	var relPath string
	relPath, err = filepath.Rel(root, p)
	if err != nil {
		logger.Error(err)
		return err
	}

	if relPath == "," {
		logger.Debug("Don't add relative root")
		return nil
	}

	if !c.UseFullpath {
		p = filepath.Join(filepath.Base(root), relPath)
	} 
	
	f, err := os.Open(p)
	if err != nil {
		logger.Error(err)
		return err
	}

	c.counterLock.Lock()
	c.files++
	c.bytes += fi.Size()
	c.counterLock.Unlock()

	c.FileCh <- f

	return nil
}

func (c *Car) addFile(root, p string) (bool, error) {
	if strings.HasSuffix(root, p) {
		logger.Debugf("%s | %s, don't add if source is the source directory", root, p)
		return false, nil
	}

	b, err := c.includeFile(root, p)
	if err != nil {
		return false, err
	}
	if !b {
		logger.Debugf("don't include %q", p)
		return false, nil
	}


	b, err = c.excludeFile(root, p)
	if err != nil {
		return false, err
	}
	if b {
		logger.Debugf("exclude %q", p)
		return false, nil
	}

	return true, nil
}

func (c *Car) includeFile(root, p string) (bool, error) {
	logger.Infof("%sroot: %s c.IncludeAnchored %s", root, p, c.IncludeAnchored)
	if c.IncludeAnchored != "" {
		logger.Info(filepath.Base(p))
		if strings.HasPrefix(filepath.Base(c.IncludeAnchored), p) {
			logger.Info("has prefix")
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

	if c.IncludeExtCount == 0 {
		return false, nil
	}

	for _, ext := range c.IncludeExt {
		if strings.HasSuffix(filepath.Base(p), "." + ext) {
			return true, nil
		}
	}

	return false, nil
}


func (c *Car) excludeFile(root, p string) (bool, error) {
	logger.Infof("%s c.ExcludeAnchored %s", p, c.ExcludeAnchored)
	if c.ExcludeAnchored != "" {
		logger.Info(filepath.Base(p))
		if strings.HasPrefix(filepath.Base(p), c.ExcludeAnchored) {
			logger.Info("has prefix")
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

	for _, ext := range c.ExcludeExt {
		if strings.HasSuffix(filepath.Base(p), "." + ext) {
			return true, nil
		}
	}


	return false, nil
}


func ParseFormat(s string) (Format, error) {
	switch s {
	case "gzip", "tar.gz", "tgz":
		return FmtGzip, nil
	case "tar":
		return FmtTar, nil
	case "zip":
		return FmtZip, nil
	}
	
	return FmtUnsupported, FmtUnsupported.NotSupportedError()
}

func formattedNow() string {
	return time.Now().Local().Format(TimeFormat)
}

func getFileParts(s string) (dir, file, ext string, err error) {
	// see if there is path involved, if there is, get the last part of it
	dir, filename := filepath.Split(s)	
	parts := strings.Split(filename, ".")
	l := len(parts)
	switch l {
	case 2:
		file := parts[0]
		ext := parts[1]
		return dir, file, ext, nil
	case 1:
		file := parts[0]
		return dir, file, ext, nil
	case 0:
		err := fmt.Errorf("no destination filename found in %s", s)
		logger.Error(err)
		return dir, file, ext, err
	default:
		// join all but the last parts together with a "."
		file := strings.Join(parts[0:l-1], ".")
		ext := parts[l-1]
		return dir, file, ext, nil
	}

	err = fmt.Errorf("unable to determine destination filename and extension")
	logger.Error(err)
	return dir, file, ext, err
}
