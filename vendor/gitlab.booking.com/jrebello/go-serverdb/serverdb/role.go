package serverdb

import (
	"fmt"
	"net/http"
)

type Role struct {
	APIObject
	Description      string   `json:"description,omitempty"`
	ManagedBy        string   `json:"managed_by,omitempty"`
	ManagedByForeman bool     `json:"managed_by_foreman"`
	PuppetService    string   `json:"puppet_service,omitempty"`
	RequiresCleanup  bool     `json:"requires_cleanup,omitempty"`
	Server           []string `json:"server,omitempty"`
}

type ListRoles struct {
	APIListObject
	Roles []Role `json:"objects",omitempty`
}

type ListRolesOptions struct {
	Name []string `url:"name__in",omitempty`
}

func (c *Client) GetRoleByPath(path string) (Role, *http.Response, error) {
	serverdbError := new(ServerdbError)
	role := new(Role)
	resp, err := c.sling.New().Base(baseURLForPath).Get(path).Receive(role, serverdbError)
	if err == nil && serverdbError.Err != "" {
		err = serverdbError
	}
	return *role, resp, err
}

func (c *Client) GetRolesByName(name []string) ([]Role, *http.Response, error) {
	serverdbError := new(ServerdbError)
	roles := new(ListRoles)
	query := ListRolesOptions{
		Name: name,
	}
	resp, err := c.sling.New().Get("serverrole/").QueryStruct(query).Receive(roles, serverdbError)
	if err == nil && serverdbError.Err != "" {
		err = serverdbError
	}
	return roles.Roles, resp, err

}

func (c *Client) GetRoleByName(name string) (Role, *http.Response, error) {
	names := []string{name}
	roles, resp, err := c.GetRolesByName(names)
	if len(roles) < 1 {
		return Role{}, nil, fmt.Errorf("Role with name [%s] not found", name)
	}
	if len(roles) > 1 {
		return Role{}, nil, fmt.Errorf("We were expecting a single role for the name [%s], but we got [%d] [%v]", name, len(roles), roles)
	}
	return roles[0], resp, err
}
