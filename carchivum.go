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
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"code.google.com/p/go.crypto/ssh"
)

const (
	UnsupportedFmt Format = iota // Not a supported format
	GzipFmt                      // Gzip compression format; always a tar
	TarFmt                       // Tar format; normally used
	Tar1Fmt                      // Tar1 header format; normalizes to FmtTar
	Tar2Fmt                      // Tar1 header format; normalizes to FmtTar
	ZipFmt                       // Zip archive
	ZipEmptyFmt                  // Empty Zip Archive
	ZipSpannedFmt                // Spanned Zip Archive
	Bzip2Fmt                     // Bzip2 compression
	LZHFmt                       // LZH compression
	LZWFmt                       // LZW compression
	LZ4Fmt                       // LZ4 compression
	RARFmt                       // RAR 5.0 and later compression
	RAROldFmt                    // Rar pre 1.5 compression
)

// Format is a type for file format constants.
type (
	Format int
)

var unsetTime time.Time
var CreateDir bool

// Header information for common archive/compression formats.
// Zip includes: zip, jar, odt, ods, odp, docx, xlsx, pptx, apx, odf, ooxml
var (
	headerGzip       = []byte{0x1f, 0x8b}
	headerTar1       = []byte{0x75, 0x73, 0x74, 0x61, 0x72, 0x00, 0x30, 0x30} // offset: 257
	headerTar2       = []byte{0x75, 0x73, 0x74, 0x61, 0x72, 0x00, 0x20, 0x00} // offset: 257
	headerZip        = []byte{0x50, 0x4b, 0x03, 0x04}
	headerZipEmpty   = []byte{0x50, 0x4b, 0x05, 0x06}
	headerZipSpanned = []byte{0x50, 0x4b, 0x07, 0x08}
	headerBzip2      = []byte{0x42, 0x5a, 0x68}
	headerLZH        = []byte{0x1F, 0xa0}
	headerLZW        = []byte{0x1F, 0x9d}
	headerLZ4        = []byte{0x18, 0x4d, 0x22, 0x04}
	headerRAR        = []byte{0x52, 0x61, 0x72, 0x21, 0x1a, 0x07, 0x01, 0x00}
	headerRAROld     = []byte{0x52, 0x61, 0x72, 0x21, 0x1a, 0x07, 0x00}
)

// getFileFormat determines what format the file is in by checking the file's
// header information.
func getFileFormat(r io.ReaderAt) (Format, error) {
	h := make([]byte, 8, 8)
	r.ReadAt(h, 0)
	if bytes.Equal(headerGzip, h[0:2]) {
		return GzipFmt, nil
	}
	if bytes.Equal(headerZip, h[0:4]) {
		return ZipFmt, nil
	}
	if bytes.Equal(headerLZW, h[0:2]) {
		return LZWFmt, nil
	}
	if bytes.Equal(headerLZ4, h[0:4]) {
		return LZ4Fmt, nil
	}
	// partially supported
	if bytes.Equal(headerBzip2, h[0:3]) {
		return Bzip2Fmt, nil
	}
	// unsupported
	if bytes.Equal(headerRAROld, h[0:7]) {
		return UnsupportedFmt, RAROldFmt.NotSupportedError()
	}
	if bytes.Equal(headerRAR, h[0:8]) {
		return UnsupportedFmt, RARFmt.NotSupportedError()
	}
	if bytes.Equal(headerZipEmpty, h[0:4]) {
		return UnsupportedFmt, ZipEmptyFmt.NotSupportedError()
	}
	if bytes.Equal(headerZipSpanned, h[0:4]) {
		return UnsupportedFmt, ZipSpannedFmt.NotSupportedError()
	}
	if bytes.Equal(headerLZH, h[0:2]) {
		return UnsupportedFmt, LZHFmt.NotSupportedError()
	}
	r.ReadAt(h, 257)
	if bytes.Equal(headerTar1, h) || bytes.Equal(headerTar2, h) {
		return TarFmt, nil
	}
	return UnsupportedFmt, UnsupportedFmt.NotSupportedError()
}

