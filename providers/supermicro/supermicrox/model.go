package supermicrox

// ConfigSyslog declares payload to configure syslog parameters.
// /cgi/op.cgi
type ConfigSyslog struct {
	Op          string `url:"op"`          // op=config_syslog
	SyslogIP1   string `url:"syslogip1"`   // syslogip1=10.01.12.1
	SyslogIP2   string `url:"syslogip2"`   // syslogip1=10.01.12.2
	SyslogIP3   string `url:"syslogip3"`   // syslogip1=10.01.12.3
	SyslogPort1 int    `url:"syslogport1"` // syslogport1=514
	SyslogPort2 int    `url:"syslogport2"` // syslogport2=0
	SyslogPort3 int    `url:"syslogport3"` // syslogport3=0
	Enable      bool   `url:"enable,int"`  // enable=1
}

// ConfigDateTime declares payload to configure time parameters.
// /cgi/op.cgi
type ConfigDateTime struct {
	Op                 string `url:"op"`             // op=config_date_time
	Timezone           int    `url:"timezone"`       // timezone=-7200
	DstEn              bool   `url:"dst_en,int"`     // dst_en=0
	Enable             string `url:"ntp"`            // ntp=on
	NtpServerPrimary   string `url:"ntp_server_pri"` // ntp_server_pri=ntp0.example.com
	NtpServerSecondary string `url:"ntp_server_2nd"` // ntp_server_2nd=ntp1.example.com
	Year               int    `url:"year"`           // year=2018
	Month              int    `url:"month"`          // month=6
	Day                int    `url:"day"`            // day=1
	Hour               int    `url:"hour"`           // hour=05
	Minute             int    `url:"min"`            // min=49
	Second             int    `url:"sec"`            // sec=42
	TimeStamp          string `url:"time_stamp"`     // time_stamp=Fri%20Jun%2001%202018%2009%3A58%3A19%20GMT%2B0200%20(CEST)
}

// ConfigUser declares payload to configure User accounts.
// /cgi/config_user.cgi
type ConfigUser struct {
	Username     string `url:"username"`
	UserID       int    `url:"original_username"` // username integer
	Password     string `url:"password,omitempty"`
	NewPrivilege int    `url:"new_privilege,omitempty"` // 4 == administrator, 3 == operator
}

// ConfigLdap declares payload to configure LDAP.
// /cgi/op.cgi
type ConfigLdap struct {
	Op           string `url:"op"`        // op=config_ldap
	Enable       string `url:"en_ldap"`   // en_ldap=on
	EnableSsl    bool   `url:"enSSL,int"` // enSSL=1
	LdapIP       string `url:"ldapip"`    // ldapip=10.252.13.5
	BaseDn       string `url:"basedn"`    // basedn=cn=Supermicro,cn=bmcUsers
	LdapPort     int    `url:"ldapport"`  // ldapport=636
	BindDn       string `url:"bind_dn"`   // bind_dn=undefined <- default value
	BindPassword string `url:"bind_pwd"`  // bind_pwd=******** <- default value
}

// ConfigPort declares payload to configure services.
type ConfigPort struct {
	Op                string `url:"op"`                // op=config_port
	HTTPPort          int    `url:"HTTP_PORT"`         // HTTP_PORT=80
	HTTPSPort         int    `url:"HTTPS_PORT"`        // HTTPS_PORT=443
	IkvmPort          int    `url:"IKVM_PORT"`         // IKVM_PORT=5900
	VMPort            int    `url:"VM_PORT"`           // VM_PORT=623  <- virtual media port
	SSHPort           int    `url:"SSH_PORT"`          // SSH_PORT=22
	WsmanPort         int    `url:"WSMAN_PORT"`        // WSMAN_PORT=5985
	SnmpPort          int    `url:"SNMP_PORT"`         // SNMP_PORT=161
	httpEnable        bool   `url:"HTTP_SERVICE,int"`  // HTTP_SERVICE=1
	httpsEnable       bool   `url:"HTTPS_SERVICE,int"` // HTTPS_SERVICE=1
	IkvmEnable        bool   `url:"IKVM_SERVICE,int"`  // IKVM_SERVICE=1
	VMEnable          bool   `url:"VM_SERVICE,int"`    // VM_SERVICE=1
	SSHEnable         bool   `url:"SSH_SERVICE,int"`   // SSH_SERVICE=1
	SnmpEnable        bool   `url:"SNMP_SERVICE,int"`  // SNMP_SERVICE=1
	WsmanEnable       bool   `url:"WSMAN_SERVICE,int"` // WSMAN_SERVICE=0
	SslRedirectEnable bool   `url:"SSL_REDIRECT,int"`  // SSL_REDIRECT=1
}

// CapturePreview declares payload to capture screen previews.
type CapturePreview struct {
	IkvmPreview string `url:"IKVM_PREVIEW.XML"` // IKVM_PREVIEW.XML=(0,0)
	TimeStamp   string `url:"time_stamp"`       // time_stamp=Wed Oct 17 2018 15:56:08 GMT+0200 (CEST)
}

// URLRedirect declares payload sent when capturing screen previews
type URLRedirect struct {
	URLName   string `url:"url_name"`   // url_name=Snapshot
	URLType   string `url:"url_type"`   // url_type=img
	TimeStamp string `url:"time_stamp"` // time_stamp=Wed Oct 17 2018 15:56:08 GMT+0200 (CEST)
}
