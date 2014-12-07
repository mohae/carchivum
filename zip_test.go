package carchivum

import (
	_ "io/ioutil"
	_ "os"
)

/*
func TestZipBytes(t *testing.T) {
	tests := []struct {
		name          string
		value         string
		expectedBytes int
		expectedErr   string
	}{
		{"", "", 0, "z"},
		{"empty.zip", "", 0, "z"},
		{"basic.zip", "This is a test string to zip. It really isn't much, but it's enough to test", 0, ""},
	}

	for _, test := range tests {
		z, err := ZipBytes([]byte(test.value), test.name)
		if test.expectedErr == "" {
			assert.Nil(t, err)
		} else {
			assert.Equal(t, test.expectedErr, err.Error())
			continue
		}

		//zz, err := z.Bytes()
		assert.NotNil(t, err)
		assert.Equal(t, test.expectedBytes, len(z))
	}

}
*/
/*
func TestZipBytes2(t *testing.T) {
	b := []byte("Test ZipBytes using this string. Its short, but long enough to prove things work")

	zb, err := ZipBytes(b, "test")
	assert.Nil(t, err)
	assert.Equal(t, 8, len(zb))
	assert.Equal(t, "adfaf", string(zb))
}
*/
