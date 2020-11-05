package ilo

// Users struct declares payload to un/marshal user accounts.
type Users struct {
	UsersInfo []UserInfo `json:"users"`
}

// DirectoryGroupAccts struct declares directory group account payload.
type DirectoryGroupAccts struct {
	Groups []DirectoryGroups `json:"group_accts"`
}

// LicenseInfo declares License information payload.
//POST https://10.183.244.173/json/license_info
type LicenseInfo struct {
	Key        string `json:"key,omitempty"`
	Method     string `json:"method,omitempty"`
	SessionKey string `json:"session_key,omitempty"`
}

// UserInfo struct declares payload for a user account.
// Add/Modify/Delete a user account
// POST
// https://10.193.251.48/json/user_info
type UserInfo struct {
	ID               int    `json:"id,omitempty"`
	LoginName        string `json:"login_name,omitempty"`
	UserName         string `json:"user_name,omitempty"`
	Password         string `json:"password,omitempty"`
	RemoteConsPriv   int    `json:"remote_cons_priv,omitempty"`
	VirtualMediaPriv int    `json:"virtual_media_priv,omitempty"`
	ResetPriv        int    `json:"reset_priv,omitempty"`
	ConfigPriv       int    `json:"config_priv,omitempty"`
	UserPriv         int    `json:"user_priv,omitempty"`
	LoginPriv        int    `json:"login_priv,omitempty"`
	Method           string `json:"method"` //mod_user, add_user, del_user
	UserID           int    `json:"user_id,omitempty"`
	SessionKey       string `json:"session_key,omitempty"`
}

// RemoteSyslog struct declares Syslog configuration payload.
// Set syslog params
// POST
// https://10.193.251.48/json/remote_syslog
type RemoteSyslog struct {
	SyslogEnable int    `json:"syslog_enable"`
	SyslogPort   int    `json:"syslog_port"`
	Method       string `json:"method"` //syslog_save,
	SyslogServer string `json:"syslog_server"`
	SessionKey   string `json:"session_key,omitempty"`
}

// NetworkSntp struct declares network services configuration payload.
// /json/network_sntp
type NetworkSntp struct {
	Interface                   int    `json:"interface"`
	PendingChange               int    `json:"pending_change"`
	NicWcount                   int    `json:"nic_wcount"`
	TzWcount                    int    `json:"tz_wcount"`
	Ipv4Disabled                int    `json:"ipv4_disabled"`
	Ipv6Disabled                int    `json:"ipv6_disabled"`
	DhcpEnabled                 int    `json:"dhcp_enabled"`
	Dhcp6Enabled                int    `json:"dhcp6_enabled"`
	UseDhcpSuppliedTimeServers  int    `json:"use_dhcp_supplied_time_servers"`
	UseDhcp6SuppliedTimeServers int    `json:"use_dhcp6_supplied_time_servers"`
	Sdn1WCount                  int    `json:"sdn1_wcount"`
	Sdn2WCount                  int    `json:"sdn2_wcount"`
	SntpServer1                 string `json:"sntp_server1"`
	SntpServer2                 string `json:"sntp_server2"`
	TimePropagate               int    `json:"time_propagate"` //propagate time from OA to blade
	OurZone                     int    `json:"our_zone"`       //368 - see Timezones
	Method                      string `json:"method"`         //set_sntp
	SessionKey                  string `json:"session_key,omitempty"`
}

// Directory struct declares LDAP configuration payload.
// /json/directory
//{"server_address":"ldap.example.com","method":"mod_dir_config","session_key":"51b01f402d65eb2f42342f6d67832989","server_port":637,"user_contexts":["ou=People,dc=example,dc=con"],"authentication_enabled":1,"enable_group_acct":1,"enable_kerberos":0,"local_user_acct":1,"enable_generic_ldap":1}
type Directory struct {
	ServerAddress         string   `json:"server_address"`
	ServerPort            int      `json:"server_port"`
	UserContexts          []string `json:"user_contexts"`
	AuthenticationEnabled int      `json:"authentication_enabled"`
	LocalUserAcct         int      `json:"local_user_acct"` //enable local user accounts
	EnableGroupAccount    int      `json:"enable_group_acct"`
	EnableKerberos        int      `json:"enable_kerberos"`
	KerberosKdcAddress    string   `json:"kerberos_kdc_address,omitempty"`
	KerberosRealm         string   `json:"kerberos_realm,omitempty"`
	EnableGenericLdap     int      `json:"enable_generic_ldap"`
	Method                string   `json:"method"`
	SessionKey            string   `json:"session_key"`
}

