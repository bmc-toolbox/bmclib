package goipmi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClone(t *testing.T) {
	orig, err := New("user", "pass", "host", 623, WithCipherSuite("17"))
	require.NoError(t, err)

	clone, err := orig.Clone()
	require.NoError(t, err)

	// The clone must carry a fresh, independent client so operations on it
	// never touch the original's session.
	assert.NotSame(t, orig, clone)
	assert.NotSame(t, orig.client, clone.client)

	// All connection parameters must be preserved.
	assert.Equal(t, orig.Username, clone.Username)
	assert.Equal(t, orig.Password, clone.Password)
	assert.Equal(t, orig.Host, clone.Host)
	assert.Equal(t, orig.Port, clone.Port)
	assert.Equal(t, orig.cipherSuite, clone.cipherSuite)
}
