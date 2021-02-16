package discover

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/go-logr/logr"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/errors"
	"github.com/bmc-toolbox/bmclib/internal/httpclient"
	"github.com/bmc-toolbox/bmclib/providers/asrockrack"
)

type Probev2 struct {
	host     string
	username string
	password string
	client   *http.Client
	logger   logr.Logger
}

func NewProbev2(opts *Options) (*Probev2, error) {

	c, err := httpclient.Build()
	if err != nil {
		return nil, err
	}

	return &Probev2{
		client:   c,
		logger:   opts.Logger,
		host:     opts.Host,
		username: opts.Username,
		password: opts.Password,
	}, nil
}

func (p *Probev2) asRockRack(ctx context.Context) (interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://%s", p.host), nil)
	if err != nil {
		return nil, err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body) // nolint

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// ensure the response we got included a png
	if resp.StatusCode == 200 && bytes.Contains(payload, []byte("ASRockRack")) {
		p.logger.V(1).Info("step", "ScanAndConnect", "host", p.host, "vendor", string(devices.ASRockRack), "msg", "it's an ASRockRack")
		return asrockrack.New(ctx, p.host, p.username, p.password, p.logger)
	}

	return nil, errors.ErrDeviceNotMatched
}
