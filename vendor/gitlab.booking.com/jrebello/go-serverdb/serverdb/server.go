package serverdb

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type Server struct {
	APIObject
	Description string   `json:"description,omitempty"`
	Arch        string   `json:"arch"`
	Cluster     string   `json:"cluster,omitempty"`
	Os          string   `json:"os"`
	Status      string   `json:"status"`
	Environment string   `json:"environment"`
	Memory      int      `json:"memory_in_gb"`
	Model       string   `json:"model"`
	Profile     string   `json:"profile"`
	Project     string   `json:"project"`
	Raid        string   `json:"raid,omitempty"`
	Rack        string   `json:"rack"`
	Role        []string `json:"role"`
	Site        string   `json:"site"`
}

type ListServers struct {
	APIListObject
	Servers []Server `json:"objects",omitempty`
}

type ListServersOptions struct {
	ChassisId      int      `url:"chassis,omitempty"`
	Name           string   `url:"name__exact,omitempty"`
	Status         string   `url:"status__exact,omitempty"`
	SiteShortName  string   `url:"site__shortname__exact,omitempty"`
	Environment    string   `url:"environment__exact,omitempty"`
	EnvironmentIn  []string `url:"environment__in,omitempty"`
	Format         string   `url:"format",omitempty`
	ProfileGroup   string   `url:"profile__group__name__in,omitempty"`
	MinMemory      string   `url:"memory_in_gb__gte,omitempty"`
	Model          int      `url:"model,omitempty"`
	Rack           string   `url:"rack__name,omitempty"`
	Role           string   `url:"role__name__exact,omitempty"`
	RoleIdIn       []int    `url:"role__id__in,omitempty"`
	Site           int      `url:"site__exact,omitempty"`
	NameStartsWith string   `url:"name__startswith,omitempty"`
	PortSpeed      int      `url:"interface__network__port_speed__exact,omitempty"`
	Limit          int      `url:"-"`
	ServerDbLimit  int      `url:"limit,omitempty"`
}

type ProvisionOptions struct {
	Name    string `url:"name"`
	Role    []int  `url:"role"`
	Os      string `url:"os"`
	Project int    `url:"project"`
	Raid    string `url:"raid,omitempty"`
}

type ProvisionConstraints struct {
	ProvisionOptions

	MaxOnRack          int
	RackConflictPrefix string
	NamePattern        string
	Site               int
	ProfileGroup       string
	Environment        string
	PortSpeed          int
}

type UpdateServerOptions struct {
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
	Os          string `json:"os,omitempty"`
	Cluster     string `json:"cluster,omitempty"`
	Project     string `json:"project,omitempty"`
}

type ByRackCount struct {
	servers []Server
	count   map[string]int
}

// List Servers
func (c *Client) ListServers(listOptions ListServersOptions) (ListServers, *http.Response, error) {
	serverdbError := new(ServerdbError)
	tmpServers := new(ListServers)
	servers := new(ListServers)
	resp, err := c.sling.New().Get("server/").QueryStruct(listOptions).Receive(servers, serverdbError)
	if err == nil && serverdbError.Err != "" {
		err = serverdbError
	}
	next := servers.Meta.Next
	read := len(servers.Servers)
	if c.Debug() {
		log.Printf("ListServers: making another query?. Read: [%d] Limit: [%d], HasNext: [%s]", read, listOptions.Limit, next)
	}

	for (listOptions.Limit == -1 || listOptions.Limit > read) && next != nil {
		if c.Debug() {
			log.Printf("ListServers: making another query. Read: [%d] Limit: [%d], HasNext: [%s]", read, listOptions.Limit, next)
		}
		resp, err = c.sling.New().Base(baseURLForPath).Get(*next).Receive(tmpServers, serverdbError)
		if err == nil && serverdbError.Err != "" {
			err = serverdbError
		}

		next = tmpServers.Meta.Next
		servers.Servers = append(servers.Servers, tmpServers.Servers...)
		read += len(tmpServers.Servers)
	}
	if c.Debug() {
		log.Printf("ListServers: The end. Read: [%d] Limit: [%d], HasNext: [%s], Len: [%d]", read, listOptions.Limit, next, len(servers.Servers))
	}

	return *servers, resp, err
}

//Read a server by ID
func (c *Client) ReadServerById(id int) (Server, *http.Response, error) {
	serverdbError := new(ServerdbError)
	server := new(Server)
	path := fmt.Sprintf("server/%d", id)
	resp, err := c.sling.New().Get(path).Receive(server, serverdbError)
	if err == nil && serverdbError.Err != "" {
		err = serverdbError
	}
	return *server, resp, err
}

