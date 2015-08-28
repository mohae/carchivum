package carchivum

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGzipTar(t *testing.T) {
	tmpDir, err := CreateTempFiles()
	assert.Nil(t, err)
	newT := NewTar()
	newT.Format = GzipFmt
	tarName := filepath.Join(tmpDir, "test.tgz")
	// Test CreateTar
	cnt, err := newT.Create(tarName, filepath.Join(tmpDir, "test"))
	assert.Nil(t, err)
	assert.Equal(t, 5, cnt)
	// Check the created tarfile
	tFi, err := os.Stat(tarName)
	assert.Nil(t, err)
	// check range because the returned size can vary by a few bytes.
	if tFi.Size() < 230 || tFi.Size() > 236 {
		t.Errorf("Expected Filesize to be 233 +- 3 bytes, got %d", tFi.Size())
	}

	// Test Extract T ar
	srcF, err := os.Open(tarName)
	assert.Nil(t, err)
	defer srcF.Close()
	err = newT.Extract(srcF, filepath.Join(tmpDir, "tmp"))
	assert.Nil(t, err)
	// see that the extracte files are there and are as expected.
	eDir := filepath.Join(tmpDir, "tmp")
	fB, err := ioutil.ReadFile(filepath.Join(eDir, "test/test1.txt"))
	assert.Nil(t, err)
	assert.Equal(t, "some content\n", string(fB))

	fB, err = ioutil.ReadFile(filepath.Join(eDir, "test/test2.txt"))
	assert.Nil(t, err)
	assert.Equal(t, "some more content\n", string(fB))

	fB, err = ioutil.ReadFile(filepath.Join(eDir, "test/dir/test1.txt"))
	assert.Nil(t, err)
	assert.Equal(t, "different content\n", string(fB))

	fB, err = ioutil.ReadFile(filepath.Join(eDir, "test/dir/test2.txt"))
	assert.Nil(t, err)
	assert.Equal(t, "might be different content\n", string(fB))
}

func TestZTar(t *testing.T) {
	tmpDir, err := CreateTempFiles()
	assert.Nil(t, err)
	newT := NewTar()
	newT.Format = LZWFmt
	tarName := filepath.Join(tmpDir, "test.tz2")
	// Test CreateTar
	cnt, err := newT.Create(tarName, filepath.Join(tmpDir, "test"))
	assert.Nil(t, err)
	assert.Equal(t, 5, cnt)
	// Check the created tarfile
	tFi, err := os.Stat(tarName)
	assert.Nil(t, err)
	// check range because the returned size can vary by a few bytes.
	if tFi.Size() < 410 || tFi.Size() > 430 {
		t.Errorf("Expected Filesize to be 420 +- 10 bytes, got %d", tFi.Size())
	}
	// Test Extract T ar
	srcF, err := os.Open(tarName)
	assert.Nil(t, err)
	defer srcF.Close()
	err = newT.Extract(srcF, filepath.Join(tmpDir, "tmp"))
	assert.Nil(t, err)
	// see that the extracte files are there and are as expected.
	eDir := filepath.Join(tmpDir, "tmp")
	fB, err := ioutil.ReadFile(filepath.Join(eDir, "test/test1.txt"))
	assert.Nil(t, err)
	assert.Equal(t, "some content\n", string(fB))

	fB, err = ioutil.ReadFile(filepath.Join(eDir, "test/test2.txt"))
	assert.Nil(t, err)
	assert.Equal(t, "some more content\n", string(fB))

	fB, err = ioutil.ReadFile(filepath.Join(eDir, "test/dir/test1.txt"))
	assert.Nil(t, err)
	assert.Equal(t, "different content\n", string(fB))

	fB, err = ioutil.ReadFile(filepath.Join(eDir, "test/dir/test2.txt"))
	assert.Nil(t, err)
	assert.Equal(t, "might be different content\n", string(fB))
}
