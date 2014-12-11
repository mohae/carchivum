// extract.go
package carchivum

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
)

// Extract extracts the source to the destination. If no destination is specified
// the files will be extracted relative to '.'. If CreateDir, a directory with the
// same name as the archive filename will be created and the files will be
// extracted there.
//
// The format is detected from the file itself; the extension is not used as a hint.
func Extract(src, dst string) (message string, err error) {
	if src == "" {
		return "", fmt.Errorf("expected source, none provided")
	}

	// Open file
	srcF, err := os.Open(src)
	if err != nil {
		logger.Error(err)
		return "", err
	}
	defer srcF.Close() // don't error check as it may be closed
	// set dst
	if dst == "" && CreateDir {
		// make the destination folder the same as the src name
		_, f, _, err := getFileParts(src)
		if err != nil {
			logger.Error(err)
			return "", err
		}
		dst = f
	}

	// Check format
	format, err := GetFileFormat(srcF)

	// Extract accordingly
	switch format {
	case FmtGzip:
		err = extractGzip(srcF, dst)
	case FmtZip:
		srcF.Close()
		err = extractZipFile(src, dst)
	case FmtTar:
		err = extractTar(srcF, dst)
	default:
		return "not implemented", nil
	}

	if err != nil {
		logger.Error(err)
		return "", err
	}
	return fmt.Sprintf("%q extracted to %q", src, dst), nil
}

func extractTar(src io.Reader, dst string) error {
	t := NewTar()
	err := t.Extract(src, dst)
	if err != nil {
		logger.Error(err)
	}
	return err
}

func extractGzip(src *os.File, dst string) error {
	gR, err := gzip.NewReader(src)
	if err != nil {
		return err
	}
	defer func() {
		cerr := gR.Close()
		if cerr != nil {
			logger.Error(cerr)
			if err == nil {
				err = cerr
			}
		}
	}()

	return extractTar(gR, dst)
}

func extractZipFile(src, dst string) error {
	z := NewZip()
	return z.ExtractFile(src, dst)
}
