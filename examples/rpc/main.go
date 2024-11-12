package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/metal-toolbox/bmclib"
	"github.com/metal-toolbox/bmclib/logging"
	"github.com/metal-toolbox/bmclib/providers/rpc"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Start the test consumer
	go testConsumer(ctx)
	time.Sleep(100 * time.Millisecond)

	log := logging.ZeroLogger("info")
	opts := []bmclib.Option{
		bmclib.WithLogger(log),
		bmclib.WithPerProviderTimeout(5 * time.Second),
		bmclib.WithRPCOpt(rpc.Provider{
			ConsumerURL: "http://localhost:8800",
			// Opts are not required.
			Opts: rpc.Opts{
				HMAC: rpc.HMACOpts{
					Secrets: rpc.Secrets{rpc.SHA256: {"superSecret1"}},
				},
				Signature: rpc.SignatureOpts{
					HeaderName:             "X-Bespoke-Signature",
					IncludedPayloadHeaders: []string{"X-Bespoke-Timestamp"},
				},
				Request: rpc.RequestOpts{
					TimestampHeader: "X-Bespoke-Timestamp",
				},
			},
		}),
	}
	host := "127.0.1.1"
	user := "admin"
	pass := "admin"
	c := bmclib.NewClient(host, user, pass, opts...)
	if err := c.Open(ctx); err != nil {
		panic(err)
	}
	defer c.Close(ctx)

	state, err := c.GetPowerState(ctx)
	if err != nil {
		panic(err)
	}
	log.Info("power state", "state", state)
	log.Info("metadata for GetPowerState", "metadata", c.GetMetadata())

	ok, err := c.SetPowerState(ctx, "on")
	if err != nil {
		panic(err)
	}
	log.Info("set power state", "ok", ok)
	log.Info("metadata for SetPowerState", "metadata", c.GetMetadata())

	<-ctx.Done()
}

func testConsumer(ctx context.Context) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		req := rpc.RequestPayload{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		rp := rpc.ResponsePayload{
			ID:   req.ID,
			Host: req.Host,
		}
		switch req.Method {
		case rpc.PowerGetMethod:
			rp.Result = "on"
		case rpc.PowerSetMethod:

		case rpc.BootDeviceMethod:

		case rpc.PingMethod:
			rp.Result = "pong"
		default:
			w.WriteHeader(http.StatusNotFound)
		}
		b, _ := json.Marshal(rp)
		w.Write(b)
	})

	return http.ListenAndServe(":8800", nil)
}
