package ilo

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/spf13/viper"
)

var (
	mux     *http.ServeMux
	server  *httptest.Server
	Answers = map[string][]byte{
		"/xmldata": []byte(`
			<RIMP>
			<HSI>
			<SBSN>CZ3605020D</SBSN>
			<SPN>ProLiant DL380 Gen9</SPN>
			<UUID>719064CZ3605020D</UUID>
			<SP>1</SP>
			<cUUID>30393137-3436-5A43-3336-303530323044</cUUID>
			<VIRTUAL>
			<STATE>Inactive</STATE>
			<VID>
			<BSN></BSN>
			<cUUID></cUUID>
			</VID>
			</VIRTUAL>
			<PRODUCTID> 719064-B21</PRODUCTID>
			<NICS>
			<NIC>
			<PORT>1</PORT>
			<DESCRIPTION>iLO 4</DESCRIPTION>
			<LOCATION>Embedded</LOCATION>
			<MACADDR>94:57:a5:60:aa:ca</MACADDR>
			<IPADDR>10.193.251.54</IPADDR>
			<STATUS>OK</STATUS>
			</NIC>
			<NIC>
			<PORT>2</PORT>
			<DESCRIPTION>iLO 4</DESCRIPTION>
			<LOCATION>Embedded</LOCATION>
			<MACADDR>94:57:a5:60:aa:cb</MACADDR>
			<IPADDR>Unknown</IPADDR>
			<STATUS>Disabled</STATUS>
			</NIC>
			<NIC>
			<PORT>1</PORT>
			<DESCRIPTION>HPE Ethernet 1Gb 4-port 331i Adapter - NIC</DESCRIPTION>
			<LOCATION>Embedded</LOCATION>
			<MACADDR>14:02:ec:33:1d:30</MACADDR>
			<IPADDR>N/A</IPADDR>
			<STATUS>Unknown</STATUS>
			</NIC>
			<NIC>
			<PORT>2</PORT>
			<DESCRIPTION>HPE Ethernet 1Gb 4-port 331i Adapter - NIC</DESCRIPTION>
			<LOCATION>Embedded</LOCATION>
			<MACADDR>14:02:ec:33:1d:31</MACADDR>
			<IPADDR>N/A</IPADDR>
			<STATUS>Unknown</STATUS>
			</NIC>
			<NIC>
			<PORT>3</PORT>
			<DESCRIPTION>HPE Ethernet 1Gb 4-port 331i Adapter - NIC</DESCRIPTION>
			<LOCATION>Embedded</LOCATION>
			<MACADDR>14:02:ec:33:1d:32</MACADDR>
			<IPADDR>N/A</IPADDR>
			<STATUS>Unknown</STATUS>
			</NIC>
			<NIC>
			<PORT>4</PORT>
			<DESCRIPTION>HPE Ethernet 1Gb 4-port 331i Adapter - NIC</DESCRIPTION>
			<LOCATION>Embedded</LOCATION>
			<MACADDR>14:02:ec:33:1d:33</MACADDR>
			<IPADDR>N/A</IPADDR>
			<STATUS>Unknown</STATUS>
			</NIC>
			<NIC>
			<PORT>1</PORT>
			<DESCRIPTION>HPE Ethernet 10Gb 2-port 562FLR-SFP+ Adpt</DESCRIPTION>
			<LOCATION>Embedded</LOCATION>
			<MACADDR>14:02:ec:6c:95:20</MACADDR>
			<IPADDR>N/A</IPADDR>
			<STATUS>OK</STATUS>
			</NIC>
			<NIC>
			<PORT>2</PORT>
			<DESCRIPTION>HPE Ethernet 10Gb 2-port 562FLR-SFP+ Adpt</DESCRIPTION>
			<LOCATION>Embedded</LOCATION>
			<MACADDR>14:02:ec:6c:95:28</MACADDR>
			<IPADDR>N/A</IPADDR>
			<STATUS>Unknown</STATUS>
			</NIC>
			</NICS>
			</HSI>
			<MP>
			<ST>1</ST>
			<PN>Integrated Lights-Out 4 (iLO 4)</PN>
			<FWRI>2.54</FWRI>
			<BBLK></BBLK>
			<HWRI>ASIC: 17</HWRI>
			<SN>ILOCZ3605020D</SN>
			<UUID>ILO719064CZ3605020D</UUID>
			<IPM>1</IPM>
			<SSO>0</SSO>
			<PWRM>1.0.9</PWRM>
			<ERS>0</ERS>
			<EALERT>1</EALERT>
			</MP>
			<SPATIAL>
			<DISCOVERY_RACK>Not Supported</DISCOVERY_RACK>
			<DISCOVERY_DATA>Server does not detect Location Discovery Services</DISCOVERY_DATA>
			<TAG_VERSION>0</TAG_VERSION>
			<RACK_ID>0</RACK_ID>
			<RACK_ID_PN>0</RACK_ID_PN>
			<RACK_DESCRIPTION>0</RACK_DESCRIPTION>
			<RACK_UHEIGHT>0</RACK_UHEIGHT>
			<UPOSITION>0</UPOSITION>
			<ULOCATION>0</ULOCATION>
			<cUUID>30393137-3436-5A43-3336-303530323044</cUUID>
			<UHEIGHT>2.00</UHEIGHT>
			<UOFFSET>0</UOFFSET>
			</SPATIAL>
			<HEALTH>
			<STATUS>2</STATUS>
			</HEALTH>
			</RIMP>
		`),
		"/json/login_session":      []byte(`OK`),
		"/json/overview":           []byte(`{"server_name":"bbmi","product_name":"ProLiant DL380 Gen9","serial_num":"CZ3605020D","virtual_serial_num":null,"product_id":"719064-B21","uuid":"30393137-3436-5A43-3336-303530323044","virtual_uuid":null,"system_rom":"P89 v2.42 (04/25/2017)","system_rom_date":"04/25/2017","backup_rom_date":"09/13/2016","license":"iLO Advanced","ilo_fw_version":"2.54 Jun 15 2017","ilo_fw_bootleg":"","nic":0,"ip_address":"10.193.251.54","ipv6_link_local":"FE80::9657:A5FF:FE60:AACA","system_health":"OP_STATUS_OK","uid_led":"UID_OFF","power":"ON","date":"Thu Nov  2 10:56:58 2017","https_port":443,"ilo_name":".machine.example.com","removable_hw":[{"tpm_status":"NOT_PRESENT","module_type":"UNSPECIFIED","sd_card":"NOT_PRESENT"}],"option_ROM_measuring":"Disabled","has_reset_priv":1,"chassis_sn":"","isUEFI":1,"ers_state":"ERS_INACTIVE"}`),
		"/json/mem_info":           []byte(`{"hostpwr_state":"ON","mem_type_configured":"MEM_ADVANCED_ECC","mem_type_active":"MEM_ADVANCED_ECC","mem_type_available":[{"available_type":"MEM_ADVANCED_ECC"},{"available_type":"MEM_RANK_SPARE"},{"available_type":"MEM_MIRROR_INTRA"}],"mem_status":"MEM_ADVANCED_ECC","mem_condition":"OP_STATUS_OK","mem_hot_plug":"MEM_UNKNOWN","mem_op_speed":1866,"mem_os_mem_size":0,"mem_total_mem_size":98304,"mem_riv_state":"MEM_UNKNOWN","mem_data_stale":0,"mem_boards":[{"brd_idx":0,"brd_slot_num":0,"brd_cpu_num":1,"brd_riser_num":0,"brd_online_status":"MEM_OTHER","brd_error_status":"MEM_OTHER","brd_locked":"MEM_OTHER","brd_num_of_sockets":12,"brd_os_mem_size":0,"brd_total_mem_size":49152,"brd_condition":"OP_STATUS_UNKNOWN","brd_hot_plug":"MEM_OTHER","brd_oper_freq":1866,"brd_oper_volt":1200},{"brd_idx":1,"brd_slot_num":1,"brd_cpu_num":2,"brd_riser_num":0,"brd_online_status":"MEM_OTHER","brd_error_status":"MEM_OTHER","brd_locked":"MEM_OTHER","brd_num_of_sockets":12,"brd_os_mem_size":0,"brd_total_mem_size":49152,"brd_condition":"OP_STATUS_UNKNOWN","brd_hot_plug":"MEM_OTHER","brd_oper_freq":1866,"brd_oper_volt":1200}],"mem_modules":[{"mem_mod_idx":0,"mem_brd_num":0,"mem_cpu_num":1,"mem_riser_num":0,"mem_mod_num":1,"mem_mod_size":16384,"mem_mod_type":"MEM_DIMM_DDR4","mem_mod_tech":"MEM_RDIMM","mem_mod_frequency":2133,"mem_mod_status":"MEM_GOOD_IN_USE","mem_mod_condition":"MEM_OK","mem_mod_smartmem":"MEM_SMART","mem_mod_part_num":"752369-081","mem_mod_min_volt":1200,"mem_mod_ranks":2},{"mem_mod_idx":1,"mem_brd_num":0,"mem_cpu_num":1,"mem_riser_num":0,"mem_mod_num":2,"mem_mod_size":0,"mem_mod_type":"MEM_DIMM_DDR4","mem_mod_tech":"MEM_OTHER","mem_mod_frequency":0,"mem_mod_status":"MEM_NOT_PRESENT","mem_mod_condition":"MEM_OTHER","mem_mod_smartmem":"MEM_NO","mem_mod_part_num":"NOT AVAILABLE","mem_mod_min_volt":0,"mem_mod_ranks":0},{"mem_mod_idx":2,"mem_brd_num":0,"mem_cpu_num":1,"mem_riser_num":0,"mem_mod_num":3,"mem_mod_size":0,"mem_mod_type":"MEM_DIMM_DDR4","mem_mod_tech":"MEM_OTHER","mem_mod_frequency":0,"mem_mod_status":"MEM_NOT_PRESENT","mem_mod_condition":"MEM_OTHER","mem_mod_smartmem":"MEM_NO","mem_mod_part_num":"NOT AVAILABLE","mem_mod_min_volt":0,"mem_mod_ranks":0},{"mem_mod_idx":3,"mem_brd_num":0,"mem_cpu_num":1,"mem_riser_num":0,"mem_mod_num":4,"mem_mod_size":0,"mem_mod_type":"MEM_DIMM_DDR4","mem_mod_tech":"MEM_OTHER","mem_mod_frequency":0,"mem_mod_status":"MEM_NOT_PRESENT","mem_mod_condition":"MEM_OTHER","mem_mod_smartmem":"MEM_NO","mem_mod_part_num":"NOT AVAILABLE","mem_mod_min_volt":0,"mem_mod_ranks":0},{"mem_mod_idx":4,"mem_brd_num":0,"mem_cpu_num":1,"mem_riser_num":0,"mem_mod_num":5,"mem_mod_size":0,"mem_mod_type":"MEM_DIMM_DDR4","mem_mod_tech":"MEM_OTHER","mem_mod_frequency":0,"mem_mod_status":"MEM_NOT_PRESENT","mem_mod_condition":"MEM_OTHER","mem_mod_smartmem":"MEM_NO","mem_mod_part_num":"NOT AVAILABLE","mem_mod_min_volt":0,"mem_mod_ranks":0},{"mem_mod_idx":5,"mem_brd_num":0,"mem_cpu_num":1,"mem_riser_num":0,"mem_mod_num":6,"mem_mod_size":0,"mem_mod_type":"MEM_DIMM_DDR4","mem_mod_tech":"MEM_OTHER","mem_mod_frequency":0,"mem_mod_status":"MEM_NOT_PRESENT","mem_mod_condition":"MEM_OTHER","mem_mod_smartmem":"MEM_NO","mem_mod_part_num":"NOT AVAILABLE","mem_mod_min_volt":0,"mem_mod_ranks":0},{"mem_mod_idx":6,"mem_brd_num":0,"mem_cpu_num":1,"mem_riser_num":0,"mem_mod_num":7,"mem_mod_size":0,"mem_mod_type":"MEM_DIMM_DDR4","mem_mod_tech":"MEM_OTHER","mem_mod_frequency":0,"mem_mod_status":"MEM_NOT_PRESENT","mem_mod_condition":"MEM_OTHER","mem_mod_smartmem":"MEM_NO","mem_mod_part_num":"NOT AVAILABLE","mem_mod_min_volt":0,"mem_mod_ranks":0},{"mem_mod_idx":7,"mem_brd_num":0,"mem_cpu_num":1,"mem_riser_num":0,"mem_mod_num":8,"mem_mod_size":0,"mem_mod_type":"MEM_DIMM_DDR4","mem_mod_tech":"MEM_OTHER","mem_mod_frequency":0,"mem_mod_status":"MEM_NOT_PRESENT","mem_mod_condition":"MEM_OTHER","mem_mod_smartmem":"MEM_NO","mem_mod_part_num":"NOT AVAILABLE","mem_mod_min_volt":0,"mem_mod_ranks":0},{"mem_mod_idx":8,"mem_brd_num":0,"mem_cpu_num":1,"mem_riser_num":0,"mem_mod_num":9,"mem_mod_size":16384,"mem_mod_type":"MEM_DIMM_DDR4","mem_mod_tech":"MEM_RDIMM","mem_mod_frequency":2133,"mem_mod_status":"MEM_GOOD_IN_USE","mem_mod_condition":"MEM_OK","mem_mod_smartmem":"MEM_SMART","mem_mod_part_num":"752369-081","mem_mod_min_volt":1200,"mem_mod_ranks":2},{"mem_mod_idx":9,"mem_brd_num":0,"mem_cpu_num":1,"mem_riser_num":0,"mem_mod_num":10,"mem_mod_size":0,"mem_mod_type":"MEM_DIMM_DDR4","mem_mod_tech":"MEM_OTHER","mem_mod_frequency":0,"mem_mod_status":"MEM_NOT_PRESENT","mem_mod_condition":"MEM_OTHER","mem_mod_smartmem":"MEM_NO","mem_mod_part_num":"NOT AVAILABLE","mem_mod_min_volt":0,"mem_mod_ranks":0},{"mem_mod_idx":10,"mem_brd_num":0,"mem_cpu_num":1,"mem_riser_num":0,"mem_mod_num":11,"mem_mod_size":0,"mem_mod_type":"MEM_DIMM_DDR4","mem_mod_tech":"MEM_OTHER","mem_mod_frequency":0,"mem_mod_status":"MEM_NOT_PRESENT","mem_mod_condition":"MEM_OTHER","mem_mod_smartmem":"MEM_NO","mem_mod_part_num":"NOT AVAILABLE","mem_mod_min_volt":0,"mem_mod_ranks":0},{"mem_mod_idx":11,"mem_brd_num":0,"mem_cpu_num":1,"mem_riser_num":0,"mem_mod_num":12,"mem_mod_size":16384,"mem_mod_type":"MEM_DIMM_DDR4","mem_mod_tech":"MEM_RDIMM","mem_mod_frequency":2133,"mem_mod_status":"MEM_GOOD_IN_USE","mem_mod_condition":"MEM_OK","mem_mod_smartmem":"MEM_SMART","mem_mod_part_num":"752369-081","mem_mod_min_volt":1200,"mem_mod_ranks":2},{"mem_mod_idx":12,"mem_brd_num":0,"mem_cpu_num":2,"mem_riser_num":0,"mem_mod_num":1,"mem_mod_size":16384,"mem_mod_type":"MEM_DIMM_DDR4","mem_mod_tech":"MEM_RDIMM","mem_mod_frequency":2133,"mem_mod_status":"MEM_GOOD_IN_USE","mem_mod_condition":"MEM_OK","mem_mod_smartmem":"MEM_SMART","mem_mod_part_num":"752369-081","mem_mod_min_volt":1200,"mem_mod_ranks":2},{"mem_mod_idx":13,"mem_brd_num":0,"mem_cpu_num":2,"mem_riser_num":0,"mem_mod_num":2,"mem_mod_size":0,"mem_mod_type":"MEM_DIMM_DDR4","mem_mod_tech":"MEM_OTHER","mem_mod_frequency":0,"mem_mod_status":"MEM_NOT_PRESENT","mem_mod_condition":"MEM_OTHER","mem_mod_smartmem":"MEM_NO","mem_mod_part_num":"NOT AVAILABLE","mem_mod_min_volt":0,"mem_mod_ranks":0},{"mem_mod_idx":14,"mem_brd_num":0,"mem_cpu_num":2,"mem_riser_num":0,"mem_mod_num":3,"mem_mod_size":0,"mem_mod_type":"MEM_DIMM_DDR4","mem_mod_tech":"MEM_OTHER","mem_mod_frequency":0,"mem_mod_status":"MEM_NOT_PRESENT","mem_mod_condition":"MEM_OTHER","mem_mod_smartmem":"MEM_NO","mem_mod_part_num":"NOT AVAILABLE","mem_mod_min_volt":0,"mem_mod_ranks":0},{"mem_mod_idx":15,"mem_brd_num":0,"mem_cpu_num":2,"mem_riser_num":0,"mem_mod_num":4,"mem_mod_size":0,"mem_mod_type":"MEM_DIMM_DDR4","mem_mod_tech":"MEM_OTHER","mem_mod_frequency":0,"mem_mod_status":"MEM_NOT_PRESENT","mem_mod_condition":"MEM_OTHER","mem_mod_smartmem":"MEM_NO","mem_mod_part_num":"NOT AVAILABLE","mem_mod_min_volt":0,"mem_mod_ranks":0},{"mem_mod_idx":16,"mem_brd_num":0,"mem_cpu_num":2,"mem_riser_num":0,"mem_mod_num":5,"mem_mod_size":0,"mem_mod_type":"MEM_DIMM_DDR4","mem_mod_tech":"MEM_OTHER","mem_mod_frequency":0,"mem_mod_status":"MEM_NOT_PRESENT","mem_mod_condition":"MEM_OTHER","mem_mod_smartmem":"MEM_NO","mem_mod_part_num":"NOT AVAILABLE","mem_mod_min_volt":0,"mem_mod_ranks":0},{"mem_mod_idx":17,"mem_brd_num":0,"mem_cpu_num":2,"mem_riser_num":0,"mem_mod_num":6,"mem_mod_size":0,"mem_mod_type":"MEM_DIMM_DDR4","mem_mod_tech":"MEM_OTHER","mem_mod_frequency":0,"mem_mod_status":"MEM_NOT_PRESENT","mem_mod_condition":"MEM_OTHER","mem_mod_smartmem":"MEM_NO","mem_mod_part_num":"NOT AVAILABLE","mem_mod_min_volt":0,"mem_mod_ranks":0},{"mem_mod_idx":18,"mem_brd_num":0,"mem_cpu_num":2,"mem_riser_num":0,"mem_mod_num":7,"mem_mod_size":0,"mem_mod_type":"MEM_DIMM_DDR4","mem_mod_tech":"MEM_OTHER","mem_mod_frequency":0,"mem_mod_status":"MEM_NOT_PRESENT","mem_mod_condition":"MEM_OTHER","mem_mod_smartmem":"MEM_NO","mem_mod_part_num":"NOT AVAILABLE","mem_mod_min_volt":0,"mem_mod_ranks":0},{"mem_mod_idx":19,"mem_brd_num":0,"mem_cpu_num":2,"mem_riser_num":0,"mem_mod_num":8,"mem_mod_size":0,"mem_mod_type":"MEM_DIMM_DDR4","mem_mod_tech":"MEM_OTHER","mem_mod_frequency":0,"mem_mod_status":"MEM_NOT_PRESENT","mem_mod_condition":"MEM_OTHER","mem_mod_smartmem":"MEM_NO","mem_mod_part_num":"NOT AVAILABLE","mem_mod_min_volt":0,"mem_mod_ranks":0},{"mem_mod_idx":20,"mem_brd_num":0,"mem_cpu_num":2,"mem_riser_num":0,"mem_mod_num":9,"mem_mod_size":16384,"mem_mod_type":"MEM_DIMM_DDR4","mem_mod_tech":"MEM_RDIMM","mem_mod_frequency":2133,"mem_mod_status":"MEM_GOOD_IN_USE","mem_mod_condition":"MEM_OK","mem_mod_smartmem":"MEM_SMART","mem_mod_part_num":"752369-081","mem_mod_min_volt":1200,"mem_mod_ranks":2},{"mem_mod_idx":21,"mem_brd_num":0,"mem_cpu_num":2,"mem_riser_num":0,"mem_mod_num":10,"mem_mod_size":0,"mem_mod_type":"MEM_DIMM_DDR4","mem_mod_tech":"MEM_OTHER","mem_mod_frequency":0,"mem_mod_status":"MEM_NOT_PRESENT","mem_mod_condition":"MEM_OTHER","mem_mod_smartmem":"MEM_NO","mem_mod_part_num":"NOT AVAILABLE","mem_mod_min_volt":0,"mem_mod_ranks":0},{"mem_mod_idx":22,"mem_brd_num":0,"mem_cpu_num":2,"mem_riser_num":0,"mem_mod_num":11,"mem_mod_size":0,"mem_mod_type":"MEM_DIMM_DDR4","mem_mod_tech":"MEM_OTHER","mem_mod_frequency":0,"mem_mod_status":"MEM_NOT_PRESENT","mem_mod_condition":"MEM_OTHER","mem_mod_smartmem":"MEM_NO","mem_mod_part_num":"NOT AVAILABLE","mem_mod_min_volt":0,"mem_mod_ranks":0},{"mem_mod_idx":23,"mem_brd_num":0,"mem_cpu_num":2,"mem_riser_num":0,"mem_mod_num":12,"mem_mod_size":16384,"mem_mod_type":"MEM_DIMM_DDR4","mem_mod_tech":"MEM_RDIMM","mem_mod_frequency":2133,"mem_mod_status":"MEM_GOOD_IN_USE","mem_mod_condition":"MEM_OK","mem_mod_smartmem":"MEM_SMART","mem_mod_part_num":"752369-081","mem_mod_min_volt":1200,"mem_mod_ranks":2}],"memory":[{"mem_dev_loc":"PROC 1 DIMM 1","mem_size":16384,"mem_speed":2133},{"mem_dev_loc":"PROC 1 DIMM 2","mem_size":0,"mem_speed":0},{"mem_dev_loc":"PROC 1 DIMM 3","mem_size":0,"mem_speed":0},{"mem_dev_loc":"PROC 1 DIMM 4","mem_size":0,"mem_speed":0},{"mem_dev_loc":"PROC 1 DIMM 5","mem_size":0,"mem_speed":0},{"mem_dev_loc":"PROC 1 DIMM 6","mem_size":0,"mem_speed":0},{"mem_dev_loc":"PROC 1 DIMM 7","mem_size":0,"mem_speed":0},{"mem_dev_loc":"PROC 1 DIMM 8","mem_size":0,"mem_speed":0},{"mem_dev_loc":"PROC 1 DIMM 9","mem_size":16384,"mem_speed":2133},{"mem_dev_loc":"PROC 1 DIMM 10","mem_size":0,"mem_speed":0},{"mem_dev_loc":"PROC 1 DIMM 11","mem_size":0,"mem_speed":0},{"mem_dev_loc":"PROC 1 DIMM 12","mem_size":16384,"mem_speed":2133},{"mem_dev_loc":"PROC 2 DIMM 1","mem_size":16384,"mem_speed":2133},{"mem_dev_loc":"PROC 2 DIMM 2","mem_size":0,"mem_speed":0},{"mem_dev_loc":"PROC 2 DIMM 3","mem_size":0,"mem_speed":0},{"mem_dev_loc":"PROC 2 DIMM 4","mem_size":0,"mem_speed":0},{"mem_dev_loc":"PROC 2 DIMM 5","mem_size":0,"mem_speed":0},{"mem_dev_loc":"PROC 2 DIMM 6","mem_size":0,"mem_speed":0},{"mem_dev_loc":"PROC 2 DIMM 7","mem_size":0,"mem_speed":0},{"mem_dev_loc":"PROC 2 DIMM 8","mem_size":0,"mem_speed":0},{"mem_dev_loc":"PROC 2 DIMM 9","mem_size":16384,"mem_speed":2133},{"mem_dev_loc":"PROC 2 DIMM 10","mem_size":0,"mem_speed":0},{"mem_dev_loc":"PROC 2 DIMM 11","mem_size":0,"mem_speed":0},{"mem_dev_loc":"PROC 2 DIMM 12","mem_size":16384,"mem_speed":2133}]}`),
		"/json/proc_info":          []byte(`{"hostpwr_state":"ON","processors":[{"proc_socket":"Proc 1","proc_name":"Intel(R) Xeon(R) CPU E5-2620 v3 @ 2.40GHz","proc_status":"OP_STATUS_OK","proc_speed":2400,"proc_num_cores_enabled":6,"proc_num_cores":6,"proc_num_threads":12,"proc_mem_technology":"64-bit Capable","proc_num_l1cache":384,"proc_num_l2cache":1536,"proc_num_l3cache":15360},{"proc_socket":"Proc 2","proc_name":"Intel(R) Xeon(R) CPU E5-2620 v3 @ 2.40GHz","proc_status":"OP_STATUS_OK","proc_speed":2400,"proc_num_cores_enabled":6,"proc_num_cores":6,"proc_num_threads":12,"proc_mem_technology":"64-bit Capable","proc_num_l1cache":384,"proc_num_l2cache":1536,"proc_num_l3cache":15360}]}`),
		"/json/power_summary":      []byte(`{"hostpwr_state":"ON","last_avg_pwr_accum":143,"last_5min_avg":141,"last_5min_peak":148,"_24hr_average":139,"_24hr_peak":167,"_24hr_min":138,"_24hr_max_cap":0,"_24hr_max_temp":13,"_20min_average":143,"_20min_peak":149,"_20min_min":140,"_20min_max_cap":0,"max_measured_wattage":283,"min_measured_wattage":0,"volts":229,"power_cap":0,"power_cap_mode":"off","power_regulator_mode":"max","power_supply_capacity":1000,"power_supply_input_power":145,"num_valid_history_samples":288,"num_valid_fast_history_samples":120,"powerreg":1}`),
		"/json/health_temperature": []byte(`{"hostpwr_state":"ON","in_post":11,"temperature":[{"label":"01-Inlet Ambient","xposition":15,"yposition":0,"location":"Ambient","status":"OP_STATUS_OK","currentreading":13,"caution":42,"critical":50,"temp_unit":"Celsius"},{"label":"02-CPU 1","xposition":11,"yposition":5,"location":"CPU","status":"OP_STATUS_OK","currentreading":40,"caution":70,"critical":0,"temp_unit":"Celsius"},{"label":"03-CPU 2","xposition":4,"yposition":5,"location":"CPU","status":"OP_STATUS_OK","currentreading":40,"caution":70,"critical":0,"temp_unit":"Celsius"},{"label":"04-P1 DIMM 1-6","xposition":9,"yposition":5,"location":"Memory","status":"OP_STATUS_OK","currentreading":28,"caution":89,"critical":0,"temp_unit":"Celsius"},{"label":"05-P1 DIMM 7-12","xposition":14,"yposition":5,"location":"Memory","status":"OP_STATUS_OK","currentreading":31,"caution":89,"critical":0,"temp_unit":"Celsius"},{"label":"06-P2 DIMM 1-6","xposition":1,"yposition":5,"location":"Memory","status":"OP_STATUS_OK","currentreading":22,"caution":89,"critical":0,"temp_unit":"Celsius"},{"label":"07-P2 DIMM 7-12","xposition":6,"yposition":5,"location":"Memory","status":"OP_STATUS_OK","currentreading":28,"caution":89,"critical":0,"temp_unit":"Celsius"},{"label":"08-HD Max","xposition":10,"yposition":0,"location":"System","status":"OP_STATUS_OK","currentreading":35,"caution":60,"critical":0,"temp_unit":"Celsius"},{"label":"09-Exp Bay Drive","xposition":12,"yposition":0,"location":"System","status":"OP_STATUS_ABSENT","currentreading":0,"caution":75,"critical":0,"temp_unit":"Celsius"},{"label":"10-Chipset","xposition":13,"yposition":10,"location":"System","status":"OP_STATUS_OK","currentreading":37,"caution":105,"critical":0,"temp_unit":"Celsius"},{"label":"11-PS 1 Inlet","xposition":1,"yposition":10,"location":"Power Supply","status":"OP_STATUS_OK","currentreading":18,"caution":0,"critical":0,"temp_unit":"Celsius"},{"label":"12-PS 2 Inlet","xposition":4,"yposition":10,"location":"Power Supply","status":"OP_STATUS_OK","currentreading":25,"caution":0,"critical":0,"temp_unit":"Celsius"},{"label":"13-VR P1","xposition":10,"yposition":1,"location":"System","status":"OP_STATUS_OK","currentreading":35,"caution":115,"critical":120,"temp_unit":"Celsius"},{"label":"14-VR P2","xposition":4,"yposition":1,"location":"System","status":"OP_STATUS_OK","currentreading":33,"caution":115,"critical":120,"temp_unit":"Celsius"},{"label":"15-VR P1 Mem","xposition":9,"yposition":1,"location":"System","status":"OP_STATUS_OK","currentreading":25,"caution":115,"critical":120,"temp_unit":"Celsius"},{"label":"16-VR P1 Mem","xposition":13,"yposition":1,"location":"System","status":"OP_STATUS_OK","currentreading":27,"caution":115,"critical":120,"temp_unit":"Celsius"},{"label":"17-VR P2 Mem","xposition":2,"yposition":1,"location":"System","status":"OP_STATUS_OK","currentreading":26,"caution":115,"critical":120,"temp_unit":"Celsius"},{"label":"18-VR P2 Mem","xposition":6,"yposition":1,"location":"System","status":"OP_STATUS_OK","currentreading":25,"caution":115,"critical":120,"temp_unit":"Celsius"},{"label":"19-PS 1 Internal","xposition":1,"yposition":13,"location":"Power Supply","status":"OP_STATUS_OK","currentreading":40,"caution":0,"critical":0,"temp_unit":"Celsius"},{"label":"20-PS 2 Internal","xposition":4,"yposition":13,"location":"Power Supply","status":"OP_STATUS_OK","currentreading":40,"caution":0,"critical":0,"temp_unit":"Celsius"},{"label":"21-PCI 1","xposition":13,"yposition":13,"location":"I/O Board","status":"OP_STATUS_ABSENT","currentreading":0,"caution":100,"critical":0,"temp_unit":"Celsius"},{"label":"22-PCI 2","xposition":13,"yposition":13,"location":"I/O Board","status":"OP_STATUS_ABSENT","currentreading":0,"caution":100,"critical":0,"temp_unit":"Celsius"},{"label":"23-PCI 3","xposition":13,"yposition":13,"location":"I/O Board","status":"OP_STATUS_ABSENT","currentreading":0,"caution":100,"critical":0,"temp_unit":"Celsius"},{"label":"24-PCI 4","xposition":5,"yposition":12,"location":"I/O Board","status":"OP_STATUS_ABSENT","currentreading":0,"caution":100,"critical":0,"temp_unit":"Celsius"},{"label":"25-PCI 5","xposition":5,"yposition":12,"location":"I/O Board","status":"OP_STATUS_ABSENT","currentreading":0,"caution":100,"critical":0,"temp_unit":"Celsius"},{"label":"26-PCI 6","xposition":5,"yposition":12,"location":"I/O Board","status":"OP_STATUS_ABSENT","currentreading":0,"caution":100,"critical":0,"temp_unit":"Celsius"},{"label":"27-HD Controller","xposition":8,"yposition":8,"location":"I/O Board","status":"OP_STATUS_OK","currentreading":55,"caution":100,"critical":0,"temp_unit":"Celsius"},{"label":"28-LOM Card","xposition":14,"yposition":14,"location":"I/O Board","status":"OP_STATUS_OK","currentreading":70,"caution":100,"critical":0,"temp_unit":"Celsius"},{"label":"29-LOM","xposition":7,"yposition":14,"location":"System","status":"OP_STATUS_ABSENT","currentreading":0,"caution":100,"critical":0,"temp_unit":"Celsius"},{"label":"30-Front Ambient","xposition":9,"yposition":0,"location":"Ambient","status":"OP_STATUS_OK","currentreading":22,"caution":65,"critical":0,"temp_unit":"Celsius"},{"label":"31-PCI 1 Zone.","xposition":13,"yposition":13,"location":"I/O Board","status":"OP_STATUS_OK","currentreading":25,"caution":70,"critical":75,"temp_unit":"Celsius"},{"label":"32-PCI 2 Zone.","xposition":13,"yposition":13,"location":"I/O Board","status":"OP_STATUS_OK","currentreading":26,"caution":70,"critical":75,"temp_unit":"Celsius"},{"label":"33-PCI 3 Zone.","xposition":13,"yposition":13,"location":"I/O Board","status":"OP_STATUS_OK","currentreading":26,"caution":70,"critical":75,"temp_unit":"Celsius"},{"label":"34-PCI 4 Zone","xposition":5,"yposition":12,"location":"I/O Board","status":"OP_STATUS_ABSENT","currentreading":0,"caution":70,"critical":75,"temp_unit":"Celsius"},{"label":"35-PCI 5 Zone","xposition":5,"yposition":12,"location":"I/O Board","status":"OP_STATUS_ABSENT","currentreading":0,"caution":70,"critical":75,"temp_unit":"Celsius"},{"label":"36-PCI 6 Zone","xposition":5,"yposition":12,"location":"I/O Board","status":"OP_STATUS_ABSENT","currentreading":0,"caution":70,"critical":75,"temp_unit":"Celsius"},{"label":"37-HD Cntlr Zone","xposition":11,"yposition":7,"location":"I/O Board","status":"OP_STATUS_OK","currentreading":36,"caution":75,"critical":0,"temp_unit":"Celsius"},{"label":"38-I/O Zone","xposition":14,"yposition":11,"location":"System","status":"OP_STATUS_OK","currentreading":29,"caution":75,"critical":80,"temp_unit":"Celsius"},{"label":"39-P/S 2 Zone","xposition":3,"yposition":7,"location":"System","status":"OP_STATUS_OK","currentreading":29,"caution":70,"critical":0,"temp_unit":"Celsius"},{"label":"40-Battery Zone","xposition":7,"yposition":10,"location":"System","status":"OP_STATUS_OK","currentreading":28,"caution":75,"critical":80,"temp_unit":"Celsius"},{"label":"41-iLO Zone","xposition":9,"yposition":14,"location":"System","status":"OP_STATUS_OK","currentreading":31,"caution":90,"critical":95,"temp_unit":"Celsius"},{"label":"42-Rear HD Max","xposition":9,"yposition":14,"location":"System","status":"OP_STATUS_ABSENT","currentreading":0,"caution":60,"critical":0,"temp_unit":"Celsius"},{"label":"43-Storage Batt","xposition":5,"yposition":1,"location":"System","status":"OP_STATUS_OK","currentreading":17,"caution":60,"critical":0,"temp_unit":"Celsius"},{"label":"44-Fuse","xposition":3,"yposition":14,"location":"Power Supply","status":"OP_STATUS_OK","currentreading":28,"caution":100,"critical":0,"temp_unit":"Celsius"}]}`),
		"/json/license":            []byte(`{"key":"3353M-XKMML-D7H3P-XV794-3DXMM","name":"iLO Advanced","type":"Perpetual","expires":"","seats":0}`),
		"/json/power_supplies":     []byte(`{"supplies":[{"unhealthy":0,"enabled":1,"mismatch":0,"ps_bay":1,"ps_present":"PS_YES","ps_condition":"PS_OK","ps_error_code":"PS_GOOD_IN_USE","ps_ipdu_capable":"PS_NO","ps_hotplug_capable":"PS_YES","ps_model":"720478-B21","ps_spare":"754377-001","ps_serial_num":"5DMWA0CLL9E56R","ps_max_cap_watts":500,"ps_fw_ver":"1.00","ps_input_volts":230,"ps_output_watts":73,"avg":72,"max":74,"supply":true,"bbu":false,"charge":0,"age":0,"battery_health":0},{"unhealthy":0,"enabled":1,"mismatch":0,"ps_bay":2,"ps_present":"PS_YES","ps_condition":"PS_OK","ps_error_code":"PS_GOOD_IN_USE","ps_ipdu_capable":"PS_NO","ps_hotplug_capable":"PS_YES","ps_model":"720478-B21","ps_spare":"754377-001","ps_serial_num":"5DMWA0CLL9E5SU","ps_max_cap_watts":500,"ps_fw_ver":"1.00","ps_input_volts":228,"ps_output_watts":70,"avg":70,"max":72,"supply":true,"bbu":false,"charge":0,"age":0,"battery_health":0}],"present_power_reading":143}`),
		"/json/health_phy_drives":  []byte(`{"hostpwr_state":"ON","in_post":0,"ams_ready":"AMS_UNAVAILABLE","data_state":"DATA_NOT_AVAILABLE","next_page":null,"phy_drive_arrays":[{"physical_drives":[{"name":"Physical Drive in Port 1I Box 1 Bay 1","status":"OP_STATUS_OK","serial_no":"S403CRXK0000E7227365","model":"EG1200JEMDA","capacity":"1200 GB","location":"Port 1I Box 1 Bay 1","fw_version":"HPD6","phys_status":"PHYS_OK","drive_type":"PHY_ARRAY","encr_stat":"ENCR_NOT_ENCR","phys_idx":0,"drive_mediatype":"HDD"},{"name":"Physical Drive in Port 1I Box 1 Bay 2","status":"OP_STATUS_OK","serial_no":"S403D7J40000E722A3MT","model":"EG1200JEMDA","capacity":"1200 GB","location":"Port 1I Box 1 Bay 2","fw_version":"HPD6","phys_status":"PHYS_OK","drive_type":"PHY_ARRAY","encr_stat":"ENCR_NOT_ENCR","phys_idx":1,"drive_mediatype":"HDD"}],"storage_type":"SMART_ARRAY_CONTROLLER_TYPE","name":"Controller on System Board","status":"OP_STATUS_OK","hw_status":"OP_STATUS_OK","serial_no":"PDNLU0MLM55058","model":"Smart Array P246br Controller","fw_version":"5.52","accel_cond":"OP_STATUS_OK","accel_serial":"PDNLU0MLM55058","accel_tot_mem":"1048576 KB","has_accel":1,"encr_stat":"ENCR_NOT_ENABLED","encr_self_stat":"OP_STATUS_OK","encr_csp_stat":"OP_STATUS_OK","has_encrypt":1,"enclosures":[{"name":"Drive Enclosure Port 1I Box 1","status":"OP_STATUS_OK","ports":"2"}]}]}`),
	}
)

