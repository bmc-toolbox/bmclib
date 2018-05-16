package c7000

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// wraps the XML to be sent in the SOAP envelope
func wrapXML(element interface{}, sessionKey string) (doc Envelope) {

	body := Body{Content: element}
	doc = Envelope{
		SOAPENV: "http://www.w3.org/2003/05/soap-envelope",
		Xsi:     "http://www.w3.org/2001/XMLSchema-instance",
		Xsd:     "http://www.w3.org/2001/XMLSchema",
		Wsu:     "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd",
		Wsse:    "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd",
		Hpoa:    "hpoa.xsd",
		Body:    body,
	}

	if sessionKey != "" {

		doc.Header = Header{Security: Security{
			MustUnderstand: "true",
			HpOaSessionKeyToken: HpOaSessionKeyToken{
				OaSessionKey: OaSessionKey{Text: sessionKey},
			},
		},
		}
	}

	return doc
}

func (c *C7000) postXML(data []byte) (statusCode int, body []byte, err error) {

	//THIS BIT IS SUPER IMPORTANT - dont use the hostname for HP OAs!
	//Joel spent 3-4 hours debugging why, requests would fail intermittently becaues it was connecting
	//to the standby ! :<

	//ip := "10.193.251.22"
	//fmt.Println("IP:", c.ip)
	u, err := url.Parse(fmt.Sprintf("https://%s/hpoa", c.ip))
	if err != nil {
		return 0, []byte{}, err
	}

	req, err := http.NewRequest("POST", u.String(), bytes.NewReader(data))
	if err != nil {
		return 0, []byte{}, err
	}
	//	req.Header.Add("Content-Type", "application/soap+xml; charset=utf-8")
	req.Header.Add("Content-Type", "text/plain;charset=UTF-8")

	//fmt.Printf("---\n%s\n\n---\n", bytes.NewReader(data))
	//XXX to debug
	//fmt.Printf("--> %s\n", bytes.NewReader(data))
	//return 0, []byte{}, err
	//return err
	resp, err := c.client.Do(req)
	if err != nil {
		return resp.StatusCode, []byte{}, err
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, []byte{}, err
	}

	//fmt.Printf("%+v\n", body)
	return resp.StatusCode, body, err
}
