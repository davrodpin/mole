package tunnel

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"

	"github.com/awnumar/memguard"
	"golang.org/x/crypto/ssh"
)

// PemKeyParser translates pem keys to a signature signer.
type PemKeyParser interface {
	// Parse returns a key signer to create signatures that verify against a
	// public key.
	Parse() (*ssh.Signer, error)
}

// PemKey holds data related to PEM keys
type PemKey struct {
	// Data holds the data for a PEM private key
	Data []byte

	// passphrase used to parse a PEM encoded private key
	passphrase *memguard.LockedBuffer
}

func NewPemKey(keyPath, passphrase string) (*PemKey, error) {
	data, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	k := &PemKey{Data: data}

	if passphrase != "" {
		k.updatePassphrase([]byte(passphrase))
	}

	return k, nil
}

// IsEncrypted inspects the key data block to tell if it is whether encrypted
// or not.
func (k PemKey) IsEncrypted() (bool, error) {
	p, err := decodePemKey(k.Data)
	if err != nil {
		return false, err
	}

	return x509.IsEncryptedPEMBlock(p), nil //nolint
}

// Parse translates a pem key to a signer to create signatures that verify
// against a public key.
func (k *PemKey) Parse() (ssh.Signer, error) {
	var signer ssh.Signer

	enc, err := k.IsEncrypted()
	if err != nil {
		return nil, err
	}

	if enc {
		if k.passphrase == nil {
			return nil, fmt.Errorf("can't read protected ssh key because no passphrase was provided")
		}

		signer, err = ssh.ParsePrivateKeyWithPassphrase(k.Data, k.passphrase.Bytes())
		if err != nil {
			return nil, err
		}
	} else {
		signer, err = ssh.ParsePrivateKey(k.Data)
		if err != nil {
			return nil, err
		}
	}

	return signer, nil
}

// HandlePassphrase securely records a passphrase given by a callback to the
// memory.
func (k *PemKey) HandlePassphrase(handler func() ([]byte, error)) error {
	enc, err := k.IsEncrypted()
	if err != nil {
		return fmt.Errorf("error while reading ssh key: %v", err)
	}

	if !enc {
		return nil
	}

	pp, err := handler()
	if err != nil {
		return fmt.Errorf("error while reading password: %v", err)
	}

	k.updatePassphrase(pp)

	return nil
}

func (k *PemKey) updatePassphrase(pp []byte) {
	if k.passphrase != nil {
		k.passphrase.Destroy()
	}

	if len(pp) < 1 {
		k.passphrase = nil
		return
	}

	k.passphrase = memguard.NewBufferFromBytes(pp)
}

func decodePemKey(data []byte) (*pem.Block, error) {
	p, r := pem.Decode(data)

	if p == nil && len(r) > 0 {
		return nil, fmt.Errorf("error while parsing key: no PEM data found")
	}

	if len(r) != 0 {
		return nil, fmt.Errorf("extra data in encoded key")
	}

	return p, nil
}
