package carchivum

import (
	"archive/zip"
	"bytes"
	_ "io/ioutil"
	_ "os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/mohae/customjson"
)

var stringMarshaller = customjson.NewMarshalString()
/*
func TestZipBytes(t *testing.T) {
	tests := []struct{
		name string
		value string
		expectedBytes	int
		expectedErr	string
	}{
		{"", "", 0, "z"},
		{"empty.zip", "", 0, "z"},
		{"basic.zip", "This is a test string to zip. It really isn't much, but it's enough to test", 0, ""},
	}


	for _, test := range tests {
		z, err := ZipBytes(test.name, []byte(test.value))
		if test.expectedErr == "" {
			assert.Nil(t, err)
		} else {
			assert.Equal(t, test.expectedErr, err.Error())
			continue
		}
		
		zz, err := z.Byte()
		assert.NotNil(t, err)
		assert.Equal(t, test.expectedBytes, zz)
	}

}
*/
/*
func TestZipFiles(t *testing.T) {
	tests := []struct{
		value []string
		expected []byte
		expectedErr interface{}
		expectedBytes int
	} {
		{nil, nil, "no files to zip", 0},
		{[]string{"invalid.file"}, nil, "open invalid.file: no such file or directory", 0},
		{[]string{"test_files/tmbg-ana-ng.txt"}, []byte("testa"), nil, 100},
		{[]string{"test_files/tmbg-ana-ng.txt", "test_files/pink-floyd/meddle/echos.txt", "test_files/pink-floyd/meddle/san-tropez.txt"}, []byte("testa"), nil, 100},
//		{[]string{"test_files"}, nil, &os.PathError{Op:"read", Path:"test_files", Err: 0x15}, 0},	
	}

	for _, test := range tests {
		res, err := ZipFiles(test.value...)
		if test.expectedErr == nil {
			assert.Nil(t, err)
		} else {
			assert.Equal(t, test.expectedErr, err.Error())
			continue
		}

		assert.Equal(t, test.expectedBytes, len(res))
		assert.Equal(t, test.expected, res)
	}
}
*/
func TestZipBytes(t *testing.T) {
	b := []byte("Test ZipBytes using this string. Its short, but long enough to prove things work")
	
	zb, err := ZipBytes(b, "test")
	assert.Nil(t, err)
	assert.Equal(t, 8, len(zb))
	assert.Equal(t, "adfaf", string(zb))
}
/*
func TestAddFile(t *testing.T) {
	tests := []struct{
		value string
		expectedErr string
		expectedB int
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
