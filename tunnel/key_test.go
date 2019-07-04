package tunnel

import (
	"io/ioutil"
	"testing"
)

func passwordHandler(password string) func() ([]byte, error) {
	return func() ([]byte, error) { return []byte(password), nil }
}

func TestPemKey(t *testing.T) {
	tests := []struct {
		keyPath    string
		encrypted  bool
		passphrase string
	}{
		{
			"testdata/dotssh/id_rsa",
			false,
			"",
		},
		{
			"testdata/dotssh/id_rsa_encrypted",
			true,
			"mole",
		},
	}

	for _, test := range tests {
		key, err := NewPemKey(test.keyPath, test.passphrase)
		if err != nil {
			t.Errorf("test failed for key %s: %v", test.keyPath, err)
		}

		enc, err := key.IsEncrypted()
		if err != nil {
			t.Errorf("test failed for key %s: %v", test.keyPath, err)
		}

		if test.encrypted != enc {
			t.Errorf("test for encryption check on %s failed : expected: %t, result: %t", test.keyPath, test.encrypted, enc)
		}

		_, err = key.Parse()
		if err != nil {
			t.Errorf("test failed for key %s: %v", test.keyPath, err)
		}
	}
}

func TestHandlePassword(t *testing.T) {
	tests := []struct {
		keyPath    string
		passphrase string
	}{
		{
			"testdata/dotssh/id_rsa",
			"",
		},
		{
			"testdata/dotssh/id_rsa_encrypted",
			"mole",
		},
	}

	for _, test := range tests {
		data, err := ioutil.ReadFile(test.keyPath)
		if err != nil {
			t.Errorf("can't read key file %s: %v", test.keyPath, err)
		}

		key := &PemKey{Data: data}

		key.HandlePassphrase(func() ([]byte, error) {
			return []byte(test.passphrase), nil
		})

		enc, err := key.IsEncrypted()
		if err != nil {
			t.Errorf("test failed for key %s: %v", test.keyPath, err)
		}

		if enc {
			if test.passphrase != string(key.passphrase.Bytes()) {
				t.Errorf("passphrases don't match for key %s: expected: %s, result: %s", test.keyPath, test.passphrase, string(key.passphrase.Bytes()))
			}
		} else {
			if nil != key.passphrase {
				t.Errorf("passphared is suppoed to be nil for %s", test.keyPath)
			}
		}
	}
}
