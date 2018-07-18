package serverdb

import (
	"fmt"
	"net/http"
)

type Project struct {
	APIObject
	Active bool `json:"active"`
}

type ListProjects struct {
	APIListObject
	Projects []Project `json:"objects"`
}

type ListProjectOptions struct {
	Name string `url:"name,omitempty"`
}

// List projects by name
func (c *Client) GetProjectByName(name string) (Project, *http.Response, error) {
	serverdbError := new(ServerdbError)
	projects := new(ListProjects)
	query := ListProjectOptions{
		Name: name,
	}
	resp, err := c.sling.New().Get("project/").QueryStruct(query).Receive(projects, serverdbError)
	if err == nil && serverdbError.Err != "" {
		err = serverdbError
	}
	if len(projects.Projects) == 0 {
		return Project{}, nil, fmt.Errorf("No project with this name : [%s]", name)
	}
	return projects.Projects[0], resp, err
}

func (c *Client) GetProjectByPath(path string) (Project, *http.Response, error) {
	serverdbError := new(ServerdbError)
	project := new(Project)
	resp, err := c.sling.New().Base(baseURLForPath).Get(path).Receive(project, serverdbError)
	if err == nil && serverdbError.Err != "" {
		err = serverdbError
	}
	return *project, resp, err
}

func (c *Client) GetProjectById(id int) (Project, *http.Response, error) {
	serverdbError := new(ServerdbError)
	project := new(Project)
	path := fmt.Sprintf("project/%d", id)
	resp, err := c.sling.New().Get(path).Receive(project, serverdbError)
	if err == nil && serverdbError.Err != "" {
		err = serverdbError
	}
	return *project, resp, err
}
