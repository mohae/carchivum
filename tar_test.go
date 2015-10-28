package carchivum

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	magicnum "github.com/mohae/magicnum/mcompress"
)

func TestGzipTar(t *testing.T) {
	tmpDir, err := CreateTempFiles()
	if err != nil {
		t.Errorf("Expected creation of temp files to result in no error, got %q", err)
		return
	}
	newT := NewTar(filepath.Join(tmpDir, "test.tgz"))
	newT.Format = magicnum.Gzip
	// Test CreateTar
	cnt, err := newT.Create(newT.Name, filepath.Join(tmpDir, "test"))
	if err != nil {
		t.Errorf("Expected creation of tar to result in no error, got %q", err)
	}
	if cnt != 5 {
		t.Errorf("Expected a count of 5; got %d", cnt)
	}
	// Check the created tarfile
	tFi, err := os.Stat(newT.Name)
	if err != nil {
		t.Errorf("expected stat of created tar to not result in an error, got %q", err)
	}
	// check range because the returned size can vary by a few bytes.
	if tFi.Size() == 0 {
		t.Error("Expected Filesize to be  > 0. it wasn't")
	}

	// Test Extract Tar
	srcF, err := os.Open(newT.Name)
	if err != nil {
		t.Errorf("expected open of %q to not result in an error, got %q", newT.Name, err)
	}
	defer srcF.Close()
	eDir := filepath.Join(tmpDir, "extract")
	newT.OutDir = eDir
	err = newT.ExtractArchive(srcF)
	if err != nil {
		t.Errorf("expected extract of tar to not result in an error, got %q", err)
	}
	// see that the extracte files are there and are as expected.
	filepath.Join(tmpDir, "tmp")

	tests := []struct {
		path     string
		expected string
	}{
		{"test/test1.txt", "some content\n"},
		{"test/test2.txt", "some more content\n"},
		{"test/dir/test1.txt", "different content\n"},
		{"test/dir/test2.txt", "might be different content\n"},
	}
	for i, test := range tests {
		fB, err := ioutil.ReadFile(filepath.Join(eDir, test.path))
		if err != nil {
			t.Errorf("%d: expected read of %q to not error; got %q", i, test.path, err)
		}
		if test.expected != string(fB) {
			t.Errorf("%d: expected file to contents to be %q got %q", i, test.expected, string(fB))
		}
	}
}
