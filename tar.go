package carchivum

// tar implements the tape archive format.

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"compress/lzw"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/MichaelTJones/walk"
	"github.com/pierrec/lz4"
)

// Tar is a struct for a tar, tape archive.
type Tar struct {
	Car
	writer *tar.Writer
	Format
	sources []string
}

// NewTar returns an initialized Tar struct ready for use.
func NewTar() *Tar {
	return &Tar{Car: Car{t0: time.Now()}, Format: defaultFormat, sources: []string{}}
}

// Create creates a tarfile from the passed src('s) and saves it to the dst.
func (t *Tar) Create(dst string, src ...string) (cnt int, err error) {
	// If there isn't a destination, return err
	if dst == "" {
		err = fmt.Errorf("destination required to create a tar archive")
		log.Print(err)
		return 0, err
	}

	// If there aren't any sources, return err
	if len(src) == 0 {
		err = fmt.Errorf("a source is required to create a tar archive")
		log.Print(err)
		return 0, err
	}

	t.sources = src
	// See if we can create the destination file before processing
	tball, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, 0744)
	if err != nil {
		log.Print(err)
		return 0, err
	}
	defer func() {
		cerr := tball.Close()
		if cerr != nil && err == nil {
			log.Print(cerr)
			err = cerr
		}
	}()
	switch t.Format {
	case GzipFmt:
		err = t.CreateGzip(tball)
		if err != nil {
			log.Print(err)
			return 0, err
		}
	case Bzip2Fmt:
		err = fmt.Errorf("Bzip2 compression is not supported")
		return 0, err
	case LZWFmt:
		err = t.CreateZ(tball)
		if err != nil {
			log.Print(err)
			return 0, err
		}
	case LZ4Fmt:
		err = t.CreateLZ4(tball)
		if err != nil {
			log.Print(err)
			return 0, err
		}
	default:
		err = fmt.Errorf("Unsupported compression format: %s", t.Format.String())
		return 0, err
	}

	if t.DeleteArchived {
		err := t.removeFiles()
		if err != nil {
			err = fmt.Errorf("an error was encountered while deleting the archived files; some files may not be deleted: %q", err)
			log.Print(err)
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
			log.Print(err)
			return err
		}
	}
	return nil
}

// CreateGzip creates a GZip using the passed writer.
func (t *Tar) CreateGzip(w io.Writer) (err error) {
	zw := gzip.NewWriter(w)
	// Close the file with error handling
	defer func() {
		cerr := zw.Close()
		if cerr != nil && err == nil {
			log.Print(cerr)
			err = cerr
		}
	}()

	err = t.writeTar(zw)
	return err
}

// CreateLZW compresses using LZW and LSB order using the passed writer.
// TODO: address order so that it doesn't necessarily default to LSB
func (t *Tar) CreateZ(w io.Writer) (err error) {
	zw := lzw.NewWriter(w, lzw.LSB, 8)
	// Close the file with error handling
	defer func() {
		cerr := zw.Close()
		if cerr != nil && err == nil {
			log.Print(cerr)
			err = cerr
		}
	}()

	err = t.writeTar(zw)
	return err
}

// CreateLZ4 creates a LZ4 compressed tarball using the passed writer.
func (t *Tar) CreateLZ4(w io.Writer) (err error) {
	lzW := lz4.NewWriter(w)
	err = t.writeTar(lzW)
	lzW.Close()
	return err
}

func (t *Tar) writeTar(w io.Writer) (err error) {
	t.writer = tar.NewWriter(w)
	defer func() {
		cerr := t.writer.Close()
		if cerr != nil && err == nil {
			log.Print(cerr)
			err = cerr
		}
	}()

	t.FileCh = make(chan *os.File)

	wait, err := t.Write()
	if err != nil {
		log.Print(err)
		return err
	}

	var fullPath string
	visitor := func(p string, fi os.FileInfo, err error) error {
		return t.AddFile(fullPath, p, fi, err)
	}

	var wg sync.WaitGroup
	wg.Add(len(t.sources) - 1)
	for _, source := range t.sources {
		fullPath, err = filepath.Abs(source)
		if err != nil {
			log.Print(err)
			return err
		}

		err = walk.Walk(fullPath, visitor)
		if err != nil {
			log.Print(err)
			return err
		}
	}

	wg.Wait()
	close(t.FileCh)
	wait.Wait()

	return err

}

// Write adds the files received on the channel to the tarball.
func (t *Tar) Write() (*sync.WaitGroup, error) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() error {
		defer wg.Done()

		for f := range t.FileCh {
			info, err := f.Stat()
			if err != nil {
				log.Print(err)
				return err
			}
			if info.IsDir() {
				continue
			}

			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				log.Print(err)
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
			} else {
				header.Mode = int64(info.Mode().Perm())
			}

			header.ModTime = info.ModTime()

			err = t.writer.WriteHeader(header)
			if err != nil {
				log.Print(err)
				return err
			}

			_, err = io.Copy(t.writer, f)
			if err != nil {
				log.Print(err)
				return err
			}

			err = f.Close()
			if err != nil {
				log.Print(err)
				return err
			}
		}
		return nil
	}()
	return &wg, nil
}

