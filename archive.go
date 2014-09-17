package carchivum

import (
	"bufio"
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/MichaelTJones/walk"
	utilp "github.com/mohae/utilitybelt/path"
	log "github.com/Sirupsen/logrus"
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

	// cType si the compression type to be used (extension).
	cType string

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
	// The file's path
	p string
	
	// The file's FileInfo
	fi os.FileInfo
}

func NewArchive() *car {
	archive := &car{}
	archive.appendDate = appendDate
	archive.appendOnFilenameCollision = appendOnFilenameCollision
	archive.separator = separator
	archive.dateFormat = dateFormat
	archive.cType = "tar"
	archive.format = "gzip"
	archive.t0 = time.Now()

	return archive
}

func (c *car) Create(destination string, sources ...string) (message string, err error) {
	// Validate the destination, e.g. try and create a file there with the name
	// TODO make this flexible so destinations other than files can be done
	f, err := c.createOutputFile(destination)
	if err != nil {
		log.Error(err)
		return message, err
	}
	defer func() {
		cerr := f.Close()
		if cerr != nil {
			log.Error(cerr)
			// don't overwrite an existing error
			if err == nil {
				err = cerr
			}
		}
	}()


	// Process the sources
	err = c.AddSources(sources...)
	if err != nil {
		log.Error(err)
		return message, err
	}

	// use the appropriate archive
	switch c.format {
	case "tar":
		err = c.tar()
	case "zip":
		fmt.Println("Zip not implemented")
//		err = c.Zip()
	default:
		err = fmt.Errorf("unknown archive format: %s", c.format)
	}

	if err != nil {
		log.Error(err)
		return message, err
	}
	
	message = "created done, we hope"
	return message, err
}

