package redfishwrapper

import (
	"context"
	"encoding/json"

	bmclibErrs "github.com/metal-toolbox/bmclib/errors"
	"github.com/pkg/errors"
	"github.com/stmcginnis/gofish/redfish"
)

// ClearSystemEventLog clears all of the LogServices logs
func (c *Client) ClearSystemEventLog(ctx context.Context) (err error) {
	if err := c.SessionActive(); err != nil {
		return errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	chassis, err := c.client.Service.Chassis()
	if err != nil {
		return err
	}

	for _, c := range chassis {
		logServices, err := c.LogServices()
		if err != nil {
			return err
		}

		for _, logService := range logServices {
			err = logService.ClearLog()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetSystemEventLog returns the SystemEventLogEntries
func (c *Client) GetSystemEventLog(ctx context.Context) (entries [][]string, err error) {
	if err := c.SessionActive(); err != nil {
		return nil, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	managers, err := c.client.Service.Managers()
	if err != nil {
		return nil, err
	}

	for _, m := range managers {
		logServices, err := m.LogServices()
		if err != nil {
			return nil, err
		}

		for _, logService := range logServices {
			lentries, err := logService.Entries()
			if err != nil {
				return nil, err
			}

			for _, entry := range lentries {
				entries = append(entries, []string{
					entry.ID,
					entry.Created,
					entry.Description,
					entry.Message,
				})
			}
		}
	}

	return entries, nil
}

// GetSystemEventLogRaw returns the raw SEL
func (c *Client) GetSystemEventLogRaw(ctx context.Context) (eventlog string, err error) {
	var allEntries []*redfish.LogEntry

	if err := c.SessionActive(); err != nil {
		return "", errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	managers, err := c.client.Service.Managers()
	if err != nil {
		return "", err
	}

	for _, m := range managers {
		logServices, err := m.LogServices()
		if err != nil {
			return "", err
		}

		for _, logService := range logServices {
			lentries, err := logService.Entries()
			if err != nil {
				return "", err
			}

			allEntries = append(allEntries, lentries...)
		}
	}

	rawEntries, err := json.Marshal(allEntries)
	if err != nil {
		return "", err
	}

	return string(rawEntries), nil
}
