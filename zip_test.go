package carchivum

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

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
		{"basic.txt", "This is a test string to zip. It really isn't much, but it's enough to test", 75, []byte("\x50\x4b\x03\x04\x14\x00\x08\x00\x08\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x09\x00\x00\x00\x62\x61\x73\x69\x63\x2e\x74\x78\x74\x14\xcb\xb1\x0d\x85\x20\x14\x86\xd1\x55\xbe\x8e\xe6\xe5\xed\x61\xef\x02\x68\x08\xdc\x04\xc1\x70\x7f\x0a\x9d\xde\xd0\x9f\xb3\x17\x73\xcc\x89\x28\xb9\x70\x0d\x6b\x19\x75\x5e\xbb\xff\x6c\x62\xa4\x58\xeb\x83\x79\x0b\xe2\x9a\x67\xf9\x71\x4c\x61\x0a\x4e\x6a\x7d\xe6\xb2\xf4\xca\x5f\x00\x00\x00\xff\xff\x50\x4b\x07\x08\xe4\x7e\xd5\x77\x4a\x00\x00\x00\x4b\x00\x00\x00\x50\x4b\x01\x02\x14\x00\x14\x00\x08\x00\x08\x00\x00\x00\x00\x00\xe4\x7e\xd5\x77\x4a\x00\x00\x00\x4b\x00\x00\x00\x09\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x62\x61\x73\x69\x63\x2e\x74\x78\x74\x50\x4b\x05\x06\x00\x00\x00\x00\x01\x00\x01\x00\x37\x00\x00\x00\x81\x00\x00\x00\x00\x00"), ""},
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
