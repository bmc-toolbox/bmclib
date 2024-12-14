package bmc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// Opener interface for opening a connection to a BMC
type Opener interface {
	Open(ctx context.Context) error
}

// Closer interface for closing a connection to a BMC
type Closer interface {
	Close(ctx context.Context) error
}

// connectionProviders is an internal struct to correlate an implementation/provider and its name
type connectionProviders struct {
	name   string
	closer Closer
}

// OpenConnectionFromInterfaces will try all opener interfaces and remove failed ones.
// The reason failed ones need to be removed is so that when other methods are called (like powerstate)
// implementations that have connections wont nil pointer error when their connection fails.
func OpenConnectionFromInterfaces(ctx context.Context, timeout time.Duration, providers []interface{}) (opened []interface{}, metadata Metadata, err error) {
	metadata = newMetadata()

	// Return immediately if the context is done.
	select {
	case <-ctx.Done():
		return nil, metadata, multierror.Append(err, ctx.Err())
	default:
	}

	// Create a context with the specified timeout. This is done for backward compatibility but
	// we should consider removing the timeout parameter alltogether given the context will
	// container the timeout.
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// result facilitates communication of data between the concurrent opener goroutines and
	// the the parent goroutine.
	type result struct {
		ProviderName string
		Opener       Opener
		Err          error
	}

	// Create a channel to communicate results between opener goroutines and the parent goroutine.
	results := make(chan result)

	// Use a WaitGroup to control closing of the results channel when all opener goroutines finish.
	var wg sync.WaitGroup

	// For every provider, launch a goroutine that attempts to open a connection and report
	// back via the results channel what happened.
	for _, elem := range providers {
		if elem == nil {
			continue
		}
		switch p := elem.(type) {
		case Opener:
			providerName := getProviderName(elem)
			metadata.ProvidersAttempted = append(metadata.ProvidersAttempted, providerName)

			wg.Add(1)
			go func(provider Opener, providerName string) {
				defer wg.Done()
				res := result{ProviderName: providerName, Opener: provider}

				if err := provider.Open(ctx); err != nil {
					res.Err = errors.WithMessagef(err, "provider: %v", providerName)
				}

				results <- res
			}(p, providerName)

		default:
			err = multierror.Append(err, fmt.Errorf("not a Opener implementation: %T", p))
		}
	}

	// Launch the goroutine to close the results channel ensuring we can exit the for-range over
	// the results channel below.
	go func() { wg.Wait(); close(results) }()

	// Gather and handle results from the opener goroutines.
	for res := range results {
		if res.Err != nil {
			err = multierror.Append(err, res.Err)
			metadata.FailedProviderDetail[res.ProviderName] = res.Err.Error()
			continue
		}

		opened = append(opened, res.Opener)
		metadata.SuccessfulOpenConns = append(metadata.SuccessfulOpenConns, res.ProviderName)
	}

	if len(opened) == 0 {
		return nil, metadata, multierror.Append(err, errors.New("no Opener implementations found"))
	}

	return opened, metadata, nil
}

// closeConnection closes a connection to a BMC, trying all interface implementations passed in
func closeConnection(ctx context.Context, c []connectionProviders) (metadata Metadata, err error) {
	metadata = newMetadata()
	var connClosed bool

	for _, elem := range c {
		if elem.closer == nil {
			continue
		}
		metadata.ProvidersAttempted = append(metadata.ProvidersAttempted, elem.name)
		closeErr := elem.closer.Close(ctx)
		if closeErr != nil {
			err = multierror.Append(err, errors.WithMessagef(closeErr, "provider: %v", elem.name))
			metadata.FailedProviderDetail[elem.name] = closeErr.Error()
			continue
		}
		connClosed = true
		metadata.SuccessfulCloseConns = append(metadata.SuccessfulCloseConns, elem.name)
	}
	if connClosed {
		return metadata, nil
	}
	return metadata, multierror.Append(err, errors.New("failed to close connection"))
}

// CloseConnectionFromInterfaces identifies implementations of the Closer() interface and and passes the found implementations to the closeConnection() wrapper
func CloseConnectionFromInterfaces(ctx context.Context, generic []interface{}) (metadata Metadata, err error) {
	metadata = newMetadata()

	closers := make([]connectionProviders, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
		temp := connectionProviders{name: getProviderName(elem)}
		switch p := elem.(type) {
		case Closer:
			temp.closer = p
			closers = append(closers, temp)
		default:
			e := fmt.Sprintf("not a Closer implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(closers) == 0 {
		return metadata, multierror.Append(err, errors.New("no Closer implementations found"))
	}
	return closeConnection(ctx, closers)
}
