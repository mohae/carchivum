package carchivum

import (
	"archive/tar"
	"compress/gzip"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

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
