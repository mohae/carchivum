package carchivum

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/MichaelTJones/walk"
	magicnum "github.com/mohae/magicnum/compress"
	"github.com/pierrec/lz4"
)

// Tar is a struct for a tar, tape archive.
type Tar struct {
	Car
	*tar.Writer
	magicnum.Format
	sources []string
}

// NewTar returns an initialized Tar struct ready for use.
func NewTar(n string) *Tar {
	return &Tar{Car: Car{Name: n, t0: time.Now()}, Format: defaultFormat, sources: []string{}}
}

// Create creates a tarfile from the passed src('s) and saves it to the dst.
func (t *Tar) Create(src ...string) (cnt int, err error) {
	// If there isn't a destination, return err
	if t.Name == "" {
		return 0, fmt.Errorf("destination required to create a tar archive")
	}
	// If there aren't any sources, return err
	if len(src) == 0 {
		return 0, fmt.Errorf("a source is required to create a tar archive")
	}
	t.sources = src
	// See if we can create the destination file before processing
	tball, err := os.OpenFile(t.Name, os.O_RDWR|os.O_CREATE, 0744)
	if err != nil {
		return 0, err
	}
	defer func() {
		cerr := tball.Close()
		if cerr != nil && err == nil {
			err = cerr
		}
	}()
	switch t.Format {
	case magicnum.GZip:
		err = t.CreateGZip(tball)
		if err != nil {
			return 0, err
		}
	case magicnum.BZip2:
		err = fmt.Errorf("Bzip2 compression is not supported")
		return 0, err
	case magicnum.LZ4:
		err = t.CreateLZ4(tball)
		if err != nil {
			return 0, err
		}
	default:
		err = fmt.Errorf("Unsupported compression format: %s", t.Format.String())
		return 0, err
	}
	if t.DeleteArchived {
		err := t.removeFiles()
		if err != nil {
			return 0, fmt.Errorf("an error was encountered while deleting the archived files; some files may not be deleted: %s", err)
		}
	}
	t.setDelta()
	return int(t.Car.files), nil
}

