package carchivum

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/dotcloud/tar"
	"github.com/stretchr/testify/assert"
)

type testFile struct {
	name    string
	content []byte
}

var TestFiles []testFile

func initTestFiles() {
	TestFiles = make([]testFile, 0)
	TestFiles = append(TestFiles, testFile{name: "test/test1.txt", content: []byte("some content\n")})
	TestFiles = append(TestFiles, testFile{name: "test/test2.txt", content: []byte("some more content\n")})
	TestFiles = append(TestFiles, testFile{name: "test/dir/test1.txt", content: []byte("different content\n")})
	TestFiles = append(TestFiles, testFile{name: "test/dir/test2.txt", content: []byte("might be different content\n")})
}

func CreateTempTgz() (string, error) {
	tmpDir := os.TempDir()
	testTar := filepath.Join(tmpDir, "test.tgz")
	f, _ := os.OpenFile(testTar, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
	// create a tgz file
	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	for _, testF := range TestFiles {
		hdr := &tar.Header{
			Name: testF.name,
			Size: int64(len(testF.content)),
		}
		err := tw.WriteHeader(hdr)
		if err != nil {
			return testTar, err
		}
		_, err = tw.Write(testF.content)
		if err != nil {
			return testTar, err
		}
	}
	return testTar, nil
}

func CreateTempFiles() (dir string, err error) {
	initTestFiles()
	tmpDir, _ := ioutil.TempDir("", "car")
	err = os.Chdir(tmpDir)
	if err != nil {
		return "", err
	}
	err = os.MkdirAll("test/dir", 0755)
	if err != nil {
		return "", err
	}
	for _, f := range TestFiles {
		err = ioutil.WriteFile(f.name, f.content, 0755)
		if err != nil {
			return tmpDir, err
		}
	}
	return tmpDir, nil
}

func TestGetFileFormat(t *testing.T) {
	tests := []struct {
		bytes       []byte
		typ         Format
		expectedErr string
	}{
		{[]byte{0x1f, 0x8b}, FmtGzip, ""},
		{[]byte{0x75, 0x73, 0x74, 0x61, 0x72, 0x00, 0x30, 0x30}, FmtTar, ""},
		{[]byte{0x75, 0x73, 0x74, 0x61, 0x72, 0x00, 0x20, 0x00}, FmtTar, ""},
		{[]byte{0x50, 0x4b, 0x03, 0x04}, FmtZip, ""},
		{[]byte{0x50, 0x4b, 0x05, 0x06}, FmtZipEmpty, "empty zip archive not supported"},
		{[]byte{0x50, 0x4b, 0x07, 0x08}, FmtZipSpanned, "spanned zip archive not supported"},
		{[]byte{0x42, 0x5a, 0x68}, FmtBzip2, "bzip2 not supported"},
		{[]byte{0x1f, 0xa0}, FmtLZH, "LZH not supported"},
		{[]byte{0x1f, 0x9d}, FmtLZW, "LZW not supported"},
		{[]byte{0x52, 0x61, 0x72, 0x21, 0x1a, 0x07, 0x01, 0x00}, FmtRAR, "RAR post 5.0 not supported"},
		{[]byte{0x52, 0x61, 0x72, 0x21, 0x1a, 0x07, 0x00}, FmtRAROld, "RAR pre 1.5 not supported"},
	}
	for _, test := range tests {
		r := bytes.NewReader(test.bytes)
		format, err := GetFileFormat(r)
		if test.expectedErr != "" {
			assert.Equal(t, test.expectedErr, err.Error())
			continue
		}

		assert.Nil(t, err)
		assert.Equal(t, test.typ, format)
	}

}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		value       string
		expected    Format
		expectedErr string
	}{
		{"gzip", FmtGzip, ""},
		{"tar.gz", FmtGzip, ""},
		{"tgz", FmtGzip, ""},
		{"tar", FmtTar, ""},
		{"zip", FmtZip, ""},
		{"z", FmtUnsupported, "unsupported not supported"},
	}

	for _, test := range tests {
		f, err := ParseFormat(test.value)
		if test.expectedErr != "" {
			assert.Equal(t, test.expectedErr, err.Error())
			assert.Equal(t, FmtUnsupported, f)
			continue
		}
		assert.Nil(t, err)
		assert.Equal(t, test.expected, f)
	}
}

func TestGetFileParts(t *testing.T) {
	tests := []struct {
		value            string
		expectedDir      string
		expectedFilename string
		expectedExt      string
		expectedErr      string
	}{
		{"", "", "", "", ""},
		{"test", "", "test", "", ""},
		{"test.tar", "", "test", "tar", ""},
		{"/dir/name/test.tar", "/dir/name/", "test", "tar", ""},
		{"dir/name/test.tar", "dir/name/", "test", "tar", ""},
		{"../dir/name/test.tar", "../dir/name/", "test", "tar", ""},
	}

	for _, test := range tests {
		dir, fname, ext, err := getFileParts(test.value)
		if test.expectedErr != "" {
			assert.Equal(t, test.expectedErr, err.Error())
			continue
		}
		assert.Nil(t, err)
		assert.Equal(t, test.expectedDir, dir)
		assert.Equal(t, test.expectedFilename, fname)
		assert.Equal(t, test.expectedExt, ext)

	}
}

/*
func TestAddFile(t *testing.T) {
	tests := []struct {
		value       string
		expectedErr string
		expectedB   int
	}{
		{"", "open : no such file or directory", 0},
		{"test", "open test: no such file or directory", 0},
		{"test_files/pixies/born-in-chicago.txt", "", 609},
	}

	for _, test := range tests {
		buf := new(bytes.Buffer)
		w := zip.NewWriter(buf)

		b, err := addFile(w, test.value)
		if test.expectedErr != "" {
			assert.Equal(t, test.expectedErr, err.Error())
			continue
		}

		assert.Nil(t, err)
		assert.Equal(t, test.expectedB, b)

		w.Close()
	}
}
*/
