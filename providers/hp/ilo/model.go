package ilo

type Users struct {
	UsersInfo []UserInfo `json:"users"`
}

// Add/Modify/Delete a user account
// POST
// https://10.193.251.48/json/user_info
type UserInfo struct {
	Id               int    `json:"id,int,omitempty"`
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
	UserId           int    `json:"user_id,int,omitempty"`
	SessionKey       string `json:"session_key,omitempty"`
}

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

//Important timezone ints taken from https://10.193.251.48/html/network_sntp.html?intf=0
var Timezones = map[string]int{
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
