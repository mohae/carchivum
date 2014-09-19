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

func (c *car) CreateTar(w io.Writer) (err error) {
	// 
	logger.Infof("creating tarball: %s", c.name)
//	var ext string
	// Find out the compression type and wrap the tBall with it
	switch c.compressionType {
	case "gzip", "tgz", "tar.gz", "cgz", "car.gz":
		err := c.writeTarGzip(w)
		if err != nil {
			logger.Error(err)
			return err
		}
	default:
		err := fmt.Errorf("Unsupported compression type: %q", c.compressionType)
		logger.Error(err)
		return err
	}

	

	return nil
}

func (c *car) writeTar(w io.Writer) (err error) {
	tw := tar.NewWriter(w)
	defer func() {
		cerr := tw.Close()
		if cerr != nil && err == nil {
			logger.Error(cerr)
			err = cerr
		}
		logger.Debug("closed tar writer")
	}()

//	var i int
	var f file
	for _, f = range c.Files {
		err := c.tarFile(tw, f)
		if err != nil {
			logger.Error(err)
			return err
		}
	}

	return err
	
}

func (c *car)  writeTarGzip(w io.Writer) (err error) {
	gw := gzip.NewWriter(w)
	// Close the file with error handling
	defer func() {
		cerr := gw.Close()
		if cerr != nil && err == nil {
			logger.Error(cerr)
			err = cerr
		}
		logger.Debug("Closed gzip writer")
	}()

	err = c.writeTar(gw)
	return err
}

func (c *car) tarFile(tW *tar.Writer, f file) (err error) {
	logger.Debugf("%+v", f)

	file, err := os.Open(f.fullPath)
	if err != nil {
		logger.Error(err)
		return err
	}
	// Close the file with error handling
	defer func() {
		cerr := file.Close()
		if cerr != nil && err == nil {
			logger.Error(cerr)
			err = cerr
		}
	}()
	
// Get the current stats for the file
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

	tHeader.Name = f.fullPath

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
