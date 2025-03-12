package rpc

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"net/http"
	"strings"
)

type Hashes map[Algorithm][]hash.Hash

// createSignaturePayload a signature payload is created by appending header values to the request body.
// there is no delimiter between the body and the header values and all header values.
func createSignaturePayload(body []byte, h http.Header) []byte {
	// add headers to signature payload, no space between values.
	for _, val := range h {
		body = append(body, []byte(strings.Join(val, ""))...)
	}

	return body
}

// sign signs the data with all the given hashes and returns the signatures.
func sign(data []byte, h Hashes, prefixSigDisabled bool) (Signatures, error) {
	sigs := map[Algorithm][]string{}
	for algo, hshs := range h {
		for _, hsh := range hshs {
			if _, err := hsh.Write(data); err != nil {
				return nil, err
			}
			sig := hex.EncodeToString(hsh.Sum(nil))
			if !prefixSigDisabled {
				sig = fmt.Sprintf("%s=%s", algo, sig)
			}
			sigs[algo] = append(sigs[algo], sig)
			// reset so Sign can be called multiple times. Otherwise, the next call will append to the previous one.
			hsh.Reset()
		}
	}

	return sigs, nil
}

// ToShort returns the short version of an algorithm.
func (a Algorithm) ToShort() Algorithm {
	switch a {
	case SHA256:
		return SHA256Short
	case SHA512:
		return SHA512Short
	default:
		return a
	}
}

// NewSHA256 returns a map of SHA256 HMACs from the given secrets.
func NewSHA256(secret ...string) Hashes {
	var hsh []hash.Hash
	for _, s := range secret {
		hsh = append(hsh, hmac.New(sha256.New, []byte(s)))
	}
	return Hashes{SHA256: hsh}
}

// NewSHA512 returns a map of SHA512 HMACs from the given secrets.
func NewSHA512(secret ...string) Hashes {
	var hsh []hash.Hash
	for _, s := range secret {
		hsh = append(hsh, hmac.New(sha512.New, []byte(s)))
	}
	return Hashes{SHA512: hsh}
}

func mergeHashes(hs ...Hashes) Hashes {
	m := Hashes{}
	for _, h := range hs {
		for k, v := range h {
			m[k] = append(m[k], v...)
		}
	}
	return m
}

// CreateHashes creates a new hash for all secrets provided.
func CreateHashes(s Secrets) map[Algorithm][]hash.Hash {
	h := map[Algorithm][]hash.Hash{}
	for algo, secrets := range s {
		switch algo {
		case SHA256, SHA256Short:
			h = mergeHashes(h, NewSHA256(secrets...))
		case SHA512, SHA512Short:
			h = mergeHashes(h, NewSHA512(secrets...))
		}
	}

	return h
}
