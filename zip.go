package carchivum

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/MichaelTJones/walk"
)

// Zip handles .zip archives.
type Zip struct {
	Car
	writer  *zip.Writer
	fwriter *os.File
}

// NewZip returns an initialized Zip struct ready for use.
func NewZip() *Zip {
	return &Zip{
		Car: Car{t0: time.Now()},
	}
}

// Create creates a zip file from src in the dst
func (z *Zip) Create(dst string, src ...string) (cnt int, err error) {
	logger.Debug("Create Zipfile")

	// If there isn't a destination, return err
	if dst == "" {
		err = fmt.Errorf("destination required to create a zip archive")
		logger.Error(err)
		return 0, err
	}

	// If there aren't any sources, return err
	if len(dtv) == 0 {
		err = fmt.Errorf("a source is required to create a zip archive")
		logger.Error(err)
		return 0, err
	}

	// See if we can create the destination file before processing
	z.fwriter, err = os.OpenFile(dst, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		logger.Error(err)
		return 0, err
	}
	defer z.fwriter.Close()

	buf := new(bytes.Buffer)
	z.writer = zip.NewWriter(buf)
	defer z.writer.Close()

	logger.Debug("Setup channel")
	// Set up the file queue and its drain.
	z.FileCh = make(chan *os.File)
	wait, err := z.write()
	if err != nil {
		logger.Error(err)
		return 0, err
	}

	var fullPath string
	// Walk the sources, add each file to the queue.
	// This isn't limited as a large number of sources is not expected.
	//
	visitor := func(p string, fi os.FileInfo, err error) error {
		return z.AddFile(fullPath, p, fi, err)
	}

	var wg sync.WaitGroup
	wg.Add(len(src) - 1)
	for _, source := range src {
		logger.Debug(source)
		// first get the absolute, its needed either way
		fullPath, err = filepath.Abs(source)
		if err != nil {
			logger.Error(err)
			return 0, err
		}

		err = walk.Walk(fullPath, visitor)
		if err != nil {
			logger.Error(err)
			return 0, err
		}
	}

	logger.Debug("wg wait")
	wg.Wait()
	logger.Debug("before closing channel and waiting")

	close(z.FileCh)
	wait.Wait()

	z.writer.Close()

	// Copy the zip
	_, err = z.fwriter.Write(buf.Bytes())
	if err != nil {
		logger.Error(err)
		return 0, err
	}

	z.fwriter.Close()
	z.setDelta()
	return int(z.Car.files), nil
}

// ZipBytes takes a string and bytes and returns a zip archive of the bytes
// using the name.
func ZipBytes(b []byte, name string) (n int, zipped []byte, err error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	defer w.Close() // defer for convenience, though it may already be closed
	f, err := w.Create(name)
	if err != nil {
		logger.Error(err)
		return 0, zipped, err
	}

	n, err = f.Write(b)
	if err != nil {
		logger.Error(err)
		return n, zipped, err
	}
	w.Close() // we need to close it to get the bytes.
	return n, buf.Bytes(), err
}

func copyTo(w io.Writer, z *zip.File) (int64, error) {
	f, err := z.Open()
	if err != nil {
		return 0, err
	}
	defer f.Close()

	return io.Copy(w, f)
}

//
// Because zip can't be parallized because  `Create/CreateHEader` implicitly
// closes the writer and I don't feel like writing a parallized zip writer,
// we spawn a new goroutine for each file to read and pipe them to the zipper
// goroutine.
//
// SEE where to add defer
func (z *Zip) write() (*sync.WaitGroup, error) {
	logger.Debug("Write channel...")
	var wg sync.WaitGroup
	wg.Add(1)
	go func() error {
		defer wg.Done()

		for f := range z.FileCh {
			defer f.Close()
			logger.Debugf("write %s", f.Name())
			info, err := f.Stat()
			if err != nil {
				logger.Error(err)
				return err
			}

			if info.IsDir() {
				continue
			}

			header, err := zip.FileInfoHeader(info)
			if err != nil {
				logger.Error(err)
				return err
			}

			header.Name = f.Name()

			fw, err := z.writer.CreateHeader(header)
			if err != nil {
				logger.Error(err)
				return err
			}

			_, err = io.Copy(fw, f)
			if err != nil {
				logger.Error(err)
				return err
			}

		}

		return nil
	}()

	return &wg, nil
}

// ExtractFile extracts the content of src, a zip archive, to dst.
func (z *Zip) ExtractFile(src, dst string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		logger.Error(err)
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			logger.Error(err)
			return err
		}
		err = os.MkdirAll(filepath.Join(dst, filepath.Dir(f.Name)), 0755)
		if err != nil {
			logger.Error(err)
			return err
		}
		dF, err := os.Create(filepath.Join(dst, f.Name))
		if err != nil {
			logger.Error(err)
			return err
		}
		_, err = io.Copy(dF, rc)
		if err != nil {
			logger.Error(err)
			return err
		}
		rc.Close()
		dF.Close()
	}
	return nil
}