func (f Format) String() string {
	switch f {
	case GzipFmt:
		return "gzip"
	case TarFmt, Tar1Fmt, Tar2Fmt:
		return "tar"
	case ZipFmt:
		return "zip"
	case ZipEmptyFmt:
		return "empty zip archive"
	case ZipSpannedFmt:
		return "spanned zip archive"
	case Bzip2Fmt:
		return "bzip2"
	case LZHFmt:
		return "lzh"
	case LZWFmt:
		return "lzw"
	case LZ4Fmt:
		return "lz4"
	case RARFmt:
		return "rar post 5.0"
	case RAROldFmt:
		return "rar pre 1.5"
	}
	return "unsupported"
}

func FormatFromString(s string) Format {
	s = strings.ToLower(s)
	switch s {
	case "gzip":
		return GzipFmt
	case "tar":
		return TarFmt
	case "zip":
		return ZipFmt
	case "bzip2":
		return Bzip2Fmt
	case "lzh":
		return LZHFmt
	case "lzw":
		return LZWFmt
	case "lz4":
		return LZ4Fmt
	case "rar":
		return RARFmt
	}
	return UnsupportedFmt
}

// NotSupportedError returns a formatted error string
func (f Format) NotSupportedError() error {
	return fmt.Errorf("%s not supported", f.String())
}

var defaultFormat = GzipFmt

// Options
//var AppendDate bool
//var o utputNameTimeFormat string = time.RFC3339
//var UseNano bool
//var Separator string = "-"
//var MakeUnique bool = false

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
	Name        string
	UseLongExt  bool
	UseFullpath bool
	// Create operation modifiers
	Owner int
	Group int
	os.FileMode
	// Extract operation modifiers

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

func Extract(src, dst string) error {
	// determine the type of archive
	f, err := os.Open(src)
	if err != nil {
		log.Print(err)
		return err
	}
	// find its format
	format, err := getFileFormat(f)
	if err != nil {
		log.Print(err)
		return err
	}
	if format == UnsupportedFmt {
		err := fmt.Errorf("%s: %s is not a supported format", src, format.String())
		log.Print(err)
		return err
	}
	// if dst != "" see if it exists. If it doesn't create it
	if dst != "" {
		fi, err := os.Stat(dst)
		if err != nil {
			// wasn't found make it
			err := os.MkdirAll(dst, 0744)
			if err != nil {
				log.Print(err)
				return err
			}
			goto typeSwitch
		}
		if !fi.IsDir() {
			err := fmt.Errorf("cannot extract to %q: not a directory", dst)
			log.Print(err)
			return err
		}
	}

typeSwitch:
	if format == ZipFmt {
		// Close for now, since Extract expects src and dst name
		// TODO change it so zip expected a reader
		f.Close()
		zip := NewZip()
		err := zip.Extract(dst, src)
		if err != nil {
			log.Print(err)
		}
		return err
	}
	tar := NewTar()
	tar.Format = format
	err = tar.Extract(dst, f)
	if err != nil {
		log.Print(err)
	}
	return err
}

// ParseFormat takes a string and returns the
func ParseFormat(s string) (Format, error) {
	switch s {
	case "gzip", "tar.gz", "tgz":
		return GzipFmt, nil
	case "tar":
		return TarFmt, nil
	case "lzw", "taz", "tz", "tar.Z":
		return LZWFmt, nil
	case "bz2", "tbz", "tb2", "tbz2", "tar.bz2":
		return Bzip2Fmt, nil
	case "lz4", "tar.lz4", "tz4":
		return LZ4Fmt, nil
	case "zip":
		return ZipFmt, nil
	}
	return UnsupportedFmt, UnsupportedFmt.NotSupportedError()
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

func Push() error {
	cfg := &ssh.ClientConfig{
		User: "asd",
		Auth: []ssh.AuthMethod{ssh.Password("asd")},
	}
	_ = cfg
	return nil
}
