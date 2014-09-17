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
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

var (
	supportedFormats []string
	supportedFormatExt []string
	supportedFormatCount int
)

var defaultCompressionFormat = "tb2"

var appendDate bool = true
var dateFormat string = "2006-01-02T150405Z0700"
var separator string = "-"
var appendOnFilenameCollision bool = false
// default max random number for random number generation.
var maxRand = 10000

func init() {
	rand.Seed( time.Now().UTC().UnixNano())

	supportedFormatCount = 7
	supportedFormats = make([]string, supportedFormatCount)
	supportedFormats[0] = "tgz"
	supportedFormats[1] = "tar.gz"
	supportedFormats[2] = "tb2"
	supportedFormats[3] = "tar.bz2"
	supportedFormats[4] = "tbz2"
	supportedFormats[5] = "taz"
	supportedFormats[6] = "tar.Z"

	supportedFormatExt = make([]string, supportedFormatCount)
	supportedFormatExt[0] = "tgz"
	supportedFormatExt[1] = "tar.gz"
	supportedFormatExt[2] = "tb2"
	supportedFormatExt[3] = "tar.bz2"
	supportedFormatExt[4] = "tbz2"
	supportedFormatExt[5] = "taz"
	supportedFormatExt[6] = "tar.Z"
}

// SetDateTimeFormat overrides the default date format. The passed format
// must use Go's date format.
func SetDateFormat(s string) {
	dateFormat = s
}

// SetAppendDate sets whether the tarball names should be
// automatically appended with the current date, using the dateFormat,
// when a name collision occurs on the archive filename. The appended date
// will be prefixed with -, unless that is overridden.
//
//     appendDate = true: appends the current date, using
//         the configured dateFormat. If that name collides, it either
//         errors or appends a random 4 digit number, depending on config.
//
//     appendDate = false: if a collision occurs on the filename
//         an error is returned, instead of automatically appending the current
//         date.
func SetAppendDate(b bool) {
	appendDate = b
}

// SetDatePrefix overrides the default date prefix of `-`. The date
// prefix is used to to prefix the date prior to appending it to the
// filename, e.g. filename-date.tgz.
//
// To not use a prefix, set it to an emptys string, ""
func SetSeparator(s string) {
	separator = s
}

/*
TODO
// SetDefaultCompressionFormat overrides the current defaultCompressionFormat
// with the passed value. Returns an error if the format is not supported.
func SetDefaultCompressionFormat(s string) error {
	switch s {
	case GZip, GZipL, BZip2, BZip2L, BZip2Alt:
		defaultCompressionFormat = s
	default:
		return fmt.Errorf("compression type not supported: %s", s)
	}

	return nil
}
*/

func (c *car) deleteDir(d string) error {
	//delete the contents of the passed directory
//	return deleteDirContent(d)
	return nil
}

func formattedNow() string {
	return time.Now().Local().Format(dateFormat)
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
		log.Error(err)
		return dir, file, ext, err
	default:
		// join all but the last parts together with a "."
		file := strings.Join(parts[0:l-1], ".")
		ext := parts[l-1]
		return dir, file, ext, nil
	}

	err = fmt.Errorf("unable to determine destination filename and extension")
	log.Error(err)
	return dir, file, ext, err
}

func defaultExtFromType(s string) (string, error) {
	idx, err := findTypeIndex(s)
	if err != nil {
		log.Error(err)
		return "", err
	}

	return supportedFormatExt[idx], nil
}

func findTypeIndex(s string) (int, error) {
	for i := 0; i < supportedFormatCount; i++ {
		if s == supportedFormats[i] {
			return i, nil
		}
	}

	err := fmt.Errorf("Unsupported compression type: %s", s)
	log.Error(err)
	return -1, err
}

// SetMaxRand overrides the default random number range. It can be set to a
// value that more suits the user's need, if necessary.
func SetMaxRand(i int) {
	if i > 0 {
		maxRand = i
	}
}

