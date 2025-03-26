package bmc

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockNMISender struct {
	err error
}

func (m *mockNMISender) SendNMI(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return m.err
	}
}

func (m *mockNMISender) Name() string {
	return "mock"
}

func TestSendNMIFromInterface(t *testing.T) {
	testCases := []struct {
		name             string
		mockSenders      []interface{}
		errMsg           string
		isTimedout       bool
		expectedMetadata Metadata
	}{
		{
			name:        "success",
			mockSenders: []interface{}{&mockNMISender{}},
			expectedMetadata: Metadata{
				SuccessfulProvider:   "mock",
				ProvidersAttempted:   []string{"mock"},
				FailedProviderDetail: make(map[string]string),
			},
		},
		{
			name: "success with multiple senders",
			mockSenders: []interface{}{
				nil,
				"foo",
				&mockNMISender{err: errors.New("err from sender")},
				&mockNMISender{},
			},
			expectedMetadata: Metadata{
				SuccessfulProvider:   "mock",
				ProvidersAttempted:   []string{"mock", "mock"},
				FailedProviderDetail: map[string]string{"mock": "err from sender"},
			},
		},
		{
			name:        "not an nmisender",
			mockSenders: []interface{}{nil},
			errMsg:      "not an NMISender",
			expectedMetadata: Metadata{
				FailedProviderDetail: make(map[string]string),
			},
		},
		{
			name:        "no nmisenders",
			mockSenders: []interface{}{},
			errMsg:      "no NMISender implementations found",
			expectedMetadata: Metadata{
				FailedProviderDetail: make(map[string]string),
			},
		},
		{
			name:        "timed out",
			mockSenders: []interface{}{&mockNMISender{}},
			isTimedout:  true,
			errMsg:      "context deadline exceeded",
			expectedMetadata: Metadata{
				ProvidersAttempted:   []string{"mock"},
				FailedProviderDetail: map[string]string{"mock": "context deadline exceeded"},
			},
		},
		{
			name:        "error from nmisender",
			mockSenders: []interface{}{&mockNMISender{err: errors.New("foobar")}},
			errMsg:      "foobar",
			expectedMetadata: Metadata{
				ProvidersAttempted:   []string{"mock"},
				FailedProviderDetail: map[string]string{"mock": "foobar"},
			},
		},
		{
			name:        "error when fail to send",
			mockSenders: []interface{}{&mockNMISender{err: errors.New("err from sender")}},
			errMsg:      "failed to send NMI",
			expectedMetadata: Metadata{
				ProvidersAttempted:   []string{"mock"},
				FailedProviderDetail: map[string]string{"mock": "err from sender"},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			timeout := time.Second * 60
			if tt.isTimedout {
				timeout = 0
			}

			metadata, err := SendNMIFromInterface(context.Background(), timeout, tt.mockSenders)

			if tt.errMsg == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.errMsg)
			}

			assert.Equal(t, tt.expectedMetadata, metadata)
		})
	}
}
