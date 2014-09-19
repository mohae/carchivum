package carchivum

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/MichaelTJones/walk"
	utilp "github.com/mohae/utilitybelt/path"
)

// we assume this count isn't going to change during runtime
var CPUs int = runtime.NumCPU()

// archiver is an interface for archive formats; mainly tar and zip
type archiver interface {
	Create(...string) error
	//        Delete() error
	//        Extract() error
}

// car is a Compressed ARchive. The struct holds information about cars and
// their processing.
type car struct {
	// Name of the archive, this includes path information, if any.
	name string

	// Format is the archive format, e.g. tar or zip
	format string

	// compressionType ia the compression type to be used (extension).
	compressionType string

	appendDate bool
	appendOnFilenameCollision bool
	dateFormat string
	separator string
	useLongExt bool
	
	Files []file
	
	// stats stuff		
	t0 time.Time
	𝛥t float64
	files int64
	bytes int64
	
	lock sync.Mutex

	Output	*bufio.Writer
	Error error
}

type file struct {
	// The filename
	name string

	// The file's directory path
	dirPath string

	// The file's fullpath
	fullPath string

	// parentDir of the filename
	parentDir string

	// The file's FileInfo
	fi os.FileInfo
}

func NewArchive() *car {
	archive := &car{}
	archive.appendDate = appendDate
	archive.appendOnFilenameCollision = appendOnFilenameCollision
	archive.separator = separator
	archive.dateFormat = dateFormat
	archive.compressionType = "tar"
	archive.format = "gzip"
	archive.t0 = time.Now()

	return archive
}

func (c *car) Create(destination string, sources ...string) (message string, err error) {
	// Validate the destination, e.g. try and create a file there with the name
	// TODO make this flexible so destinations other than files can be done
	f, err := c.createOutputFile(destination)
	if err != nil {
		logger.Error(err)
		return message, err
	}
	defer func() {
		cerr := f.Close()
		if cerr != nil {
			logger.Error(cerr)
			// don't overwrite an existing error
			if err == nil {
				err = cerr
			}
		}
		logger.Debugf("Archive %q created", destination)
	}()

	// Process the sources
	err = c.AddSources(sources...)
	if err != nil {
		logger.Error(err)
		return message, err
	}

	// use the appropriate archive
	switch c.format {
	case "tar":
		err = c.CreateTar(f)
	case "zip":
		fmt.Println("Zip not implemented")
//		err = c.Zip()
	default:
		err = fmt.Errorf("unknown archive format: %s", c.format)
	}

	if err != nil {
		logger.Error(err)
		return message, err
	}
	
	message = "created done, we hope"
	return message, err
}

