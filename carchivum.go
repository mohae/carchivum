// creates compressed tarballs
//    
// Go's `archive` package, supports `tar` and `zip`
// Go's `comopress` package supports: bzip2, flate, gzip, lzw, zlib
//
// Carchivum supports zip and tar. For tar, archiver also supports
// the following compression:
//
// When using archiver, compression is not optional.
package carchivum 

import (
	_ "compress/gzip"
	"errors"
	_ "fmt"
	"io"
	"os"
	_ "path"
	"sync"
	"time"

	"github.com/dotcloud/tar"
//	"github.com/mohae/goutils"
)

const (
	GZip = "tgz"
        GZipL = "tar.gz"
	BZip2 = "tb2"
        BZip2L = "tar.bz2"
        BZip2Alt = "tbz2"
        Compress = "taz"
        CompressL = "tar.Z"
)
var defaultCompressionFormat = "tb2"

var appendDatetimeOnCollision = true
var datetimeFormat = "2006-01-02T150405Z0700"
var datetimePrefix = "-"

// SetDateTimeFormat overrides the default datetime format. The passed format
// must use Go's datetime format.
func SetDatetimeFormat(s string) {
	datetimeFormat = s
}

// SetAppendDatetimeOnCollisioni sets whether the tarball names should be 
// automatically appended with the current datetime, using the datetimeFormat,
// when a name collision occurs on the archive filename. The appended datetime
// will be prefixed with -, unless that is overridden. 
// 
//     appendDatetimeOnCollision = true: appends the current datetime, using
//         the configured datetimeFormat. If that name collides, it either
//         errors or appends a random 4 digit number, depending on config.
//
//     appendDatetimeOnCollision = false: if a collision occurs on the filename
//         an error is returned, instead of automatically appending the current
//         datetime.
func SetAppendDatetimeOnCollision(b bool) {
	appendDatetimeOnCollision = b
}

// SetDatetimePrefix overrides the default datetime prefix of `-`. The datetime
// prefix is used to to prefix the datetime prior to appending it to the
// filename, e.g. filename-datetime.tgz.
// 
// To not use a prefix, set it to an emptys string, ""
func SetDatetimePrefix(s string) {
	datetimePrefix = s
}

// SetDefaultCompressionFormat overrides the current defaultCompressionFormat
// with the passed value. Returns an error if the format is not supported.
func SetDefaultCompressionFormat(s string) error {
	switch s {
	case GZip, GZipL, BZip2, BZip2L, BZip2Alt:
		defaultCompressionFormat = s
		return nil
	default:
		return errors.New(s + " is not supported.")
	}

	return nil
}


// archiver is an interface for archive formats; mainly tar and zip
type archiver interface{
	Create(...string) error
//        Delete() error
//        Extract() error
}

// car is a Compressed ARchive. The struct holds information about cars.
type car struct {
	// Name of the archive, this includes path information, if any.
	Name string

	// Compression type to be used (extension).
	Type string

	// keep track of count of readers
	ReaderCount int

	// scanneQueue queues up scans of directories.
	ScannerQueue chan string

	// readerQueue queues up reads of files.
	ReaderQueue chan string

	// blockedQueue queues up blocked items for unblocked processing
	BlockedQueue chan string

//	Output	*bufio.Writer

	// synchronicity synchronizes stuff
	Synchronicity sync.WaitGroup

	// List of files to add to the archive.
	Fi	[]os.FileInfo
}

/*
func (a *archive) addFile(tw *tar.Writer, fi os.FileInfo) error {
	var err error 

	// It should exist, since we are 
	file, err = os.Open(fi.Name())
	if err != nil {
		return err
	}
	defer file.Close()

	var fileStat os.FileInfo
	fileStat, err= file.Stat()
	if err != nil {
		return err
	}

	// Don't add directories--they result in tar header errors.
	fileMode := fileStat.Mode()
	if fileMode.IsDir() {
		return nil
	}

	// Create the tar header stuff.
	tarHeader := new(tar.Header)
	tarHeader.Name = fi.Name()
	tarHeader.Size = fileStat.Size()
	tarHeader.Mode = int64(fileStat.Mode())
	tarHeader.ModTime = fileStat.ModTime()

	// Write the file header to the tarball.
	err = tw.WriteHeader(tarHeader)
	if err != nil {
		return err
	}

	// Add the file to the tarball.
	_, err = io.Copy(tw, file)
	if err != nil {
		return err
	}

	return nil
}
*/
/*
// ArchiveAndDelete creates a compressed archive of the passed sources using the
// passed filename in the destination directory. If the archive is successfully
// created and written to the target, the archived targets are deleted.
// Target(s) is variadic.
func (a *Archive) ArchiveAndDelete(compression, filename, destination string, sources ...string) error {
	if filename == "" || filename == "./" || filename == "." {
		return errors.New(fmt.Sprintf("Filename was empty or invalid: %s", filename))
	}

	if len(targets) <= 0  {
		return errors.New(fmt.Sprintf("No source files or directories were specified. Unable to create archive"))
	}
 
	if compression == "" {
		compressiond = defaultCompression
	}
	
	// See if the requested compression exists


	// See if src exists, if it doesn't then don't do anything
	_, err := os.Stat(p)
	if err != nil {

		// Nothing to do if it doesn't exist
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}


	// build the tarball file name; let archive worry about whether the
	// destination can be written to	
	err = a.archive(tarball, sources)
	if err != nil {
		return err
	}

	// Delete the old artifacts.
	err = a.delete(sources)
	if err != nil {
		return err
	}

	return nil
}
*/

/*
// AddPaths adds the contents of the passed paths. Additions are automatically
// recursive so if a path is a directory, all the contents of that directory
// will be added.
func (a *archiver) AddSources(sources ...string) error {
	// Make file already exists
        if a.Fi {
		return errors.New("An error occurred while trying to work with " + "_____________" )
        }

	// loop through the paths and add them; errors on invalid paths and such
	for _, source := range sources {
		// Get a list of directory contents
		err := a.DirWalk(p)
		if err != nil {
			return err
		}

		if len(a.Files) <= 1 {
			return nil
		}

	}
}
*/

/*
// CreateArchive adds each file to the archive and filters the archive through
// its configured compression format. The resulting archive is written as a
// byte stream to its destination.
func (a *Archive) CreateArchive() {
	// The tarball gets compressed with gzip
	gw := gzip.NewWriter(tBall)
	defer gw.Close()

	// Create the tar writer.
	tW := tar.NewWriter(gw)
	defer tW.Close()

	// Go through each file in the path and add it to the archive
	var i int
	var f file

	for i, f = range a.Files {
		err := a.addFile(tW, goutils.appendSlash(relPath)+f.p)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Archive) deletePriorBuild(p string) error {
	//delete the contents of the passed directory
	return deleteDirContent(p)
}
*/

func formattedNow() string {
	// Time in ISO 8601 like format. The difference being the : have been
	// removed from the time.
	return time.Now().Local().Format(timeFormat)
}

func newArchive(appendDatetime bool, datetimePrefix, datetimeFormat string) *Archiver {
	archive :=  &Archive{}
	archive.appendDatetime = appendDatetime
	archive.datetimePrefix = datetimePrefix

	// Only override the datetime format if it isn't empty
	if datetimeFormat != "" {
		archive.datetimeFormat == datetimeformat
	}

	return archive
}

