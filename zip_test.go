package carchivum

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/mohae/customjson"
	"github.com/stretchr/testify/assert"
)

var stringMarshaller = customjson.NewMarshalString()

func TestZipBytes(t *testing.T) {
	tests := []struct {
		name           string
		value          string
		expectedLen    int
		expectedZipped []byte
		expectedErr    string
	}{
		{"", "", 0, []byte("PK\x03\x04\x14\x00\b\x00\b\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x01\x00\x00\xff\xffPK\a\b\x00\x00\x00\x00\x05\x00\x00\x00\x00\x00\x00\x00PK\x01\x02\x14\x00\x14\x00\b\x00\b\x00\x00\x00\x00\x00\x00\x00\x00\x00\x05\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00PK\x05\x06\x00\x00\x00\x00\x01\x00\x01\x00.\x00\x00\x003\x00\x00\x00\x00\x00"), ""},
		{"empty.txt", "", 0, []byte("PK\x03\x04\x14\x00\b\x00\b\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\t\x00\x00\x00empty.txt\x01\x00\x00\xff\xffPK\a\b\x00\x00\x00\x00\x05\x00\x00\x00\x00\x00\x00\x00PK\x01\x02\x14\x00\x14\x00\b\x00\b\x00\x00\x00\x00\x00\x00\x00\x00\x00\x05\x00\x00\x00\x00\x00\x00\x00\t\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00empty.txtPK\x05\x06\x00\x00\x00\x00\x01\x00\x01\x007\x00\x00\x00<\x00\x00\x00\x00\x00"), ""},
		{"basic.txt", "This is a test string to zip. It really isn't much, but it's enough to test", 75, []byte("PK\x03\x04\x14\x00\b\x00\b\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\t\x00\x00\x00basic.txt\n\xc9\xc8,V\x00\xa2D\x85\x92\xd4\xe2\x12\x85⒢̼t\x85\x92|\x85\xaa\xcc\x02=\x05\xcf\x12\x85\xa2\xd4Ĝ\x9cJ\xa0\x92<\xf5\x12\x85\xdc\xd2\xe4\f\x1d\x85\xa4\xd2\x12\x85\xcc\x12\xf5b\x85Լ\xfc\xd2\xf4\f\x90j\x90f@\x00\x00\x00\xff\xffPK\a\b\xe4~\xd5wK\x00\x00\x00K\x00\x00\x00PK\x01\x02\x14\x00\x14\x00\b\x00\b\x00\x00\x00\x00\x00\xe4~\xd5wK\x00\x00\x00K\x00\x00\x00\t\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00basic.txtPK\x05\x06\x00\x00\x00\x00\x01\x00\x01\x007\x00\x00\x00\x82\x00\x00\x00\x00\x00"), ""},
	}

	for _, test := range tests {
		l, zb, err := ZipBytes([]byte(test.value), test.name)
		if test.expectedErr != "" {
			assert.Equal(t, test.expectedErr, err.Error())
			continue
		}
		assert.Nil(t, err)
		assert.Equal(t, test.expectedLen, l)
		assert.Equal(t, test.expectedZipped, zb)
	}

}

func TestZip(t *testing.T) {
	tmpDir, err := CreateTempFiles()
	assert.Nil(t, err)
	newZ := NewZip()
	newZ.Car.Name = filepath.Join(tmpDir, "test.zip")
	// Test Create
	zDir := filepath.Join(tmpDir, "test")
	cnt, err := newZ.Create(zDir)
	assert.Nil(t, err)
	assert.Equal(t, 5, cnt)
	// Check the created file
	fi, err := os.Stat(newZ.Car.Name)
	assert.Nil(t, err)
	assert.Equal(t, 594, fi.Size())

	// Test Extract Zip
	eDir := filepath.Join(tmpDir, "extract")
	err = os.Mkdir(eDir, 0755)
	assert.Nil(t, err)
	err = newZ.Extract(eDir, newZ.Car.Name)
	assert.Nil(t, err)
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
