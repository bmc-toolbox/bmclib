package idrac9

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/bmc-toolbox/bmclib/providers"
	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"
	"github.com/pkg/errors"
)

// Conn for dell idrac9 connections
type Conn struct {
	Host      string
	Port      string
	User      string
	Pass      string
	Log       logr.Logger
	conn      *http.Client
	xsrfToken string
}

const (
	ProviderName     = "idrac9"
	ProviderProtocol = "webgui"
)

// Features implemented by Dell's idrac9 provider:
var Features = registrar.Features{
	providers.FeatureUserCreate,
	providers.FeatureUserUpdate,
	providers.FeatureUserRead,
	providers.FeatureUserDelete,
}

func (c *Conn) Name() string {
	return ProviderName
}

func (c *Conn) Open(ctx context.Context) error {
	idrac := &IDrac9{ip: c.Host, username: c.User, password: c.Pass, log: c.Log}
	err := idrac.httpLogin()
	if err != nil {
		return err
	}
	if idrac.httpClient == nil {
		return errors.New("error opening connection")
	}
	c.conn = idrac.httpClient
	c.xsrfToken = idrac.xsrfToken
	return nil
}

func (c *Conn) Close(ctx context.Context) error {
	idrac := &IDrac9{ip: c.Host, username: c.User, password: c.Pass, log: c.Log}
	idrac.xsrfToken = c.xsrfToken
	idrac.httpClient = c.conn
	if idrac.httpClient != nil {
		if _, _, err := idrac.delete("sysmgmt/2015/bmc/session"); err != nil {
			return err
		}
	}
	return nil
}

// Compatible tests whether a BMC is compatible with the idrac9 provider
func (c *Conn) Compatible(ctx context.Context) bool {
	url := fmt.Sprintf("https://%s/sysmgmt/2015/bmc/info", c.Host)
	log := c.Log.WithValues("url", url)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.V(0).Error(err, "error creating http request")
		return false
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.V(0).Error(err, "error making http request")
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.V(0).Error(errors.New("error checking compatibility"), "status code not 200", "status_code", resp.StatusCode)
		return false
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.V(0).Error(err, "error reading response body")
		return false
	}
	models := []string{"PowerEdge M640", "PowerEdge R640", "PowerEdge R6415", "PowerEdge R6515", "PowerEdge R740xd"}
	for _, subStr := range models {
		if bytes.Contains(body, []byte(subStr)) {
			return true
		}
	}
	log.V(0).Error(errors.New("error checking compatibility"), "url did not contain expected string", "expected_strings", models)
	return false
}

func (c *Conn) UserCreate(ctx context.Context, user, pass, role string) (ok bool, err error) {
	idrac := &IDrac9{ip: c.Host, username: c.User, password: c.Pass, log: c.Log}
	idrac.xsrfToken = c.xsrfToken
	idrac.httpClient = c.conn
	// check if user already exists and capture any open user slots
	users, err := idrac.queryUsers()
	if err != nil {
		return false, errors.Wrap(err, "Unable to query existing users.")
	}
	var availableID int
	for id, usr := range users {
		if usr.UserName == user {
			return false, errors.New("user already exists")
		}
		if id != 1 && usr.UserName == "" {
			availableID = id
		}
	}

	// check if there's an open user slot available
	if availableID == 0 {
		return false, errors.New("all user account slots are in use, remove an account before adding a new one")
	}

	var userToCreate User
	userToCreate.Enable = "Enabled"
	userToCreate.SolEnable = "Enabled"
	userToCreate.UserName = user
	userToCreate.Password = pass
	// configure the user with a role
	if role == "admin" {
		userToCreate.Privilege = "511"
		userToCreate.IpmiLanPrivilege = "Administrator"
	} else {
		userToCreate.Privilege = "499"
		userToCreate.IpmiLanPrivilege = "Operator"
	}

	// create the user
	err = idrac.putUser(availableID, userToCreate)
	if err != nil {
		return false, errors.Wrap(err, "error creating user")
	}

	return true, nil
}

func (c *Conn) UserUpdate(ctx context.Context, user, pass, role string) (ok bool, err error) {
	idrac := &IDrac9{ip: c.Host, username: c.User, password: c.Pass, log: c.Log}
	idrac.xsrfToken = c.xsrfToken
	idrac.httpClient = c.conn
	// check if user exists and capture its ID
	users, err := idrac.queryUsers()
	if err != nil {
		return false, errors.Wrap(err, "Unable to query existing users.")
	}
	var id int
	for idx, usr := range users {
		if usr.UserName == user {
			id = idx
		}
	}
	if id == 0 {
		return false, errors.New("user not found")
	}

	// create the user payload
	var userPayload User
	userPayload.Enable = "Enabled"
	userPayload.SolEnable = "Enabled"
	userPayload.UserName = user
	userPayload.Password = pass
	// configure the user with a role
	if role == "admin" {
		userPayload.Privilege = "511"
		userPayload.IpmiLanPrivilege = "Administrator"
	} else {
		userPayload.Privilege = "499"
		userPayload.IpmiLanPrivilege = "Operator"
	}

	// create the user
	err = idrac.putUser(id, userPayload)
	if err != nil {
		return false, errors.Wrap(err, "error updating user")
	}
	return true, nil
}

func (c *Conn) UserDelete(ctx context.Context, user string) (ok bool, err error) {
	idrac := &IDrac9{ip: c.Host, username: c.User, password: c.Pass, log: c.Log}
	idrac.xsrfToken = c.xsrfToken
	idrac.httpClient = c.conn
	// get the user ID from a name
	users, err := idrac.queryUsers()
	if err != nil {
		return false, errors.Wrap(err, "Unable to query existing users.")
	}
	var userID int
	for id, usr := range users {
		if usr.UserName == user {
			userID = id
			break
		}
	}
	if userID == 0 {
		return true, nil
	}

	// delete the user
	endpoint := fmt.Sprintf("sysmgmt/2017/server/user?userid=%d", userID)
	statusCode, response, err := idrac.delete(endpoint)
	if err != nil {
		return false, errors.Wrap(err, string(response))
	}
	if statusCode < 200 || statusCode > 299 {
		return false, fmt.Errorf("error deleting user: %v", string(response))
	}

	return true, nil
}

func (c *Conn) UserRead(ctx context.Context) (users []map[string]string, err error) {
	idrac := &IDrac9{ip: c.Host, username: c.User, password: c.Pass, log: c.Log}
	idrac.xsrfToken = c.xsrfToken
	idrac.httpClient = c.conn
	// get the user ID from a name
	existingUsers, err := idrac.queryUsers()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to query existing users.")
	}
	for id, usr := range existingUsers {
		if usr.UserName == "" {
			continue
		}
		var temp map[string]string
		userJson, err := json.Marshal(usr)
		if err != nil {
			return nil, errors.Wrap(err, "error reading users")
		}
		err = json.Unmarshal(userJson, &temp)
		if err != nil {
			return nil, errors.Wrap(err, "error reading users")
		}

		temp["ID"] = strconv.Itoa(id)
		users = append(users, temp)
	}
	return users, nil
}