//Update a server
func (c *Client) UpdateServer(id int, updateOptions UpdateServerOptions) (*http.Response, error) {
	serverdbError := new(ServerdbError)
	path := fmt.Sprintf("server/%d/", id)
	resp, err := c.sling.New().Patch(path).BodyJSON(updateOptions).Receive(nil, serverdbError)
	if err == nil && serverdbError.Err != "" {
		err = serverdbError
	}
	return resp, err
}

//Update a server, remove it from it's cluster
func (c *Client) UpdateServerRemoveCluster(id int) (*http.Response, error) {
	serverdbError := new(ServerdbError)
	updateOptions := struct {
		Cluster *string `json:"cluster"`
	}{
		nil,
	}
	path := fmt.Sprintf("server/%d/", id)
	resp, err := c.sling.New().Patch(path).BodyJSON(updateOptions).Receive(nil, serverdbError)
	if err == nil && serverdbError.Err != "" {
		err = serverdbError
	}
	return resp, err
}

// Provision a server with ID..
func (c *Client) ProvisionServer(id int, p ProvisionOptions) (Server, *http.Response, error) {
	serverdbError := new(ServerdbError)
	actionResult := new(ActionResult)
	path := fmt.Sprintf("server/%d/%s/", id, "provision")
	resp, err := c.sling.New().Post(path).BodyForm(p).Receive(actionResult, serverdbError)
	if err == nil && serverdbError.Err != "" {
		err = serverdbError
	}
	if err != nil {
		return Server{}, nil, err
	}
	server, _, err := c.ReadServerById(id)
	return server, resp, err
}

func nameMatch(pattern, name string) (string, error) {
	ret := ""
	if len(name) != len(pattern) {
		return "", fmt.Errorf("pattern and name don't have the same lenght: [%s] [%s]", pattern, name)
	}
	for idx, chr := range []byte(pattern) {
		n := name[idx]
		if n == chr {
			continue
		}
		if string(chr) != "#" {
			return "", fmt.Errorf("pattern is not matching the name: [%s] [%s]", pattern, name)
		}
		ret += string(n)
	}
	return ret, nil
}

func nextName(names []int) int {
	sort.Ints(names)
	name := 1
	if len(names) == 0 {
		return name
	}
	existing := make(map[int]bool)
	for _, n := range names {
		existing[n] = true
	}
	name = names[0]
	if name%10 != 1 && name%10 != 0 {
		return name - 1
	}

	for _, n := range names {
		if n != name && !existing[name] {
			return name
		}
		name = n + 1
	}
	return name
}

func fillPattern(pattern string, idx int) (string, error) {
	strName := strconv.Itoa(idx)
	l := strings.Count(pattern, "#")
	diff := l - len(strName)
	if diff < 0 {
		return "", fmt.Errorf("The pattern is not big enough for the server index")
	}
	s := strings.Repeat("0", diff)
	name := s + strName
	patt := strings.Repeat("#", l)
	return strings.Replace(pattern, patt, name, 1), nil
}

func (r ByRackCount) Len() int {
	return len(r.servers)
}

func (r ByRackCount) Swap(i, j int) {
	r.servers[i], r.servers[j] = r.servers[j], r.servers[i]
}

func (r ByRackCount) Less(i, j int) bool {
	return r.count[r.servers[i].Rack] < r.count[r.servers[j].Rack]
}