// DirectoryGroups declares LDAP groups configuration payload.
// /json/directory_groups
//{"dn":"cn=hp,cn=bmcUsers","new_dn":"cn=hp,cn=bmcUsers","sid":"","login_priv":1,"remote_cons_priv":1,"virtual_media_priv":1,"reset_priv":1,"config_priv":0,"user_priv":0,"method":"mod_group","session_key":"bc2dae77e36a45fbeffce0bddd2ccabe"}
type DirectoryGroups struct {
	Dn               string `json:"dn"`
	NewDn            string `json:"new_dn,omitempty"` //same as Dn, unless being modified
	Sid              string `json:"sid,omitempty"`
	LoginPriv        int    `json:"login_priv,omitempty"`
	RemoteConsPriv   int    `json:"remote_cons_priv,omitempty"`
	VirtualMediaPriv int    `json:"virtual_media_priv,omitempty"`
	ResetPriv        int    `json:"reset_priv,omitempty"`
	ConfigPriv       int    `json:"config_priv,omitempty"`
	UserPriv         int    `json:"user_priv,omitempty"`
	Method           string `json:"method"` //add_group, mod_group, del_group
	SessionKey       string `json:"session_key"`
}

// Generate CSR
// POST /json/csr
type csr struct {
	Country          string `json:"country"`
	State            string `json:"state"`
	Locality         string `json:"locality"`
	OrganizationName string `json:"organization_name"`
	OrganizationUnit string `json:"organization_unit"`
	CommonName       string `json:"common_name"`
	IncludeIP        int    `json:"include_ip"`
	Method           string `json:"method"`
	SessionKey       string `json:"session_key"`
}

// The CSR response
type csrResponse struct {
	CsrPEM string `json:"csr_pem"`
}

// Cert import
// POST json/certificate
type certImport struct {
	Method          string `json:"method"`
	CertificateData string `json:"certificate_data"`
	SessionKey      string `json:"session_key"`
}

// AccessSettings declares BMC network service ports
// Updating these params requires the BMC to be reset.
type AccessSettings struct {
	SSHStatus                    int           `json:"ssh_status"`
	SSHPort                      int           `json:"ssh_port"`
	HTTPPort                     int           `json:"http_port"`
	HTTPSPort                    int           `json:"https_port"`
	RemoteConsolePort            int           `json:"remote_console_port"`
	VirtualMediaPort             int           `json:"virtual_media_port"`
	IpmiLanStatus                int           `json:"ipmi_lan_status"`
	IpmiPort                     int           `json:"ipmi_port"`
	SNMPSettings                 *SNMPSettings `json:"snmp_settings"`
	SessionTimeout               int           `json:"session_timeout"`
	IloFunctEnabled              int           `json:"ilo_funct_enabled"`
	IloFunctRequired             int           `json:"ilo_funct_required"`
	RbsuEnabled                  int           `json:"rbsu_enabled"`
	F8LoginRequired              int           `json:"f8_login_required"`
	RbsuPostIP                   int           `json:"rbsu_post_ip"`
	SerialCliStatus              int           `json:"serial_cli_status"`
	SystemNoUart                 int           `json:"system_no_uart"`
	SerialCliSpeed               int           `json:"serial_cli_speed"`
	VspLogging                   int           `json:"vsp_logging"`
	AuthenticationFailureLogging int           `json:"authentication_failure_logging"`
	MinPassword                  int           `json:"min_password"`
	AuthFailureDelayTime         int           `json:"auth_failure_delay_time"`
	AuthNodelayFailures          int           `json:"auth_nodelay_failures"`
	ServerName                   string        `json:"server_name"`
	ServerFqdn                   string        `json:"server_fqdn"`
	DefaultLang                  string        `json:"default_lang"`
	SessionKey                   string        `json:"session_key"`
	Method                       string        `json:"method"`
}

// SNMPSettings declares BMC SNMP params
type SNMPSettings struct {
	SnmpPort            int `json:"snmp_port"`
	TrapPort            int `json:"trap_port"`
	SnmpExternalDisable int `json:"snmp_external_disable"`
}

// NetworkIPv4 sets IPv4 network settings
// The BMC would require a reset if these params are updated.
type NetworkIPv4 struct {
	Interface                   int    `json:"interface"`
	PendingChange               int    `json:"pending_change"`
	DhcpEnabled                 int    `json:"dhcp_enabled"`
	UseDhcpSuppliedGateway      int    `json:"use_dhcp_supplied_gateway"`
	UseDhcpSuppliedDNS          int    `json:"use_dhcp_supplied_dns"`
	UseDhcpSuppliedWins         int    `json:"use_dhcp_supplied_wins"`
	UseDhcpSuppliedStaticRoutes int    `json:"use_dhcp_supplied_static_routes"`
	UseDhcpSuppliedDomainName   int    `json:"use_dhcp_supplied_domain_name"`
	UseDhcpSuppliedTimeServers  int    `json:"use_dhcp_supplied_time_servers"`
	IPAddress                   string `json:"ip_address"`
	SubnetMask                  string `json:"subnet_mask"`
	GatewayIPAddress            string `json:"gateway_ip_address"`
	PingGateway                 int    `json:"ping_gateway"`
	RegWinsServer               int    `json:"reg_wins_server"`
	DNS                         []ipv4 `json:"dns"`
	Wins                        []ipv4 `json:"wins"`
	RegDdnsServer               int    `json:"reg_ddns_server"`
	StaticRouteDest             []ipv4 `json:"static_route_dest"`
	StaticRouteMask             []ipv4 `json:"static_route_mask"`
	StaticRouteGate             []ipv4 `json:"static_route_gate"`
	Wcount                      int    `json:"wcount"`
	SessionKey                  string `json:"session_key"`
	Method                      string `json:"method"`
}