// Delete is not implemented
func (t *Tar) Delete() error {
	return nil
}

// Extract extracts the files from src and writes them to the dst.
func (t *Tar) Extract(src io.Reader, dst string) error {
	switch t.Format {
	case GzipFmt:
		return t.ExtractTgz(src, dst)
	case Bzip2Fmt:
		return t.ExtractTbz(src, dst)
	case LZWFmt:
		return t.ExtractZ(src, dst)
	case LZ4Fmt:
		return t.ExtractLZ4(src, dst)
	default:
		return UnsupportedFmt.NotSupportedError()
	}
	return nil
}

// ExtractTar extracts a tar file using the passed reader
func (t *Tar) ExtractTar(r io.Reader, dst string) (err error) {
	tr := tar.NewReader(r)

	for {
		header, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Print(err)
			return err
		}

		fname := header.Name
		// extract is always relative to cwd, for now
		fname = filepath.Join(dst, fname)
		fmt.Printf("%s %s\n", fname, strconv.Itoa(int(header.Mode)))
		switch header.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(fname, 0744)
			if err != nil {
				log.Print(err)
				return err
			}
			// set the final element to the appropriate permission
			err = os.Chmod(fname, os.FileMode(header.Mode))
			if err != nil {
				log.Print(err)
				return err
			}
		case tar.TypeReg:
			// create the parent directory if necessary
			pdir := filepath.Dir(fname)
			err = os.MkdirAll(pdir, 0744)
			if err != nil {
				log.Print(err)
				return err
			}
			if err != nil {
				log.Print(err)
				return err
			}
			w, err := os.Create(fname)
			if err != nil {
				log.Print(err)
				return err
			}
			io.Copy(w, tr)
			err = os.Chmod(fname, os.FileMode(header.Mode))
			if err != nil {
				log.Print(err)
				return err
			}
			w.Close()
		default:
			err = fmt.Errorf("Unable to extract type: %c in file %s", header.Typeflag, fname)
			log.Print(err)
			return err
		}
	}

	return nil
}

// ExtractGzip reads a GZip using the passed reader.
func (t *Tar) ExtractGzip(r io.Reader, dst string) (err error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		log.Print(err)
		return err
	}

	// Close the file with error handling
	defer func() {
		cerr := gr.Close()
		if cerr != nil && err == nil {
			log.Print(cerr)
			err = cerr
		}
	}()

	err = t.ExtractTar(gr, dst)
	return err
}

// ExtractTgz extracts GZip'd tarballs.
func (t *Tar) ExtractTgz(src io.Reader, dst string) error {
	gr, err := gzip.NewReader(src)
	if err != nil {
		log.Print(err)
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
			log.Print(err)
			return err
		}
		if hdr.Name == "." {
			continue // skip .
		}
		err = extractTarFile(hdr, dst, tr)
		if err != nil {
			log.Print(err)
			return err
		}
	}
	return nil
}

// ExtractTbz extracts Bzip2 compressed tarballs.
func (t *Tar) ExtractTbz(src io.Reader, dst string) error {
	zr := bzip2.NewReader(src)

	tr := tar.NewReader(zr)
	//defer tr.Close()
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // break at eof
		}
		if err != nil {
			log.Print(err)
			return err
		}
		if hdr.Name == "." {
			continue // skip .
		}
		err = extractTarFile(hdr, dst, tr)
		if err != nil {
			log.Print(err)
			return err
		}
	}
	return nil
}

// ExtractZ extracts tarballs compressed with LZW, typically .Z extension.
// TODO fix so that order and width get properly set. Assuming order isn't
// good.
func (t *Tar) ExtractZ(src io.Reader, dst string) error {
	zr := lzw.NewReader(src, lzw.LSB, 8)
	defer zr.Close()

	tr := tar.NewReader(zr)
	//defer tr.Close()
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // break at eof
		}
		if err != nil {
			log.Print(err)
			return err
		}
		if hdr.Name == "." {
			continue // skip .
		}
		err = extractTarFile(hdr, dst, tr)
		if err != nil {
			log.Print(err)
			return err
		}
	}
	return nil
}

// ExtractLZ4 extracts LZ4 compressed tarballs.
func (t *Tar) ExtractLZ4(src io.Reader, dst string) error {
	lzR := lz4.NewReader(src)
	tr := tar.NewReader(lzR)
	err := t.ExtractTar(tr, dst)
	return err
}

func extractTarFile(hdr *tar.Header, dst string, in io.Reader) error {
	fP := filepath.Join(dst, hdr.Name)
	fI := hdr.FileInfo()

	err := os.MkdirAll(filepath.Dir(fP), fI.Mode())
	if err != nil {
		log.Print(err)
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
		log.Print(err)
		return err
	}
	defer dF.Close()

	_, err = io.Copy(dF, in)
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}
