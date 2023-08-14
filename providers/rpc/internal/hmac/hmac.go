package hmac

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
)

// Conf is the hmac configuration for signing data.
type Conf struct {
	// Hashes is a map of algorithms to a slice of hash.Hash (these are the hashed secrets). The slice is used to support multiple secrets.
	Hashes hashes
	// PrefixSig determines whether the algorithm will be prefixed to the signature. Example: sha256=abc123
	PrefixSig bool
}

type hashes map[Algorithm][]hash.Hash

type Algorithm string

const (
	// SHA256 is the SHA256 algorithm.
	SHA256 Algorithm = "sha256"
	// SHA256Short is the short version of the SHA256 algorithm.
	SHA256Short Algorithm = "256"
	// SHA512 is the SHA512 algorithm.
	SHA512 Algorithm = "sha512"
	// SHA512Short is the short version of the SHA512 algorithm.
	SHA512Short Algorithm = "512"
)

// NewHMAC returns a new HMAC.
func NewHMAC() *Conf {
	h := &Conf{
		Hashes:    hashes{},
		PrefixSig: true,
	}

	return h
}

// Sign takes the given data and signs it with the HMAC from h.
func (h *Conf) Sign(data []byte) (map[Algorithm][]string, error) {
	sigs := map[Algorithm][]string{}
	for algo, hshs := range h.Hashes {
		for _, hsh := range hshs {
			if _, err := hsh.Write(data); err != nil {
				return nil, err
			}
			sig := hex.EncodeToString(hsh.Sum(nil))
			if h.PrefixSig {
				sig = fmt.Sprintf("%s=%s", algo, sig)
			}
			sigs[algo] = append(sigs[algo], sig)
			// reset so Sign can be called multiple times. Otherwise, the next call will append to the previous one.
			hsh.Reset()
		}
	}

	return sigs, nil
}

func (h *Conf) AddSecretSHA256(secrets ...string) {
	h.Hashes = mergeHashes(h.Hashes, newSHA256(secrets...))
}

func (h *Conf) AddSecretSHA512(secrets ...string) {
	h.Hashes = mergeHashes(h.Hashes, newSHA512(secrets...))
}

// newSHA256 returns a map of SHA256 HMACs from the given secrets.
func newSHA256(secret ...string) hashes {
	var hsh []hash.Hash
	for _, s := range secret {
		hsh = append(hsh, hmac.New(sha256.New, []byte(s)))
	}
	return hashes{SHA256: hsh}
}

// newSHA512 returns a map of SHA512 HMACs from the given secrets.
func newSHA512(secret ...string) hashes {
	var hsh []hash.Hash
	for _, s := range secret {
		hsh = append(hsh, hmac.New(sha512.New, []byte(s)))
	}
	return hashes{SHA512: hsh}
}

func mergeHashes(hs ...hashes) hashes {
	m := hashes{}
	for _, h := range hs {
		for k, v := range h {
			m[k] = append(m[k], v...)
		}
	}
	return m
}
