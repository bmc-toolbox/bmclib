package serverdb

import (
	"fmt"
	"net/http"
)

type Site struct {
	APIObject
	ShortName string `json:"shortname"`
	Type      string `json:"type,omitempty"`
}

type ListSites struct {
	APIListObject
	Sites []Site `json:"objects",omitempty`
}

type ListSitesOptions struct {
	ShortName []string `url:"shortname__in",omitempty`
}

func (c *Client) GetSiteByPath(path string) (Site, *http.Response, error) {
	serverdbError := new(ServerdbError)
	site := new(Site)
	resp, err := c.sling.New().Base(baseURLForPath).Get(path).Receive(site, serverdbError)
	if err == nil && serverdbError.Err != "" {
		err = serverdbError
	}
	return *site, resp, err
}

func (c *Client) GetSitesByName(name []string) ([]Site, *http.Response, error) {
	serverdbError := new(ServerdbError)
	sites := new(ListSites)
	query := ListSitesOptions{
		ShortName: name,
	}
	resp, err := c.sling.New().Get("site/").QueryStruct(query).Receive(sites, serverdbError)
	if err == nil && serverdbError.Err != "" {
		err = serverdbError
	}
	return sites.Sites, resp, err
}

func (c *Client) GetSiteByShortName(name string) (Site, *http.Response, error) {
	names := []string{name}
	sites, resp, err := c.GetSitesByName(names)
	if err != nil {
		return Site{}, nil, err
	}
	if len(sites) != 1 {
		return Site{}, nil, fmt.Errorf("We received %d sites while searching for [%s]", len(sites), name)
	}
	return sites[0], resp, err
}