func (c *car) createOutputFile(s string) (file *os.File, err error) {
	// There is a small chance that things will mutatue on us since we are
	// just stat'ing the file's existence. But that's a risk we'll take,
	// for now.

	// First see if file exists
	_, err = os.Stat(s)
	if err == nil {
		if c.appendOnFilenameCollision {
			dir, file, ext, err := getFileParts(s)
			if err != nil {
				log.Error(err)
				return nil, err
			}
	
			// get the default extension for this format type				
			if ext == "" {
				ext, err = defaultExtFromType(c.cType)
				if err != nil {	
					log.Error(err)
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
				log.Error(err)
				return nil, err
			}
		} else {
			err = fmt.Errorf("file exists, unable to create destination file for the archive: %s")
			log.Error(err)
			return nil, err
		}
	}	

	// Create the archive file
	file, err = os.Create(s)
	if err != nil {
		log.Error(err)
	}

	return file, err	
}
/*
// ArchiveAndDelete creates a compressed archive of the passed sources using the
// passed filename in the destination directory. If the archive is successfully
// created and written to the target, the archived targets are deleted.
// Target(s) is variadic.
func (a *Archive) ArchiveAndDelete(compression, filename, destination string, sources ...string) error {
	if filename == "" || filename == "./" || filename == "." {
		return fmt.Errorf(fmt.Sprintf("Filename was empty or invalid: %s", filename))
	}

	if len(targets) <= 0  {
		return fmt.Errorf(fmt.Sprintf("No source files or directories were specified. Unable to create archive"))
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

// AddSources adds the sources to the archive. This does not perform the walk
// of resources, just the add.
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
	if source == "" {
		// If nothing was passed. do nothing. 
		// TODO is this an error state? probably not
		return nil
	}

	// See if the path exists
	exists, err := utilp.PathExists(source)
	if err != nil {
		log.Error(err)
		return err
	}

	if !exists{
		err := fmt.Errorf("Unable to inventory directory contents; path does not exist: %s", source)
		log.Error(err)
		return err
	}

	// get the absolute path
	fullPath, err := filepath.Abs(source)
	if err != nil {
		log.Error(err)
		return err
	}

	// setup the callback function
	visitor := func(p string, fi os.FileInfo, err error) error {
		return c.addFilename(fullPath, p, fi, err)
	}

	err = walk.Walk(fullPath, visitor)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (c *car) addFilename(fullpath string, p string, fi os.FileInfo, err error) error {
	// See if the path exists
	var exists bool
	exists, err = utilp.PathExists(p)
	if err != nil {
		log.Error(err)
		return err
	}

	if !exists {
		err = fmt.Errorf("file does not exist: %s", p)
		log.Error(err)
		return err
	}

	// Get the relative information.
	rel, err := filepath.Rel(fullpath, p)
	if err != nil {
		log.Error(err)
		return err
	}

	if rel == "." {
		log.Info("Don't add the relative root")
		return nil
	}

	// Add the file information.
	c.Files = append(c.Files, file{p: rel, fi: fi})
	return nil
}



func (c *car) SetCompressionType(s string) error {
	if c.compressionTypeIsSupported(s) {
		c.cType = s
		return nil
	}

	err := fmt.Errorf("unsupported compression type: %s", s)
	log.Error(err)
	return err
}

func (c *car) SetFormat(s string) error {
	if s == "tar" || s == "zip" {
		c.format = s
		return nil
	} 

	err := fmt.Errorf("unsupported archive format: %s", s)
	log.Error(err)
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

func (c *car) SetDateFormat(s string) error {
	if s == "" {
		return nil
	}
	const longForm = "Jan 2, 2006 at 3:04pm (MST)"
	// see if the format is valid, if it errors, return the error
	_, err := time.Parse(s, longForm)
	if err != nil {
		log.Error(err)
		return err
	}

	c.dateFormat = s

	return nil
}

func (c *car) setTimeDelta() {
	c.𝛥t = float64(time.Since(c.t0)) / 1e9
}

func (c *car) tar() error {
	// 
	log.Infof("creating tarball: %s", c.name)
	
	// create the tarwriter
	tBall, err := os.Create(c.name)
	if err != nil {
		log.Fatal(err)
		fmt.Println("Post Fatal output, pre return: So what does log.Fatal really do if this output made it?")
		return err
	}
	defer func() {
		cerr := tBall.Close()
		if cerr != nil {
			log.Error(cerr)
			// don't overwrite an existing error
			if err == nil {
				err = cerr
			}
		}
	}()

	var compressor io.Writer

	var ext string
	// Find out the compression type and wrap the tBall with it
	switch c.cType {
	case "gzip", "tgz", "tar.gz", "cgz", "car.gz":
		if c.useLongExt {
			ext = ".car.gz"
		} else {
			ext = ".cgz"
		}
		compressor = gzip.NewWriter(tBall)
	
	// todo work out extension stuff, if necessary

		
/*
	case "zlib", "taz", "tar.z", "caz", "car.z" {
		if c.useLongExt {
			ext = "car.z"
		} else {
			ext = "caz"
		}
		compressor = zlib.NewWriter(tBall)
*/
	
	default:
		err := fmt.Errorf("unknown compression type: %s", c.cType)
		log.Fatal(err)
		return err
	}

	_ = ext
	// Wrap the compressor with a tar writer
	tW := tar.NewWriter(compressor)
	defer func() {
		cerr := tW.Close()
		if cerr != nil {
			log.Error(cerr)
			// don't overwrite an existing error
			if err == nil {
				err = cerr
			}
		}
	}()

	var i int
	var f file
	for i, f = range c.Files {
		c.tarFile(tW, f.p)
		if err != nil {
			log.Fatal(err)
			return err
		}
	}

	log.Debugf("Archive created: %d files totalling %d bytes processed of %s files inventoried", i, c.bytes, c.files)

	return nil
}

func (c *car) tarFile(tW *tar.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		log.Error(err)
		return err
	}
	defer file.Close()
	
	var fileStat os.FileInfo

	fileStat, err = file.Stat()
	if err != nil {
		log.Error(err)
		return err
	}

	fileMode := fileStat.Mode()
	if fileMode.IsDir() {
		return nil
	}

	// Initialize the header based on the fileinfo
	// this call assumes it isn't a symlink
	// TODO handle symlink, unless the walk skips them too
	tHeader, err := tar.FileInfoHeader(fileStat, "")
	if err != nil {
		log.Error(err)
		return err
	}

	err = tW.WriteHeader(tHeader)
	if err != nil {
		log.Error(err)
		return err
	}

	b, err := io.Copy(tW, file)
	if err != nil {
		log.Error(err)
		return err
	}

	c.lock.Lock()
	c.files++
	c.bytes += b
	c.lock.Unlock()

	return nil
}
