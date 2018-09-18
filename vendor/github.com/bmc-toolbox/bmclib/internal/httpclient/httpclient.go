package httpclient

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/net/publicsuffix"
)

// Build builds a client session with our default parameters
func Build() (client *http.Client, err error) {
	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
		Dial: (&net.Dialer{
			Timeout:   120 * time.Second,
			KeepAlive: 120 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   120 * time.Second,
		ResponseHeaderTimeout: 120 * time.Second,
	}

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return client, err
	}

	client = &http.Client{
		Timeout:   time.Second * 120,
		Transport: tr,
		Jar:       jar,
	}

	return client, err
}

// DumpInvalidPayload is here to help identify unknown or broken payload messages
func DumpInvalidPayload(endpoint string, host string, payload []byte) (err error) {
	// TODO(jumartinez): We need to also add the reference for this payload or it's useless
	if viper.GetBool("collector.dump_invalid_payloads") {
		log.WithFields(log.Fields{"operation": "dump invalid payload", "host": host}).Info("dump invalid payload")

		t := time.Now()
		timeStamp := t.Format("20060102150405")

		dumpPath := viper.GetString("collector.dump_invalid_payload_path")
		err = os.MkdirAll(path.Join(dumpPath, host), 0755)
		if err != nil {
			return err
		}

		file, err := os.OpenFile(path.Join(dumpPath, host, timeStamp), os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			log.WithFields(log.Fields{"operation": "dump invalid payload", "host": host, "error": err}).Error("dump invalid payload")
			return err
		}

		_, err = file.Write([]byte(fmt.Sprintf("%s\n", endpoint)))
		if err != nil {
			log.WithFields(log.Fields{"operation": "dump invalid payload", "host": host, "error": err}).Error("dump invalid payload")
			return err
		}

		_, err = file.Write(payload)
		if err != nil {
			log.WithFields(log.Fields{"operation": "dump invalid payload", "host": host, "error": err}).Error("dump invalid payload")
			return err
		}
		file.Sync()
		file.Close()
	}

	return err
}

// StandardizeProcessorName makes the processor name standard across vendors
func StandardizeProcessorName(name string) string {
	return strings.ToLower(strings.TrimSuffix(strings.TrimSpace(strings.Split(name, "@")[0]), " 0"))
}
