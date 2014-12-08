package carchivum

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateTar(t *testing.T) {
	tmpDir, err := CreateTempFiles()
	assert.Nil(t, err)
	newT := NewTar()
	newT.format = FmtGzip
	tarName := filepath.Join(tmpDir, "test.tgz")
	// Create the tar file
	cnt, err := newT.Create(tarName, filepath.Join(tmpDir, "test"))
	assert.Nil(t, err)
	assert.Equal(t, 5, cnt)
	// Check the created tarfile
	tFi, err := os.Stat(tarName)
	assert.Nil(t, err)
	assert.Equal(t, 233, tFi.Size()) // This is coming back 233 or 234; I'm doing something wrong here
}