func setup() (bmc *Ilo, err error) {
	viper.SetDefault("debug", true)
	mux = http.NewServeMux()
	server = httptest.NewTLSServer(mux)
	ip := strings.TrimPrefix(server.URL, "https://")
	username := "super"
	password := "test"

	for url := range Answers {
		url := url
		mux.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
			cookie := http.Cookie{Name: "sessionKey", Value: "sessionKey_test"}
			http.SetCookie(w, &cookie)
			w.Write(Answers[url])
		})
	}

	bmc, err = New(ip, username, password)
	if err != nil {
		return bmc, err
	}

	return bmc, err
}

func tearDown() {
	server.Close()
}

func TestIloSerial(t *testing.T) {
	expectedAnswer := "cz3605020d"

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.Serial()
	if err != nil {
		t.Fatalf("Found errors calling bmc.Serial %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestIloModel(t *testing.T) {
	expectedAnswer := "ProLiant DL380 Gen9"

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.Model()
	if err != nil {
		t.Fatalf("Found errors calling bmc.Model %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestIloBmcType(t *testing.T) {
	expectedAnswer := "ilo4"

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer := bmc.BmcType()
	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestIloBmcVersion(t *testing.T) {
	expectedAnswer := "2.54"

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.BmcVersion()
	if err != nil {
		t.Fatalf("Found errors calling bmc.BmcVersion %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestIloName(t *testing.T) {
	expectedAnswer := "bbmi"

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.Name()
	if err != nil {
		t.Fatalf("Found errors calling bmc.Name %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestIloStatus(t *testing.T) {
	expectedAnswer := "OK"

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.Status()
	if err != nil {
		t.Fatalf("Found errors calling bmc.Status %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestIloMemory(t *testing.T) {
	expectedAnswer := 96

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.Memory()
	if err != nil {
		t.Fatalf("Found errors calling bmc.Memory %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestIloCPU(t *testing.T) {
	expectedAnswerCPUType := "intel(r) xeon(r) cpu e5-2620 v3"
	expectedAnswerCPUCount := 2
	expectedAnswerCore := 6
	expectedAnswerHyperthread := 12

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	cpuType, cpuCount, core, ht, err := bmc.CPU()
	if err != nil {
		t.Fatalf("Found errors calling bmc.CPU %v", err)
	}

	if cpuType != expectedAnswerCPUType {
		t.Errorf("Expected cpuType answer %v: found %v", expectedAnswerCPUType, cpuType)
	}

	if cpuCount != expectedAnswerCPUCount {
		t.Errorf("Expected cpuCount answer %v: found %v", expectedAnswerCPUCount, cpuCount)
	}

	if core != expectedAnswerCore {
		t.Errorf("Expected core answer %v: found %v", expectedAnswerCore, core)
	}

	if ht != expectedAnswerHyperthread {
		t.Errorf("Expected ht answer %v: found %v", expectedAnswerHyperthread, ht)
	}

	tearDown()
}

func TestIloBiosVersion(t *testing.T) {
	expectedAnswer := "P89 v2.42 (04/25/2017)"

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.BiosVersion()
	if err != nil {
		t.Fatalf("Found errors calling bmc.BiosVersion %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestIloPowerKW(t *testing.T) {
	expectedAnswer := 0.1416015625

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.PowerKw()
	if err != nil {
		t.Fatalf("Found errors calling bmc.PowerKW %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestIloTempC(t *testing.T) {
	expectedAnswer := 13

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.TempC()
	if err != nil {
		t.Fatalf("Found errors calling bmc.Temp %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestIloNics(t *testing.T) {
	expectedAnswer := []*devices.Nic{
		{
			MacAddress: "94:57:a5:60:aa:ca",
			Name:       "bmc",
		},
		{
			MacAddress: "94:57:a5:60:aa:cb",
			Name:       "bmc",
		},
		{
			MacAddress: "14:02:ec:33:1d:30",
			Name:       "HPE Ethernet 1Gb 4-port 331i Adapter - NIC",
		},
		{
			MacAddress: "14:02:ec:33:1d:31",
			Name:       "HPE Ethernet 1Gb 4-port 331i Adapter - NIC",
		},
		{
			MacAddress: "14:02:ec:33:1d:32",
			Name:       "HPE Ethernet 1Gb 4-port 331i Adapter - NIC",
		},
		{
			MacAddress: "14:02:ec:33:1d:33",
			Name:       "HPE Ethernet 1Gb 4-port 331i Adapter - NIC",
		},
		{
			MacAddress: "14:02:ec:6c:95:20",
			Name:       "HPE Ethernet 10Gb 2-port 562FLR-SFP+ Adpt",
		},
		{
			MacAddress: "14:02:ec:6c:95:28",
			Name:       "HPE Ethernet 10Gb 2-port 562FLR-SFP+ Adpt",
		},
	}

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	nics, err := bmc.Nics()
	if err != nil {
		t.Fatalf("Found errors calling bmc.Nics %v", err)
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

func TestIloLicense(t *testing.T) {
	expectedName := "iLO Advanced"
	expectedLicType := "Perpetual"

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	name, licType, err := bmc.License()
	if err != nil {
		t.Fatalf("Found errors calling bmc.License %v", err)
	}

	if name != expectedName {
		t.Errorf("Expected name %v: found %v", expectedName, name)
	}

	if licType != expectedLicType {
		t.Errorf("Expected name %v: found %v", expectedLicType, licType)
	}

	tearDown()
}

func TestIloPsu(t *testing.T) {
	expectedAnswer := []*devices.Psu{
		{
			Serial:     "5dmwa0cll9e56r",
			CapacityKw: 0.5,
			Status:     "OK",
			PowerKw:    0.073,
		},
		{
			Serial:     "5dmwa0cll9e5su",
			CapacityKw: 0.5,
			Status:     "OK",
			PowerKw:    0.07,
		},
	}

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test discrete %v", err)
	}

	psus, err := bmc.Psus()
	if err != nil {
		t.Fatalf("Found errors calling discrete.Psus %v", err)
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

func TestIloDisks(t *testing.T) {
	expectedAnswer := []*devices.Disk{
		{
			Serial:    "s403crxk0000e7227365",
			Type:      "HDD",
			Size:      "1200 GB",
			Model:     "eg1200jemda",
			Location:  "Port 1I Box 1 Bay 1",
			Status:    "OK",
			FwVersion: "hpd6",
		},
		{
			Serial:    "s403d7j40000e722a3mt",
			Type:      "HDD",
			Size:      "1200 GB",
			Model:     "eg1200jemda",
			Location:  "Port 1I Box 1 Bay 2",
			Status:    "OK",
			FwVersion: "hpd6",
		},
	}

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test hpChassissetup %v", err)
	}

	disks, err := bmc.Disks()
	if err != nil {
		t.Fatalf("Found errors calling chassis.Disks %v", err)
	}

	if len(disks) != len(expectedAnswer) {
		t.Fatalf("Expected %v disks: found %v disks", len(expectedAnswer), len(disks))
	}

	for pos, disk := range disks {
		if disk.Serial != expectedAnswer[pos].Serial ||
			disk.Type != expectedAnswer[pos].Type ||
			disk.Size != expectedAnswer[pos].Size ||
			disk.Status != expectedAnswer[pos].Status ||
			disk.Model != expectedAnswer[pos].Model ||
			disk.FwVersion != expectedAnswer[pos].FwVersion ||
			disk.Location != expectedAnswer[pos].Location {
			t.Errorf("Expected answer %v: found %v", expectedAnswer[pos], disk)
		}
	}

	tearDown()
}

func TestIloIsBlade(t *testing.T) {
	expectedAnswer := false

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.IsBlade()
	if err != nil {
		t.Fatalf("Found errors calling bmc.isBlade %v", err)
	}

	if expectedAnswer != answer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestIloPoweState(t *testing.T) {
	expectedAnswer := "on"

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.PowerState()
	if err != nil {
		t.Fatalf("Found errors calling bmc.PowerState %v", err)
	}

	if expectedAnswer != answer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestIloInterface(t *testing.T) {
	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	_ = devices.Bmc(bmc)
	_ = devices.Configure(bmc)
	tearDown()
}

func TestUpdateCredentials(t *testing.T) {
	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	bmc.UpdateCredentials("newUsername", "newPassword")

	if bmc.username != "newUsername" {
		t.Fatalf("Expected username to be updated to 'newUsername' but is: %s", bmc.username)
	}

	if bmc.password != "newPassword" {
		t.Fatalf("Expected password to be updated to 'newPassword' but is: %s", bmc.password)
	}

	tearDown()
}
