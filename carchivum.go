// creates compressed archives.
//
// Go's `archive` package, supports `tar` and `zip`
// Go's `compress` package supports: bzip2, flate, gzip, lzw, zlib
//
// Carchivum supports zip and tar. For tar, archiver also supports
// the following compression:
//
// When using archiver, compression is not optional.
package carchivum

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	Uncompressed Compression = iota
	Gzip
)

const (
	InvalidFormat Format = iota
	TarFormat
	ZipFormat
)

type (
	Compression int
	Format int
)

var defaultCompression Compression = Gzip
var defaultFormat Format = TarFormat

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

	if strings.HasSuffix(root, p) {
		logger.Debugf("%s | %s, don't add if source is the source directory", root, p)
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

func ParseType(s string) (Compression, error) {
	switch s {
	case "gzip", "tar.gz", "tgz":
		return Gzip, nil
	}

	return Uncompressed, fmt.Errorf("unsupported compression: %s", s)
}

func ParseFormat(s string) (Format, error) {
	switch s {
	case "tar":
		return TarFormat, nil
	case "zip":
		return ZipFormat, nil
	}
	
	return InvalidFormat, fmt.Errorf("invalid archive format: %s", s)
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


