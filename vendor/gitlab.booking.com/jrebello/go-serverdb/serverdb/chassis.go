package serverdb

import (
	"fmt"
	"net/http"
)

type ChassisStatus struct {
	Status string `json:"status"`
}

type ListChassis struct {
	Meta    ListChassisMeta `json:"meta"`
	Objects []Chassis       `json:"objects"`
}

type Chassis struct {
	Serial string `json:"serialnumber"`
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type ListChassisMeta struct {
	Limit      int    `json:"limit"`
	Next       string `json:"next"`
	Offset     int    `json:"offset"`
	Previous   string `json:"previous"`
	TotalCount int    `json:"total_count"`
}

type ListChassisOptions struct {
	Status  string `url:"status,omitempty"`           //needs-setup etc
	Serial  string `url:"serialnumber,omitempty"`     // chassis/?format=json&serialnumber=51F3dk2
	Serials string `url:"serialnumber__in,omitempty"` // chassis/?format=json&serialnumber__in=51F3dk2,CZ3225ML1V
	Owner   int    `url:"owner,omitempty"`            //1 == booking
	Limit   int    `url:"limit,omitempty"`
	Offset  int    `url:"offset,omitempty"`
	Format  string `url:"format,omitempty"` //format=json required.
}

func (c *Client) ListChassis(listOptions ListChassisOptions) (ListChassis, *http.Response, error) {

	serverdbError := new(ServerdbError)
	chassisList := ListChassis{}

	resp, err := c.sling.New().Get("chassis/").QueryStruct(listOptions).Receive(&chassisList, serverdbError)
	if err == nil && serverdbError.Err != "" {
		err = serverdbError
	}

	return chassisList, resp, err
}

//sets a chassis status in serverdb
func (c *Client) UpdateChassisStatus(serverId int, status string) (*http.Response, error) {

	serverdbError := new(ServerdbError)
	body := struct {
		Status string `json:"status",omitempty`
	}{
		status,
	}

	endpoint := fmt.Sprintf("chassis/%d/?format=json", serverId)
	resp, err := c.sling.New().Patch(endpoint).BodyJSON(body).Receive(nil, serverdbError)
	if err == nil && serverdbError.Err != "" {
		err = serverdbError
	}

	return resp, err

}
