package rpc

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
)

// hmacConf is the hmacConf configuration for signing data.
type hmacConf struct {
	// Hashes is a map of algorithms to a slice of hash.Hash (these are the hashed secrets). The slice is used to support multiple secrets.
	Hashes map[Algorithm][]hash.Hash
	// PrefixSig determines whether the algorithm will be prefixed to the signature. Example: sha256=abc123
	PrefixSig bool
}

// newHMAC returns a new HMAC.
func newHMAC() hmacConf {
	h := hmacConf{
		Hashes:    map[Algorithm][]hash.Hash{},
		PrefixSig: true,
	}

	return h
}

// Sign takes the given data and signs it with the HMAC from h.
func (h hmacConf) Sign(data []byte) (map[Algorithm][]string, error) {
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

// newSHA256 returns a map of SHA256 HMACs from the given secrets.
func newSHA256(secret ...string) map[Algorithm][]hash.Hash {
	var hsh []hash.Hash
	for _, s := range secret {
		hsh = append(hsh, hmac.New(sha256.New, []byte(s)))
	}
	return map[Algorithm][]hash.Hash{SHA256: hsh}
}

// newSHA512 returns a map of SHA512 HMACs from the given secrets.
func newSHA512(secret ...string) map[Algorithm][]hash.Hash {
	var hsh []hash.Hash
	for _, s := range secret {
		hsh = append(hsh, hmac.New(sha512.New, []byte(s)))
	}
	return map[Algorithm][]hash.Hash{SHA512: hsh}
}

func mergeHashes(hashes ...map[Algorithm][]hash.Hash) map[Algorithm][]hash.Hash {
	m := map[Algorithm][]hash.Hash{}
	for _, h := range hashes {
		for k, v := range h {
			m[k] = append(m[k], v...)
		}
	}
	return m
}
