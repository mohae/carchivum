package carchivum

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtract(t *testing.T) {
	fname, err := CreateTempTgz()
	assert.Nil(t, err)
	tDir := os.TempDir()
	_ = os.Chdir(tDir)
	//tball := NewTar()
	msg, err := Extract(fname, tDir)
	assert.Nil(t, err)
	assert.Equal(t, "\"/tmp/test.tgz\" extracted to \"/tmp\"", msg)
	for _, tfiles := range TestFiles {
		f, err := os.Open(tfiles.name)
		assert.Nil(t, err)
		f.Close()
	}

}
