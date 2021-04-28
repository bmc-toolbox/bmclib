package redfish

import (
	"context"

	rf "github.com/stmcginnis/gofish/redfish"
)

// VirtualMediaOptions holds virtual media parameters
type VirtualMediaOptions struct {
	MediaType string // CD || USBStick
	MediaURI  string // http://
	Username  string
	Password  string
}

// InsertVirtualMedia inserts virtual media on the host via the BMC
func (r *Conn) InsertVirtualMedia(ctx context.Context, options *VirtualMediaOptions) error {
	managers, err := r.conn.Service.Managers()
	if err != nil {
		return err
	}

	for _, m := range managers {
		media, err := m.VirtualMedia()
		if err != nil {
			return err
		}

		for _, m := range media {
			if options.MediaType != string(rf.CDMediaType) {
				continue
			}
			err := m.InsertMedia(options.MediaURI, true, true)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// EjectVirtualMedia ejects any mounted virtual media
func (r *Conn) EjectVirtualMedia(ctx context.Context) error {
	managers, err := r.conn.Service.Managers()
	if err != nil {
		return err
	}

	for _, m := range managers {
		media, err := m.VirtualMedia()
		if err != nil {
			return err
		}

		for _, m := range media {
			if !stringInSlice(string(rf.CDMediaType), m.MediaTypes) {
				continue
			}
			err := m.EjectMedia()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func stringInSlice(s string, sl []rf.VirtualMediaType) bool {
	for _, e := range sl {
		if s == string(e) {
			return true
		}
	}

	return false
}
