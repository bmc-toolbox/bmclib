package serverdb

import (
	"fmt"
	"net/http"
)

type ServerProperty struct {
	ShouldDeprovision bool `json:"should_deprovision,omitempty"`
}

func (c *Client) SetServerProperty(id int, propertyPath string, value ServerProperty) (*http.Response, error) {
	path := fmt.Sprintf("server/%d/property/%s/", id, propertyPath)
	serverdbError := new(ServerdbError)
	resp, err := c.sling.New().Base(baseURLV2).Post(path).BodyJSON(value).Receive(nil, serverdbError)
	if err == nil && serverdbError.Err != "" {
		err = serverdbError
	}
	return resp, err
}

func (c *Client) GetServerProperty(id int, propertyPath string) (*ServerProperty, *http.Response, error) {
	path := fmt.Sprintf("server/%d/property/%s/", id, propertyPath)
	serverdbError := new(ServerdbError)
	value := new(ServerProperty)
	resp, err := c.sling.New().Base(baseURLV2).Get(path).Receive(value, serverdbError)
	if err == nil && serverdbError.Err != "" {
		err = serverdbError
	}
	return value, resp, err
}

func (c *Client) GetServerPropertyByServerName(name string, path string) (*ServerProperty, *http.Response, error) {
	server, rsp, err := c.GetServerByName(name)
	if err != nil {
		return nil, rsp, err
	}
	id := server.ID
	return c.GetServerProperty(id, path)
}
