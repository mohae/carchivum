// tar implements the tape archive format.
package carchivum

import (
	"archive/tar"
	_ "compress/gzip"
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
