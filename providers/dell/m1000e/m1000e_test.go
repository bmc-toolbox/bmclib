package m1000e

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ncode/bmclib/devices"
	"github.com/spf13/viper"
)

var (
	mux                *http.ServeMux
	server             *httptest.Server
	dellChassisAnswers = map[string]map[string][]byte{
		"/cgi-bin/webcgi/login": {
			"default": []byte(``),
		},
		"/cgi-bin/webcgi/logout": {
			"default": []byte(``),
		},
		"/cgi-bin/webcgi/cmc_status": {
			"default": []byte(`<?xml version="1.0"?>
				<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/html4/strict.dtd">
				<html xmlns="http://www.w3.org/1999/xhtml" xmlns:fo="http://www.w3.org/1999/XSL/Format">
				  <head>
					<meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
					<title>Chassis Controller Status</title>
					<link rel="stylesheet" type="text/css" href="/cmc/css/stylesheet.css?0818" />
					<script language="javascript" src="/cmc/js/validate.js?0818"></script>
					<script language="javascript" src="/cmc/js/Clarity.js?0818"></script>
					<script type="text/javascript" src="/cmc/js/context_help.js?0818"></script>
					<script language="javascript">
						UpdateHelpIdAndState(context_help("Chassis Controller Status"), true);
					  </script>
				  </head>
				  <body class="data-area">
					<div class="data-area" id="dataarea">
					  <a name="top" id="top"></a>
					  <div xmlns="" id="pullstrip" onmousedown="javascript:pullTab.ra_resizeStart(event, this);">
						<div id="pulltab"></div>
					  </div>
					  <div xmlns="" id="rightside"></div>
					  <div xmlns="" class="data-area-page-title">
						<span id="pageTitle">Chassis Controller Status</span>
						<div class="toolbar">
						  <a id="A2" name="printbutton" class="print" href="javascript:window.print();" title="Print"></a>
						  <a id="A5" name="refresh" class="refresh" href="javascript:top.globalnav.f_refresh();" title="Refresh"></a>
						  <a id="A6" name="help" class="help" href="javascript:top.globalnav.f_help();" title="Help"></a>
						</div>
						<div class="da-line"></div>
					  </div>
					  <div class="data-area-jump-bar">
						<span class="data-area-jump-items">Jump to:<a id="jb1" href="#general_info" class="data-area-jump-bar">General Information</a>|<a id="jb2" href="#common_network_info" class="data-area-jump-bar">Common Network Information</a>|<a id="jb3" href="#ipv4_info" class="data-area-jump-bar">IPv4 Information</a>|<a id="jb4" href="#ipv6_info" class="data-area-jump-bar">IPv6 Information</a></span>
						<div class="jumpbar-line"></div>
					  </div>
					  <div xmlns="" class="table_container">
						<div class="backtotop">
						  <a href="#top">Back to top</a>
						</div>
						<a name="general_info" id="general_info"></a>
						<div class="table_title">General Information</div>
						<table class="container">
						  <thead>
							<tr>
							  <td class="topleft" width="3px"></td>
							  <td class="top borderright" width="49%">Attribute</td>
							  <td class="top">Value</td>
							  <td class="topright"></td>
							</tr>
						  </thead>
						  <tbody>
							<tr>
							  <td class="left" width="3px"></td>
							  <td class="contents borderright borderbottom">Health</td>
							  <td class="contents borderbottom" id="cmcHealth">
								<img name="icon" src="/cmc/images/ok.png" id="icon" />
							  </td>
							  <td class="right"></td>
							</tr>
							<tr class="fill">
							  <td class="left" width="3px"></td>
							  <td class="contents borderright borderbottom">Date/Time</td>
							  <td class="contents borderbottom" id="TIME">Thu  2 Nov 2017 08:54:48 PM CST6CDT</td>
							  <td class="right"></td>
							</tr>
							<tr>
							  <td class="left" width="3px"></td>
							  <td class="contents borderright borderbottom">Active CMC Location</td>
							  <td class="contents borderbottom" id="CMC_Active_Slot">CMC-1</td>
							  <td class="right"></td>
							</tr>
							<tr class="fill">
							  <td class="left" width="3px"></td>
							  <td class="contents borderright borderbottom">Redundancy Mode</td>
							  <td class="contents borderbottom" id="CMC_Redundancy_Mode">Full Redundancy</td>
							  <td class="right"></td>
							</tr>
							<tr>
							  <td class="left" width="3px"></td>
							  <td class="contents borderright borderbottom">Primary Firmware Version</td>
							  <td class="contents borderbottom" id="cmc_fw_version">6.00</td>
							  <td class="right"></td>
							</tr>
							<tr class="fill">
							  <td class="left" width="3px"></td>
							  <td class="contents borderright borderbottom">Firmware Last Updated</td>
							  <td class="contents borderbottom" id="last_update">Wed Oct 18 16:25:38 2017</td>
							  <td class="right"></td>
							</tr>
							<tr>
							  <td class="left" width="3px"></td>
							  <td class="contents borderright borderbottom">Hardware Version</td>
							  <td class="contents borderbottom" id="cmc_hw_version">A12</td>
							  <td class="right"></td>
							</tr>
						  </tbody>
						  <tfoot>
							<tr>
							  <td class="bottomleft" width="3px"></td>
							  <td class="bottom" colspan="2"></td>
							  <td class="bottomright"></td>
							</tr>
						  </tfoot>
						</table>
					  </div>
					  <div xmlns="" class="table_container">
						<div class="backtotop">
						  <a href="#top">Back to top</a>
						</div>
						<a name="common_network_info" id="common_network_info"></a>
						<div class="table_title">Common Network Information</div>
						<table class="container">
						  <thead>
							<tr>
							  <td class="topleft" width="3px"></td>
							  <td class="top borderright" width="49%">Attribute</td>
							  <td class="top">Value</td>
							  <td class="topright"></td>
							</tr>
						  </thead>
						  <tbody>
							<tr>
							  <td class="left" width="3px"></td>
							  <td class="contents borderright borderbottom">MAC Address</td>
							  <td class="contents borderbottom" id="mac_addr">18:66:DA:9D:CD:CD</td>
							  <td class="right"></td>
							</tr>
							<tr class="fill">
							  <td class="left" width="3px"></td>
							  <td class="contents borderright borderbottom">DNS Domain Name</td>
							  <td class="contents borderbottom" id="DNSCurrentDomainName">machine.example.com</td>
							  <td class="right"></td>
							</tr>
							<tr>
							  <td class="left" width="3px"></td>
							  <td class="contents borderright borderbottom">Use DHCP for DNS Domain Name</td>
							  <td class="contents borderbottom" id="DNS_use_dhcp_domain">Yes</td>
							  <td class="right"></td>
							</tr>
						  </tbody>
						  <tfoot>
							<tr>
							  <td class="bottomleft" width="3px"></td>
							  <td class="bottom" colspan="2"></td>
							  <td class="bottomright"></td>
							</tr>
						  </tfoot>
						</table>
					  </div>
					  <div xmlns="" class="table_container">
						<div class="backtotop">
						  <a href="#top">Back to top</a>
						</div>
						<a name="ipv4_info" id="ipv4_info"></a>
						<div class="table_title">IPv4 Information</div>
						<table class="container">
						  <thead>
							<tr>
							  <td class="topleft" width="3px"></td>
							  <td class="top borderright" width="49%">Attribute</td>
							  <td class="top">Value</td>
							  <td class="topright"></td>
							</tr>
						  </thead>
						  <tbody>
							<tr>
							  <td class="left" width="3px"></td>
							  <td class="contents borderright borderbottom">IPv4 Enabled</td>
							  <td class="contents borderbottom" id="CurrentIPv4Enabled">Yes</td>
							  <td class="right"></td>
							</tr>
							<tr class="fill">
							  <td class="left" width="3px"></td>
							  <td class="contents borderright borderbottom">DHCP Enabled</td>
							  <td class="contents borderbottom" id="NETWORK_NIC_dhcp_IPv6_enable">Yes</td>
							  <td class="right"></td>
							</tr>
							<tr>
							  <td class="left" width="3px"></td>
							  <td class="contents borderright borderbottom">IP Address</td>
							  <td class="contents borderbottom" id="ipaddr">10.193.251.36</td>
							  <td class="right"></td>
							</tr>
							<tr class="fill">
							  <td class="left" width="3px"></td>
							  <td class="contents borderright borderbottom">Subnet Mask</td>
							  <td class="contents borderbottom" id="netmask">255.255.255.0</td>
							  <td class="right"></td>
							</tr>
							<tr>
							  <td class="left" width="3px"></td>
							  <td class="contents borderright borderbottom">Gateway</td>
							  <td class="contents borderbottom" id="gateway">10.193.251.254</td>
							  <td class="right"></td>
							</tr>
							<tr class="fill">
							  <td class="left" width="3px"></td>
							  <td class="contents borderright borderbottom">Use DHCP to obtain DNS server addresses</td>
							  <td class="contents borderbottom" id="DNS_dhcp_enable">Yes</td>
							  <td class="right"></td>
							</tr>
							<tr>
							  <td class="left" width="3px"></td>
							  <td class="contents borderright borderbottom">Preferred DNS Server</td>
							  <td class="contents borderbottom" id="DNSCurrentServer1">10.252.13.1</td>
							  <td class="right"></td>
							</tr>
							<tr class="fill">
							  <td class="left" width="3px"></td>
							  <td class="contents borderright borderbottom">Alternate DNS Server</td>
							  <td class="contents borderbottom" id="DNSCurrentServer2">10.252.13.2</td>
							  <td class="right"></td>
							</tr>
						  </tbody>
						  <tfoot>
							<tr>
							  <td class="bottomleft" width="3px"></td>
							  <td class="bottom" colspan="2"></td>
							  <td class="bottomright"></td>
							</tr>
						  </tfoot>
						</table>
					  </div>
					  <div xmlns="" class="table_container">
						<div class="backtotop">
						  <a href="#top">Back to top</a>
						</div>
						<a name="ipv6_info" id="ipv6_info"></a>
						<div class="table_title">IPv6 Information</div>
						<table class="container">
						  <thead>
							<tr>
							  <td class="topleft" width="3px"></td>
							  <td class="top borderright" width="49%">Attribute</td>
							  <td class="top">Value</td>
							  <td class="topright"></td>
							</tr>
						  </thead>
						  <tbody>
							<tr>
							  <td class="left" width="3px"></td>
							  <td class="contents borderright borderbottom">IPv6 Enabled</td>
							  <td class="contents borderbottom" id="CurrentIPv6Enabled">Yes</td>
							  <td class="right"></td>
							</tr>
							<tr class="fill">
							  <td class="left" width="3px"></td>
							  <td class="contents borderright borderbottom">Autoconfiguration Enabled</td>
							  <td class="contents borderbottom" id="CMCIPv6AutoconfigEnable">Yes</td>
							  <td class="right"></td>
							</tr>
							<tr>
							  <td class="left" width="3px"></td>
							  <td class="contents borderright borderbottom">Link Local Address</td>
							  <td class="contents borderbottom" id="CurrentIPv6LinkAddress">fe80::1a66:daff:fe9d:cdcd/64</td>
							  <td class="right"></td>
							</tr>
							<tr class="fill">
							  <td class="left" width="3px"></td>
							  <td class="contents borderright borderbottom">IPv6 Address 
					1</td>
							  <td class="contents borderbottom" id="ipv6addr1">::</td>
							  <td class="right"></td>
							</tr>
							<tr>
							  <td class="left" width="3px"></td>
							  <td class="contents borderright borderbottom">Gateway</td>
							  <td class="contents borderbottom" id="gateway">::</td>
							  <td class="right"></td>
							</tr>
							<tr class="fill">
							  <td class="left" width="3px"></td>
							  <td class="contents borderright borderbottom">Use DHCPv6 to obtain DNS Server Addresses</td>
							  <td class="contents borderbottom" id="DNS_dhcp_IPv6_enable">Yes</td>
							  <td class="right"></td>
							</tr>
							<tr>
							  <td class="left" width="3px"></td>
							  <td class="contents borderright borderbottom">Preferred DNS Server</td>
							  <td class="contents borderbottom" id="DNSCurrentServer1">::</td>
							  <td class="right"></td>
							</tr>
							<tr class="fill">
							  <td class="left" width="3px"></td>
							  <td class="contents borderright borderbottom">Alternate DNS Server</td>
							  <td class="contents borderbottom" id="DNSCurrentServer2">::</td>
							  <td class="right"></td>
							</tr>
						  </tbody>
						  <tfoot>
							<tr>
							  <td class="bottomleft" width="3px"></td>
							  <td class="bottom" colspan="2"></td>
							  <td class="bottomright"></td>
							</tr>
						  </tfoot>
						</table>
					  </div>
					</div>
				  </body>
				</html>
				`),
		},
		"/cgi-bin/webcgi/json": {
			"groupinfo":       []byte(`{"ChassisGroup":{"ChassisGroupLicensed":1,"ChassisGroupName":"","ChassisGroupRoleStr":"No Role","ChassisGroupLeaderChassisName":"","ChassisGroupLeaderNode":"cmc-51F3DK2","ChassisGroupRole":0},"0":{"ChassisGroupMemberHealthBlob":{"health_status":{"blade":3,"global":3,"power":3,"lcdActiveErrorSev":3,"lcdActiveError":"No Errors","cmc":3,"iom":3,"rear":"\/graphics\/rearImage21428.png","front":"\/graphics\/frontImage21428.png","lcd":3,"fans":3,"temp":3,"lcdStatus":1,"kvm":3},"getcmccel":{"celsev_1":"2","celdesc_1":"A firmware or software incompatibility was corrected between CMC in slot 1 and CMC in slot 2.","celsev_2":"2","celdesc_4":"The chassis management controller (CMC) is not redundant.","celdate_6":"Wed Oct 18 2017 16:25:36","celdate_0":"Wed Oct 18 2017 16:29:32","celdesc_2":"The chassis management controller (CMC) is not redundant.","celdate_9":"Wed Oct 18 2017 16:22:32","celdate_2":"Wed Oct 18 2017 16:29:29","celdesc_9":"Chassis management controller (CMC) redundancy is lost.","celdate_1":"Wed Oct 18 2017 16:29:31","celsev_9":"4","celdate_7":"Wed Oct 18 2017 16:22:40","celdate_4":"Wed Oct 18 2017 16:25:39","celsev_0":"2","celdesc_8":"The chassis management controller (CMC) is redundant.","celdate_8":"Wed Oct 18 2017 16:22:35","celsev_7":"4","celdate_3":"Wed Oct 18 2017 16:27:39","celdesc_7":"Chassis management controller (CMC) redundancy is lost.","celsev_8":"2","celsev_5":"4","celdesc_6":"The chassis management controller (CMC) is redundant.","celsev_6":"2","celsev_3":"4","celdesc_3":"Chassis management controller (CMC) redundancy is lost.","celdesc_0":"The chassis management controller (CMC) is redundant.","celdate_5":"Wed Oct 18 2017 16:25:37","celsev_4":"2","celdesc_5":"A firmware or software incompatibility is detected between CMC in slot 2 and CMC in slot 1."},"blades_status":{"SlotName_host":{"BLADESLOT_NAME_usehostname":1},"1":{"bladeTemperature":"12","storageSelectedFabric":0,"bladeVKVMLicensed":1,"bladeSystemName":"provision-test-13163.ams4.example.com","bladePresent":1,"idracURL":"https:\/\/10.193.251.5:443","bladeLogDescription":"No Errors","bladeIMCStatus":0,"bladeFwUpdatable":1,"bladeVKVMSupported":1,"nic":{"1":{"bladeNicName":"Intel(R) Ethernet 10G 2P X520-k bNDC - EC:F4:BB:EB:A2:48","bladeNicVer":"18.0.17"},"0":{"bladeNicName":"Intel(R) Ethernet 10G 2P X520-k bNDC - EC:F4:BB:EB:A2:4A","bladeNicVer":"18.0.17"}},"storageNumDrives":0,"bladeIdracExternalManaged":0,"bladeLEDColor":0,"bladeBudgetAlloc":172,"storageNumControllers":0,"bladePriority":1,"bladePartNum":"0PHY8D","bladeDiagsVer":"4239A33","bladeMasterSlot":1,"pwrAccumulate":"1777.4","bladeSlotName":"provision-test-13163.ams4.example.com","bladeUSCVer":"2.41.40.40","WSMAN_Ready":1,"bladeNicEnable":1,"bladeCurrentConsumption":172,"pwrMaxTime":"Mon 21 Dec 2015 06:08:02 AM","bladeOsdrvVer":"15.10.02","bladeTotalMem":"128.0","bladeCpuInfo":"2 x Intel(R) Xeon(R) CPU E5-2620 v3 @ 2.40GHz","pwrStartTime":"Mon 21 Dec 2015 04:11:30 AM","isConfigured":0,"bladeSvcTag":"4NY1H92","bladeLogSeverity":3,"pwrMaxConsump":354,"bladePwrCtlSupported":1,"bladeBIOSver":"2.4.2","actualPwrConsump":158,"bladeSlot":"1","bladeManufacturer":"70163","isStorageBlade":0,"bladeTempUpperCriticalThreshold":"47","raid":{"1":{"bladeRaidVer":"D416","bladeRaidName":"Disk 0 in Backplane 1 of Integrated RAID Controller 1"},"0":{"bladeRaidVer":"25.5.3.0005","bladeRaidName":"PERC H730P Mini"},"2":{"bladeRaidVer":"2.25","bladeRaidName":"BP13G+ 0:1"}},"bladeBootOnce":1,"bladeModel":"PowerEdge M630","bladeMinBudgetAlloc":120,"bladeFwVer":"2.41.40.40 (07)","bladeTempUpperNonCriticalThreshold":"42","bladeMaxConsumption":354,"bladeLEDState":16,"bladeFirstBootDevice":0,"bladeWidth":0,"bladeCPLDVer":"1.0.5","bladePowerStatus":1,"bladeTempLowerNonCriticalThreshold":"3","bladeHealth":3,"bladeFormFactor":114,"bladeTempLowerCriticalThreshold":"-7","bladeOS":"","bladeName":"provision-test-13163.ams4.example.com","bladeSerialNum":"CN701635BS003Q"},"15":{"bladeTemperature":-1,"bladeSvcTag":"N\/A","bladeMasterSlot":15,"bladePresent":0,"bladeSlotName":"SLOT-15","bladeSlot":"15","bladeHealth":-1,"idracURL":"","bladePowerStatus":-1,"bladeName":"SLOT-15","bladePriority":1},"3":{"bladeTemperature":"13","storageSelectedFabric":0,"bladeVKVMLicensed":1,"bladeSystemName":"localhost.localdomain","bladePresent":1,"idracURL":"https:\/\/10.193.251.17:443","bladeLogDescription":"No Errors","bladeIMCStatus":0,"bladeFwUpdatable":1,"bladeVKVMSupported":1,"nic":{"1":{"bladeNicName":"Intel(R) Ethernet 10G 2P X520-k bNDC - 24:6E:96:0D:42:8E","bladeNicVer":"17.5.10"},"0":{"bladeNicName":"Intel(R) Ethernet 10G 2P X520-k bNDC - 24:6E:96:0D:42:8C","bladeNicVer":"17.5.10"}},"storageNumDrives":0,"bladeIdracExternalManaged":0,"bladeLEDColor":0,"bladeBudgetAlloc":373,"storageNumControllers":0,"bladePriority":1,"bladePartNum":"0PHY8D","bladeDiagsVer":"4239A33","bladeMasterSlot":3,"pwrAccumulate":"1799.2","bladeSlotName":"localhost.localdomain","bladeUSCVer":"2.41.40.40","WSMAN_Ready":1,"bladeNicEnable":1,"bladeCurrentConsumption":373,"pwrMaxTime":"Tue 04 Apr 2017 01:47:16 PM","bladeOsdrvVer":"15.10.02","bladeTotalMem":"128.0","bladeCpuInfo":"2 x Intel(R) Xeon(R) CPU E5-2690 v3 @ 2.60GHz","pwrStartTime":"Thu 25 Feb 2016 03:39:09 PM","isConfigured":0,"bladeSvcTag":"FQG0VB2","bladeLogSeverity":3,"pwrMaxConsump":554,"bladePwrCtlSupported":1,"bladeBIOSver":"2.4.2","actualPwrConsump":224,"bladeSlot":"3","bladeManufacturer":"70163","isStorageBlade":0,"bladeTempUpperCriticalThreshold":"47","raid":{"1":{"bladeRaidVer":"TT31","bladeRaidName":"Disk 0 in Backplane 1 of Integrated RAID Controller 1"},"0":{"bladeRaidVer":"25.5.2.0001","bladeRaidName":"PERC H730P Mini"},"3":{"bladeRaidVer":"2.25","bladeRaidName":"BP13G+ 0:1"},"2":{"bladeRaidVer":"TT31","bladeRaidName":"Disk 1 in Backplane 1 of Integrated RAID Controller 1"}},"bladeBootOnce":1,"bladeModel":"PowerEdge M630","bladeMinBudgetAlloc":127,"bladeFwVer":"2.41.40.40 (07)","bladeTempUpperNonCriticalThreshold":"42","bladeMaxConsumption":554,"bladeLEDState":16,"bladeFirstBootDevice":0,"bladeWidth":0,"bladeCPLDVer":"1.0.5","bladePowerStatus":1,"bladeTempLowerNonCriticalThreshold":"3","bladeHealth":3,"bladeFormFactor":114,"bladeTempLowerCriticalThreshold":"-7","bladeOS":"","bladeName":"localhost.localdomain","bladeSerialNum":"CN7016361N00JJ"},"2":{"bladeTemperature":"12","storageSelectedFabric":0,"bladeVKVMLicensed":1,"bladeSystemName":"devratefamilysearchproxy-1001.ams4.example.com","bladePresent":1,"idracURL":"https:\/\/10.193.251.10:443","bladeLogDescription":"No Errors","bladeIMCStatus":0,"bladeFwUpdatable":1,"bladeVKVMSupported":1,"nic":{"1":{"bladeNicName":"Intel(R) Ethernet 10G 2P X520-k bNDC - EC:F4:BB:E5:1F:C8","bladeNicVer":"18.0.17"},"0":{"bladeNicName":"Intel(R) Ethernet 10G 2P X520-k bNDC - EC:F4:BB:E5:1F:CA","bladeNicVer":"18.0.17"}},"storageNumDrives":0,"bladeIdracExternalManaged":0,"bladeLEDColor":0,"bladeBudgetAlloc":190,"storageNumControllers":0,"bladePriority":1,"bladePartNum":"0PHY8D","bladeDiagsVer":"4239A33","bladeMasterSlot":2,"pwrAccumulate":"446.0","bladeSlotName":"devratefamilysearchproxy-1001.ams4.example.com","bladeUSCVer":"2.41.40.40","WSMAN_Ready":1,"bladeNicEnable":1,"bladeCurrentConsumption":190,"pwrMaxTime":"Wed 13 Jul 2016 08:56:14 PM","bladeOsdrvVer":"15.07.07","bladeTotalMem":"128.0","bladeCpuInfo":"2 x Intel(R) Xeon(R) CPU E5-2620 v3 @ 2.40GHz","pwrStartTime":"Tue 29 Sep 2015 10:43:26 AM","isConfigured":0,"bladeSvcTag":"7467Z72","bladeLogSeverity":3,"pwrMaxConsump":362,"bladePwrCtlSupported":1,"bladeBIOSver":"2.4.2","actualPwrConsump":147,"bladeSlot":"2","bladeManufacturer":"70163","isStorageBlade":0,"bladeTempUpperCriticalThreshold":"47","raid":{"1":{"bladeRaidVer":"DM06","bladeRaidName":"Disk 0 in Backplane 1 of Integrated RAID Controller 1"},"0":{"bladeRaidVer":"25.5.3.0005","bladeRaidName":"PERC H730P Mini"},"3":{"bladeRaidVer":"2.25","bladeRaidName":"BP13G+ 0:1"},"2":{"bladeRaidVer":"DM06","bladeRaidName":"Disk 1 in Backplane 1 of Integrated RAID Controller 1"}},"bladeBootOnce":1,"bladeModel":"PowerEdge M630","bladeMinBudgetAlloc":120,"bladeFwVer":"2.41.40.40 (07)","bladeTempUpperNonCriticalThreshold":"42","bladeMaxConsumption":362,"bladeLEDState":16,"bladeFirstBootDevice":0,"bladeWidth":0,"bladeCPLDVer":"1.0.5","bladePowerStatus":1,"bladeTempLowerNonCriticalThreshold":"3","bladeHealth":3,"bladeFormFactor":114,"bladeTempLowerCriticalThreshold":"-7","bladeOS":"CentOS Linux","bladeName":"devratefamilysearchproxy-1001.ams4.example.com","bladeSerialNum":"CN7016358K00JA"},"5":{"bladeTemperature":-1,"bladeSvcTag":"N\/A","bladeMasterSlot":5,"bladePresent":0,"bladeSlotName":"SLOT-05","bladeSlot":"5","bladeHealth":-1,"idracURL":"","bladePowerStatus":-1,"bladeName":"SLOT-05","bladePriority":1},"4":{"bladeTemperature":-1,"bladeSvcTag":"N\/A","bladeMasterSlot":4,"bladePresent":0,"bladeSlotName":"SLOT-04","bladeSlot":"4","bladeHealth":-1,"idracURL":"","bladePowerStatus":-1,"bladeName":"SLOT-04","bladePriority":1},"7":{"bladeTemperature":-1,"bladeSvcTag":"N\/A","bladeMasterSlot":7,"bladePresent":0,"bladeSlotName":"SLOT-07","bladeSlot":"7","bladeHealth":-1,"idracURL":"","bladePowerStatus":-1,"bladeName":"SLOT-07","bladePriority":1},"6":{"bladeTemperature":-1,"bladeSvcTag":"N\/A","bladeMasterSlot":6,"bladePresent":0,"bladeSlotName":"SLOT-06","bladeSlot":"6","bladeHealth":-1,"idracURL":"","bladePowerStatus":-1,"bladeName":"SLOT-06","bladePriority":1},"14":{"bladeTemperature":-1,"bladeSvcTag":"N\/A","bladeMasterSlot":14,"bladePresent":0,"bladeSlotName":"SLOT-14","bladeSlot":"14","bladeHealth":-1,"idracURL":"","bladePowerStatus":-1,"bladeName":"SLOT-14","bladePriority":1},"8":{"bladeTemperature":-1,"bladeSvcTag":"N\/A","bladeMasterSlot":8,"bladePresent":0,"bladeSlotName":"SLOT-08","bladeSlot":"8","bladeHealth":-1,"idracURL":"","bladePowerStatus":-1,"bladeName":"SLOT-08","bladePriority":1},"16":{"bladeTemperature":-1,"bladeSvcTag":"N\/A","bladeMasterSlot":16,"bladePresent":0,"bladeSlotName":"SLOT-16","bladeSlot":"16","bladeHealth":-1,"idracURL":"","bladePowerStatus":-1,"bladeName":"SLOT-16","bladePriority":1},"9":{"bladeTemperature":-1,"bladeSvcTag":"N\/A","bladeMasterSlot":9,"bladePresent":0,"bladeSlotName":"SLOT-09","bladeSlot":"9","bladeHealth":-1,"idracURL":"","bladePowerStatus":-1,"bladeName":"SLOT-09","bladePriority":1},"13":{"bladeTemperature":-1,"bladeSvcTag":"N\/A","bladeMasterSlot":13,"bladePresent":0,"bladeSlotName":"SLOT-13","bladeSlot":"13","bladeHealth":-1,"idracURL":"","bladePowerStatus":-1,"bladeName":"SLOT-13","bladePriority":1},"12":{"bladeTemperature":-1,"bladeSvcTag":"N\/A","bladeMasterSlot":12,"bladePresent":0,"bladeSlotName":"SLOT-12","bladeSlot":"12","bladeHealth":-1,"idracURL":"","bladePowerStatus":-1,"bladeName":"SLOT-12","bladePriority":1},"11":{"bladeTemperature":-1,"bladeSvcTag":"N\/A","bladeMasterSlot":11,"bladePresent":0,"bladeSlotName":"SLOT-11","bladeSlot":"11","bladeHealth":-1,"idracURL":"","bladePowerStatus":-1,"bladeName":"SLOT-11","bladePriority":1},"10":{"bladeTemperature":-1,"bladeSvcTag":"N\/A","bladeMasterSlot":10,"bladePresent":0,"bladeSlotName":"SLOT-10","bladeSlot":"10","bladeHealth":-1,"idracURL":"","bladePowerStatus":-1,"bladeName":"SLOT-10","bladePriority":1}},"fans_status":{"1":{"FanPowerOff":2,"FanRPMTach2":2993,"FanOperationalStatus":2,"FanUpperNonCriticalThreshold":-1,"FanUpperCriticalThreshold":10390,"FanPresence":1,"FanActiveCooling":1,"FanRPMTach1":2873,"FanLowerNonCriticalThreshold":-1,"fanActiveErrorSev":3,"FanID":1,"fanActiveError":"No Errors","FanHealthState":5,"FanName":"Fan-1","FanMaximumWattage":57,"FanType":6,"FanEnabledState":2,"FanPowerOn":18,"FanLowerCriticalThreshold":1748,"FanPWMMax":10390},"3":{"FanPowerOff":2,"FanRPMTach2":4233,"FanOperationalStatus":2,"FanUpperNonCriticalThreshold":-1,"FanUpperCriticalThreshold":10390,"FanPresence":1,"FanActiveCooling":1,"FanRPMTach1":4103,"FanLowerNonCriticalThreshold":-1,"fanActiveErrorSev":3,"FanID":3,"fanActiveError":"No Errors","FanHealthState":5,"FanName":"Fan-3","FanMaximumWattage":57,"FanType":6,"FanEnabledState":2,"FanPowerOn":18,"FanLowerCriticalThreshold":2485,"FanPWMMax":10390},"2":{"FanPowerOff":2,"FanRPMTach2":4222,"FanOperationalStatus":2,"FanUpperNonCriticalThreshold":-1,"FanUpperCriticalThreshold":10390,"FanPresence":1,"FanActiveCooling":1,"FanRPMTach1":4089,"FanLowerNonCriticalThreshold":-1,"fanActiveErrorSev":3,"FanID":2,"fanActiveError":"No Errors","FanHealthState":5,"FanName":"Fan-2","FanMaximumWattage":57,"FanType":6,"FanEnabledState":2,"FanPowerOn":18,"FanLowerCriticalThreshold":2485,"FanPWMMax":10390},"5":{"FanPowerOff":2,"FanRPMTach2":4268,"FanOperationalStatus":2,"FanUpperNonCriticalThreshold":-1,"FanUpperCriticalThreshold":10390,"FanPresence":1,"FanActiveCooling":1,"FanRPMTach1":4076,"FanLowerNonCriticalThreshold":-1,"fanActiveErrorSev":3,"FanID":5,"fanActiveError":"No Errors","FanHealthState":5,"FanName":"Fan-5","FanMaximumWattage":57,"FanType":6,"FanEnabledState":2,"FanPowerOn":18,"FanLowerCriticalThreshold":2485,"FanPWMMax":10390},"4":{"FanPowerOff":2,"FanRPMTach2":3002,"FanOperationalStatus":2,"FanUpperNonCriticalThreshold":-1,"FanUpperCriticalThreshold":10390,"FanPresence":1,"FanActiveCooling":1,"FanRPMTach1":2871,"FanLowerNonCriticalThreshold":-1,"fanActiveErrorSev":3,"FanID":4,"fanActiveError":"No Errors","FanHealthState":5,"FanName":"Fan-4","FanMaximumWattage":57,"FanType":6,"FanEnabledState":2,"FanPowerOn":18,"FanLowerCriticalThreshold":1748,"FanPWMMax":10390},"7":{"FanPowerOff":2,"FanRPMTach2":2973,"FanOperationalStatus":2,"FanUpperNonCriticalThreshold":-1,"FanUpperCriticalThreshold":10390,"FanPresence":1,"FanActiveCooling":1,"FanRPMTach1":2860,"FanLowerNonCriticalThreshold":-1,"fanActiveErrorSev":3,"FanID":7,"fanActiveError":"No Errors","FanHealthState":5,"FanName":"Fan-7","FanMaximumWattage":57,"FanType":6,"FanEnabledState":2,"FanPowerOn":18,"FanLowerCriticalThreshold":1748,"FanPWMMax":10390},"6":{"FanPowerOff":2,"FanRPMTach2":4204,"FanOperationalStatus":2,"FanUpperNonCriticalThreshold":-1,"FanUpperCriticalThreshold":10390,"FanPresence":1,"FanActiveCooling":1,"FanRPMTach1":4078,"FanLowerNonCriticalThreshold":-1,"fanActiveErrorSev":3,"FanID":6,"fanActiveError":"No Errors","FanHealthState":5,"FanName":"Fan-6","FanMaximumWattage":57,"FanType":6,"FanEnabledState":2,"FanPowerOn":18,"FanLowerCriticalThreshold":2485,"FanPWMMax":10390},"9":{"FanPowerOff":2,"FanRPMTach2":4215,"FanOperationalStatus":2,"FanUpperNonCriticalThreshold":-1,"FanUpperCriticalThreshold":10390,"FanPresence":1,"FanActiveCooling":1,"FanRPMTach1":4086,"FanLowerNonCriticalThreshold":-1,"fanActiveErrorSev":3,"FanID":9,"fanActiveError":"No Errors","FanHealthState":5,"FanName":"Fan-9","FanMaximumWattage":57,"FanType":6,"FanEnabledState":2,"FanPowerOn":18,"FanLowerCriticalThreshold":2485,"FanPWMMax":10390},"8":{"FanPowerOff":2,"FanRPMTach2":4226,"FanOperationalStatus":2,"FanUpperNonCriticalThreshold":-1,"FanUpperCriticalThreshold":10390,"FanPresence":1,"FanActiveCooling":1,"FanRPMTach1":4068,"FanLowerNonCriticalThreshold":-1,"fanActiveErrorSev":3,"FanID":8,"fanActiveError":"No Errors","FanHealthState":5,"FanName":"Fan-8","FanMaximumWattage":57,"FanType":6,"FanEnabledState":2,"FanPowerOn":18,"FanLowerCriticalThreshold":2485,"FanPWMMax":10390},"ECM":{"ECMConfigurable":0,"THERMAL_ecm":"0","ChassisECMStatusValue":9,"MixedModeFanInfo":""},"10":{"FanPowerOff":0,"FanRPMTach2":0,"FanOperationalStatus":0,"FanUpperNonCriticalThreshold":0,"FanUpperCriticalThreshold":0,"FanPresence":0,"FanActiveCooling":0,"FanRPMTach1":0,"FanLowerNonCriticalThreshold":0,"fanActiveErrorSev":3,"FanID":0,"fanActiveError":"No Errors","FanHealthState":0,"FanName":"","FanMaximumWattage":0,"FanType":0,"FanEnabledState":0,"FanPowerOn":0,"FanLowerCriticalThreshold":0,"FanPWMMax":0}},"cmc_status":{"CMC_Active_Slot":1,"CMC_Standby_Present":1,"CMC_Standby_Version":"6.00","cmcStbyActiveErrorSev":3,"cmcStbyActiveError":"No Errors","cmcActiveError":"No Errors","cmcActiveErrorSev":3,"NETWORK_NIC_ipv4_enable":"1","CurrentIPv6Address1":"","CurrentIPAddress":"10.193.251.36","NETWORK_NIC_IPV6_enable":"1","CMC_Local_State":1,"CMC_Redundancy_Mode":1},"ioms_status":{"1":{"iomMasterLocation":"A1","iomPoweredOn":1,"iomManageable":0,"iomBypassModeIsEnabled":0,"iomManufacturer":"DELL","iomMaxBootTime":20,"iomLinkStatus":{"53":0,"43":0,"51":0,"41":0,"47":0,"37":0,"45":0,"35":0,"49":0,"39":0,"29":0,"1":0,"3":0,"2":0,"5":0,"4":0,"7":0,"6":0,"9":0,"8":0,"27":0,"17":0,"13":0,"21":0,"11":0,"23":0,"42":0,"52":0,"40":0,"50":0,"36":0,"46":0,"34":0,"44":0,"48":0,"28":0,"38":0,"56":0,"55":0,"54":0,"33":0,"32":0,"31":0,"30":0,"15":0,"18":0,"19":0,"25":0,"14":0,"24":0,"16":0,"26":0,"20":0,"12":0,"22":0,"10":0},"iomHwVer":"A04","iomFlexIOModStatus":1,"iomThermTripStatus":0,"iomActiveErrorSev":3,"iomFlexIOMod2Type":0,"iomSubnetMask":"0.0.0.0","iomGateway":"0.0.0.0","iomPort":"","iomIpAddress":"0.0.0.0","iomBlankPresent":0,"iomBootStatus":1,"iomFlexIOMod1Type":0,"iomMode":0,"isF10Switch":0,"iomAggType":0,"iomPsocVer":"12","iomLedState":0,"iomPresent":1,"iomActiveError":"No Errors","iomCmdRevision":0,"iomMacAddress":"00:00:00:00:00:00","iomGui":0,"iomMaxPowerNeeded":73,"iomLocation":"A1","iomOpInProgress":0,"iomPtNum":"0PNDP6","iomStackRole":2,"iomFabricType":16,"iomDhcpEnabled":0,"iomHealth":3,"iomName":"Dell 10GbE KR PTM             ","iomSvcTag":"0000000","iomSwManage":0,"iomTempEnum":2,"iomSlaveStatus":0},"iom_bypass_mode":{"bypass_enabled_ioms":"","num_bypass_enabled_ioms":0},"3":{"iomMasterLocation":"B1","iomPoweredOn":-1,"iomManageable":0,"iomBypassModeIsEnabled":0,"iomManufacturer":"","iomMaxBootTime":1,"iomLinkStatus":{"53":0,"43":0,"51":0,"41":0,"47":0,"37":0,"45":0,"35":0,"49":0,"39":0,"29":0,"1":0,"3":0,"2":0,"5":0,"4":0,"7":0,"6":0,"9":0,"8":0,"27":0,"17":0,"13":0,"21":0,"11":0,"23":0,"42":0,"52":0,"40":0,"50":0,"36":0,"46":0,"34":0,"44":0,"48":0,"28":0,"38":0,"56":0,"55":0,"54":0,"33":0,"32":0,"31":0,"30":0,"15":0,"18":0,"19":0,"25":0,"14":0,"24":0,"16":0,"26":0,"20":0,"12":0,"22":0,"10":0},"iomHwVer":"","iomFlexIOModStatus":1,"iomThermTripStatus":0,"iomActiveErrorSev":3,"iomFlexIOMod2Type":0,"iomSubnetMask":"0.0.0.0","iomGateway":"0.0.0.0","iomPort":"","iomIpAddress":"0.0.0.0","iomBlankPresent":0,"iomBootStatus":0,"iomFlexIOMod1Type":0,"iomMode":0,"isF10Switch":0,"iomAggType":0,"iomPsocVer":"","iomLedState":0,"iomPresent":0,"iomActiveError":"No Errors","iomCmdRevision":0,"iomMacAddress":"00:00:00:00:00:00","iomGui":0,"iomMaxPowerNeeded":0,"iomLocation":"B1","iomOpInProgress":0,"iomPtNum":"","iomStackRole":2,"iomFabricType":0,"iomDhcpEnabled":0,"iomHealth":3,"iomName":"","iomSvcTag":"","iomSwManage":0,"iomTempEnum":0,"iomSlaveStatus":0},"2":{"iomMasterLocation":"A2","iomPoweredOn":-1,"iomManageable":0,"iomBypassModeIsEnabled":0,"iomManufacturer":"","iomMaxBootTime":1,"iomLinkStatus":{"53":0,"43":0,"51":0,"41":0,"47":0,"37":0,"45":0,"35":0,"49":0,"39":0,"29":0,"1":0,"3":0,"2":0,"5":0,"4":0,"7":0,"6":0,"9":0,"8":0,"27":0,"17":0,"13":0,"21":0,"11":0,"23":0,"42":0,"52":0,"40":0,"50":0,"36":0,"46":0,"34":0,"44":0,"48":0,"28":0,"38":0,"56":0,"55":0,"54":0,"33":0,"32":0,"31":0,"30":0,"15":0,"18":0,"19":0,"25":0,"14":0,"24":0,"16":0,"26":0,"20":0,"12":0,"22":0,"10":0},"iomHwVer":"","iomFlexIOModStatus":1,"iomThermTripStatus":0,"iomActiveErrorSev":3,"iomFlexIOMod2Type":0,"iomSubnetMask":"0.0.0.0","iomGateway":"0.0.0.0","iomPort":"","iomIpAddress":"0.0.0.0","iomBlankPresent":0,"iomBootStatus":0,"iomFlexIOMod1Type":0,"iomMode":0,"isF10Switch":0,"iomAggType":0,"iomPsocVer":"","iomLedState":0,"iomPresent":0,"iomActiveError":"No Errors","iomCmdRevision":0,"iomMacAddress":"00:00:00:00:00:00","iomGui":0,"iomMaxPowerNeeded":0,"iomLocation":"A2","iomOpInProgress":0,"iomPtNum":"","iomStackRole":2,"iomFabricType":0,"iomDhcpEnabled":0,"iomHealth":3,"iomName":"","iomSvcTag":"","iomSwManage":0,"iomTempEnum":0,"iomSlaveStatus":0},"5":{"iomMasterLocation":"C1","iomPoweredOn":-1,"iomManageable":0,"iomBypassModeIsEnabled":0,"iomManufacturer":"","iomMaxBootTime":1,"iomLinkStatus":{"53":0,"43":0,"51":0,"41":0,"47":0,"37":0,"45":0,"35":0,"49":0,"39":0,"29":0,"1":0,"3":0,"2":0,"5":0,"4":0,"7":0,"6":0,"9":0,"8":0,"27":0,"17":0,"13":0,"21":0,"11":0,"23":0,"42":0,"52":0,"40":0,"50":0,"36":0,"46":0,"34":0,"44":0,"48":0,"28":0,"38":0,"56":0,"55":0,"54":0,"33":0,"32":0,"31":0,"30":0,"15":0,"18":0,"19":0,"25":0,"14":0,"24":0,"16":0,"26":0,"20":0,"12":0,"22":0,"10":0},"iomHwVer":"","iomFlexIOModStatus":1,"iomThermTripStatus":0,"iomActiveErrorSev":3,"iomFlexIOMod2Type":0,"iomSubnetMask":"0.0.0.0","iomGateway":"0.0.0.0","iomPort":"","iomIpAddress":"0.0.0.0","iomBlankPresent":0,"iomBootStatus":0,"iomFlexIOMod1Type":0,"iomMode":0,"isF10Switch":0,"iomAggType":0,"iomPsocVer":"","iomLedState":0,"iomPresent":0,"iomActiveError":"No Errors","iomCmdRevision":0,"iomMacAddress":"00:00:00:00:00:00","iomGui":0,"iomMaxPowerNeeded":0,"iomLocation":"C1","iomOpInProgress":0,"iomPtNum":"","iomStackRole":2,"iomFabricType":0,"iomDhcpEnabled":0,"iomHealth":3,"iomName":"","iomSvcTag":"","iomSwManage":0,"iomTempEnum":0,"iomSlaveStatus":0},"MAX_IOMS":6,"4":{"iomMasterLocation":"B2","iomPoweredOn":-1,"iomManageable":0,"iomBypassModeIsEnabled":0,"iomManufacturer":"","iomMaxBootTime":1,"iomLinkStatus":{"53":0,"43":0,"51":0,"41":0,"47":0,"37":0,"45":0,"35":0,"49":0,"39":0,"29":0,"1":0,"3":0,"2":0,"5":0,"4":0,"7":0,"6":0,"9":0,"8":0,"27":0,"17":0,"13":0,"21":0,"11":0,"23":0,"42":0,"52":0,"40":0,"50":0,"36":0,"46":0,"34":0,"44":0,"48":0,"28":0,"38":0,"56":0,"55":0,"54":0,"33":0,"32":0,"31":0,"30":0,"15":0,"18":0,"19":0,"25":0,"14":0,"24":0,"16":0,"26":0,"20":0,"12":0,"22":0,"10":0},"iomHwVer":"","iomFlexIOModStatus":1,"iomThermTripStatus":0,"iomActiveErrorSev":3,"iomFlexIOMod2Type":0,"iomSubnetMask":"0.0.0.0","iomGateway":"0.0.0.0","iomPort":"","iomIpAddress":"0.0.0.0","iomBlankPresent":0,"iomBootStatus":0,"iomFlexIOMod1Type":0,"iomMode":0,"isF10Switch":0,"iomAggType":0,"iomPsocVer":"","iomLedState":0,"iomPresent":0,"iomActiveError":"No Errors","iomCmdRevision":0,"iomMacAddress":"00:00:00:00:00:00","iomGui":0,"iomMaxPowerNeeded":0,"iomLocation":"B2","iomOpInProgress":0,"iomPtNum":"","iomStackRole":2,"iomFabricType":0,"iomDhcpEnabled":0,"iomHealth":3,"iomName":"","iomSvcTag":"","iomSwManage":0,"iomTempEnum":0,"iomSlaveStatus":0},"6":{"iomMasterLocation":"C2","iomPoweredOn":-1,"iomManageable":0,"iomBypassModeIsEnabled":0,"iomManufacturer":"","iomMaxBootTime":1,"iomLinkStatus":{"53":0,"43":0,"51":0,"41":0,"47":0,"37":0,"45":0,"35":0,"49":0,"39":0,"29":0,"1":0,"3":0,"2":0,"5":0,"4":0,"7":0,"6":0,"9":0,"8":0,"27":0,"17":0,"13":0,"21":0,"11":0,"23":0,"42":0,"52":0,"40":0,"50":0,"36":0,"46":0,"34":0,"44":0,"48":0,"28":0,"38":0,"56":0,"55":0,"54":0,"33":0,"32":0,"31":0,"30":0,"15":0,"18":0,"19":0,"25":0,"14":0,"24":0,"16":0,"26":0,"20":0,"12":0,"22":0,"10":0},"iomHwVer":"","iomFlexIOModStatus":1,"iomThermTripStatus":0,"iomActiveErrorSev":3,"iomFlexIOMod2Type":0,"iomSubnetMask":"0.0.0.0","iomGateway":"0.0.0.0","iomPort":"","iomIpAddress":"0.0.0.0","iomBlankPresent":0,"iomBootStatus":0,"iomFlexIOMod1Type":0,"iomMode":0,"isF10Switch":0,"iomAggType":0,"iomPsocVer":"","iomLedState":0,"iomPresent":0,"iomActiveError":"No Errors","iomCmdRevision":0,"iomMacAddress":"00:00:00:00:00:00","iomGui":0,"iomMaxPowerNeeded":0,"iomLocation":"C2","iomOpInProgress":0,"iomPtNum":"","iomStackRole":2,"iomFabricType":0,"iomDhcpEnabled":0,"iomHealth":3,"iomName":"","iomSvcTag":"","iomSwManage":0,"iomTempEnum":0,"iomSlaveStatus":0}},"ikvm_status":{"lkvmHwVer":"A03","lkvmActiveErrorSev":3,"lkvmName":"Avocent iKVM Switch","lkvmFrntPanelEnabled":1,"lkvmFwStatus":11,"lkvmBlankPresent":0,"lkvmPtNum":"0K036D","lkvmAciPortTiered":0,"lkvmMaxPowerNeeded":15,"lkvmSvcTag":"","lkvmActiveError":"No Errors","lkvmMaxBootTime":1,"lkvmRearPanelConnected":0,"lkvmFwVer":"01.00.01.01","lkvmManufacturer":"DELL","lkvmPoweredOn":1,"lkvmKvmTelnetEnabled":1,"lkvmHealth":3,"lkvmBoardStatus":0,"lkvmFrntPanelConnected":0,"lkvmPresent":1},"chassis_status":{"CHASSIS_fresh_air":"0","FIPS_Mode":0,"CHASSIS_name":"CMC-51F3DK2","CHASSIS_asset_tag":"00000","RO_cmc_fw_version_string":"6.00","RO_chassis_productname":"PowerEdge M1000e","RO_chassis_service_tag":"51F3DK2"},"psu_status":{"ChassisEPPFault":3145779,"acPowerHi_btuhr":"16692 BTU\/h","ChassisEPPEngaged":0,"psuCount":4,"CHASSIS_POWER_epp_enable":0,"acPowerCapacity":"5944 W","RedundancyReserve":"5944 W","CHASSIS_POWER_SBPMMode":0,"acPowerPotential_btuhr":"5275 BTU\/h","acPowerWarn":"0 W","acEnergyStartTime":"20:36:52 06\/05\/2017","psu_3":{"psuSvctag":"","psuCapacity":0,"psuPartNum":"","psuPresent":0,"psuActiveErrorSev":3,"psuActiveError":"No Errors","psuHealth":10,"psuAcCurrent":"0.0 N\/A","psuState":1,"psuAcVolts":"0.0"},"acPowerHiStartTime":"16:27:22 10\/18\/2017","chassisPowerState":6,"ChassisEPPPercentAvailable":0,"ServerAllocation":"735 W","acEnergy":"620.4 kWh","dcPowerCapacity":"5068 W","acPowerSurplus_btuhr":"51656 BTU\/h","CHASSIS_POWER_UPSMode":0,"psuMax":6,"psu_6":{"psuSvctag":"","psuCapacity":2700,"psuPartNum":"0TJJ3M","psuPresent":1,"psuActiveErrorSev":3,"psuActiveError":"No Errors","psuHealth":0,"psuAcCurrent":"1.1 A","psuState":3,"psuAcVolts":"229.8"},"acPowerCurrentReading":"3.7 A","ChassisEPPUsed_btuhr":"0 BTU\/h","AvailablePower":"4768 W","psuDynEng":0,"psu_4":{"psuSvctag":"","psuCapacity":0,"psuPartNum":"","psuPresent":0,"psuActiveErrorSev":3,"psuActiveError":"No Errors","psuHealth":10,"psuAcCurrent":"0.0 N\/A","psuState":1,"psuAcVolts":"0.0"},"cmcPowerHealth":3,"ChassisEPPAvailable_btuhr":"0 BTU\/h","CHASSIS_type_nebs":0,"acPower_btuhr":"2129 BTU\/h","acPowerIdle":"624 W","psu_5":{"psuSvctag":"","psuCapacity":2700,"psuPartNum":"0TJJ3M","psuPresent":1,"psuActiveErrorSev":3,"psuActiveError":"No Errors","psuHealth":0,"psuAcCurrent":"0.9 A","psuState":3,"psuAcVolts":"230.8"},"psuRedundancy":1,"acPowerLoStartTime":"16:27:22 10\/18\/2017","acPowerSurplus":"15139 W","ChassisEPPAvailable":"0 W","ChassisEPPStatusValue":0,"acPower":"624 W","acPowerPotential":"1546 W","acPowerBudget":"16685 W","CHASSIS_POWER_budget_percent_ac":"100 %","psu_2":{"psuSvctag":"","psuCapacity":2700,"psuPartNum":"0TJJ3M","psuPresent":1,"psuActiveErrorSev":3,"psuActiveError":"No Errors","psuHealth":0,"psuAcCurrent":"0.9 A","psuState":3,"psuAcVolts":"231.8"},"acEnergyTime":"20:43:06 11\/02\/2017","RemainingPower":"0 W","BaseConsumption":"279 W","psuMask":51,"LoadSharingLoss":"0 W","acPowerIdle_btuhr":"2129 BTU\/h","psuRedundancyState":1,"ChassisEPPPercentUsed":0,"WaterMarkClearedTime":"1508362042 W","ChassisEPPUsed":"0 W","CHASSIS_POWER_budget_btuhr_ac":"56931 BTU\/h","acPowerLoTime":"19:03:38 11\/02\/2017","acPowerHiTime":"19:04:00 11\/02\/2017","CHASSIS_POWER_warning_btuhr_ac":"0 BTU\/h","acPowerLo":"164 W","acPowerLo_btuhr":"559 BTU\/h","acPowerHi":"4892 W","psu_1":{"psuSvctag":"","psuCapacity":2700,"psuPartNum":"0TJJ3M","psuPresent":1,"psuActiveErrorSev":3,"psuActiveError":"No Errors","psuHealth":0,"psuAcCurrent":"0.8 A","psuState":3,"psuAcVolts":"230.0"},"CHASSIS_type_freshair":0},"active_alerts":{"LCD":{"1":{"noncriticalCount":0,"infoCount":0,"criticalCount":0}},"iKVM":{"1":{"noncriticalCount":0,"infoCount":0,"criticalCount":0}},"IOM":{"1":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"3":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"2":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"5":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"4":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"6":{"noncriticalCount":0,"infoCount":0,"criticalCount":0}},"PSU":{"1":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"3":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"2":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"5":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"4":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"6":{"noncriticalCount":0,"infoCount":0,"criticalCount":0}},"Fan":{"1":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"3":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"2":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"5":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"4":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"7":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"6":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"9":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"8":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"10":{"noncriticalCount":0,"infoCount":0,"criticalCount":0}},"CMC":{"1":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"2":{"noncriticalCount":0,"infoCount":0,"criticalCount":0}},"Chassis":{"1":{"noncriticalCount":0,"infoCount":0,"criticalCount":0}},"Server":{"43":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"41":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"47":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"37":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"45":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"35":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"39":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"29":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"1":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"3":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"2":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"5":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"4":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"7":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"6":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"9":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"8":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"27":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"17":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"13":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"21":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"11":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"23":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"42":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"40":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"36":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"46":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"34":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"44":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"48":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"28":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"38":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"33":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"32":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"31":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"30":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"15":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"18":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"19":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"25":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"14":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"24":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"16":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"26":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"20":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"12":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"22":{"noncriticalCount":0,"infoCount":0,"criticalCount":0},"10":{"noncriticalCount":0,"infoCount":0,"criticalCount":0}}}},"ChassisGroupMemberIp":"cmc-51F3DK2","ChassisGroupMemberLastErrorStr":"No Error","ChassisGroupMemberState":0,"ChassisGroupMemberId":0,"ChassisGroupMemberUpdateTime":"","ChassisGroupMemberLastError":"0x0000","ChassisGroupMemberStateStr":"No Error"}}`),
			"temp-sensors":    []byte(`{"1":{"TempHealth":5,"TempUpperCriticalThreshold":40,"TempSensorID":1,"TempCurrentValue":17,"TempLowerCriticalThreshold":-1,"TempPresence":1,"TempSensorName":"Ambient_Temp"},"blades_status":{"SlotName_host":{"BLADESLOT_NAME_usehostname":1},"1":{"bladeTemperature":"12","storageSelectedFabric":0,"bladeVKVMLicensed":1,"bladeSystemName":"provision-test-13163.ams4.example.com","bladePresent":1,"idracURL":"https:\/\/10.193.251.5:443","bladeLogDescription":"No Errors","bladeIMCStatus":0,"bladeFwUpdatable":1,"bladeVKVMSupported":1,"nic":{"1":{"bladeNicName":"Intel(R) Ethernet 10G 2P X520-k bNDC - EC:F4:BB:EB:A2:48","bladeNicVer":"18.0.17"},"0":{"bladeNicName":"Intel(R) Ethernet 10G 2P X520-k bNDC - EC:F4:BB:EB:A2:4A","bladeNicVer":"18.0.17"}},"storageNumDrives":0,"bladeIdracExternalManaged":0,"bladeLEDColor":0,"bladeBudgetAlloc":190,"storageNumControllers":0,"bladePriority":1,"bladePartNum":"0PHY8D","bladeDiagsVer":"4239A33","bladeMasterSlot":1,"pwrAccumulate":"1777.4","bladeSlotName":"provision-test-13163.ams4.example.com","bladeUSCVer":"2.41.40.40","WSMAN_Ready":1,"bladeNicEnable":1,"bladeCurrentConsumption":190,"pwrMaxTime":"Mon 21 Dec 2015 06:08:02 AM","bladeOsdrvVer":"15.10.02","bladeTotalMem":"128.0","bladeCpuInfo":"2 x Intel(R) Xeon(R) CPU E5-2620 v3 @ 2.40GHz","pwrStartTime":"Mon 21 Dec 2015 04:11:30 AM","isConfigured":0,"bladeSvcTag":"4NY1H92","bladeLogSeverity":3,"pwrMaxConsump":354,"bladePwrCtlSupported":1,"bladeBIOSver":"2.4.2","actualPwrConsump":143,"bladeSlot":"1","bladeManufacturer":"70163","isStorageBlade":0,"bladeTempUpperCriticalThreshold":"47","raid":{"1":{"bladeRaidVer":"D416","bladeRaidName":"Disk 0 in Backplane 1 of Integrated RAID Controller 1"},"0":{"bladeRaidVer":"25.5.3.0005","bladeRaidName":"PERC H730P Mini"},"2":{"bladeRaidVer":"2.25","bladeRaidName":"BP13G+ 0:1"}},"bladeBootOnce":1,"bladeModel":"PowerEdge M630","bladeMinBudgetAlloc":120,"bladeFwVer":"2.41.40.40 (07)","bladeTempUpperNonCriticalThreshold":"42","bladeMaxConsumption":354,"bladeLEDState":16,"bladeFirstBootDevice":0,"bladeWidth":0,"bladeCPLDVer":"1.0.5","bladePowerStatus":1,"bladeTempLowerNonCriticalThreshold":"3","bladeHealth":3,"bladeFormFactor":114,"bladeTempLowerCriticalThreshold":"-7","bladeOS":"","bladeName":"provision-test-13163.ams4.example.com","bladeSerialNum":"CN701635BS003Q"},"15":{"bladeTemperature":-1,"bladeSvcTag":"N\/A","bladeMasterSlot":15,"bladePresent":0,"bladeSlotName":"SLOT-15","bladeSlot":"15","bladeHealth":-1,"idracURL":"","bladePowerStatus":-1,"bladeName":"SLOT-15","bladePriority":1},"3":{"bladeTemperature":"13","storageSelectedFabric":0,"bladeVKVMLicensed":1,"bladeSystemName":"localhost.localdomain","bladePresent":1,"idracURL":"https:\/\/10.193.251.17:443","bladeLogDescription":"No Errors","bladeIMCStatus":0,"bladeFwUpdatable":1,"bladeVKVMSupported":1,"nic":{"1":{"bladeNicName":"Intel(R) Ethernet 10G 2P X520-k bNDC - 24:6E:96:0D:42:8E","bladeNicVer":"17.5.10"},"0":{"bladeNicName":"Intel(R) Ethernet 10G 2P X520-k bNDC - 24:6E:96:0D:42:8C","bladeNicVer":"17.5.10"}},"storageNumDrives":0,"bladeIdracExternalManaged":0,"bladeLEDColor":0,"bladeBudgetAlloc":373,"storageNumControllers":0,"bladePriority":1,"bladePartNum":"0PHY8D","bladeDiagsVer":"4239A33","bladeMasterSlot":3,"pwrAccumulate":"1799.3","bladeSlotName":"localhost.localdomain","bladeUSCVer":"2.41.40.40","WSMAN_Ready":1,"bladeNicEnable":1,"bladeCurrentConsumption":373,"pwrMaxTime":"Tue 04 Apr 2017 01:47:16 PM","bladeOsdrvVer":"15.10.02","bladeTotalMem":"128.0","bladeCpuInfo":"2 x Intel(R) Xeon(R) CPU E5-2690 v3 @ 2.60GHz","pwrStartTime":"Thu 25 Feb 2016 03:39:09 PM","isConfigured":0,"bladeSvcTag":"FQG0VB2","bladeLogSeverity":3,"pwrMaxConsump":554,"bladePwrCtlSupported":1,"bladeBIOSver":"2.4.2","actualPwrConsump":224,"bladeSlot":"3","bladeManufacturer":"70163","isStorageBlade":0,"bladeTempUpperCriticalThreshold":"47","raid":{"1":{"bladeRaidVer":"TT31","bladeRaidName":"Disk 0 in Backplane 1 of Integrated RAID Controller 1"},"0":{"bladeRaidVer":"25.5.2.0001","bladeRaidName":"PERC H730P Mini"},"3":{"bladeRaidVer":"2.25","bladeRaidName":"BP13G+ 0:1"},"2":{"bladeRaidVer":"TT31","bladeRaidName":"Disk 1 in Backplane 1 of Integrated RAID Controller 1"}},"bladeBootOnce":1,"bladeModel":"PowerEdge M630","bladeMinBudgetAlloc":127,"bladeFwVer":"2.41.40.40 (07)","bladeTempUpperNonCriticalThreshold":"42","bladeMaxConsumption":554,"bladeLEDState":16,"bladeFirstBootDevice":0,"bladeWidth":0,"bladeCPLDVer":"1.0.5","bladePowerStatus":1,"bladeTempLowerNonCriticalThreshold":"3","bladeHealth":3,"bladeFormFactor":114,"bladeTempLowerCriticalThreshold":"-7","bladeOS":"","bladeName":"localhost.localdomain","bladeSerialNum":"CN7016361N00JJ"},"2":{"bladeTemperature":"12","storageSelectedFabric":0,"bladeVKVMLicensed":1,"bladeSystemName":"devratefamilysearchproxy-1001.ams4.example.com","bladePresent":1,"idracURL":"https:\/\/10.193.251.10:443","bladeLogDescription":"No Errors","bladeIMCStatus":0,"bladeFwUpdatable":1,"bladeVKVMSupported":1,"nic":{"1":{"bladeNicName":"Intel(R) Ethernet 10G 2P X520-k bNDC - EC:F4:BB:E5:1F:C8","bladeNicVer":"18.0.17"},"0":{"bladeNicName":"Intel(R) Ethernet 10G 2P X520-k bNDC - EC:F4:BB:E5:1F:CA","bladeNicVer":"18.0.17"}},"storageNumDrives":0,"bladeIdracExternalManaged":0,"bladeLEDColor":0,"bladeBudgetAlloc":190,"storageNumControllers":0,"bladePriority":1,"bladePartNum":"0PHY8D","bladeDiagsVer":"4239A33","bladeMasterSlot":2,"pwrAccumulate":"446.0","bladeSlotName":"devratefamilysearchproxy-1001.ams4.example.com","bladeUSCVer":"2.41.40.40","WSMAN_Ready":1,"bladeNicEnable":1,"bladeCurrentConsumption":190,"pwrMaxTime":"Wed 13 Jul 2016 08:56:14 PM","bladeOsdrvVer":"15.07.07","bladeTotalMem":"128.0","bladeCpuInfo":"2 x Intel(R) Xeon(R) CPU E5-2620 v3 @ 2.40GHz","pwrStartTime":"Tue 29 Sep 2015 10:43:26 AM","isConfigured":0,"bladeSvcTag":"7467Z72","bladeLogSeverity":3,"pwrMaxConsump":362,"bladePwrCtlSupported":1,"bladeBIOSver":"2.4.2","actualPwrConsump":147,"bladeSlot":"2","bladeManufacturer":"70163","isStorageBlade":0,"bladeTempUpperCriticalThreshold":"47","raid":{"1":{"bladeRaidVer":"DM06","bladeRaidName":"Disk 0 in Backplane 1 of Integrated RAID Controller 1"},"0":{"bladeRaidVer":"25.5.3.0005","bladeRaidName":"PERC H730P Mini"},"3":{"bladeRaidVer":"2.25","bladeRaidName":"BP13G+ 0:1"},"2":{"bladeRaidVer":"DM06","bladeRaidName":"Disk 1 in Backplane 1 of Integrated RAID Controller 1"}},"bladeBootOnce":1,"bladeModel":"PowerEdge M630","bladeMinBudgetAlloc":120,"bladeFwVer":"2.41.40.40 (07)","bladeTempUpperNonCriticalThreshold":"42","bladeMaxConsumption":362,"bladeLEDState":16,"bladeFirstBootDevice":0,"bladeWidth":0,"bladeCPLDVer":"1.0.5","bladePowerStatus":1,"bladeTempLowerNonCriticalThreshold":"3","bladeHealth":3,"bladeFormFactor":114,"bladeTempLowerCriticalThreshold":"-7","bladeOS":"CentOS Linux","bladeName":"devratefamilysearchproxy-1001.ams4.example.com","bladeSerialNum":"CN7016358K00JA"},"5":{"bladeTemperature":-1,"bladeSvcTag":"N\/A","bladeMasterSlot":5,"bladePresent":0,"bladeSlotName":"SLOT-05","bladeSlot":"5","bladeHealth":-1,"idracURL":"","bladePowerStatus":-1,"bladeName":"SLOT-05","bladePriority":1},"4":{"bladeTemperature":-1,"bladeSvcTag":"N\/A","bladeMasterSlot":4,"bladePresent":0,"bladeSlotName":"SLOT-04","bladeSlot":"4","bladeHealth":-1,"idracURL":"","bladePowerStatus":-1,"bladeName":"SLOT-04","bladePriority":1},"7":{"bladeTemperature":-1,"bladeSvcTag":"N\/A","bladeMasterSlot":7,"bladePresent":0,"bladeSlotName":"SLOT-07","bladeSlot":"7","bladeHealth":-1,"idracURL":"","bladePowerStatus":-1,"bladeName":"SLOT-07","bladePriority":1},"6":{"bladeTemperature":-1,"bladeSvcTag":"N\/A","bladeMasterSlot":6,"bladePresent":0,"bladeSlotName":"SLOT-06","bladeSlot":"6","bladeHealth":-1,"idracURL":"","bladePowerStatus":-1,"bladeName":"SLOT-06","bladePriority":1},"14":{"bladeTemperature":-1,"bladeSvcTag":"N\/A","bladeMasterSlot":14,"bladePresent":0,"bladeSlotName":"SLOT-14","bladeSlot":"14","bladeHealth":-1,"idracURL":"","bladePowerStatus":-1,"bladeName":"SLOT-14","bladePriority":1},"8":{"bladeTemperature":-1,"bladeSvcTag":"N\/A","bladeMasterSlot":8,"bladePresent":0,"bladeSlotName":"SLOT-08","bladeSlot":"8","bladeHealth":-1,"idracURL":"","bladePowerStatus":-1,"bladeName":"SLOT-08","bladePriority":1},"16":{"bladeTemperature":-1,"bladeSvcTag":"N\/A","bladeMasterSlot":16,"bladePresent":0,"bladeSlotName":"SLOT-16","bladeSlot":"16","bladeHealth":-1,"idracURL":"","bladePowerStatus":-1,"bladeName":"SLOT-16","bladePriority":1},"9":{"bladeTemperature":-1,"bladeSvcTag":"N\/A","bladeMasterSlot":9,"bladePresent":0,"bladeSlotName":"SLOT-09","bladeSlot":"9","bladeHealth":-1,"idracURL":"","bladePowerStatus":-1,"bladeName":"SLOT-09","bladePriority":1},"13":{"bladeTemperature":-1,"bladeSvcTag":"N\/A","bladeMasterSlot":13,"bladePresent":0,"bladeSlotName":"SLOT-13","bladeSlot":"13","bladeHealth":-1,"idracURL":"","bladePowerStatus":-1,"bladeName":"SLOT-13","bladePriority":1},"12":{"bladeTemperature":-1,"bladeSvcTag":"N\/A","bladeMasterSlot":12,"bladePresent":0,"bladeSlotName":"SLOT-12","bladeSlot":"12","bladeHealth":-1,"idracURL":"","bladePowerStatus":-1,"bladeName":"SLOT-12","bladePriority":1},"11":{"bladeTemperature":-1,"bladeSvcTag":"N\/A","bladeMasterSlot":11,"bladePresent":0,"bladeSlotName":"SLOT-11","bladeSlot":"11","bladeHealth":-1,"idracURL":"","bladePowerStatus":-1,"bladeName":"SLOT-11","bladePriority":1},"10":{"bladeTemperature":-1,"bladeSvcTag":"N\/A","bladeMasterSlot":10,"bladePresent":0,"bladeSlotName":"SLOT-10","bladeSlot":"10","bladeHealth":-1,"idracURL":"","bladePowerStatus":-1,"bladeName":"SLOT-10","bladePriority":1}},"SensorCount":1}`),
			"blades-wwn-info": []byte(`{"cmc_privileges":{"cfg":1,"serveradmin":1},"slot_mac_wwn":{"slot_mac_wwn_list":{"pwwnSlotSelected9":0,"pwwnSlotSelected8":0,"14":{"bladeSlotName":"SLOT-14","is_full_height":0,"BM_EXT_DC_INFO":{"A":{"FabricLocation":"A","ShowA1A2":0,"isSelected":0,"isInstalled":0,"A1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:AA","portVMAC":"40:5C:FD:BC:E6:AA","portMacType":16},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:AB","portVMAC":"40:5C:FD:BC:E6:AB","portMacType":16}},"dcModelName":"","A2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:AD","portVMAC":"40:5C:FD:BC:E6:AD","portMacType":16},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:AC","portVMAC":"40:5C:FD:BC:E6:AC","portMacType":16},"port_list2":"3 4 "},"dcFabricType":0,"dcNPorts":4},"Fabric_Per_Slot_List":"A B C ","C":{"FabricLocation":"C","ShowA1A2":0,"isSelected":0,"isInstalled":0,"dcNPorts":4,"dcModelName":"","C1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:B2","portVMAC":"40:5C:FD:BC:E6:B2","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:B3","portVMAC":"40:5C:FD:BC:E6:B3","portMacType":0}},"dcFabricType":0,"C2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:B5","portVMAC":"40:5C:FD:BC:E6:B5","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:B4","portVMAC":"40:5C:FD:BC:E6:B4","portMacType":0},"port_list2":"3 4 "}},"B":{"FabricLocation":"B","ShowA1A2":0,"isSelected":0,"isInstalled":0,"B2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:B1","portVMAC":"40:5C:FD:BC:E6:B1","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:B0","portVMAC":"40:5C:FD:BC:E6:B0","portMacType":0},"port_list2":"3 4 "},"dcModelName":"","B1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:AE","portVMAC":"40:5C:FD:BC:E6:AE","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:AF","portVMAC":"40:5C:FD:BC:E6:AF","portMacType":0}},"dcFabricType":0,"dcNPorts":4}},"is_not_double_height":{"AddIdracTag":1,"FabricLocation":"iDRAC","isSelected":0,"isInstalled":"1","portPMAC":"40:5C:FD:BC:E6:A9","dcModelName":"iDRAC","portFMAC":"Not Installed","dcFabricType":0,"portMacType":"Management"}},"15":{"bladeSlotName":"SLOT-15","is_full_height":0,"BM_EXT_DC_INFO":{"A":{"FabricLocation":"A","ShowA1A2":0,"isSelected":0,"isInstalled":0,"A1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:B7","portVMAC":"40:5C:FD:BC:E6:B7","portMacType":16},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:B8","portVMAC":"40:5C:FD:BC:E6:B8","portMacType":16}},"dcModelName":"","A2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:BA","portVMAC":"40:5C:FD:BC:E6:BA","portMacType":16},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:B9","portVMAC":"40:5C:FD:BC:E6:B9","portMacType":16},"port_list2":"3 4 "},"dcFabricType":0,"dcNPorts":4},"Fabric_Per_Slot_List":"A B C ","C":{"FabricLocation":"C","ShowA1A2":0,"isSelected":0,"isInstalled":0,"dcNPorts":4,"dcModelName":"","C1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:BF","portVMAC":"40:5C:FD:BC:E6:BF","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:C0","portVMAC":"40:5C:FD:BC:E6:C0","portMacType":0}},"dcFabricType":0,"C2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:C2","portVMAC":"40:5C:FD:BC:E6:C2","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:C1","portVMAC":"40:5C:FD:BC:E6:C1","portMacType":0},"port_list2":"3 4 "}},"B":{"FabricLocation":"B","ShowA1A2":0,"isSelected":0,"isInstalled":0,"B2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:BE","portVMAC":"40:5C:FD:BC:E6:BE","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:BD","portVMAC":"40:5C:FD:BC:E6:BD","portMacType":0},"port_list2":"3 4 "},"dcModelName":"","B1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:BB","portVMAC":"40:5C:FD:BC:E6:BB","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:BC","portVMAC":"40:5C:FD:BC:E6:BC","portMacType":0}},"dcFabricType":0,"dcNPorts":4}},"is_not_double_height":{"AddIdracTag":1,"FabricLocation":"iDRAC","isSelected":0,"isInstalled":"1","portPMAC":"40:5C:FD:BC:E6:B6","dcModelName":"iDRAC","portFMAC":"Not Installed","dcFabricType":0,"portMacType":"Management"}},"pwwnSlotSelected15":0,"pwwnSlotSelected4":0,"pwwnSlotSelected16":0,"pwwnSlotSelected14":0,"13":{"bladeSlotName":"SLOT-13","is_full_height":0,"BM_EXT_DC_INFO":{"A":{"FabricLocation":"A","ShowA1A2":0,"isSelected":0,"isInstalled":0,"A1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:9D","portVMAC":"40:5C:FD:BC:E6:9D","portMacType":16},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:9E","portVMAC":"40:5C:FD:BC:E6:9E","portMacType":16}},"dcModelName":"","A2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:A0","portVMAC":"40:5C:FD:BC:E6:A0","portMacType":16},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:9F","portVMAC":"40:5C:FD:BC:E6:9F","portMacType":16},"port_list2":"3 4 "},"dcFabricType":0,"dcNPorts":4},"Fabric_Per_Slot_List":"A B C ","C":{"FabricLocation":"C","ShowA1A2":0,"isSelected":0,"isInstalled":0,"dcNPorts":4,"dcModelName":"","C1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:A5","portVMAC":"40:5C:FD:BC:E6:A5","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:A6","portVMAC":"40:5C:FD:BC:E6:A6","portMacType":0}},"dcFabricType":0,"C2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:A8","portVMAC":"40:5C:FD:BC:E6:A8","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:A7","portVMAC":"40:5C:FD:BC:E6:A7","portMacType":0},"port_list2":"3 4 "}},"B":{"FabricLocation":"B","ShowA1A2":0,"isSelected":0,"isInstalled":0,"B2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:A4","portVMAC":"40:5C:FD:BC:E6:A4","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:A3","portVMAC":"40:5C:FD:BC:E6:A3","portMacType":0},"port_list2":"3 4 "},"dcModelName":"","B1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:A1","portVMAC":"40:5C:FD:BC:E6:A1","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:A2","portVMAC":"40:5C:FD:BC:E6:A2","portMacType":0}},"dcFabricType":0,"dcNPorts":4}},"is_not_double_height":{"AddIdracTag":1,"FabricLocation":"iDRAC","isSelected":0,"isInstalled":"1","portPMAC":"40:5C:FD:BC:E6:9C","dcModelName":"iDRAC","portFMAC":"Not Installed","dcFabricType":0,"portMacType":"Management"}},"pwwnSlotSelected13":0,"pwwnSlotSelected7":0,"pwwnSlotSelected12":0,"pwwnSlotSelected2":0,"pwwnSlotSelected6":0,"pwwnSlotSelected11":0,"4":{"bladeSlotName":"SLOT-04","is_full_height":0,"BM_EXT_DC_INFO":{"A":{"FabricLocation":"A","ShowA1A2":0,"isSelected":0,"isInstalled":0,"A1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:28","portVMAC":"40:5C:FD:BC:E6:28","portMacType":16},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:29","portVMAC":"40:5C:FD:BC:E6:29","portMacType":16}},"dcModelName":"","A2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:2B","portVMAC":"40:5C:FD:BC:E6:2B","portMacType":16},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:2A","portVMAC":"40:5C:FD:BC:E6:2A","portMacType":16},"port_list2":"3 4 "},"dcFabricType":0,"dcNPorts":4},"Fabric_Per_Slot_List":"A B C ","C":{"FabricLocation":"C","ShowA1A2":0,"isSelected":0,"isInstalled":0,"dcNPorts":4,"dcModelName":"","C1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:30","portVMAC":"40:5C:FD:BC:E6:30","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:31","portVMAC":"40:5C:FD:BC:E6:31","portMacType":0}},"dcFabricType":0,"C2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:33","portVMAC":"40:5C:FD:BC:E6:33","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:32","portVMAC":"40:5C:FD:BC:E6:32","portMacType":0},"port_list2":"3 4 "}},"B":{"FabricLocation":"B","ShowA1A2":0,"isSelected":0,"isInstalled":0,"B2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:2F","portVMAC":"40:5C:FD:BC:E6:2F","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:2E","portVMAC":"40:5C:FD:BC:E6:2E","portMacType":0},"port_list2":"3 4 "},"dcModelName":"","B1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:2C","portVMAC":"40:5C:FD:BC:E6:2C","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:2D","portVMAC":"40:5C:FD:BC:E6:2D","portMacType":0}},"dcFabricType":0,"dcNPorts":4}},"is_not_double_height":{"AddIdracTag":1,"FabricLocation":"iDRAC","isSelected":0,"isInstalled":"1","portPMAC":"40:5C:FD:BC:E6:27","dcModelName":"iDRAC","portFMAC":"Not Installed","dcFabricType":0,"portMacType":"Management"}},"pwwnSlotSelected5":0,"pwwnSlotSelected10":0,"3":{"bladeSlotName":"localhost.localdomain","is_full_height":0,"BM_EXT_DC_INFO":{"A":{"FabricLocation":"A","ShowA1A2":0,"isSelected":0,"isInstalled":1,"A1":{"1":{"is_vmac_enabled":0,"partition_status":2,"IOMapped":"","portFMAC":"24:6E:96:0D:42:8C","portPMAC":"40:5C:FD:BC:E6:1B","portVMAC":"Not Installed","portMacType":3},"port_list1":"1 2 3 ","3":{"is_vmac_enabled":0,"partition_status":3,"IOMapped":"","portFMAC":"20:01:24:6E:96:0D:42:8D","portPMAC":"20:01:40:5C:FD:BC:E6:1C","portVMAC":"Not Installed","portMacType":128},"2":{"is_vmac_enabled":0,"partition_status":3,"IOMapped":"","portFMAC":"24:6E:96:0D:42:8D","portPMAC":"40:5C:FD:BC:E6:1C","portVMAC":"Not Installed","portMacType":129}},"dcModelName":"Intel(R) 10G 2P X520-k bNDC   ","A2":{"5":{"is_vmac_enabled":0,"partition_status":3,"IOMapped":"","portFMAC":"24:6E:96:0D:42:8F","portPMAC":"40:5C:FD:BC:E6:1E","portVMAC":"Not Installed","portMacType":129},"4":{"is_vmac_enabled":0,"partition_status":2,"IOMapped":"","portFMAC":"24:6E:96:0D:42:8E","portPMAC":"40:5C:FD:BC:E6:1D","portVMAC":"Not Installed","portMacType":3},"port_list2":"4 5 6 ","6":{"is_vmac_enabled":0,"partition_status":3,"IOMapped":"","portFMAC":"20:01:24:6E:96:0D:42:8F","portPMAC":"20:01:40:5C:FD:BC:E6:1E","portVMAC":"Not Installed","portMacType":128}},"dcFabricType":3,"dcNPorts":6},"Fabric_Per_Slot_List":"A B C ","C":{"FabricLocation":"C","ShowA1A2":0,"isSelected":0,"isInstalled":0,"dcNPorts":4,"dcModelName":"","C1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"Not Installed","portPMAC":"40:5C:FD:BC:E6:23","portVMAC":"40:5C:FD:BC:E6:23","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"Not Installed","portPMAC":"40:5C:FD:BC:E6:24","portVMAC":"40:5C:FD:BC:E6:24","portMacType":0}},"dcFabricType":0,"C2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"Not Installed","portPMAC":"40:5C:FD:BC:E6:26","portVMAC":"40:5C:FD:BC:E6:26","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"Not Installed","portPMAC":"40:5C:FD:BC:E6:25","portVMAC":"40:5C:FD:BC:E6:25","portMacType":0},"port_list2":"3 4 "}},"B":{"FabricLocation":"B","ShowA1A2":0,"isSelected":0,"isInstalled":0,"B2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"Not Installed","portPMAC":"40:5C:FD:BC:E6:22","portVMAC":"40:5C:FD:BC:E6:22","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"Not Installed","portPMAC":"40:5C:FD:BC:E6:21","portVMAC":"40:5C:FD:BC:E6:21","portMacType":0},"port_list2":"3 4 "},"dcModelName":"","B1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"Not Installed","portPMAC":"40:5C:FD:BC:E6:1F","portVMAC":"40:5C:FD:BC:E6:1F","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"Not Installed","portPMAC":"40:5C:FD:BC:E6:20","portVMAC":"40:5C:FD:BC:E6:20","portMacType":0}},"dcFabricType":0,"dcNPorts":4}},"is_not_double_height":{"AddIdracTag":1,"FabricLocation":"iDRAC","isSelected":0,"isInstalled":"1","portPMAC":"40:5C:FD:BC:E6:1A","dcModelName":"iDRAC","portFMAC":"10:98:36:9D:8F:B7","dcFabricType":0,"portMacType":"Management"}},"2":{"bladeSlotName":"devratefamilysearchproxy-1001.ams4.example.com","is_full_height":0,"BM_EXT_DC_INFO":{"A":{"FabricLocation":"A","ShowA1A2":0,"isSelected":0,"isInstalled":1,"A1":{"1":{"is_vmac_enabled":0,"partition_status":2,"IOMapped":"","portFMAC":"EC:F4:BB:E5:1F:C8","portPMAC":"40:5C:FD:BC:E6:0E","portVMAC":"Not Installed","portMacType":3},"port_list1":"1 2 3 ","3":{"is_vmac_enabled":0,"partition_status":3,"IOMapped":"","portFMAC":"20:01:EC:F4:BB:E5:1F:C9","portPMAC":"20:01:40:5C:FD:BC:E6:0F","portVMAC":"Not Installed","portMacType":128},"2":{"is_vmac_enabled":0,"partition_status":3,"IOMapped":"","portFMAC":"EC:F4:BB:E5:1F:C9","portPMAC":"40:5C:FD:BC:E6:0F","portVMAC":"Not Installed","portMacType":129}},"dcModelName":"Intel(R) 10G 2P X520-k bNDC   ","A2":{"5":{"is_vmac_enabled":0,"partition_status":3,"IOMapped":"","portFMAC":"EC:F4:BB:E5:1F:CB","portPMAC":"40:5C:FD:BC:E6:11","portVMAC":"Not Installed","portMacType":129},"4":{"is_vmac_enabled":0,"partition_status":2,"IOMapped":"","portFMAC":"EC:F4:BB:E5:1F:CA","portPMAC":"40:5C:FD:BC:E6:10","portVMAC":"Not Installed","portMacType":3},"port_list2":"4 5 6 ","6":{"is_vmac_enabled":0,"partition_status":3,"IOMapped":"","portFMAC":"20:01:EC:F4:BB:E5:1F:CB","portPMAC":"20:01:40:5C:FD:BC:E6:11","portVMAC":"Not Installed","portMacType":128}},"dcFabricType":3,"dcNPorts":6},"Fabric_Per_Slot_List":"A B C ","C":{"FabricLocation":"C","ShowA1A2":0,"isSelected":0,"isInstalled":0,"dcNPorts":4,"dcModelName":"","C1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"Not Installed","portPMAC":"40:5C:FD:BC:E6:16","portVMAC":"40:5C:FD:BC:E6:16","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"Not Installed","portPMAC":"40:5C:FD:BC:E6:17","portVMAC":"40:5C:FD:BC:E6:17","portMacType":0}},"dcFabricType":0,"C2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"Not Installed","portPMAC":"40:5C:FD:BC:E6:19","portVMAC":"40:5C:FD:BC:E6:19","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"Not Installed","portPMAC":"40:5C:FD:BC:E6:18","portVMAC":"40:5C:FD:BC:E6:18","portMacType":0},"port_list2":"3 4 "}},"B":{"FabricLocation":"B","ShowA1A2":0,"isSelected":0,"isInstalled":0,"B2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"Not Installed","portPMAC":"40:5C:FD:BC:E6:15","portVMAC":"40:5C:FD:BC:E6:15","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"Not Installed","portPMAC":"40:5C:FD:BC:E6:14","portVMAC":"40:5C:FD:BC:E6:14","portMacType":0},"port_list2":"3 4 "},"dcModelName":"","B1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"Not Installed","portPMAC":"40:5C:FD:BC:E6:12","portVMAC":"40:5C:FD:BC:E6:12","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"Not Installed","portPMAC":"40:5C:FD:BC:E6:13","portVMAC":"40:5C:FD:BC:E6:13","portMacType":0}},"dcFabricType":0,"dcNPorts":4}},"is_not_double_height":{"AddIdracTag":1,"FabricLocation":"iDRAC","isSelected":0,"isInstalled":"1","portPMAC":"40:5C:FD:BC:E6:0D","dcModelName":"iDRAC","portFMAC":"10:98:36:9B:AC:33","dcFabricType":0,"portMacType":"Management"}},"5":{"bladeSlotName":"SLOT-05","is_full_height":0,"BM_EXT_DC_INFO":{"A":{"FabricLocation":"A","ShowA1A2":0,"isSelected":0,"isInstalled":0,"A1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:35","portVMAC":"40:5C:FD:BC:E6:35","portMacType":16},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:36","portVMAC":"40:5C:FD:BC:E6:36","portMacType":16}},"dcModelName":"","A2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:38","portVMAC":"40:5C:FD:BC:E6:38","portMacType":16},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:37","portVMAC":"40:5C:FD:BC:E6:37","portMacType":16},"port_list2":"3 4 "},"dcFabricType":0,"dcNPorts":4},"Fabric_Per_Slot_List":"A B C ","C":{"FabricLocation":"C","ShowA1A2":0,"isSelected":0,"isInstalled":0,"dcNPorts":4,"dcModelName":"","C1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:3D","portVMAC":"40:5C:FD:BC:E6:3D","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:3E","portVMAC":"40:5C:FD:BC:E6:3E","portMacType":0}},"dcFabricType":0,"C2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:40","portVMAC":"40:5C:FD:BC:E6:40","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:3F","portVMAC":"40:5C:FD:BC:E6:3F","portMacType":0},"port_list2":"3 4 "}},"B":{"FabricLocation":"B","ShowA1A2":0,"isSelected":0,"isInstalled":0,"B2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:3C","portVMAC":"40:5C:FD:BC:E6:3C","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:3B","portVMAC":"40:5C:FD:BC:E6:3B","portMacType":0},"port_list2":"3 4 "},"dcModelName":"","B1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:39","portVMAC":"40:5C:FD:BC:E6:39","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:3A","portVMAC":"40:5C:FD:BC:E6:3A","portMacType":0}},"dcFabricType":0,"dcNPorts":4}},"is_not_double_height":{"AddIdracTag":1,"FabricLocation":"iDRAC","isSelected":0,"isInstalled":"1","portPMAC":"40:5C:FD:BC:E6:34","dcModelName":"iDRAC","portFMAC":"Not Installed","dcFabricType":0,"portMacType":"Management"}},"pwwnSlotSelected1":0,"7":{"bladeSlotName":"SLOT-07","is_full_height":0,"BM_EXT_DC_INFO":{"A":{"FabricLocation":"A","ShowA1A2":0,"isSelected":0,"isInstalled":0,"A1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:4F","portVMAC":"40:5C:FD:BC:E6:4F","portMacType":16},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:50","portVMAC":"40:5C:FD:BC:E6:50","portMacType":16}},"dcModelName":"","A2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:52","portVMAC":"40:5C:FD:BC:E6:52","portMacType":16},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:51","portVMAC":"40:5C:FD:BC:E6:51","portMacType":16},"port_list2":"3 4 "},"dcFabricType":0,"dcNPorts":4},"Fabric_Per_Slot_List":"A B C ","C":{"FabricLocation":"C","ShowA1A2":0,"isSelected":0,"isInstalled":0,"dcNPorts":4,"dcModelName":"","C1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:57","portVMAC":"40:5C:FD:BC:E6:57","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:58","portVMAC":"40:5C:FD:BC:E6:58","portMacType":0}},"dcFabricType":0,"C2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:5A","portVMAC":"40:5C:FD:BC:E6:5A","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:59","portVMAC":"40:5C:FD:BC:E6:59","portMacType":0},"port_list2":"3 4 "}},"B":{"FabricLocation":"B","ShowA1A2":0,"isSelected":0,"isInstalled":0,"B2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:56","portVMAC":"40:5C:FD:BC:E6:56","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:55","portVMAC":"40:5C:FD:BC:E6:55","portMacType":0},"port_list2":"3 4 "},"dcModelName":"","B1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:53","portVMAC":"40:5C:FD:BC:E6:53","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:54","portVMAC":"40:5C:FD:BC:E6:54","portMacType":0}},"dcFabricType":0,"dcNPorts":4}},"is_not_double_height":{"AddIdracTag":1,"FabricLocation":"iDRAC","isSelected":0,"isInstalled":"1","portPMAC":"40:5C:FD:BC:E6:4E","dcModelName":"iDRAC","portFMAC":"Not Installed","dcFabricType":0,"portMacType":"Management"}},"6":{"bladeSlotName":"SLOT-06","is_full_height":0,"BM_EXT_DC_INFO":{"A":{"FabricLocation":"A","ShowA1A2":0,"isSelected":0,"isInstalled":0,"A1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:42","portVMAC":"40:5C:FD:BC:E6:42","portMacType":16},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:43","portVMAC":"40:5C:FD:BC:E6:43","portMacType":16}},"dcModelName":"","A2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:45","portVMAC":"40:5C:FD:BC:E6:45","portMacType":16},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:44","portVMAC":"40:5C:FD:BC:E6:44","portMacType":16},"port_list2":"3 4 "},"dcFabricType":0,"dcNPorts":4},"Fabric_Per_Slot_List":"A B C ","C":{"FabricLocation":"C","ShowA1A2":0,"isSelected":0,"isInstalled":0,"dcNPorts":4,"dcModelName":"","C1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:4A","portVMAC":"40:5C:FD:BC:E6:4A","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:4B","portVMAC":"40:5C:FD:BC:E6:4B","portMacType":0}},"dcFabricType":0,"C2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:4D","portVMAC":"40:5C:FD:BC:E6:4D","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:4C","portVMAC":"40:5C:FD:BC:E6:4C","portMacType":0},"port_list2":"3 4 "}},"B":{"FabricLocation":"B","ShowA1A2":0,"isSelected":0,"isInstalled":0,"B2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:49","portVMAC":"40:5C:FD:BC:E6:49","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:48","portVMAC":"40:5C:FD:BC:E6:48","portMacType":0},"port_list2":"3 4 "},"dcModelName":"","B1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:46","portVMAC":"40:5C:FD:BC:E6:46","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:47","portVMAC":"40:5C:FD:BC:E6:47","portMacType":0}},"dcFabricType":0,"dcNPorts":4}},"is_not_double_height":{"AddIdracTag":1,"FabricLocation":"iDRAC","isSelected":0,"isInstalled":"1","portPMAC":"40:5C:FD:BC:E6:41","dcModelName":"iDRAC","portFMAC":"Not Installed","dcFabricType":0,"portMacType":"Management"}},"9":{"bladeSlotName":"SLOT-09","is_full_height":0,"BM_EXT_DC_INFO":{"A":{"FabricLocation":"A","ShowA1A2":0,"isSelected":0,"isInstalled":0,"A1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:69","portVMAC":"40:5C:FD:BC:E6:69","portMacType":16},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:6A","portVMAC":"40:5C:FD:BC:E6:6A","portMacType":16}},"dcModelName":"","A2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:6C","portVMAC":"40:5C:FD:BC:E6:6C","portMacType":16},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:6B","portVMAC":"40:5C:FD:BC:E6:6B","portMacType":16},"port_list2":"3 4 "},"dcFabricType":0,"dcNPorts":4},"Fabric_Per_Slot_List":"A B C ","C":{"FabricLocation":"C","ShowA1A2":0,"isSelected":0,"isInstalled":0,"dcNPorts":4,"dcModelName":"","C1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:71","portVMAC":"40:5C:FD:BC:E6:71","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:72","portVMAC":"40:5C:FD:BC:E6:72","portMacType":0}},"dcFabricType":0,"C2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:74","portVMAC":"40:5C:FD:BC:E6:74","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:73","portVMAC":"40:5C:FD:BC:E6:73","portMacType":0},"port_list2":"3 4 "}},"B":{"FabricLocation":"B","ShowA1A2":0,"isSelected":0,"isInstalled":0,"B2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:70","portVMAC":"40:5C:FD:BC:E6:70","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:6F","portVMAC":"40:5C:FD:BC:E6:6F","portMacType":0},"port_list2":"3 4 "},"dcModelName":"","B1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:6D","portVMAC":"40:5C:FD:BC:E6:6D","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:6E","portVMAC":"40:5C:FD:BC:E6:6E","portMacType":0}},"dcFabricType":0,"dcNPorts":4}},"is_not_double_height":{"AddIdracTag":1,"FabricLocation":"iDRAC","isSelected":0,"isInstalled":"1","portPMAC":"40:5C:FD:BC:E6:68","dcModelName":"iDRAC","portFMAC":"Not Installed","dcFabricType":0,"portMacType":"Management"}},"8":{"bladeSlotName":"SLOT-08","is_full_height":0,"BM_EXT_DC_INFO":{"A":{"FabricLocation":"A","ShowA1A2":0,"isSelected":0,"isInstalled":0,"A1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:5C","portVMAC":"40:5C:FD:BC:E6:5C","portMacType":16},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:5D","portVMAC":"40:5C:FD:BC:E6:5D","portMacType":16}},"dcModelName":"","A2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:5F","portVMAC":"40:5C:FD:BC:E6:5F","portMacType":16},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:5E","portVMAC":"40:5C:FD:BC:E6:5E","portMacType":16},"port_list2":"3 4 "},"dcFabricType":0,"dcNPorts":4},"Fabric_Per_Slot_List":"A B C ","C":{"FabricLocation":"C","ShowA1A2":0,"isSelected":0,"isInstalled":0,"dcNPorts":4,"dcModelName":"","C1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:64","portVMAC":"40:5C:FD:BC:E6:64","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:65","portVMAC":"40:5C:FD:BC:E6:65","portMacType":0}},"dcFabricType":0,"C2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:67","portVMAC":"40:5C:FD:BC:E6:67","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:66","portVMAC":"40:5C:FD:BC:E6:66","portMacType":0},"port_list2":"3 4 "}},"B":{"FabricLocation":"B","ShowA1A2":0,"isSelected":0,"isInstalled":0,"B2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:63","portVMAC":"40:5C:FD:BC:E6:63","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:62","portVMAC":"40:5C:FD:BC:E6:62","portMacType":0},"port_list2":"3 4 "},"dcModelName":"","B1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:60","portVMAC":"40:5C:FD:BC:E6:60","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:61","portVMAC":"40:5C:FD:BC:E6:61","portMacType":0}},"dcFabricType":0,"dcNPorts":4}},"is_not_double_height":{"AddIdracTag":1,"FabricLocation":"iDRAC","isSelected":0,"isInstalled":"1","portPMAC":"40:5C:FD:BC:E6:5B","dcModelName":"iDRAC","portFMAC":"Not Installed","dcFabricType":0,"portMacType":"Management"}},"16":{"bladeSlotName":"SLOT-16","is_full_height":0,"BM_EXT_DC_INFO":{"A":{"FabricLocation":"A","ShowA1A2":0,"isSelected":0,"isInstalled":0,"A1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:C4","portVMAC":"40:5C:FD:BC:E6:C4","portMacType":16},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:C5","portVMAC":"40:5C:FD:BC:E6:C5","portMacType":16}},"dcModelName":"","A2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:C7","portVMAC":"40:5C:FD:BC:E6:C7","portMacType":16},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:C6","portVMAC":"40:5C:FD:BC:E6:C6","portMacType":16},"port_list2":"3 4 "},"dcFabricType":0,"dcNPorts":4},"Fabric_Per_Slot_List":"A B C ","C":{"FabricLocation":"C","ShowA1A2":0,"isSelected":0,"isInstalled":0,"dcNPorts":4,"dcModelName":"","C1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:CC","portVMAC":"40:5C:FD:BC:E6:CC","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:CD","portVMAC":"40:5C:FD:BC:E6:CD","portMacType":0}},"dcFabricType":0,"C2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:CF","portVMAC":"40:5C:FD:BC:E6:CF","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:CE","portVMAC":"40:5C:FD:BC:E6:CE","portMacType":0},"port_list2":"3 4 "}},"B":{"FabricLocation":"B","ShowA1A2":0,"isSelected":0,"isInstalled":0,"B2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:CB","portVMAC":"40:5C:FD:BC:E6:CB","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:CA","portVMAC":"40:5C:FD:BC:E6:CA","portMacType":0},"port_list2":"3 4 "},"dcModelName":"","B1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:C8","portVMAC":"40:5C:FD:BC:E6:C8","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:C9","portVMAC":"40:5C:FD:BC:E6:C9","portMacType":0}},"dcFabricType":0,"dcNPorts":4}},"is_not_double_height":{"AddIdracTag":1,"FabricLocation":"iDRAC","isSelected":0,"isInstalled":"1","portPMAC":"40:5C:FD:BC:E6:C3","dcModelName":"iDRAC","portFMAC":"Not Installed","dcFabricType":0,"portMacType":"Management"}},"1":{"bladeSlotName":"provision-test-13163.ams4.example.com","is_full_height":0,"BM_EXT_DC_INFO":{"A":{"FabricLocation":"A","ShowA1A2":0,"isSelected":0,"isInstalled":1,"A1":{"1":{"is_vmac_enabled":0,"partition_status":2,"IOMapped":"","portFMAC":"EC:F4:BB:EB:A2:48","portPMAC":"40:5C:FD:BC:E6:01","portVMAC":"Not Installed","portMacType":3},"port_list1":"1 2 3 ","3":{"is_vmac_enabled":0,"partition_status":3,"IOMapped":"","portFMAC":"20:01:EC:F4:BB:EB:A2:49","portPMAC":"20:01:40:5C:FD:BC:E6:02","portVMAC":"Not Installed","portMacType":128},"2":{"is_vmac_enabled":0,"partition_status":3,"IOMapped":"","portFMAC":"EC:F4:BB:EB:A2:49","portPMAC":"40:5C:FD:BC:E6:02","portVMAC":"Not Installed","portMacType":129}},"dcModelName":"Intel(R) 10G 2P X520-k bNDC   ","A2":{"5":{"is_vmac_enabled":0,"partition_status":3,"IOMapped":"","portFMAC":"EC:F4:BB:EB:A2:4B","portPMAC":"40:5C:FD:BC:E6:04","portVMAC":"Not Installed","portMacType":129},"4":{"is_vmac_enabled":0,"partition_status":2,"IOMapped":"","portFMAC":"EC:F4:BB:EB:A2:4A","portPMAC":"40:5C:FD:BC:E6:03","portVMAC":"Not Installed","portMacType":3},"port_list2":"4 5 6 ","6":{"is_vmac_enabled":0,"partition_status":3,"IOMapped":"","portFMAC":"20:01:EC:F4:BB:EB:A2:4B","portPMAC":"20:01:40:5C:FD:BC:E6:04","portVMAC":"Not Installed","portMacType":128}},"dcFabricType":3,"dcNPorts":6},"Fabric_Per_Slot_List":"A B C ","C":{"FabricLocation":"C","ShowA1A2":0,"isSelected":0,"isInstalled":0,"dcNPorts":4,"dcModelName":"","C1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"Not Installed","portPMAC":"40:5C:FD:BC:E6:09","portVMAC":"40:5C:FD:BC:E6:09","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"Not Installed","portPMAC":"40:5C:FD:BC:E6:0A","portVMAC":"40:5C:FD:BC:E6:0A","portMacType":0}},"dcFabricType":0,"C2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"Not Installed","portPMAC":"40:5C:FD:BC:E6:0C","portVMAC":"40:5C:FD:BC:E6:0C","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"Not Installed","portPMAC":"40:5C:FD:BC:E6:0B","portVMAC":"40:5C:FD:BC:E6:0B","portMacType":0},"port_list2":"3 4 "}},"B":{"FabricLocation":"B","ShowA1A2":0,"isSelected":0,"isInstalled":0,"B2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"Not Installed","portPMAC":"40:5C:FD:BC:E6:08","portVMAC":"40:5C:FD:BC:E6:08","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"Not Installed","portPMAC":"40:5C:FD:BC:E6:07","portVMAC":"40:5C:FD:BC:E6:07","portMacType":0},"port_list2":"3 4 "},"dcModelName":"","B1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"Not Installed","portPMAC":"40:5C:FD:BC:E6:05","portVMAC":"40:5C:FD:BC:E6:05","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"Not Installed","portPMAC":"40:5C:FD:BC:E6:06","portVMAC":"40:5C:FD:BC:E6:06","portMacType":0}},"dcFabricType":0,"dcNPorts":4}},"is_not_double_height":{"AddIdracTag":1,"FabricLocation":"iDRAC","isSelected":0,"isInstalled":"1","portPMAC":"40:5C:FD:BC:E6:00","dcModelName":"iDRAC","portFMAC":"10:98:36:9C:BB:E9","dcFabricType":0,"portMacType":"Management"}},"pwwnSlotSelected3":0,"12":{"bladeSlotName":"SLOT-12","is_full_height":0,"BM_EXT_DC_INFO":{"A":{"FabricLocation":"A","ShowA1A2":0,"isSelected":0,"isInstalled":0,"A1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:90","portVMAC":"40:5C:FD:BC:E6:90","portMacType":16},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:91","portVMAC":"40:5C:FD:BC:E6:91","portMacType":16}},"dcModelName":"","A2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:93","portVMAC":"40:5C:FD:BC:E6:93","portMacType":16},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:92","portVMAC":"40:5C:FD:BC:E6:92","portMacType":16},"port_list2":"3 4 "},"dcFabricType":0,"dcNPorts":4},"Fabric_Per_Slot_List":"A B C ","C":{"FabricLocation":"C","ShowA1A2":0,"isSelected":0,"isInstalled":0,"dcNPorts":4,"dcModelName":"","C1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:98","portVMAC":"40:5C:FD:BC:E6:98","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:99","portVMAC":"40:5C:FD:BC:E6:99","portMacType":0}},"dcFabricType":0,"C2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:9B","portVMAC":"40:5C:FD:BC:E6:9B","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:9A","portVMAC":"40:5C:FD:BC:E6:9A","portMacType":0},"port_list2":"3 4 "}},"B":{"FabricLocation":"B","ShowA1A2":0,"isSelected":0,"isInstalled":0,"B2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:97","portVMAC":"40:5C:FD:BC:E6:97","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:96","portVMAC":"40:5C:FD:BC:E6:96","portMacType":0},"port_list2":"3 4 "},"dcModelName":"","B1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:94","portVMAC":"40:5C:FD:BC:E6:94","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:95","portVMAC":"40:5C:FD:BC:E6:95","portMacType":0}},"dcFabricType":0,"dcNPorts":4}},"is_not_double_height":{"AddIdracTag":1,"FabricLocation":"iDRAC","isSelected":0,"isInstalled":"1","portPMAC":"40:5C:FD:BC:E6:8F","dcModelName":"iDRAC","portFMAC":"Not Installed","dcFabricType":0,"portMacType":"Management"}},"11":{"bladeSlotName":"SLOT-11","is_full_height":0,"BM_EXT_DC_INFO":{"A":{"FabricLocation":"A","ShowA1A2":0,"isSelected":0,"isInstalled":0,"A1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:83","portVMAC":"40:5C:FD:BC:E6:83","portMacType":16},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:84","portVMAC":"40:5C:FD:BC:E6:84","portMacType":16}},"dcModelName":"","A2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:86","portVMAC":"40:5C:FD:BC:E6:86","portMacType":16},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:85","portVMAC":"40:5C:FD:BC:E6:85","portMacType":16},"port_list2":"3 4 "},"dcFabricType":0,"dcNPorts":4},"Fabric_Per_Slot_List":"A B C ","C":{"FabricLocation":"C","ShowA1A2":0,"isSelected":0,"isInstalled":0,"dcNPorts":4,"dcModelName":"","C1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:8B","portVMAC":"40:5C:FD:BC:E6:8B","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:8C","portVMAC":"40:5C:FD:BC:E6:8C","portMacType":0}},"dcFabricType":0,"C2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:8E","portVMAC":"40:5C:FD:BC:E6:8E","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:8D","portVMAC":"40:5C:FD:BC:E6:8D","portMacType":0},"port_list2":"3 4 "}},"B":{"FabricLocation":"B","ShowA1A2":0,"isSelected":0,"isInstalled":0,"B2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:8A","portVMAC":"40:5C:FD:BC:E6:8A","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:89","portVMAC":"40:5C:FD:BC:E6:89","portMacType":0},"port_list2":"3 4 "},"dcModelName":"","B1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:87","portVMAC":"40:5C:FD:BC:E6:87","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:88","portVMAC":"40:5C:FD:BC:E6:88","portMacType":0}},"dcFabricType":0,"dcNPorts":4}},"is_not_double_height":{"AddIdracTag":1,"FabricLocation":"iDRAC","isSelected":0,"isInstalled":"1","portPMAC":"40:5C:FD:BC:E6:82","dcModelName":"iDRAC","portFMAC":"Not Installed","dcFabricType":0,"portMacType":"Management"}},"10":{"bladeSlotName":"SLOT-10","is_full_height":0,"BM_EXT_DC_INFO":{"A":{"FabricLocation":"A","ShowA1A2":0,"isSelected":0,"isInstalled":0,"A1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:76","portVMAC":"40:5C:FD:BC:E6:76","portMacType":16},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:77","portVMAC":"40:5C:FD:BC:E6:77","portMacType":16}},"dcModelName":"","A2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:79","portVMAC":"40:5C:FD:BC:E6:79","portMacType":16},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:78","portVMAC":"40:5C:FD:BC:E6:78","portMacType":16},"port_list2":"3 4 "},"dcFabricType":0,"dcNPorts":4},"Fabric_Per_Slot_List":"A B C ","C":{"FabricLocation":"C","ShowA1A2":0,"isSelected":0,"isInstalled":0,"dcNPorts":4,"dcModelName":"","C1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:7E","portVMAC":"40:5C:FD:BC:E6:7E","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:7F","portVMAC":"40:5C:FD:BC:E6:7F","portMacType":0}},"dcFabricType":0,"C2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:81","portVMAC":"40:5C:FD:BC:E6:81","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:80","portVMAC":"40:5C:FD:BC:E6:80","portMacType":0},"port_list2":"3 4 "}},"B":{"FabricLocation":"B","ShowA1A2":0,"isSelected":0,"isInstalled":0,"B2":{"4":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:7D","portVMAC":"40:5C:FD:BC:E6:7D","portMacType":0},"3":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:7C","portVMAC":"40:5C:FD:BC:E6:7C","portMacType":0},"port_list2":"3 4 "},"dcModelName":"","B1":{"1":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:7A","portVMAC":"40:5C:FD:BC:E6:7A","portMacType":0},"port_list1":"1 2 ","2":{"is_vmac_enabled":0,"partition_status":0,"IOMapped":"","portFMAC":"","portPMAC":"40:5C:FD:BC:E6:7B","portVMAC":"40:5C:FD:BC:E6:7B","portMacType":0}},"dcFabricType":0,"dcNPorts":4}},"is_not_double_height":{"AddIdracTag":1,"FabricLocation":"iDRAC","isSelected":0,"isInstalled":"1","portPMAC":"40:5C:FD:BC:E6:75","dcModelName":"iDRAC","portFMAC":"Not Installed","dcFabricType":0,"portMacType":"Management"}}},"slot_list":"1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 "},"Fabric_Info_main":{"Fabric_Info":{"1":{"fabricLocation":1,"fabricPresent":1,"fabricType":16,"fabricSelected":1},"4":{"fabricLocation":4,"fabricPresent":0,"fabricType":0,"fabricSelected":1},"3":{"fabricLocation":3,"fabricPresent":0,"fabricType":0,"fabricSelected":1},"2":{"fabricLocation":2,"fabricPresent":0,"fabricType":0,"fabricSelected":1}},"Fabric_Info_List":"1 2 3 4 "}}`),
		},
	}
)