type ipv4 struct {
	ID     int    `json:"id"`
	Ipv4IP string `json:"ipv4_ip"`
}

// PowerRegulator declares json payload to set power regulator mode
type PowerRegulator struct {
	PowerMode  string `json:"prmode"`
	SessionKey string `json:"session_key"`
	Method     string `json:"method"`
}

// TimezonesIlo5 declares valid timezone for ilo5 devices.
var TimezonesIlo5 = map[string]int{
	"Etc/GMT+12":                     0,
	"Pacific/Midway":                 1,
	"US/Hawaii":                      2,
	"US/Alaska":                      3,
	"US/Pacific":                     4,
	"US/Mountain":                    5,
	"US/Central":                     6,
	"US/Eastern":                     7,
	"America/Caracas":                8,
	"Canada/Atlantic":                9,
	"Canada/Newfoundland":            10,
	"America/Argentina/Buenos_Aires": 11,
	"Atlantic/South_Georgia":         12,
	"Atlantic/Cape_Verde":            13,
	"Greenwich":                      14,
	"CET":                            15,
	"EET":                            16,
	"Asia/Kuwait":                    17,
	"Asia/Tehran":                    18,
	"Asia/Dubai":                     19,
	"Asia/Kabul":                     20,
	"Asia/Yekaterinburg":             21,
	"Asia/Kolkata":                   22,
	"Asia/Kathmandu":                 23,
	"Asia/Almaty":                    24,
	"Asia/Rangoon":                   25,
	"Asia/Bangkok":                   26,
	"Asia/Taipei":                    27,
	"Asia/Tokyo":                     28,
	"Australia/Adelaide":             29,
	"Australia/Canberra":             30,
	"Asia/Magadan":                   31,
	"Pacific/Auckland":               32,
	"Pacific/Fakaofo":                33,
	"Pacific/Kiritimati":             34,
	"Unspecified Time Zone":          35,
}

// TimezonesIlo4 declares valid timezones.
//Important timezone ints taken from https://10.193.251.48/html/network_sntp.html?intf=0
var TimezonesIlo4 = map[string]int{
	"CET":           368,
	"CST6CDT":       371,
	"EET":           373,
	"EST":           376,
	"EST5EDT":       377,
	"Etc/GMT":       378,
	"Etc/GMT+0":     379,
	"Etc/GMT+1":     380,
	"Etc/GMT+10":    381,
	"Etc/GMT+11":    382,
	"Etc/GMT+12":    383,
	"Etc/GMT+2":     384,
	"Etc/GMT+3":     385,
	"Etc/GMT+4":     386,
	"Etc/GMT+5":     387,
	"Etc/GMT+6":     388,
	"Etc/GMT+7":     389,
	"Etc/GMT+8":     390,
	"Etc/GMT+9":     391,
	"Etc/GMT-0":     392,
	"Etc/GMT-1":     393,
	"Etc/GMT-10":    394,
	"Etc/GMT-11":    395,
	"Etc/GMT-12":    396,
	"Etc/GMT-13":    397,
	"Etc/GMT-14":    398,
	"Etc/GMT-2":     399,
	"Etc/GMT-3":     400,
	"Etc/GMT-4":     401,
	"Etc/GMT-5":     402,
	"Etc/GMT-6":     403,
	"Etc/GMT-7":     404,
	"Etc/GMT-8":     405,
	"Etc/GMT-9":     406,
	"Etc/GMT0":      407,
	"Etc/Greenwich": 408,
	"Etc/UCT":       409,
	"Etc/Universal": 410,
	"Etc/UTC":       411,
	"GMT":           464,
	"GMT+0":         465,
	"GMT-0":         466,
	"GMT0":          467,
	"Greenwich":     468,
	"HST":           470,
	"MET":           488,
	"MST":           492,
	"MST7MDT":       493,
	"PST8PDT":       543,
	"UCT":           548,
	"Universal":     549,
	"UTC":           562,
	"WET":           564,
}
