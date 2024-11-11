package bmc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
)

type NMISender interface {
	SendNMI(ctx context.Context) error
}

func sendNMI(ctx context.Context, timeout time.Duration, sender NMISender, metadata *Metadata) error {
	senderName := getProviderName(sender)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	metadata.ProvidersAttempted = append(metadata.ProvidersAttempted, senderName)

	err := sender.SendNMI(ctx)
	if err != nil {
		metadata.FailedProviderDetail[senderName] = err.Error()
		return err
	}

	metadata.SuccessfulProvider = senderName

	return nil
}

// SendNMIFromInterface will look for providers that implement NMISender
// and attempt to call SendNMI until a provider is successful,
// or all providers have been exhausted.
func SendNMIFromInterface(
	ctx context.Context,
	timeout time.Duration,
	providers []interface{},
) (metadata Metadata, err error) {
	metadata = newMetadata()

	for _, provider := range providers {
		sender, ok := provider.(NMISender)
		if !ok {
			err = multierror.Append(err, fmt.Errorf("not an NMISender implementation: %T", provider))
			continue
		}

		sendNMIErr := sendNMI(ctx, timeout, sender, &metadata)
		if sendNMIErr != nil {
			err = multierror.Append(err, sendNMIErr)
			continue
		}
		return metadata, nil
	}

	if len(metadata.ProvidersAttempted) == 0 {
		err = multierror.Append(err, errors.New("no NMISender implementations found"))
	} else {
		err = multierror.Append(err, errors.New("failed to send NMI"))
	}

	return metadata, err
}