func (t *Tar) removeFiles() error {
	for _, file := range t.deleteList {
		err := os.Remove(file)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateGZip creates a GZip using the passed writer.
func (t *Tar) CreateGZip(w io.Writer) (err error) {
	zw := gzip.NewWriter(w)
	// Close the file with error handling
	defer func() {
		cerr := zw.Close()
		if cerr != nil && err == nil {
			err = cerr
		}
	}()
	err = t.writeTar(zw)
	return err
}

// CreateLZW compresses using LZW and LSB order using the passed writer.
// TODO: address order so that it doesn't necessarily default to LSB
// NOTE/TODO: NOT SUPPORTED for now. Need to get a better understanding of
//     its magic number, if lzw has one. Might be best to leave lzw support
//     for pdfs,
/*
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
*/

// CreateLZ4 creates a LZ4 compressed tarball using the passed writer.
func (t *Tar) CreateLZ4(w io.Writer) (err error) {
	lzW := lz4.NewWriter(w)
	err = t.writeTar(lzW)
	lzW.Close()
	return err
}

func (t *Tar) writeTar(w io.Writer) (err error) {
	t.Writer = tar.NewWriter(w)
	defer func() {
		cerr := t.Writer.Close()
		if cerr != nil && err == nil {
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
	for _, source := range t.sources {
		fullPath, err = filepath.Abs(source)
		if err != nil {
			return err
		}
		err = walk.Walk(fullPath, visitor)
		if err != nil {
			return err
		}
	}
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
				return err
			}
			if info.IsDir() {
				continue
			}
			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
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
			if t.FileMode > 0 {
				header.Mode = int64(t.FileMode)
			} else {
				header.Mode = int64(info.Mode().Perm())
			}
			header.ModTime = info.ModTime()
			err = t.Writer.WriteHeader(header)
			if err != nil {
				return err
			}
			_, err = io.Copy(t.Writer, f)
			if err != nil {
				return err
			}
			err = f.Close()
			if err != nil {
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

// Extract extracts the files from the src and writes them to the dst. The src
// is either a tar or a compressed tar.
func (t *Tar) Extract() error {
	// open the file
	f, err := os.Open(t.Name)
	if err != nil {
		return err
	}
	// find its format
	t.Format, err = magicnum.GetFormat(f)
	if err != nil {
		return err
	}
	defer f.Close()
	return t.ExtractArchive(f)
}

// ExtractArchive takes a compressed tar archive, as an io.Reader.  If the compression
// format used is supported, it will decompress and extract the contents of the tar;
// otherwise it will return an error.
func (t *Tar) ExtractArchive(src io.Reader) error {
	switch t.Format {
	case magicnum.Tar:
		return t.ExtractTar(src)
	case magicnum.GZip:
		return t.ExtractTgz(src)
	case magicnum.BZip2:
		return t.ExtractTbz(src)
	case magicnum.LZ4:
		return t.ExtractLZ4(src)
	default:
		return fmt.Errorf("%s is not a supported format", t.Format)
	}
}

// ExtractTar extracts a tar file using the passed reader
func (t *Tar) ExtractTar(src io.Reader) (err error) {
	tr := tar.NewReader(src)
	for {
		header, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		fname := header.Name
		// extract is always relative to cwd, for now
		// temporarily commented out because dst is no longer supported
		// TODO add flag for destinatiion
		fname = filepath.Join(t.OutDir, fname)
		switch header.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(fname, 0744)
			if err != nil {
				return err
			}
			// set the final element to the appropriate permission
			err = os.Chmod(fname, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
		case tar.TypeReg:
			// create the parent directory if necessary
			pdir := filepath.Dir(fname)
			err = os.MkdirAll(pdir, 0744)
			if err != nil {
				return err
			}
			if err != nil {
				return err
			}
			w, err := os.Create(fname)
			if err != nil {
				return err
			}
			io.Copy(w, tr)
			err = os.Chmod(fname, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			w.Close()
		default:
			return fmt.Errorf("Unable to extract type: %c in file %s", header.Typeflag, fname)
		}
	}
	return nil
}

// ExtractGzip reads a GZip using the passed reader.
func (t *Tar) ExtractGzip(src io.Reader) (err error) {
	gR, err := gzip.NewReader(src)
	if err != nil {
		return err
	}
	// Close the file with error handling
	defer func() {
		cerr := gR.Close()
		if cerr != nil && err == nil {
			err = cerr
		}
	}()
	return t.ExtractTar(gR)
}

// ExtractTgz extracts GZip'd tarballs.
func (t *Tar) ExtractTgz(src io.Reader) error {
	gr, err := gzip.NewReader(src)
	if err != nil {
		return err
	}
	defer gr.Close()
	return t.ExtractTar(gr)
}

// ExtractTbz extracts Bzip2 compressed tarballs.
func (t *Tar) ExtractTbz(src io.Reader) error {
	zR := bzip2.NewReader(src)
	return t.ExtractTar(zR)
}

// ExtractZ extracts tarballs compressed with LZW, typically .Z extension.
// TODO fix so that order and width get properly set. Assuming order isn't
// good.
// NOTE/TODO: NOT SUPPORTED for now I need to bet a better understanding for
//     its magic numbers, if there are any for general lzw. Maybe it's best
//     to use this just for pdf. gif, etc.
/*
func (t *Tar) ExtractZ(src io.Reader) error {
	zR := lzw.NewReader(src, lzw.LSB, 8)
	defer zR.Close()
	return t.ExtractTar(zR)
}
*/

// ExtractLZ4 extracts LZ4 compressed tarballs.
func (t *Tar) ExtractLZ4(src io.Reader) error {
	lzR := lz4.NewReader(src)
	tr := tar.NewReader(lzR)
	err := t.ExtractTar(tr)
	return err
}

func extractTarFile(hdr *tar.Header, dst string, src io.Reader) error {
	fP := filepath.Join(dst, hdr.Name)
	fI := hdr.FileInfo()
	err := os.MkdirAll(filepath.Dir(fP), fI.Mode())
	if err != nil {
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
		return err
	}
	defer dF.Close()

	_, err = io.Copy(dF, src)
	if err != nil {
		return err
	}
	return nil
}