func setup() (r *M1000e, err error) {
	viper.SetDefault("debug", true)
	mux = http.NewServeMux()
	server = httptest.NewTLSServer(mux)
	ip := strings.TrimPrefix(server.URL, "https://")
	username := "super"
	password := "test"

	for url := range dellChassisAnswers {
		url := url
		mux.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
			if url == "/cgi-bin/webcgi/json" {
				urlQuery := r.URL.Query()
				query := urlQuery.Get("method")
				w.Write(dellChassisAnswers[url][query])
			} else {
				w.Write(dellChassisAnswers[url]["default"])
			}
		})
	}

	r, err = New(ip, username, password)
	if err != nil {
		return r, err
	}

	err = r.Login()
	if err != nil {
		return r, err
	}

	return r, err
}

func tearDown() {
	server.Close()
}

func TestChassisFwVersion(t *testing.T) {
	expectedAnswer := "6.00"

	chassis, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := chassis.FwVersion()
	if err != nil {
		t.Fatalf("Found errors calling chassis.FwVersion %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestChassisPassThru(t *testing.T) {
	expectedAnswer := "10G"

	chassis, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := chassis.PassThru()
	if err != nil {
		t.Fatalf("Found errors calling chassis.PassThru %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestChassisSerial(t *testing.T) {
	expectedAnswer := "51f3dk2"

	chassis, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := chassis.Serial()
	if err != nil {
		t.Fatalf("Found errors calling chassis.Serial %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestChassisModel(t *testing.T) {
	expectedAnswer := "PowerEdge M1000e"

	chassis, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := chassis.Model()
	if err != nil {
		t.Fatalf("Found errors calling chassis.Model %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestChassisName(t *testing.T) {
	expectedAnswer := "CMC-51F3DK2"

	chassis, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := chassis.Name()
	if err != nil {
		t.Fatalf("Found errors calling chassis.Name %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestChassisStatus(t *testing.T) {
	expectedAnswer := "OK"

	chassis, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := chassis.Status()
	if err != nil {
		t.Fatalf("Found errors calling chassis.Status %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestChassisPowerKW(t *testing.T) {
	expectedAnswer := 0.624

	chassis, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := chassis.PowerKw()
	if err != nil {
		t.Fatalf("Found errors calling chassis.PowerKW %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestChassisTempC(t *testing.T) {
	expectedAnswer := 17

	chassis, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := chassis.TempC()
	if err != nil {
		t.Fatalf("Found errors calling chassis.TempC %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestChassisNics(t *testing.T) {
	expectedAnswer := []*devices.Nic{
		&devices.Nic{
			MacAddress: "18:66:da:9d:cd:cd",
			Name:       "OA1",
		},
	}

	chassis, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	nics, err := chassis.Nics()
	if err != nil {
		t.Fatalf("Found errors calling chassis.Nics %v", err)
	}

	if len(nics) != len(expectedAnswer) {
		t.Fatalf("Expected %v nics: found %v nics", len(expectedAnswer), len(nics))
	}

	for pos, nic := range nics {
		if nic.MacAddress != expectedAnswer[pos].MacAddress || nic.Name != expectedAnswer[pos].Name {
			t.Errorf("Expected answer %v: found %v", expectedAnswer[pos], nic)
		}
	}

	tearDown()
}

func TestChassisPsu(t *testing.T) {
	expectedAnswer := []*devices.Psu{
		&devices.Psu{
			Serial:     "51f3dk2_psu_1",
			CapacityKw: 2.7,
			Status:     "OK",
			PowerKw:    0.184,
		},
		&devices.Psu{
			Serial:     "51f3dk2_psu_2",
			CapacityKw: 2.7,
			Status:     "OK",
			PowerKw:    0.20862,
		},
		&devices.Psu{
			Serial:     "51f3dk2_psu_5",
			CapacityKw: 2.7,
			Status:     "OK",
			PowerKw:    0.20772000000000002,
		},
		&devices.Psu{
			Serial:     "51f3dk2_psu_6",
			CapacityKw: 2.7,
			Status:     "OK",
			PowerKw:    0.25278,
		},
	}

	chassis, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	psus, err := chassis.Psus()
	if err != nil {
		t.Fatalf("Found errors calling chassis.Psus %v", err)
	}

	if len(psus) != len(expectedAnswer) {
		t.Fatalf("Expected %v psus: found %v psus", len(expectedAnswer), len(psus))
	}

	for pos, psu := range psus {
		if psu.Serial != expectedAnswer[pos].Serial || psu.CapacityKw != expectedAnswer[pos].CapacityKw || psu.PowerKw != expectedAnswer[pos].PowerKw || psu.Status != expectedAnswer[pos].Status {
			t.Errorf("Expected answer %v: found %v", expectedAnswer[pos], psu)
		}
	}

	tearDown()
}

func TestChassisInterface(t *testing.T) {
	chassis, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	_ = devices.BmcChassis(chassis)
	tearDown()
}