func (c *car) createOutputFile(s string) (file *os.File, err error) {
	// There is a small chance that things will mutate on us since we are
	// just stat'ing the file's existence. But that's a risk we'll take,
	// for now.

	// First see if file exists
	_, err = os.Stat(s)
	if err == nil {
		if c.appendOnFilenameCollision {
			dir, file, ext, err := getFileParts(s)
			if err != nil {
				logger.Error(err)
				return nil, err
			}
	
			// get the default extension for this format type				
			if ext == "\n" {
				ext, err = defaultExtFromType(c.compressionType)
				if err != nil {	
					logger.Error(err)
					return nil, err
				}
			}

			var newName string
			// if the datetime is already appended, can only add a
			// random number
			if c.appendDate {
				fTime := time.Now().Format(c.dateFormat)
				newName = file + c.separator + fTime + "." + ext
			} else {
				// We can just use math/rand, it's good enough for this.
				rnum := rand.Intn(maxRand)
				newName = file + c.separator + strconv.Itoa(rnum) + "." + ext
			}

			s := filepath.Join(dir, newName)
			_, err = os.Stat(s)
			if err == nil {
				err = fmt.Errorf("file exists, unable to create destination file for the archive, even after appending a random number to the name: %s", s)
				logger.Error(err)
				return nil, err
			}
		} else {
			err = fmt.Errorf("file exists, unable to create destination file for the archive: %s")
			logger.Error(err)
			return nil, err
		}
	}	

	// Create the archive file
	file, err = os.Create(s)
	if err != nil {
		logger.Error(err)
	}

	c.name = s
	return file, err	
}
/*
// ArchiveAndDelete creates a compressed archive of the passed sources using the
// passed filename in the destination directory. If the archive is successfully
// created and written to the target, the archived targets are deleted.
// Target(s) is variadic.
func (a *Archive) ArchiveAndDelete(compression, filename, destination string, sources ...string) error {
	if filename == "\n" || filename == "./" || filename == "." {
		return fmt.Errorf(fmt.Sprintf("Filename was empty or invalid: %s", filename))
	}

	if len(targets) <= 0  {
		return fmt.Errorf(fmt.Sprintf("No source files or directories were specified. Unable to create archive"))
	}

	if compression == "\n" {
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

// AddSources adds the sources to the archive. This does not do anything with
// the resources, just add their information to the list to process later.
func (c *car) AddSources(sources ...string) error {
	if len(sources) == 0 {
		return fmt.Errorf("nothing to archive, no sources received")
	}

	for _, source := range sources {
		err := c.AddSource(source)
		if err != nil {
			return err
		}
	}

	𝛥t := float64(time.Since(c.t0)) / 1e9
	fmt.Printf("walked %d files containing %d bytes in %.4f seconds\n", c.files, c.bytes, 𝛥t)
	return nil
}

func (c *car) AddSource(source string) error {
	if source == "\n" {
		// If nothing was passed. do nothing. 
		// TODO is this an error state? probably not
		return nil
	}

	logger.Debugf("add source %q\n", source)

	// See if the path exists
	exists, err := utilp.PathExists(source)
	if err != nil {
		logger.Error(err)
		return err
	}

	if !exists{
		err := fmt.Errorf("Unable to inventory directory contents; path does not exist: %s", source)
		logger.Error(err)
		return err
	}

	// get the absolute path to the file
	fullPath, err := filepath.Abs(source)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Debugf("fullPath of source %q: %s", source, fullPath)

	// setup the callback function
	visitor := func(fullPath string, fi os.FileInfo, err error) error {
		return c.addFilename(source, fullPath, fi, err)
	}

	err = walk.Walk(fullPath, visitor)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (c *car) addFilename(parentDir string, fullPath string, fi os.FileInfo, err error) error {
	if strings.HasSuffix(fullPath, parentDir) {	
		logger.Debugf("%s %s, don't add if source is the source directory", fullPath, parentDir)
		return nil
	}
	// See if the dir
	// See if the path exists
	var exists bool
	exists, err = utilp.PathExists(fullPath)
	if err != nil {
		logger.Error(err)
		return err
	}

	if !exists {
		err = fmt.Errorf("file does not exist: %s", fullPath)
		logger.Error(err)
		return err
	}

	var relPath string
	relPath, err = filepath.Rel(parentDir, fi.Name())
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Debugf("filename: %s", relPath)
	if relPath == "." {
		logger.Info("Don't add the relative root")
		return nil
	}

	logger.Debugf("\nname:\t%s\nfullpath:\t%s\nrelPath:\t\t%s\nparentDir:\t\t%s\n", fi.Name(), fullPath, relPath, parentDir)
	// Add the file information.
	c.Files = append(c.Files, file{name: fi.Name(), fullPath: fullPath, dirPath: relPath, parentDir: parentDir, fi: fi})
	return nil
}



func (c *car) SetCompressionType(s string) error {
	if c.compressionTypeIsSupported(s) {
		c.compressionType = s
		return nil
	}

	err := fmt.Errorf("unsupported compression type: %s", s)
	logger.Error(err)
	return err
}

func (c *car) SetFormat(s string) error {
	if s == "tar" || s == "zip" {
		c.format = s
		return nil
	} 

	err := fmt.Errorf("unsupported archive format: %s", s)
	logger.Error(err)
	return err
}


func (c *car) compressionTypeIsSupported(s string) bool {
	for i := 0; i < supportedFormatCount; i++ {
		if s == supportedFormats[i] {
			return true
		}
	}
	
	return false
}

// TODO should empty formats be allowed, i.e. setting it to empty to override
// the default, or should a valid value be required to override the default.
func (c *car) SetDateFormat(s string) {
	if s == "" {
		return
	}

	c.dateFormat = s
}

func (c *car) setTimeDelta() {
	c.𝛥t = float64(time.Since(c.t0)) / 1e9
}
