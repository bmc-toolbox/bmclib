package redfish

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Write tests for GetSystemEventLog
func Test_GetSystemEventLog(t *testing.T) {
	entries, err := mockClient.GetSystemEventLog(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, entries)
	assert.Equal(t, 2, len(entries))
}

// Write tests for GetSystemEventLogRaw
func Test_GetSystemEventLogRaw(t *testing.T) {
	eventlog, err := mockClient.GetSystemEventLogRaw(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, eventlog)
}