// Find a server and provision it
func (c *Client) ProvisionServerWithConstraints(options ProvisionConstraints) (Server, *http.Response, error) {
	// Get an available sever
	rackedOptions := ListServersOptions{
		Status:       "racked",
		Site:         options.Site,
		Environment:  options.Environment,
		ProfileGroup: options.ProfileGroup,
		PortSpeed:    options.PortSpeed,
		Limit:        -1,
	}
	racked, _, err := c.ListServers(rackedOptions)
	if err != nil {
		return Server{}, nil, err
	}
	if c.Debug() {
		log.Printf("Racked: [%v]\n", racked)
	}

	existing := ListServers{}
	// Check for existing servers
	existingOptions := ListServersOptions{
		Site:  options.Site,
		Limit: -1,
	}

	// XXX this is a sinister hack
	if options.Environment == "prod" {
		existingOptions.EnvironmentIn = []string{"prod", "oob"}
	} else {
		existingOptions.Environment = options.Environment
	}

	existingOptions.RoleIdIn = options.Role
	if options.RackConflictPrefix != "" {
		existingOptions.NameStartsWith = options.RackConflictPrefix
	}
	if c.Debug() {
		log.Printf("Existing query: [%v]\n", existingOptions)
	}

	existing, _, err = c.ListServers(existingOptions)
	if c.Debug() {
		log.Printf("Existing: [%v]\n", existing)
	}

	// Apply rack restrictions
	countByRack := make(map[string]int)
	foundById := make(map[int]bool)
	var names []int
	sort.Sort(ByRackCount{servers: existing.Servers, count: countByRack})
	for _, s := range existing.Servers {
		// Sometimes serverdb answers with dumplicate entries in a query. Don't double count a server
		if foundById[s.ID] {
			continue
		}
		foundById[s.ID] = true
		countByRack[s.Rack]++
		matched, err := nameMatch(options.NamePattern, s.Name)
		if err == nil {
			n, err := strconv.Atoi(matched)
			if err == nil {
				names = append(names, n)
			}
		}
	}
	if c.Debug() {
		log.Printf("Racked [%v]\n - Existing [%v]\n - Names [%v]\n - CountByRack[%v]\n", racked, existing, names, countByRack)
	}

	// Shuffle racked.Servers
	rackedLen := len(racked.Servers)
	for i := 0; i < rackedLen; i++ {
		pos := rand.Intn(rackedLen)
		racked.Servers[i], racked.Servers[pos] = racked.Servers[pos], racked.Servers[i]
	}

	found := 0
	for _, s := range racked.Servers {
		if countByRack[s.Rack]+1 > options.MaxOnRack && options.MaxOnRack != 0 {
			continue
		}
		found = s.ID
	}
	if found == 0 {
		return Server{}, nil, fmt.Errorf("No available server found matching the rack restrictions")
	}

	name := options.Name
	if name == "" {
		nameFound := false
		newName := ""
		for !nameFound {
			// Now get a name for the new server
			// TODO: check if we run out of numbers
			nameIndex := nextName(names)
			newName, err = fillPattern(options.NamePattern, nameIndex)
			if err != nil {
				return Server{}, nil, err
			}
			server, rsp, err := c.MaybeGetServerByName(newName)
			if err != nil {
				return Server{}, rsp, err
			}
			if server != nil {
				names = append(names, nameIndex)
				if c.Debug() {
					log.Printf("Nameindex [%d] is already taken. Names: [%v]", nameIndex, names)
				}
			} else {
				nameFound = true
			}
		}
		name = newName
	}
	options.Name = name
	if c.Debug() {
		log.Printf("Racked [%v]\n - Existing [%v]\n - Names [%v]\n - CountByRack[%v]\n - Name[%v]\n", racked, existing, names, countByRack, name)
	}
	return c.ProvisionServer(found, options.ProvisionOptions)
}

func (c *Client) DeprovisionServer(id int) (Server, *http.Response, error) {
	serverdbError := new(ServerdbError)
	actionResult := new(ActionResult)
	path := fmt.Sprintf("server/%d/%s/", id, "deprovision")
	deprovisionParam := struct {
		Deprovision bool `url:"deprovision"`
	}{
		true,
	}
	resp, err := c.sling.New().Post(path).BodyForm(deprovisionParam).Receive(actionResult, serverdbError)
	if err == nil && serverdbError.Err != "" {
		err = serverdbError
	}
	if err != nil {
		return Server{}, nil, err
	}
	server, _, err := c.ReadServerById(id)
	return server, resp, err
}

func (c *Client) GetServerByName(name string) (Server, *http.Response, error) {
	query := ListServersOptions{
		Name: name,
	}
	server, resp, err := c.ListServers(query)
	if err != nil {
		return Server{}, nil, err
	}
	if len(server.Servers) != 1 {
		return Server{}, nil, fmt.Errorf("we were expecting only one server and we got: [%v]", server.Servers)
	}
	return server.Servers[0], resp, err
}

func (c *Client) MaybeGetServerByName(name string) (*Server, *http.Response, error) {
	query := ListServersOptions{
		Name: name,
	}
	server, resp, err := c.ListServers(query)
	if err != nil {
		return nil, nil, err
	}
	if len(server.Servers) >= 2 {
		return nil, nil, fmt.Errorf("we were expecting only one server and we got more: [%v]", server.Servers)
	}
	if len(server.Servers) == 0 {
		return nil, nil, err
	}
	return &server.Servers[0], resp, err
}
