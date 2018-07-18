package serverdb

import (
	"fmt"
	"log"
	"net/http"
)

type Cluster struct {
	APIObject
	Description   string   `json:"description,omitempty"`
	KerberosRealm string   `json:"kerberos_realm"`
	Server        []string `json:"server,omitempty"`
}

type ListClusters struct {
	APIListObject
	Objects []Cluster
}

// List ServerClusters
func (c *Client) ListClusters() (ListClusters, *http.Response, error) {
	serverdbError := new(ServerdbError)
	clusters := new(ListClusters)
	resp, err := c.sling.New().Get("servercluster/").Receive(clusters, serverdbError)
	if err == nil && serverdbError.Err != "" {
		err = serverdbError
	}
	return *clusters, resp, err
}

// Remove a ServerCluster
func (c *Client) DeleteCluster(id int) (*http.Response, error) {
	serverdbError := new(ServerdbError)
	path := fmt.Sprintf("servercluster/%d/", id)
	resp, err := c.sling.New().Delete(path).Receive(serverdbError, serverdbError)
	if err == nil && serverdbError.Err != "" {
		err = serverdbError

	}
	return resp, err
}

// Update a ServerCluster
func (c *Client) UpdateCluster(id int, description string) (*http.Response, error) {
	body := struct {
		Description string `json:"description",omitempty`
	}{
		description,
	}
	serverdbError := new(ServerdbError)
	path := fmt.Sprintf("servercluster/%d/", id)
	resp, err := c.sling.New().Patch(path).BodyJSON(body).Receive(nil, serverdbError)
	if err == nil && serverdbError.Err != "" {
		err = serverdbError
	}
	return resp, nil
}

//Read a cluster by Location header
func (c *Client) GetClusterByLocation(location string) (Cluster, *http.Response, error) {
	serverdbError := new(ServerdbError)
	cluster := new(Cluster)
	resp, err := c.sling.New().Base("").Get(location).Receive(cluster, serverdbError)
	if err == nil && serverdbError.Err != "" {
		err = serverdbError
	}
	return *cluster, resp, err
}

//Read a cluster by path
func (c *Client) GetClusterByPath(path string) (Cluster, *http.Response, error) {
	serverdbError := new(ServerdbError)
	cluster := new(Cluster)
	resp, err := c.sling.New().Base(baseURLForPath).Get(path).Receive(cluster, serverdbError)
	if err == nil && serverdbError.Err != "" {
		err = serverdbError
	}
	return *cluster, resp, err
}

//Read a cluster by id
func (c *Client) GetClusterById(id int) (Cluster, *http.Response, error) {
	serverdbError := new(ServerdbError)
	cluster := new(Cluster)
	path := fmt.Sprintf("servercluster/%d", id)
	resp, err := c.sling.New().Get(path).Receive(cluster, serverdbError)
	if err == nil && serverdbError.Err != "" {
		err = serverdbError
	}
	return *cluster, resp, err
}

//Create a new ServerCluster
func (c *Client) CreateCluster(name, description string) (Cluster, *http.Response, error) {
	body := struct {
		Name        string `json:"name"`
		Description string `json:"description",omitempty`
	}{
		name,
		description,
	}
	serverdbError := new(ServerdbError)
	rsp, err := c.sling.New().Post("servercluster/").BodyJSON(body).Receive(nil, serverdbError)
	if err == nil && serverdbError.Err != "" {
		err = serverdbError
		return Cluster{}, nil, err
	}
	header := rsp.Header
	location := header.Get("Location")
	if c.Debug() {
		log.Printf("Location: [%s]", location)
	}
	if location == "" {
		return Cluster{}, nil, fmt.Errorf("expecting a location header")
	}
	cluster, _, err := c.GetClusterByLocation(location)
	return cluster, rsp, err
}
