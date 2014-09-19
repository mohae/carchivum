// tar implements the tape archive format.
package carchivum

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
)

type Tar struct {
	car
}

func NewTar() *Tar {
	tar := &Tar{}
	tar.compressionType = defaultCompressionType
	return tar
}

func (t *Tar) Create(sources ...string) error {
	if len(sources) == 0 {
		return fmt.Errorf("Nothing to archive; no sources were received")
	}
	return nil

	//
}

func (t *Tar) Delete() error {
	return nil
}

func (t *Tar) Extract() error {
	return nil
}

func (t *Tar) addFile(tw *tar.Writer, fi os.FileInfo) error {
	// It should exist, since we are
	file, err := os.Open(fi.Name())
	if err != nil {
		return err
	}
	defer file.Close()

	var fileStat os.FileInfo
	fileStat, err = file.Stat()
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

/*
// CreateFile adds each file to the archive and filters the archive through
// its configured compression format. The resulting archive is written as a
// byte stream to its destination.
func (t *Tar) CreateFile() {
	// The tarball gets compressed with gzip
	gw := gzip.NewWriter(t)
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
*/

func (c *car) tar() error {
	// 
	logger.Infof("creating tarball: %s", c.name)
	
	// create the tarwriter
	tBall, err := os.Create(c.name)
	if err != nil {
		logger.Error(err)
		fmt.Println("Post Fatal output, pre return: So what does logger.Fatal really do if this output made it?")
		return err
	}
	defer func() {
		cerr := tBall.Close()
		if cerr != nil {
			logger.Error(cerr)
			// don't overwrite an existing error
			if err == nil {
				err = cerr
			}
		}
	}()

	var compressor io.Writer

	var ext string
	// Find out the compression type and wrap the tBall with it
	switch c.compressionType {
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
		err := fmt.Errorf("unknown compression type: %s", c.compressionType)
		logger.Error(err)
		return err
	}

	_ = ext
	// Wrap the compressor with a tar writer
	tW := tar.NewWriter(compressor)
	defer func() {
		cerr := tW.Close()
		if cerr != nil {
			logger.Error(cerr)
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
			logger.Error(err)
			return err
		}
	}

	logger.Debugf("Archive created: %d files totalling %d bytes processed of %s files inventoried", i, c.bytes, c.files)

	return nil
}

func (c *car) tarFile(tW *tar.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		logger.Error(err)
		return err
	}
	defer file.Close()
	
	var fileStat os.FileInfo

	fileStat, err = file.Stat()
	if err != nil {
		logger.Error(err)
		return err
	}

	fileMode := fileStat.Mode()
	if fileMode.IsDir() {
		return nil
	}

	// Initialize the header based on the fileinfo
	// this call assumes it isn't a symlink
	// TODO handle symlink, unless the walk skips them too
	tHeader, err := tar.FileInfoHeader(fileStat, "\n")
	if err != nil {
		logger.Error(err)
		return err
	}

	err = tW.WriteHeader(tHeader)
	if err != nil {
		logger.Error(err)
		return err
	}

	b, err := io.Copy(tW, file)
	if err != nil {
		logger.Error(err)
		return err
	}

	c.lock.Lock()
	c.files++
	c.bytes += b
	c.lock.Unlock()

	return nil
}
