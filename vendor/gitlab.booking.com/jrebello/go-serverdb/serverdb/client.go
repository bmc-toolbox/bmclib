package serverdb

import (
	"bufio"
	"fmt"
	sling "github.com/dghubble/sling"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	baseURL        = baseURLForPath + "/api/v1/"
	baseURLV2      = baseURLForPath + "/api/v2/"
	baseURLForPath = "https://serverdb.booking.com"
	badminConfig   = "/etc/bookings/badmin-cli.cfg"
)

type Client struct {
	apiKey string
	user   string
	sling  *sling.Sling
	dryRun bool
	debug  bool
}

type APIObject struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	ResourceUri string `json:"resource_uri"`
}

type apiListMeta struct {
	Limit    int     `json:"limit"`
	Next     *string `json:"next"`
	Offset   int     `json:"offset"`
	Previous *string `json:"previous"`
	Total    int     `json:"total_count"`
}

type ActionResult struct {
	Result string `json:"result,omitempty"`
	Id     int    `json:"id,omitempty"`
}

type APIListObject struct {
	Meta apiListMeta `json:"meta"`
}

type DefaultQueryParams struct {
	format string `url:format`
}

type ServerdbError struct {
	Err string `json:"error_message,omitempty"`
}

type DryRunDoer struct{}

func New(user, apiKey string) *Client {
	timeout := 20 * time.Second
	httpClient := http.Client{
		Timeout: timeout,
	}
	queryParams := DefaultQueryParams{
		format: "json",
	}
	authorizationHeader := "ApiKey " + user + ":" + apiKey
	return &Client{
		apiKey: apiKey,
		user:   user,
		dryRun: false,
		debug:  false,
		sling:  sling.New().Base(baseURL).Set("Authorization", authorizationHeader).Client(&httpClient).QueryStruct(queryParams),
	}
}

func (c *Client) SetDebug(debug bool) {
	c.debug = debug
}

func (c *Client) Debug() bool {
	return c.debug
}

// Switch the client to DryRun mode
func (c *Client) DryRun() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dump, err := httputil.DumpRequest(r, true)
		if err != nil {
			fmt.Printf("Error dumping the request: [%v]", err)
			return
		}
		fmt.Printf("********************\n %s \n****************\n", dump)
	}))
	c.sling = c.sling.Base(ts.URL)
	c.dryRun = true
}

func (c *Client) DryRunEnabled() bool {
	return c.dryRun
}

func (c *Client) httpClient() *sling.Sling {
	return c.sling
}

func (e *ServerdbError) Error() string {
	return fmt.Sprintf("Serverdb error: [%v]", e.Err)
}

func (e *ServerdbError) UnmarshalJSON(data []byte) error {
	e.Err = string(data)
	return nil
}

func ParseBadminCredentials() (string, string, error) {
	var retUser, retKey string

	file, err := os.Open(badminConfig)
	if err != nil {
		return "", "", err
	}
	scanner := bufio.NewScanner(file)
	user := regexp.MustCompile(`(?m)^user\s*=(?P<user>.*)$`)
	apiKey := regexp.MustCompile(`(?m)^auth\s*=(?P<api_key>.*)$`)
	for scanner.Scan() {
		line := scanner.Text()
		found := user.FindAllStringSubmatch(line, 2)
		if found != nil {
			retUser = strings.TrimSpace(found[0][1])
		}
		found = apiKey.FindAllStringSubmatch(line, 2)
		if found != nil {
			retKey = strings.TrimSpace(found[0][1])
		}
	}
	if err := scanner.Err(); err != nil {
		return "", "", err
	}
	if retUser == "" || retKey == "" {
		return retUser, retKey, fmt.Errorf("Couldn't find the user or the key in config: [%s] [%s]", retUser, retKey)
	}
	return retUser, retKey, nil
}

// TODO: add VerifyCredentials()
