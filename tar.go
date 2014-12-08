// tar implements the tape archive format.
package carchivum

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/MichaelTJones/walk"
)

type Tar struct {
	Car
	writer  *tar.Writer
	format  Format
	sources []string
}

func NewTar() *Tar {
	return &Tar{Car: Car{t0: time.Now()}, format: defaultFormat, sources: []string{}}
}

func (t *Tar) Create(dst string, src ...string) (cnt int, err error) {
	logger.Debug("Create Tarfile")

	// If there isn't a destination, return err
	if dst == "" {
		err = fmt.Errorf("destination required to create a tar archive")
		logger.Error(err)
		return 0, err
	}

	// If there aren't any sources, return err
	if len(src) == 0 {
		err = fmt.Errorf("a source is required to create a tar archive")
		logger.Error(err)
		return 0, err
	}

	t.sources = src
	// See if we can create the destination file before processing
	tball, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		logger.Error(err)
		return 0, err
	}
	defer func() {
		cerr := tball.Close()
		if cerr != nil && err == nil {
			logger.Error(cerr)
			err = cerr
		}
	}()

	switch t.format {
	case FmtGzip:
		err = t.CreateGzip(tball)
		if err != nil {
			logger.Error(err)
			return 0, err
		}
	}

	if t.DeleteArchived {
		err := t.removeFiles()
		if err != nil {
			err = fmt.Errorf("an error was encountered while deleting the archived files; some files may not be deleted: %q", err)
			logger.Error(err)
			return 0, err
		}
	}

	t.setDelta()
	return int(t.Car.files), nil
}

func (t *Tar) removeFiles() error {
	for _, file := range t.deleteList {
		err := os.Remove(file)
		if err != nil {
			logger.Error(err)
			return err
		}
	}
	return nil
}

func (t *Tar) Delete() error {
	return nil
}

func (t *Tar) Extract(src io.Reader, dst string) error {
	switch t.format {
	case FmtGzip:
		return t.ExtractTgz(src, dst)
	default:
		return FmtUnsupported.NotSupportedError()
	}
	return nil
}
func (t *Tar) ExtractTgz(src io.Reader, dst string) error {
	gr, err := gzip.NewReader(src)
	if err != nil {
		logger.Error(err)
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	//defer tr.Close()
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // break at eof
		}
		if err != nil {
			logger.Error(err)
			return err
		}
		if hdr.Name == "." {
			continue // skip .
		}
		err = extractTarFile(hdr, dst, tr)
		if err != nil {
			logger.Error(err)
			return err
		}
	}
	return nil
}

func extractTarFile(hdr *tar.Header, dst string, in io.Reader) error {
	fP := filepath.Join(dst, hdr.Name)
	fI := hdr.FileInfo()

	err := os.MkdirAll(filepath.Dir(fP), fI.Mode())
	if err != nil {
		logger.Error(err)
		return err
	}
	if fI.IsDir() {
		return nil
	}
	if fI.Mode()&os.ModeSymlink != 0 {
		return os.Symlink(hdr.Linkname, fP)
	}

	dF, err := os.OpenFile(fP, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fI.Mode())
	if err != nil {
		logger.Error(err)
		return err
	}

	_, err = io.Copy(dF, in)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}
func (t *Tar) CreateGzip(w io.Writer) (err error) {
	gw := gzip.NewWriter(w)
	// Close the file with error handling
	defer func() {
		cerr := gw.Close()
		if cerr != nil && err == nil {
			logger.Error(cerr)
			err = cerr
		}
		logger.Debug("Closed gtar writer")
	}()

	err = t.writeTar(gw)
	return err
}

func (t *Tar) writeTar(w io.Writer) (err error) {
	t.writer = tar.NewWriter(w)
	defer func() {
		cerr := t.writer.Close()
		if cerr != nil && err == nil {
			logger.Error(cerr)
			err = cerr
		}
		logger.Debug("closed tar writer")
	}()

	logger.Debug("Setup channel")
	t.FileCh = make(chan *os.File)

	wait, err := t.Write()
	if err != nil {
		logger.Error(err)
		return err
	}

	var fullPath string
	visitor := func(p string, fi os.FileInfo, err error) error {
		return t.AddFile(fullPath, p, fi, err)
	}

	var wg sync.WaitGroup
	wg.Add(len(t.sources) - 1)
	for _, source := range t.sources {
		logger.Debug(source)

		fullPath, err = filepath.Abs(source)
		if err != nil {
			logger.Error(err)
			return err
		}

		err = walk.Walk(fullPath, visitor)
		if err != nil {
			logger.Error(err)
			return err
		}
	}

	logger.Debug("wg wait")
	wg.Wait()
	close(t.FileCh)
	wait.Wait()

	return err

}

func (t *Tar) Write() (*sync.WaitGroup, error) {
	logger.Debug("Write channel...")
	var wg sync.WaitGroup
	wg.Add(1)
	go func() error {
		defer wg.Done()

		for f := range t.FileCh {
			logger.Debugf("write %s", f.Name())
			info, err := f.Stat()
			if err != nil {
				logger.Error(err)
				return err
			}

			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				logger.Error(err)
				return err
			}

			header.Name = f.Name()

			// See if any header overrides need to be done
			if t.Owner > 0 {
				header.Uid = t.Owner
			}

			if t.Group > 0 {
				header.Gid = t.Group
			}

			if t.Mode > 0 {
				header.Mode = int64(t.Mode)
			}

			logger.Debugf("%+v", header)

			err = t.writer.WriteHeader(header)
			if err != nil {
				logger.Error(err)
				return err
			}

			io.Copy(t.writer, f)
			if err != nil {
				logger.Error(err)
				return err
			}

			err = f.Close()
			if err != nil {
				logger.Error(err)
				return err
			}
		}

		return nil
	}()

	return &wg, nil
}
