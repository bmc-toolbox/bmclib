package supermicrox10

// op=config_syslog&syslogport1=514&syslogport2=0&syslogport3=0&syslogip1=10.156.63.58&syslogip2=&syslogip3=&enable=1
// /cgi/op.cgi
type ConfigSyslog struct {
	Op          string `url:"op"`          //op=config_syslog
	SyslogIp1   string `url:"syslogip1"`   //syslogip1=10.01.12.1
	SyslogIp2   string `url:"syslogip2"`   //syslogip1=10.01.12.2
	SyslogIp3   string `url:"syslogip3"`   //syslogip1=10.01.12.3
	SyslogPort1 int    `url:"syslogport1"` //syslogport1=514
	SyslogPort2 int    `url:"syslogport2"` //syslogport2=0
	SyslogPort3 int    `url:"syslogport3"` //syslogport3=0
	Enable      bool   `url:"enable,int"`  //enable=1
}
