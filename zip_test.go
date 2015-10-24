package carchivum

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/mohae/customjson"
)

var stringMarshaller = customjson.NewMarshalString()

func TestZipBytes(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		expectedLen int
		expected    []byte
		expectedErr string
	}{
		{"", "", 0, []byte("PK\x03\x04\x14\x00\b\x00\b\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x01\x00\x00\xff\xffPK\a\b\x00\x00\x00\x00\x05\x00\x00\x00\x00\x00\x00\x00PK\x01\x02\x14\x00\x14\x00\b\x00\b\x00\x00\x00\x00\x00\x00\x00\x00\x00\x05\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00PK\x05\x06\x00\x00\x00\x00\x01\x00\x01\x00.\x00\x00\x003\x00\x00\x00\x00\x00"), ""},
		{"empty.txt", "", 0, []byte("PK\x03\x04\x14\x00\b\x00\b\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\t\x00\x00\x00empty.txt\x01\x00\x00\xff\xffPK\a\b\x00\x00\x00\x00\x05\x00\x00\x00\x00\x00\x00\x00PK\x01\x02\x14\x00\x14\x00\b\x00\b\x00\x00\x00\x00\x00\x00\x00\x00\x00\x05\x00\x00\x00\x00\x00\x00\x00\t\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00empty.txtPK\x05\x06\x00\x00\x00\x00\x01\x00\x01\x007\x00\x00\x00<\x00\x00\x00\x00\x00"), ""},
		{"basic.txt", "This is a test string to zip. It really isn't much, but it's enough to test", 75, []byte("PK\x03\x04\x14\x00\b\x00\b\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\t\x00\x00\x00basic.txt\n\xc9\xc8,V\x00\xa2D\x85\x92\xd4\xe2\x12\x85⒢̼t\x85\x92|\x85\xaa\xcc\x02=\x05\xcf\x12\x85\xa2\xd4Ĝ\x9cJ\xa0\x92<\xf5\x12\x85\xdc\xd2\xe4\f\x1d\x85\xa4\xd2\x12\x85\xcc\x12\xf5b\x85Լ\xfc\xd2\xf4\f\x90j\x90f@\x00\x00\x00\xff\xffPK\a\b\xe4~\xd5wK\x00\x00\x00K\x00\x00\x00PK\x01\x02\x14\x00\x14\x00\b\x00\b\x00\x00\x00\x00\x00\xe4~\xd5wK\x00\x00\x00K\x00\x00\x00\t\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00basic.txtPK\x05\x06\x00\x00\x00\x00\x01\x00\x01\x007\x00\x00\x00\x82\x00\x00\x00\x00\x00"), ""},
	}

	for i, test := range tests {
		l, zb, err := ZipBytes([]byte(test.value), test.name)
		if test.expectedErr != "" {
			if err.Error() != test.expectedErr {
				t.Errorf("%d: expected error to be %q, got %q", i, test.expectedErr, err.Error())
				continue
			}
		}
		if err != nil {
			if err.Error() != test.expectedErr {
				t.Errorf("%d: expected error to be %q got %q", i, test.expectedErr, err)
				continue
			}
		}
		if test.expectedErr != "" {
			t.Errorf("%d: expected error to be %q, got nil", i, test.expectedErr)
			continue
		}
		if l != test.expectedLen {
			t.Errorf("%d: expected len to be %d, got %d", i, test.expectedLen, l)
			continue
		}
		if !bytes.Equal(test.expected, zb) {
			t.Errorf("%d: expected the compressed bytes did not equal the expected", i)
		}
	}
}

func TestZip(t *testing.T) {
	tmpDir, err := CreateTempFiles()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
		RemoveTmpDir(tmpDir)
		return
	}
	newZ := NewZip(filepath.Join(tmpDir, "test.zip"))
	// Test Create
	zDir := filepath.Join(tmpDir, "test")
	cnt, err := newZ.Create(zDir)
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
		RemoveTmpDir(tmpDir)
		return
	}
	if cnt != 5 {
		t.Errorf("Expected 5 got %d", cnt)
		RemoveTmpDir(tmpDir)
		return
	}
	// make sure it exists
	_, err = os.Stat(newZ.Car.Name)
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
		RemoveTmpDir(tmpDir)
		return
	}

	// Test Extract Zip
	eDir := filepath.Join(tmpDir, "extract")
	err = os.Mkdir(eDir, 0755)
	if err != nil {
		t.Errorf("expected error to be nil, got %q", err)
		RemoveTmpDir(tmpDir)
		return
	}
	newZ.OutDir = eDir
	err = newZ.Extract()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
		RemoveTmpDir(tmpDir)
		return
	}
	tests := []struct {
		fname    string
		expected []byte
	}{
		{"test/test1.txt", []byte("some content\n")},
		{"test/test2.txt", []byte("some more content\n")},
		{"test/dir/test1.txt", []byte("different content\n")},
		{"test/dir/test2.txt", []byte("might be different content\n")},
	}
	for i, test := range tests {
		fB, err := ioutil.ReadFile(filepath.Join(eDir, test.fname))
		if err != nil {
			t.Errorf("%d: expected error to be nil, got %q", i, err)
			continue
		}
		if !bytes.Equal(fB, test.expected) {
			t.Errorf("%d: expected %s, got %s", i, string(fB), string(test.expected))
			continue
		}
	}
	RemoveTmpDir(tmpDir)
}
