package tunnel

import (
	"io/ioutil"
	"testing"
)

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
			if test.passphrase != key.passphrase.String() {
				t.Errorf("passphrases don't match for key %s: expected: %s, result: %s", test.keyPath, test.passphrase, key.passphrase.String())
			}
		} else {
			if nil != key.passphrase {
				t.Errorf("passphared is suppoed to be nil for %s", test.keyPath)
			}
		}
	}
}

func TestUpdatePassphrase(t *testing.T) {
	key, _ := NewPemKey("testdata/dotssh/id_rsa_encrypted", "mole")

	key.updatePassphrase([]byte("hello"))
	if !key.passphrase.EqualTo([]byte("hello")) {
		t.Error("update failed")
	}

	key = new(PemKey) // nil
	key.updatePassphrase([]byte("bye"))
	if !key.passphrase.EqualTo([]byte("bye")) {
		t.Error("update failed")
	}

	key.updatePassphrase([]byte(""))
	if key.passphrase != nil {
		t.Error("expected nil passphrase")
	}
}
