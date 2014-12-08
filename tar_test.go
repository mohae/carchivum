package carchivum

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTar(t *testing.T) {
	tmpDir, err := CreateTempFiles()
	assert.Nil(t, err)
	newT := NewTar()
	newT.format = FmtGzip
	tarName := filepath.Join(tmpDir, "test.tgz")
	// Test CreateTar
	cnt, err := newT.Create(tarName, filepath.Join(tmpDir, "test"))
	assert.Nil(t, err)
	assert.Equal(t, 5, cnt)
	// Check the created tarfile
	tFi, err := os.Stat(tarName)
	assert.Nil(t, err)
	assert.Equal(t, 233, tFi.Size()) // This is coming back 233 or 234; I'm doing something wrong here

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
