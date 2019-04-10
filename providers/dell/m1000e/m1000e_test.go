package m1000e

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/spf13/viper"
)

var (
	mux                *http.ServeMux
	server             *httptest.Server
	dellChassisAnswers = map[string]map[string][]byte{
		"/cgi-bin/webcgi/general": {
			"default": []byte(`<input xmlns="" type="hidden" value="2a17b6d37baa526b75e06243d34763da" name="ST2" id="ST2" />`),
		},
		"/cgi-bin/webcgi/login": {
			"default": []byte(``),
		},
		"/cgi-bin/webcgi/logout": {
			"default": []byte(``),
		},
		"/cgi-bin/webcgi/pwr_redundancy": {
			"default": []byte(`<?xml version="1.0"?>
				<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/html4/strict.dtd">
				<html xmlns="http://www.w3.org/1999/xhtml" xmlns:fo="http://www.w3.org/1999/XSL/Format">
				  <head>
					<meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
					<title>Budget/Redundancy Configuration</title>
					<link rel="stylesheet" type="text/css" href="/cmc/css/stylesheet.css?0332" />
					<script type="text/javascript" src="/cmc/js/prototype.js?0332"></script>
					<script type="text/javascript" src="/cmc/js/Clarity.js?0332"></script>
					<script type="text/javascript" src="/cmc/js/validate.js?0332"></script>
					<script type="text/javascript" src="/cmc/js/context_help.js?0332"></script>
					<script type="text/javascript">
						  UpdateHelpIdAndState(context_help("Budget Redundancy Configuration"), true);
				
						  var strNoSelectionsMade             = "No user input or modifications detected, and no changes have been applied.";
						  var strStatusBudgetLowPerformance   = "System Input Power Cap setting too low. This will impact server performance.";
						  var strForceACPowerBudgetSetting    = "Do you want to force this setting?";
						  var strUPSWarning                   = "Enabling Max Power Conservation Mode will deactivate Extended Power Performance.\n\nMax Power Conservation Mode option will force servers into a low power, limited performance mode and disable server power up.\n\nPress OK to continue.";
						  var strEPPUnsupportedWarningPart1   = "System Input Power Cap cannot be set to less than or equal to";
						  var strEPPUnsupportedWarningPart2   = "while Extended Power Performance is enabled.";
						  var strBlacklisted1                 = "Enhanced DPSE is not supported by the current power supply configuration.";
						  var strBlacklisted2                 = "DPSE setting will be ignored.";
						  var strPerfOverRedundancyMessage    = "Checking the Server Performance Over Power Redundancy option allows server power allocations to exceed redundant power capabilities of the chassis. Are you sure you want to continue?";
						  var str110VackMessage               = "You are about to allow your chassis to be powered from a 110 volt circuit. This may overload the supply branch circuit. Are you sure you want to do this?";
						  var strUPSModeMessage               = "Checking the Max Power Conservation Mode option will force your servers into a low power, limited performance mode and disables server power up. Are you sure you want to do this?";
						  var strSBPMModeMessage              = "Checking the Server Based Power Management Mode option will set your power cap to max value, server priorities to default priority, and disables Max Power Conservervation Mode. Are you sure you want to continue?";
						  var strInvalidRemoteLoggingInterval = "Invalid remote logging interval. Valid range is from 1 to 1440 minutes.";
						  var strECMEnabled                   = "0";
						  var strUPSEcmAndEppWarning          = "Enabling Max Power Conservation Mode will deactivate Extended Power Performance and Enhanced Cooling Mode.\n\nMax Power Conservation Mode option will force servers into a low power, limited performance mode and disable server power up.\n\nPress OK to continue.";
						  var strUPSEcmWarning                = "Enabling Max Power Conservation Mode will deactivate Enhanced Cooling Mode.\n\nMax Power Conservation Mode option will force servers into a low power, limited performance mode and disable server power up.\n\nPress OK to continue.";
						  var strStatusValInvalid             = "Property Value Invalid. Try Again.";
						  var outOfRange = 0;
						  var max = 16685;              // MMS_PWRMGMT_BUDGET_AC_MAX as defined in pwrmgmt_interface.h
				
						  
						  
							function hasUserInputOrModifiedAnything()
							{
							  var pCount= document.getElementById("pCount").value;
							  var vCount = 0;
							  var propertyName;
							  var propertyValue;
							  var propertyCurrentValue;
							  var modified = false;
				
							  for (var i = 12; i <= pCount; i++) // Skip first 11 parameters
							  {
								propertyName = document.getElementById("p" + i.toString()).value;
								vCount = document.getElementById("vCount" + propertyName).value;
								if (propertyName != "ChassisPSUFailure")
								{ 
								propertyValue = Number(document.getElementById("v" + propertyName + "1").value);
								propertyCurrentValue = Number(document.getElementById(propertyName + "1").value);
				
								if (propertyCurrentValue != propertyValue)
								{
								  if (vCount > 1)
								  {
									document.getElementById(propertyName + "1").maxLength = document.getElementById(propertyName + "1").maxLength + 1; // FF work-around
									document.getElementById(propertyName + "1").value = propertyCurrentValue + "*"; // The '*' token is appended to allow CGI to identify which paramter/unit has been modified.
								  }
								  modified = true;
								}
							   }
							  }
							  return modified;
							}
				
							function formSubmit()
							{
							  // old valuesoldpsuinputpowercap
							  var oldEPPEnable=GetOriginalValue('CHASSIS_POWER_epp_enable');
							  var oldPerfMode=GetOriginalValue('CHASSIS_POWER_performance_over_redundancy');
							  var origDPSE = GetOriginalValue('psuDynEng');
							  var olddpse = Number(document.getElementById("vpsuDynEng1").value);
							  var old110V = GetOriginalValue('CHASSIS_POWER_110V_acknowledge');
							  var oldUPSMode=GetOriginalValue('CHASSIS_POWER_UPSMode');
							  var oldSBPMMode=GetOriginalValue('CHASSIS_POWER_SBPMMode');
							  var oldpsuinputpowercap = GetOriginalValue("acPowerBudget");
							  var oldpsuredundancy = GetOriginalValue("psuRedundancy");
							  // MMS_PWRMGMT_BUDGET_AC_MAX as defined in pwrmgmt_interface.h. ugg this is so bad to have at this level.  I hate this part.
							  // these details should be lower then GUI. -lbt
				
							  // new values
							  var enhNotSupported = Number(document.getElementById("vpsuAnyBlacklisted1").value);
							  var newEPPEnable=$("CHASSIS_POWER_epp_enable1").value;
							  var newPerfMode=$("CHASSIS_POWER_performance_over_redundancy1").value;
							  var dpse = Number(document.getElementById("psuDynEng1").value);
							  var ele=document.getElementById("CHASSIS_POWER_button_disable1");
							  ele.focus();
							  var new110V=$("CHASSIS_POWER_110V_acknowledge1").value;
							  var newUPSMode=$("CHASSIS_POWER_UPSMode1").value;
							  var newSBPMMode=$("CHASSIS_POWER_SBPMMode1").value;
							  var remoteLoggingInterval = $("CHASSIS_POWER_remote_logging_interval1").value;
							  var newpsuinputpowercap = $("acPowerBudget1").value;
							  var newpsuredundancy = $("psuRedundancy1").value;
							  var oldPowerCapPercentage = 0;
				
							  if(outOfRange == 1)
							  {
								 alert(strStatusValInvalid);
								 //did this to make sure we stay current with what the backend has the power at.
								 $("acPowerBudget1").value = oldpsuinputpowercap;
								 oldPowerCapPercentage = Math.round((oldpsuinputpowercap/max)*100);
								 $("acPowerBudget3").value = oldPowerCapPercentage;
							   }
							  else
							  {
								if(hasUserInputOrModifiedAnything() != 1)
								{
								  alert(strNoSelectionsMade);
								  return;
								}
							  }
				
							  if (dpse && (dpse != origDPSE )) {
								if (enhNotSupported) {
								  alert(strBlacklisted1 + "\n" + strBlacklisted2);
								}
							  }
				
							  if ( (old110V != new110V) && (new110V != 0) ) {
								if (!confirm(str110VackMessage)) return;
							  }
				
							  if (oldUPSMode != newUPSMode) {
								if(strECMEnabled != 0 && newEPPEnable != 0 && newUPSMode != 0) //ECM and EPP enable
								{
								  if (!confirm(strUPSEcmAndEppWarning)) return;
								}
								else if (newEPPEnable != 0 && newUPSMode != 0 && strECMEnabled != 1) //EPP enable and ECM disable
								{
								  if (!confirm(strUPSWarning)) return;
								}
								else if(strECMEnabled != 0  && newUPSMode != 0 && newEPPEnable != 1)//ECM enable and EPP disable
								{
								  if (!confirm(strUPSEcmWarning)) return;
								}
							  }
				
							  if (newEPPEnable != 0 && newpsuinputpowercap <= GetOriginalValue("EPPUpperCap")) {
								alert(strEPPUnsupportedWarningPart1 + " " + GetOriginalValue("EPPUpperCap") + " W (" +
								  // BTU_H_PER_WATTS 3.4121411564884; as defined in pwrmgmt_interface.h;
								  + (GetOriginalValue("EPPUpperCap") * 3.4121411564884).toFixed(0) + " BTU/h) " + strEPPUnsupportedWarningPart2);
				
								// Restore value of the input box
								// since hasUserInputOrModifiedAnything()
								// added an asterix to it.
				
								if(outOfRange == 0)
								{
								  $("acPowerBudget1").value = newpsuinputpowercap;
								  convertUnits($("acPowerBudget1"));
								}
								return;
							  }
				
							  if (strECMEnabled != 1 && newEPPEnable == 0 && (oldUPSMode != newUPSMode) && newUPSMode != 0) {
								if (!confirm(strUPSModeMessage)) return;
							  }
				
							  if ( (oldSBPMMode != newSBPMMode) && newSBPMMode != 0) {
								if (!confirm(strSBPMModeMessage)) return;
							  }
				
							  if ( (oldPerfMode != newPerfMode) && newPerfMode != 0)
							  {
								if (!confirm(strPerfOverRedundancyMessage)) return;
							  }
				
							  if ( remoteLoggingInterval < 1 || remoteLoggingInterval > 1440 )
							  {
								alert( strInvalidRemoteLoggingInterval );
								return;
							  }
				
				
							  // If the the EPP checkbox is set,
							  // enable it before submiting the form,
							  // otherwise if the checkbox was disabled,
							  // the CGI layer receive a value 0, instead of 1.
							  if ($("CHASSIS_POWER_epp_enable1").checked)
								$("CHASSIS_POWER_epp_enable1").disabled = false;
				
							  if(outOfRange == 0)
								document.dataarea.submit();
							}
				
							function selectProperty(property)
							{
							  if (property.checked)
							  {
								property.value = 1;
							  }
							  else
							  {
								property.value = 0;
							  }
							  updateDependents( property );
							}
				
							function updateDependents( property )
							{
							  if ( property.name == "CHASSIS_POWER_remote_logging_enable1" )
							  {
								if ( property.checked )
								{
								  $( "CHASSIS_POWER_remote_logging_interval1" ).disabled = false;
								}
								else
								{
								  $( "CHASSIS_POWER_remote_logging_interval1" ).disabled = true;
								}
							  }
							  if ( property.name == "CHASSIS_POWER_SBPMMode1" )
							  {
								if ( property.checked )
								{
								  $( "CHASSIS_POWER_UPSMode1" ).checked = false;
								  $( "CHASSIS_POWER_UPSMode1" ).disabled = true;
								  $( "acPowerBudget1" ).disabled = true;
								  $( "acPowerBudget2" ).disabled = true;
								  $( "acPowerBudget3" ).disabled = true;
								}
								else
								{
								  $( "CHASSIS_POWER_UPSMode1" ).disabled = false;
								  $( "acPowerBudget1" ).disabled = false;
								  $( "acPowerBudget2" ).disabled = false;
								  $( "acPowerBudget3" ).disabled = false;
								}
							  }
				
							  if ($("CHASSIS_POWER_epp_enable1").checked) {
				
								$("CHASSIS_POWER_SBPMMode1").disabled = true;
								$("psuDynEng1").disabled = true;
								$("psuRedundancy1").options[1].disabled = true;
								$("psuRedundancy1").options[2].disabled = true;
				
							  } else {
				
								$("CHASSIS_POWER_SBPMMode1").disabled = false;
								$("psuDynEng1").disabled = false;
								$("psuRedundancy1").options[1].disabled = false;
								$("psuRedundancy1").options[2].disabled = false;
				
							  }
				
							  if ($("CHASSIS_POWER_SBPMMode1").checked ||
								$("acPowerBudget1").value <= GetOriginalValue("EPPUpperCap") ||
								  $("psuRedundancy1").value != 1 || $("psuDynEng1").checked ||
									$("CHASSIS_POWER_UPSMode1").checked ||
									  GetOriginalValue("CHASSIS_type_freshair").value == 1 ||
										GetOriginalValue("allPSUsEPPCapable") != 1 || $("vChassisPSUFailure1").value == 1) {
				
								$("CHASSIS_POWER_epp_enable1").disabled = true;
				
							  } else {
				
								$("CHASSIS_POWER_epp_enable1").disabled = false;
				
							  }
							}
				
							function isNumericKey(e)
							{
							  try {
								if (!e) var e = window.event;
								var k = document.all ? e.keyCode : e.which;
								if ((k > 47 && k < 58) || k == 8 || k == 0) {
								  return true;
								}
								else { // Explicitly cancel the event on IE
								  e.preventDefault ? e.preventDefault() : e.returnValue = false;
								  return false;
								}
							  } catch(e) { }
							}
				
							function extractNumeric(property)
							{
							  try {
								return property.value.replace(/\D/g,"");
							  }
							  catch(e) { }
							}
				
						   function onKeyUp(obj)
							{
							  try
							  {
								//this doesn't just convert unit is sets the power cap items with respect to each other...it syncs the values
								convertUnits(obj);
							  }
							  catch(e) { }
							}
				
							//this logic should totally not be in the GUI level. -lbt
							function convertUnits(property)
							{
							  var m = 3.4121411564884;      // BTU_H_PER_WATTS as defined in pwrmgmt_interface.h
							  var min = 2715;               // MMS_PWRMGMT_BUDGET_AC_MIN as defined in pwrmgmt_interface.h
							  var n = property.name.substr(0, property.name.length - 1);
							  var c = document.getElementById("vCount" + n).value;
							  var p = property.name.substring(n.length);
							  var v = Number(property.value);
				
							  try {
								if (p == 1)                                                                                          // Watts
								{
								  document.getElementById(n + "2").value = Math.round(v * m);                                        // BTU/h
								  if (c == 3)
								  {
									var e = Math.round(v * 100 / max);
									document.getElementById(n + "3").value = (e >= 0 && e <= 100) ? e : '0';     // Percent
								  }
								}
								else if (p == 2)                                                                                     // BTU/h
								{
								  var t = Math.round(v / m);
								  document.getElementById(n + "1").value = t;                                                        // Watts
								  if (c == 3)
								  {
									var e = Math.round(t * 100 / max);
									document.getElementById(n + "3").value = (e >= 0 && e <= 100) ? e : '0';     // Percent
								  }
								}
								else if (p == 3)                                                      // Percent
								{
								  var t = Math.round(v * max / 100);
								  if ( v == Math.round(min/max * 100))
								  {
									 t = min;
								  }
								  document.getElementById(n + "1").value = t;                                                        // Watts
								  document.getElementById(n + "2").value = Math.round(t * m);                                        // BTU/h
								}
							  } catch(e) { }
							}
				
							// This function is a kind of override of the function contained in the GetCookieStatusScripts template in common.xsl to add the special handling needed for CR208270.
							function overrideDisplayStatus()
							{
							  var status = Get_Cookie("status");
							  var rcStatus = Get_Cookie("rcStatus");
				
							  var pName = Get_Cookie("pName");
							  var pValue = Get_Cookie("pValue");
							  if (status != null)
							  {
								if (rcStatus == 0x5517) // Special handling as per CR208270 : Setting for System Input Power Cap too low. This will affect server performance. Note the hard-wired error code.
								{
								  if (confirm(strStatusBudgetLowPerformance + '\n' +
											  strForceACPowerBudgetSetting))
								  {
									if (pName != null)
									{
									  document.getElementById(pName).maxLength = document.getElementById(pName).maxLength + 1; // FF work-around
									  document.getElementById(pName).value = pValue.replace('*', '#');  // The hash sign will be used to guide the cgi code as to which MMS power function to call.
				
									  document.dataarea.submit();
									 }
								  }
								}
								else
								{
								   alert(TranslateStatus(rcStatus)); // Display other status messages and error codes as normal.
								}
							  }
				
							  Delete_Cookie("pValue", "", "");
							  Delete_Cookie("pName", "", "");
							  Delete_Cookie("status", "", "");
							  Delete_Cookie("rcStatus", "", "");
							  FirstFocus();
							}
				
							function GetOriginalValue(propName)
							{
							  var pCount= document.getElementById("pCount").value;
				
							  for (var i = 4; i <= pCount; i++) // Skip first three parameters
							  {
								var propertyName = document.getElementById("p" + i.toString()).value;
								if (propName == propertyName)
								{
								  var vCount = document.getElementById("vCount" + propertyName).value;
								  var propertyValue = Number(document.getElementById("v" + propertyName + "1").value);
								  return propertyValue;
								}
							  }
							}
				
							
						  </script>
					<script xmlns="">
					  var vDesc = "Detailed Description:";
					  var vAction = "Recommended Action:";
					  var vExtension = "Extension of ";
					  var vunMapped = "Unmapped";
					  var vIsNoble      = "1";
					  
				
				  
					  function common_init(affirm)
					  {
						DisplayStatus(affirm);
				
						// Check for the RESET cookie.
						var reset = Get_Cookie("RESET_CMC");
						if (reset == 1)
						{
						  Delete_Cookie("RESET_CMC", "", "");
						  Set_Cookie( "logout", 'common.xsl: Reset CMC', 2, "", "");
						  top.document.location.replace("/cgi-bin/webcgi/logout");
						  Delete_Cookie("sid", "", ""); // comes after so that the session id is cleared from SSN cache
						}
						FirstFocus();
					  }
				
					  function Get_Cookie( name )
					  {
						var start = document.cookie.indexOf( name + "=" );
						var len = start + name.length + 1;
						if ( ( !start ) &&
						( name != document.cookie.substring( 0, name.length ) ) )
						{
						return null;
						}
						if ( start == -1 ) return null;
						var end = document.cookie.indexOf( ";", len );
						if ( end == -1 ) end = document.cookie.length;
						return unescape( document.cookie.substring( len, end ) );
					  }
				
					  function Delete_Cookie( name, path, domain ) {
						if ( Get_Cookie( name ) ) document.cookie = name + "=" +
						( ( path ) ? ";path=" + path : "") +
						( ( domain ) ? ";domain=" + domain : "" ) +
						";expires=Thu, 01-Jan-1970 00:00:01 GMT";
				
						// the cookie didn't get deleted so maybe we need to specify the path with an ending '/'
						if ( Get_Cookie( name ) && (path.match(/\/$/) == null) ) document.cookie = name + "=" +
						( ( path ) ? ";path=" + path + "/" : "") +
						( ( domain ) ? ";domain=" + domain : "" ) +
						";expires=Thu, 01-Jan-1970 00:00:01 GMT";
					  }
				
					  function Set_Cookie( name, value, days, path, domain )
					  {
						var expires = "";
						if (days)
						{
						  var date = new Date();
						  date.setTime(date.getTime()+(days*86400000));
						  expires = "; expires="+date.toGMTString();
						}
						document.cookie = name + "=" + unescape(value) + expires +
						  ( ( path ) ? ";path=" + path : "") +
						  ( ( domain ) ? ";domain=" + domain : "" ) + expires
					  }
				
					  function DisplayStatus(affirm)
					  {
						var status = Get_Cookie("status");
						var rcStatus = Get_Cookie("rcStatus");
				
						if ( status != null)
						{
						  if ( !( (affirm == "noAffirm") && (rcStatus == 0)))
						  {
						  alert(TranslateStatus(rcStatus));
						  }
						}
				
						Delete_Cookie("status", "", "");
						Delete_Cookie("rcStatus", "", "");
					  }
				
					  function DisplayAlert(msg)
					  {
						alert(htmlDecodeString(msg));
					  }
				
					  function DisplayMsgDescAction(msg, desc, action)
					  {
						var strMsg = msg;
						if ((desc != null) && (desc != "")) {
						  strMsg = strMsg.concat('\n\n' + vDesc + '\n' + desc);
						}
						if ((action != null) && (action != "")) {
						  strMsg = strMsg.concat('\n\n' + vAction + '\n'+ action);
						}
						DisplayAlert(strMsg);
					  }
				
					  //************************************************
					  // Initializa PCIE object Array
					  //************************************************
					  function FormatServerNamesForFullHeight(useHostName, serverSlot, slotType, serverName)
					  {
						var serverString = "";
				
						//Need to check for only slot 3 and 4
						if(serverSlot == 3 || serverSlot == 4)
						{
						  //slotType 2 for serverSlot 3, means it is full Height in slot 1 else half
						  //slotType 2 for serverSlot 4, means it is full Height in slot 2 else half
						  if(slotType == 2)
						  {
							serverString = vExtension + " " + (serverSlot-2);
						  }
						  else
						  {
							serverString = serverName;
						  }
						}
						else
						{
							serverString = serverName;
						}
				
						return serverString;
					  }
				
					  //Referred cmc_src\cmc_open\osabs\include\cmc_systemid.h to have System ID's for various CMC systems
					  //Adding similar functions for xslt/javaScript check to differentiate between different system.
					  var SYSID_NOBLE_CMC_CHASSIS = "0x0000";
					  var SYSID_PLASMA_CMC_CHASSIS = "0x0001";
					  var SYSID_PLASMA_CMC_EB = "0x0002"; //Carrier card
					  var SYSID_PLASMA_CMC_EB_PCBATEST = "0x0003"; //Carrier PCBA test
					  var SYSID_NOBLE_CMC_NEB = "0x0004"; //Carrier card
					  var SYSID_STOMP_CMC_CHASSIS_PSB = "0x0010";
					  var SYSID_STOMP_CMC_EB = "0x0020"; //Carrier card
					  var SYSID_STOMP_CMC_EB_PCBATEST = "0x0021"; //Carrier PCBA test
				
					  function IsNoble(sysid)
					  {
						var sysFlag = false;
						if((sysid == SYSID_NOBLE_CMC_CHASSIS) || (sysid == SYSID_NOBLE_CMC_NEB))
						{
						   sysFlag = true;
						}
						return sysFlag;
					  }
				
					  function IsPlasma(sysid)
					  {
						var sysFlag = false;
						if((sysid == SYSID_PLASMA_CMC_CHASSIS) || (sysid == SYSID_PLASMA_CMC_EB) || (sysid == SYSID_PLASMA_CMC_EB_PCBATEST))
						{
						   sysFlag = true;
						}
						return sysFlag;
				
					  }
				
					  function IsStomp(sysid)
					  {
						var sysFlag = false;
						if((sysid == SYSID_STOMP_CMC_CHASSIS_PSB) || (sysid == SYSID_STOMP_CMC_EB) || (sysid == SYSID_STOMP_CMC_EB_PCBATEST))
						{
						   sysFlag = true;
						}
						return sysFlag;
				
					  }
					
				
				
					  function TranslateStatus(v)
					  {
						switch(v*1)
						{
						  case 0:
							return ("Operation Successful.");
				
						  // Colossus Exposed Completion codes.
						  case 0x0027: // IPMI Command Downgraded Fabric
							return ("Error selecting fabric. Make sure the fabric's IOMs support 10Gb.");
						  case 0x0028: // IPMI Command Downgraded Midplane.
							return ("M1000e Midplane must be upgraded to support 10Gb on Fabric-A");
						  case 0x0029: // IPMI Command Downgraded Midplane.
							return ("Error selecting fabric. Make sure the fabric has 10Gb IOMs installed.");
						  case 0x00D5: // IPMI Command execution Failure.
							return ("The IOMs currently installed in Fabric A do not support this system");
				
						  // fwupdate Errors
						  case 0x1400: // FUP_ERR_NO_PRIVILEGE                                     //  0x1400 == 5120
							return ("User does not have permission to perform update");
						  case 0x1407: // FUP_ERR_MX_BAD_PARAM                                     //  0x1400 == 5127
							return ("File Not Found!");
						  case 0x1409: // FUP_INVALID_REQUEST_PREV_UPDATE_IN_PROGRESS              //  0x1409 == 5129
							return ("Previous update still in progress");
						  case 0x140B: // FUP_DEVICE_NOT_AVAILABLE                                 //  0x140B == 5131
							return ("Device unavailable for update");
						  case 0x140D: // FUP_DEVICE_PAYLOAD_TOO_BIG                               //  0x140D == 5133
							return ("Invalid firmware: The uploaded firmware image is not valid on this hardware.\n\nRetry the operation with a valid firmware image.");
						  case 0x1411: // FUP_ERR_TARGET_NOT_READY                                 // 0x1411 == 5137
							return ("Target not ready for update");
						  case 0x1416: // FUP_ERR_OP_NOT_CANCELABLE                                // 0x1416 == 5142
							return ("Cannot cancel update");
						  case 0x141C: // FUP_ERR_IMAGE_FILE_NOT_ACCESSIBLE                        // 0x141C == 5148
							return ("Image file not transferred from host");
						  case 0x1440: // FUP_ERR_CLIENT_CONNECT               // 0x1440 == 5184
						  case 0x1460: // FUP_ST_START_INVALID_REQ             // 0x1460 == 5216
						  case 0x1461: // FUP_ST_BOUND_INVALID_REQ,            // 0x1461 == 5217
						  case 0x1462: // FUP_ST_RESERVED_INVALID_REQ,         // 0x1462 == 5218
						  case 0x1463: // FUP_ST_CONNECTED_INVALID_REQ,        // 0x1463 == 5219
							return ("Firmware update operation temporarily unavailable " + v);
						  case 0x1464: // FUP_ST_INITIALIZING_INVALID_REQ,     // 0x1464 == 5220
							return ("A firmware update operation is already in progress");
				
						  // cfg Manager Errors
						  case 0x2E3E: // ERRNO_CFGMGR_LIB_RC_INVALID_ARGS
						  case 11776: // ERRNO_CFGMGR_LIB_RC_INVALID_ARGS
						  case 11805: // ERRNO_CFGMGR_LIB_RC_INVALID_INDEX
						  case 11837: // ERRNO_CFGMGR_CMCCFG_RC_ARGS
							return ("Property Argument List Error");
						  case 11814: // ERRNO_CFGMGR_CFG_RC_NOTFOUND
							return ("Property Not Found");
						  case 11838: // ERRNO_CFGMGR_CMCCFG_RC_LASTPROP
							return ("Operation Successful.");
						  case 11839: // ERRNO_CFGMGR_CMCCFG_RC_NORESPONSE
						  case 11781: // ERRNO_CFGMGR_LIB_RC_REQUEST_FAILED
							return ("No Response from Configuration Manager.");
						  case 11782: // ERRNO_CFGMGR_LIB_RC_TIMEOUT
						  case 11820: // ERRNO_CFGMGR_CFG_RC_EINPROGRESS
							return ("The requested configuration change is in progress. Please wait for completion.");
						  case 11840: // ERRNO_CFGMGR_CMCCFG_RC_VALUE_TOOBIG
						  case 11807: // ERRNO_CFGMGR_CFG_RC_INVALID_LENGTH
						  case 11817: // ERRNO_CFGMGR_CFG_RC_BUF_TOOSMALL
							return ("Buffer is invalid for this property.");
						  case 11808: // ERRNO_CFGMGR_CFG_RC_INVALID_VALUE
							return ("Property Value Invalid. Try Again.");
						  case 11818: // ERRNO_CFGMGR_CFG_RC_READONLY_VALUE
							return ("Property Value cannot be modified.");
						  case 11841: // ERRNO_CFGMGR_CMCCFG_RC_AUTHENTICATION_ERROR
							return ("User Authorization Error");
						  case 11842: // ERRNO_CFGMGR_CMCCFG_RC_UNAUTHORIZED
							return ("User not authorized for this action.");
						  case 12045: // ERRNO_CFGMGR_FC_ERR_BLADE_CONFLICT
							return ("Server(s) Must Be Powered Down Before Changing this Property.");
				
						  // Blade Manager Errors.
						  case 0x5200: // BM_ERR_MEM_ACQUIRE_FAILED=BM_BLADE_ERR_BASE
							return ("Failed to Acquire Memory for Operation.");
						  case 0x5201: // BM_ERR_INVALID_BLADE_NUM
							return ("Invalid server number");
						  case 0x5202: // BM_ERR_BLADE_ABSENT
							return ("Server Absent");
						  case 0x5108: // IPMI_ERR_IMC_NOT_READY
						  case 0x5203: // BM_ERR_IMC_NOT_READY
							return ("iDRAC Not Ready. See CMC Log for Details.");
						  case 0x5204: // BM_ERR_BLD_PWRD_ON
							return ("Server already powered on");
						  case 0x5205: // BM_ERR_IMC_CMD_FAILED
							return ("Error while sending command to iDRAC");
						  case 0x5206: // BM_ERR_INVALID_SLOT_NAME
							return ("Invalid slot name");
						  case 0x5207: // BM_ERR_INVALID_SLOT_PRIORITY
							return ("Invalid slot priority");
						  case 0x5208: // BM_ERR_INVALID_INPUT
							return ("Invalid input value.");
						  case 0x5209: // BM_ERR_HWABS_COMMN_FAILED
							return ("Error while communicating with instrumentation.");
						  case 0x520A: // BM_ERR_BLD_INFO_POP
							return ("Error populating server information.");
						  case 0x520B: // BM_ERR_INVALID_USR
							return ("User not authorized for this action.");
						  case 0x520C: // BM_ERR_ACQUIRE_EDID
							return ("Cannot Obtain Server Extended Display Identification Data (EDID)");
						  case 0x520D: // BM_ERR_INVALID_PASSWD_LEN
							return ("Invalid iDRAC Password");
						  case 0x520E: // BM_ERR_SHM_ACCESS
							return ("Hardware Error: Cannot obtain Mutex");
						  case 0x520F: // BM_ERR_READING_REQUEST,
							return ("Cannot communicate with Server.");
						  case 0x5210: // BM_ERR_ACTN_PROGRESS
							return ("Another action is being performed on this server. Try again.");
						  case 0x5211: // BM_ERR_BLD_PWRD_OFF
							return ("Server already powered off.");
						  case 0x5212: // BM_ERR_DUPLICATE_IP_ADDR
							return ("IP Address is already used on another iDRAC.");
						  case 0x5213: // BM_ERR_INVALID_SLOT_LEN
							return ("Slot Name length is invalid.");
						  case 0x5214: // BM_ERR_DUP_SLOT_NAME
							return ("Slot Name is already used.");
						  case 0x5215: // BM_ERR_CHASSIS_PWR_OFF
							return ("Cannot process command while chassis is off.");
						  case 0x5218: // BM_ERR_START_IP_OUT_OF_RANGE
							return ("Starting IP address is out of range.");
						  case 0x5219: // BM_ERR_FEATURE_UNAVAILABLE
							return ("Server feature unavailable.");
						  case 0x522D: // BM_ERR_STASH_PWR_ON
							return ("Cannot process command while the storage array is powered ON.");
				
						  // enum ce_error_t -- Relocated to base 0xC000 - Only listing the likely errors.
						  case 0xC006: // CE_STATUS_SSN_VALIDATION_FAILED
							return ("Not authorized to complete this action.");
						  case 0xC007: // CE_STATUS_SSN_ID_MISSING
							return ("Session credentials missing, cannot complete this action.");
						  case 0xC00A: // CE_STATUS_FILE_TOO_LARGE
							return ("File is too large, try again.");
						  case 0xC010: // CE_STATUS_INSUFFICIENT_PRIVILEGES
							return ("Not authorized to complete this action.");
						  case 0xC011: // CE_STATUS_TEST_EMAIL_FAILED
							return ("Unable to send test email.\nMake sure email alerts have been configured correctly\nand connectivity to the SMTP server exists.");
						  case 0xC012: // CE_STATUS_NO_DATA_RETURNED
							return ("Problem reading data.");
						  case 0xC014: // CE_STATUS_BACKUP_RESTORE_BUSY
							return ("Save or restore operation already in progress.");
						  case 0xC015: // CE_STATUS_TMP_FILESYSTEM_PROBLEM
							return ("Tmp filesystem not functional.");
						  case 0xC016: // CE_STATUS_ENCRYPTION_FAILED
							return ("Encryption or decryption failed.");
						  case 0xC017: // CE_STATUS_BACKUP_FAILED
							return ("Backup or restore failed.");
						  case 0xC018: // CE_STATUS_BACKUP_FILE_NOT_FOUND
							return ("File not found.");
						  case 0xC019: // CE_STATUS_FILE_OPEN
							return ("Cannot Open File");
						  case 0xC01A: // CE_STATUS_DATA_WRITE_PROBLEM
							return ("Data write failed.");
						  case 0xC01B: // CE_STATUS_NV_PREPARE_FAIL
							return ("Media preparation failed.");
						  case 0xC01C: // CE_STATUS_BACKUP_FILE_INVALID
							return ("Invalid file for restore operation.");
						  case 0xC01D: // CE_STATUS_NV_UNAVAILABLE
							return ("The EEPROM is unavailable. Contact Service.");
				
						  // cgic.h errors -- Relocated to base 0xC100
						  case 0xC100: // cgiFormSuccess
							return ("Operation Successful.");
						  case 0xC101: // cgiFormTruncated,
							return ("Name Truncated Due to Length.");
						  case 0xC102: // cgiFormBadType,
							return ("Invalid Value Specified.");
						  case 0xC103: // cgiFormEmpty,
							return ("No Data Specified.");
						  case 0xC104: // cgiFormNotFound,
							return ("No Data Found.");
						  case 0xC105: // cgiFormConstrained,
							return ("Value out-of-bounds.");
						  case 0xC106: // cgiFormNoSuchChoice,
							return ("Button Choice not Applicable.");
						  case 0xC107: // cgiFormMemory,
							return ("Out of Memory.");
						  case 0xC108: // cgiFormNoFileName,
							return ("No File Specified");
						  case 0xC209: // CE_STATUS_FILE_INVALID
						  case 0xC109: // cgiFormNoContentType,
							return ("Bad File Type.");
						  case 0xC10A: // cgiFormNotAFile,
							return ("File Not Found!");
						  case 0xC10B: // cgiFormOpenFailed,
							return ("Cannot Open File");
						  case 0xC10C: // cgiFormIO,
							return ("System I/O Error Occurred.");
						  case 0xC10D: // cgiFormEOF
							return ("Reached End-Of-File");
						  case 0xC110: // Enhance Chassis Log mode changed
							return ("Chassis Events setting fails. Enhanced Chassis Logging mode has been changed externally.");
				
						  //SSL Errors
						  case 0xC200: // invalid cert
							return ("Bad File Format. Try Again.");
						  case 0xC20A: // invalid cert
							return ("Bad File Format. Try Again.");
				
						  //System Lockdown Error
						  case 0xd4: // System Lockdown mode enabled
							if (vIsNoble == 1) {
							   return ("Lockdown Mode enabled. Unable to apply any changes. Refer to CMC log for Lockdown Mode enabled systems.");
							}
							else {
							   return ("Lockdown Mode enabled. Unable to apply any changes. Refer to Chassis log for Lockdown Mode enabled systems.");
							}
				
						  // IOM Errors
						  case 0x5406: // IOM_ERR_NO_IOM,
							return ("No IOM is present in this slot.");
						  case 0x5407: // IOM_ERR_IOM_OFF,
							return ("Cannot process the command while the I/O module is off.");
						  case 0x5408: // IOM_ERR_NO_POWER,
							return ("Insufficient Power Budget. Check Chassis Power Management.");
						  case 0x5409: // IOM_ERR_POWER_MGMT,
							return ("Power Management Sub-System is initializing. Try Again.");
						  case 0x540A: // IOM_ERR_PSOC_ACCESS,
						  case 0x5007: // MMS_ERR_IPC_RCV_TIMEOUT
							return ("Communication Error with IOM. Try Again.");
						  case 0x540B: // IOM_ERR_FCC,            // error due to fabric mismatch
							return ("IOM Fabric Mismatch. Cannot Perform Action.");
						  case 0x540C: // IOM_ERR_LINKT,          // link tuning mismatch
							return ("Link tuning data unavailable for the current IOM configuration. Cannot perform the action.");
						  case 0x540D: // IOM_ERR_LINKT_DATA,     // link tuning invalid data length
							return ("Bad Link Tuning Data. Cannot Perform Action.");
						  case 0x5416: // THREAD_IN_PROGRESS,     // VLAN update in progress
							return ("IOM update in progress. Wait for update to complete before applying additional settings.");
						  case 0x5418: // IOM_ERR_XML_OPERATION_STARTED
							return ("The requested operation has started. The operation will take some time to complete, failures can be viewed in the CMC Log.");
				
						  // Power Management Error Codes
						  case 0x5502: // MMS_PWRMGMT_INV_ARGS  (0x5502)
							return ("Property Argument List Error");
						  case 0x5503: // MMS_PWRMGMT_INV_PERMISSION (0x5503)
							return ("Not authorized to complete this action.");
						  case 0x5504: // MMS_PWRMGMT_INV_REQ  (0x5504)
							return ("Property Argument List Error");
						  case 0x5505: // MMS_PWRMGMT_TIMEOUT  (0x5505)
							return ("No Response from Configuration Manager.");
						  case 0x5506: // MMS_PWRMGMT_NOMEM  (0x5506)
							return ("Failed to Acquire Memory for Operation.");
						  case 0x5507: // MMS_PWRMGMT_ERR_SHM_ACCESS (0x5507)
						  case 0x550D: // MMS_PWRMGMT_ERR_SHM_CORRUPTION (0x550D)
							return ("Hardware Error: Cannot obtain Mutex");
						  case 0x5508: // MMS_PWRMGMT_SOCKET_SEND_ERROR (0x5508)
						  case 0x5509: // MMS_PWRMGMT_SOCKET_RECV_ERROR (0x5509)
							return ("No Response from Configuration Manager.");
						  case 0x550A: // MMS_PWRMGMT_ERR_NULL_SHM_ADDR (0x550A)
							return ("Hardware Error: Cannot obtain Mutex");
						  case 0x5501: // MMS_PWRMGMT_FAILED  (0x5501)
						  case 0x550C: // MMS_PWRMGMT_HAPI_FAILED  (0x550C)
						  case 0x550E: // MMS_PWRMGMT_ERR_BLADE_OP_FAILED (0x550E)
							return ("Error while communicating with instrumentation.");
						  case 0x550F: // MMS_PWRMGMT_PSU_CNT_INVALID
							return ("Error! Insufficient Power Supplies present in the system for the configuration operation to succeed.");
						  case 0x5510: // MMS_PWRMGMT_PSU_CNT_INVALID
							return ("Chassis power-off in progress. Click refresh for current status");
				
						  case 0x5511: // MMS_PWRMGMT_INV_REQ_CHASSIS_ON
							return ("Chassis power-on in progress. Click refresh for current status");
						  case 0x5512: // MMS_PWRMGMT_INV_REQ_CHASSIS_RST
							return ("Chassis reset in progress. Click refresh for current status");
						  case 0x5513: // MMS_PWRMGMT_INV_REQ_CHASSIS_CYC
							return ("Chassis power-cycle in progress. Click refresh for current status");
						  case 0x5514: // MMS_PWRMGMT_CHASSIS_SHUTDN_FAIL
							return ("Chassis power-shutdown failed. Check server status.");
						  case 0x5515: // MMS_PWRMGMT_RETRY_AFTER_60SECS
							return ("Operation in progress. Try again later.");
						  case 0x5517: // MMS_PWRMGMT_BUDGET_LOW_PERFORMANCE
							return ("System Input Power Cap setting too low. This will impact server performance.");
						  case 0x5518: // MMS_PWRMGMT_BUDGET_LOW_SHUTDOWN
							return ("System Input Power Cap setting too low. This will cause server or other components to shut down.");
						  case 0x551A: // MMS_PWRMGMT_FWUPDATE_INPROGRESS
							return ("PSU Update in progress. Try again later.");
						  case 0x5522: // MMS_PWRMGMT_AC_REDUNDANCY_CANNOT_SET
						   return ("Cannot set Grid redundancy due to insufficient PSU in power grids.");
						  case 0x5523: // power cap is set lower than power bound
						   return ("The power cap value cannot be less than the lower power bound");
				
						  // IPv6 address errors
						  case 0x1900: // NET_OP_V6ADRERRNULL       6400
							return ("IPv6 Null address not allowed.");
						  case 0x1901: // NET_OP_V6ADRERRLENGTH     6401
							return ("IPv6 Address too long.");
						  case 0x1902: // NET_OP_V6ADRERRINVCHAR    6402
							return ("IPv6 Address contains invalid character(s).");
						  case 0x1903: // NET_OP_V6ADRERRSEGMENTS   6403
							return ("IPv6 Address contains too many segments.");
						  case 0x1904: // NET_OP_V6ADRERRSEGMENTLEN 6404
							return ("IPv6 Address segment contains more than 4 characters.");
						  case 0x1905: // NET_OP_V6ADRERRDOUBLECOLON  6405
							return ("IPv6 Address contains more than one double colon.");
						  case 0x1906: // NET_OP_V6ADRERRSEGMENTNAN 6406
							return ("IPv6 Address segment is not a number.");
						  case 0x1907: // NET_OP_V6ADRERRTOOSHORT   6407
							return ("IPv6 Address less than 128 bits.");
						  case 0x1908: // NET_OP_V6ADRERRZERO       6408
							return ("IPv6 Address of zero not allowed.");
						  case 0x1909: // NET_OP_V6ADRERRMULTICAST  6409
							return ("IPv6 Multicast address not allowed (FFxx::/8).");
						  case 0x190A: // NET_OP_V6ADRBEGINCOLON    6410
							return ("IPv6 Address cannot begin with a single colon.");
						  case 0x190B: // NET_OP_V6ADRENDCOLON      6411
							return ("IPv6 Address cannot end with a single colon.");
						  case 0x1804: // INVALID IP RANGE ADDRESS   6148
							return ("The IP Range Address is invalid.");
						  case 0x1805: // INVALID IP RANGE MASK      6149
							return ("The IP Range Mask is invalid.");
				
						  // Multi Chassis Management Client Library Errors
				
						  case 0x7000: // MCMCL_ERR_EINVAL               28672
							return ("Invalid data submitted");
						  case 0x7001: // MCMCL_ERR_TIMEOUT              28673
							return ("Communication timeout");
						  case 0x7002: // MCMCL_ERR_IPC                  28674
							return ("Communication error");
						  case 0x7003: // MCMCL_ERR_NORESOURCES          28675
							return ("Insufficient resources");
						  case 0x7004: // MCMCL_ERR_NOSUCH_MEMBER        28676
							return ("No such member exists");
						  case 0x7005: // MCMCL_ERR_TOOMANY_MEMBERS      28677
							return ("Too many members");
						  case 0x7006: // MCMCL_ERR_DUPLICATE_NODE       28678
							return ("Duplicate member IP address or host name");
						  case 0x700B: // MCMCL_ERR_INCONSISTENT_ROLE    28683
							return ("Operation cannot be performed by this role");
						  case 0x700C: // MCMCL_ERR_CFG                  28684
							return ("Configuration storage error");
						  case 0x700D: // MCMCL_ERR_NO_EXT_STORAGE       28685
							return ("Extended storage required");
						  case 0x700E: // MCMCL_ERR_NO_DATA              28686
							return ("Data retrieval error");
						  case 0x700F: // MCMCL_ERR_INSUFFICIENT_PRIV    28687
							return ("Insufficient privileges");
				
						  // Blade Cloning error codes
						  case 0x727b: // BLCL_ERR_SYSCALL_ENOMEDIUM     29307
							return ("Error accessing Extended Storage. Make sure that the Extended Storage feature is enabled and healthy.");
						  case 0x7302: // BLCL_ERR_OP_INPROGRESS         29442
							return ("Server profile operation is already in progress. Wait for the completion of previous operation before starting another one.");
						  case 0x7304: // BLCL_ERR_MAX_PROFILE_LIMIT     29444
							return ("Error Saving Profile. Maximum allowed profiles reached.");
						  case 0x7305: // BLCL_ERR_PROFILE_EXIST         29445
							return ("Profile name already in use.");
						  case 0x7306: // BLCL_ERR_PROFILE_NOT_FOUND     29446
							return ("Profile not found.");
						  case 0x7307: // BLCL_ERR_PROFILE_NO_NAME       29447
							return ("No profile name and description provided.");
				
						  case 0x7400: // BLCL_CANT_READ_CONFIGURATION              29696
							return ("Error accessing the profile.");
						  case 0x7404: // BLCL_CANT_ACCESS_SERVER                   29700
							return ("Error accessing the server.");
						  case 0x7405: // BLCL_CANT_ACCESS_VIEW_DISPLAY_SETTINGS    29701
							return ("Error occurred while extracting viewable settings.");
						  case 0x7409: // BLCL_SERVER_NOT_READY                     29705
							return ("Error occurred while accessing the server - iDRAC is unresponsive. Restart the iDRAC and retry the operation.");
						  case 0x740C: // _BLCL_PROFILE_TOO_LARGE                   29708
							return ("Error Importing Profile - File size is larger than the 2MB maximum. Select a valid profile file.");
						  case 0x740D: // _BLCL_PROFILE_INVALID_FORMAT              29709
							return ("Error Importing Profile - Selected file is not a valid profile. Select a valid profile file.");
						  case 0x740E: // _BLCL_CSIOR_DISABLED                      29710
							return ("Warning: The Collect System Inventory on Restart (CSIOR) control is not enabled for servers targeted by the Blade Cloning operation. Enable CSIOR and retry the operation.");
						  case 0x7414: // _BLCL_REMOTE_SVC                          29716
							return ("Remote Service is not ready.");
				
						  case 0x7480: // VMACDB_NO_SERVICE_TAG                     29824
							return ("Service tag for the selected server is not availabe.");
						  case 0x7481: // VMACDB_INVALID_MAC_ADDR                   29825
							return ("MAC Address format is not valid.");
						  case 0x7484: // VMACDB_MAC_ADD_PARTIAL                    29828
							return ("Some of the MAC addresses are not added to the Virtual MAC Address Pool, because the MAC Addresses exists.");
						  case 0x7485: // VMACDB_MAC_REMOVE_PARTIAL                 29829
							return ("Cannot remove some of the MAC addresses from the pool, as these MAC Addresses do not exist or used in Boot Identity profile.");
						  case 0x7486: // VMACDB_MAC_STATUS_MISMATCH                29830
							return ("MAC address status does not support the requested action.");
						  case 0x7487: // VMACDB_INSUFFICIENT_MAC_ADDRESS           29831
							return ("Sufficient number of MAC addresses not available to capture the profile.");
						  case 0x7489: // VMACDB_OTHER_PROFILE_IN_USE               29833
							return ("If a profile is already applied to the server, ensure that you clear the existing profile identity before applying a new profile.");
						  case 0x748A: // VMACDB_INVALID_BI_SERVER                  29834
							return ("There is no profile associated with the server.");
						  case 0x748F: // VMACDB_ACTIVE_PROFILE_DELETE              29839
							return ("Active profiles cannot be deleted.");
						  case 0x7490: // VMACDB_ENABLE_IOID_PERSISTENCE            29840
							return ("Error in enabling IO ID and Persistence policy for the selected server.");
						  case 0x7483: // VMACDB_MAC_ADDRESS_EXISTS                 29827
							return ("MAC Address exists in the Virtual MAC Address Pool.");
						  case 0x7482: // VMACDB_MAC_ADDRESS_NOT_AVAILABLE          29826
							return ("MAC addresses are not available in the MAC Pool.");
						  case 0x748D: // VMACDB_XML_MACDB_ERROR                    29837
							return ("Virtual MAC Address pool is not valid.");
						  case 0x748B: // VMACDB_INVALID_PROFILE                    29835
							return ("Profile is not valid.");
						  case 0x7491: // VMACDB_WSMAN_FAILED                       29841
							return ("Server communication error.");
						  case 0x7493: // VMACDB_BI_OPS_IN_PROGRESS                 29843
							return ("The Boot Identity operation cannot be started because a Boot Identity operation is already in progress.");
						  case 0x7494: // VMACDB_PROFILE_USED_IN_OTHER_BLADE        29844
							return ("Selected Boot identity profile is already used by some other server.");
				
						  case 0x2018: // SSN_ERR_NO_PRIV           8216
							return ("Not authorized to complete this action.");
				
						  case 0x7500: // LIC000 (base=0x7500)
							return("The License Manager successfully performed the actions.");
						  case 0x7501: // LIC001 (base=0x7500)
							return("The License Manager command parameter used is invalid.");
						  case 0x7502: // LIC002 (base=0x7500)
							return("License Manager is unable to allocate the required resources at startup.");
						  case 0x7503: // LIC003 (base=0x7500)
							return("License Manager was unable to create and/or allocate the required resources.");
						  case 0x7504: // LIC004 (base=0x7500)
							return("An internal system error has occurred.");
						  case 0x7505: // LIC005 (base=0x7500)
							return("Unable to import as the maximum number of licenses are installed.");
						  case 0x7506: // LIC006 (base=0x7500)
							return("The license has expired.");
						  case 0x7507: // LIC007 (base=0x7500)
							return("Invalid entry: Object does not exist or cannot be found.");
						  case 0x7508: // LIC008 (base=0x7500)
							return("The license binding ID does not match the device unique identifier.");
						  case 0x7509: // LIC009 (base=0x7500)
							return("The license upgrade was unsuccessful.");
						  case 0x750A: // LIC010 (base=0x7500)
							return("Unable to import as the license is not applicable for the specified device.");
						  case 0x750B: // LIC011 (base=0x7500)
							return("A non-evaluation license cannot be replaced with an evaluation license.");
						  case 0x750C: // LIC012 (base=0x7500)
							return("The license file does not exist.");
						  case 0x750D: // LIC013 (base=0x7500)
							return("These license features are not supported by this firmware version.");
						  case 0x750E: // LIC014 (base=0x7500)
							return("Multiple backup or restore operations have been simultaneously attempted on the License Manager database.");
						  case 0x750F: // LIC015 (base=0x7500)
							return("The License Manager database restore operation failed.");
						  case 0x7510: // LIC016 (base=0x7500)
							return("Unable to meet the feature dependencies of the license.");
						  case 0x7511: // LIC017 (base=0x7500)
							return("The license file is corrupted, not unzipped, or not valid license.");
						  case 0x7512: // LIC018 (base=0x7500)
							return("The license is already imported.");
						  case 0x7513: // LIC019 (base=0x7500)
							return("A leased license cannot be imported before its start date.");
						  case 0x7514: // LIC020 (base=0x7500)
							return("Unable to import due to End User License Agreement (EULA) import upgrade warning.");
						  case 0x7515: // LIC021 (base=0x7500)
							return("Unable to import because the features contained in the evaluation license are already licensed.");
						  case 0x7516: // LIC022 (base=0x7500)
							return("The License Manager database is locked due to ongoing backup and restore operation.");
						  case 0x7517: // LIC201 (base=0x7500)
							return("License %s assigned to device %s expires in %s days.");
						  case 0x7518: // LIC202 (base=0x7500)
							return("Unable to perform the operation because device has not been powered off.\nPower off the device and retry the operation.");
						  case 0x7519: // LIC203 (base=0x7500)
							return("The license %s has encountered an error.");
						  case 0x751A: // LIC204 (base=0x7500)
							return("Unable to complete the License Manager database restore operation.");
						  case 0x751B: // LIC205 (base=0x7500)
							return("License Manager database lock has timed-out.");
						  case 0x751C: // LIC206 (base=0x7500)
							return("EULA warning: Importing license %s may violate the End-User License Agreement.");
						  case 0x751D: // LIC207 (base=0x7500)
							return("License %s on device %s has expired.");
						  case 0x751E: // LIC208 (base=0x7500)
							return("License %s successfully imported to device %s.");
						  case 0x751F: // LIC209 (base=0x7500)
							return("License %s successfully exported from the device %s.");
						  case 0x7520: // LIC210 (base=0x7500)
							return("License %s successfully deleted from the device %s.");
						  case 0x7521: // LIC211 (base=0x7500)
							return("The iDRAC feature set has changed.");
						  case 0x7522: // LIC501 (base=0x7500)
							return("A required license is missing or expired.");
						  case 0x7523: // LIC502 (base=0x7500)
							return("Features are not available.");
						  case 0x7524: // LIC503 (base=0x7500)
							return("A required license is missing or expired. The following features are not enabled: %s");
						  case 0x7525: // LIC900 (base=0x7500)
							return("The command was successful.");
						  case 0x7526: // LIC901 (base=0x7500)
							return("Invalid parameter value for %s.");
						  case 0x7527: // LIC902 (base=0x7500)
							return("Resource allocation failure.");
						  case 0x7528: // LIC903 (base=0x7500)
							return("Missing parameters %s.");
						  case 0x7529: // LIC904 (base=0x7500)
							return("Unable to connect to the network share.");
						  case 0x752A: // LIC905 (base=0x7500)
							return("The LicenseName value cannot exceed 64 characters.");
						  case 0x752B: // LIC906 (base=0x7500)
							return("License file is inaccessible on the network share.");
						  case 0x752C: // LIC907 (base=0x7500)
							return("Unable to perform the operation due to an unknown error in iDRAC.");
				
						  // Duplicate IDRAC DNS Name
						  case 0x52100:
							return ("Duplicate DNS Name Selected");
				
						  //Controller Troubleshooting Page Errors
						  case 0x80000003: // EFI_UNSUPPORTED - Different CMC Versions
							return("The passive CMC is not operating at the correct version. Update the passive CMC.");
						  case 0x80000006: // EFI_NOT_READY - ActiveControllerBeingDisabled
							return("Unable to disable RAID controller because Shared PERC8 (Integrated 2) is the active controller. \nPower cycle the chassis and retry the operation. If the error persists contact your service provider.");
						  case 0x80000015: // EFI_ABORTED - ServersOn
							return("Unable to enable or disable RAID controller because all servers have not been powered off.\nPower off all servers and retry the operation.");
						  case 0x80000014: // EFI_ALREADY_STARTED - ControllerAlreadyDisabled
							return("The opposite storage adapter is already disabled. Enable the opposite storage adater before disabling this adapter.");
				
						  default:
							return ("Operation Not Successful: Error=" + v);
						}
						return "ERROR";
					  }
				
					</script>
				  </head>
				  <body onload="javascript:overrideDisplayStatus();" onkeypress="javascript:return(f_keypress(event));">
					<a name="top" id="top"></a>
					<div xmlns="" id="pullstrip" onmousedown="javascript:pullTab.ra_resizeStart(event, this);">
					  <div id="pulltab"></div>
					</div>
					<div xmlns="" id="rightside"></div>
					<div xmlns="" class="data-area-page-title">
					  <span id="pageTitle">Budget/Redundancy Configuration</span>
					  <div class="toolbar">
						<a id="A2" name="printbutton" class="print" href="javascript:window.print();" title="Print"></a>
						<a id="A5" name="refresh" class="refresh" href="javascript:top.globalnav.f_refresh();" title="Refresh"></a>
						<a id="A6" name="help" class="help" href="javascript:top.globalnav.f_help();" title="Help"></a>
					  </div>
					  <div class="da-line"></div>
					</div>
					<div class="table_container">
					  <table class="container">
						<thead>
						  <tr>
							<td class="topleft" width="3px"></td>
							<td class="top">Information</td>
							<td class="topright"></td>
						  </tr>
						</thead>
						<tr>
						  <td class="left" width="3px"></td>
						  <td class="instructions">
							<ul>
							  <li>Setting changes on this page may not be reflected immediately. Refreshing the page after an appropriate delay will display the new values.</li>
							  <li>Remote Power logging requires Remote SysLog to be enabled. See Network Services for more information.</li>
							</ul>
						  </td>
						  <td class="right"></td>
						</tr>
						<tfoot>
						  <tr>
							<td class="bottomleft" width="3px"></td>
							<td class="bottom"></td>
							<td class="bottomright"></td>
						  </tr>
						</tfoot>
					  </table>
					</div>
					<form onsubmit="true" action="/cgi-bin/webcgi/pwr_redundancy" name="dataarea" id="dataarea" method="post" autocomplete="off">
					  <input xmlns="" type="hidden" value="6f12651ace76363bb9cff970c453c8df" name="ST2" id="ST2" />
					  <div class="table_container">
						<table class="container">
						  <thead>
							<tr>
							  <td class="topleft" width="3px"></td>
							  <td class="top borderright" width="49%">Attribute</td>
							  <td class="top" colspan="3">Value</td>
							  <td class="topright"></td>
							</tr>
						  </thead>
						  <tbody>
							<tr xmlns="" class="fill">
							  <td class="left" width="3px"></td>
							  <td nowrap="nowrap" class="contents borderright borderbottom">Enable Server Based Power Management</td>
							  <td class="contents borderbottom" colspan="3">
								<input type="checkbox" onClick="javascript:selectProperty(this);" name="CHASSIS_POWER_SBPMMode1" id="CHASSIS_POWER_SBPMMode1" value="&#10;    0&#10;  " />
							  </td>
							  <td class="right"></td>
							</tr>
							<tr xmlns="">
							  <td class="left" width="3px"></td>
							  <td nowrap="nowrap" class="contents borderright borderbottom">System Input Power Cap<br />(0-16% = 2715W - 100% = 16685W)
					  </td>
							  <td class="contents borderbottom" valign="middle"><input type="text" onkeypress="return isNumericKey(event);" onkeyup="onKeyUp(this);" onchange="this.value = extractNumeric(this); updateDependents(this);" onfocus="this.style.background='yellow'" onblur="this.style.background='white'" size="4" maxlength="5" name="acPowerBudget1" id="acPowerBudget1" value="16685" />W</td>
							  <td class="contents borderbottom" valign="middle"><input type="text" onkeypress="return isNumericKey(event);" onkeyup="onKeyUp(this);" onchange="this.value = extractNumeric(this); updateDependents(this);" onfocus="this.style.background='yellow'" onblur="this.style.background='white'" size="5" maxlength="5" name="acPowerBudget2" id="acPowerBudget2" value="56931" />BTU/h</td>
							  <td class="contents borderbottom" valign="middle"><input type="text" onkeypress="return isNumericKey(event);" onkeyup="onKeyUp(this);" onchange="this.value = extractNumeric(this); updateDependents(this);" onfocus="this.style.background='yellow'" onblur="this.style.background='white'" size="2" maxlength="3" name="acPowerBudget3" id="acPowerBudget3" value="100" />%</td>
							  <td class="right"></td>
							</tr>
							<tr xmlns="" class="fill">
							  <td class="left" width="3px"></td>
							  <td nowrap="nowrap" class="contents borderright borderbottom">Redundancy Policy</td>
							  <td class="contents borderbottom" colspan="3">
								<select onchange="updateDependents(this);" name="psuRedundancy1" id="psuRedundancy1">
								  <option value="1" selected="selected">Grid Redundancy</option>
								  <option value="2">Power Supply Redundancy</option>
								  <option value="0">No Redundancy</option>
								</select>
							  </td>
							  <td class="right"></td>
							</tr>
							<tr xmlns="">
							  <td class="left" width="3px"></td>
							  <td nowrap="nowrap" class="contents borderright borderbottom">Enable Extended Power Performance</td>
							  <td class="contents borderbottom" colspan="3">
								<input type="checkbox" onClick="javascript:selectProperty(this);" name="CHASSIS_POWER_epp_enable1" id="CHASSIS_POWER_epp_enable1" value="&#10;    0&#10;  " disabled="disabled" />
							  </td>
							  <td class="right"></td>
							</tr>
							<tr xmlns="" class="fill">
							  <td class="left" width="3px"></td>
							  <td nowrap="nowrap" class="contents borderright borderbottom">Server Performance Over Power Redundancy</td>
							  <td class="contents borderbottom" colspan="3">
								<input type="checkbox" onClick="javascript:selectProperty(this);" name="CHASSIS_POWER_performance_over_redundancy1" id="CHASSIS_POWER_performance_over_redundancy1" value="&#10;    0&#10;  " />
							  </td>
							  <td class="right"></td>
							</tr>
							<tr xmlns="">
							  <td class="left" width="3px"></td>
							  <td nowrap="nowrap" class="contents borderright borderbottom">Enable Dynamic Power Supply Engagement</td>
							  <td class="contents borderbottom" colspan="3">
								<input type="checkbox" onClick="javascript:selectProperty(this);" name="psuDynEng1" id="psuDynEng1" value="&#10;    0&#10;  " />
							  </td>
							  <td class="right"></td>
							</tr>
							<tr xmlns="" class="fill">
							  <td class="left" width="3px"></td>
							  <td nowrap="nowrap" class="contents borderright borderbottom">Disable Chassis Power Button</td>
							  <td class="contents borderbottom" colspan="3">
								<input type="checkbox" onClick="javascript:selectProperty(this);" name="CHASSIS_POWER_button_disable1" id="CHASSIS_POWER_button_disable1" value="&#10;    0&#10;  " />
							  </td>
							  <td class="right"></td>
							</tr>
							<tr xmlns="">
							  <td class="left" width="3px"></td>
							  <td nowrap="nowrap" class="contents borderright borderbottom">Allow 110 VAC Operation</td>
							  <td class="contents borderbottom" colspan="3">
								<input type="checkbox" onClick="javascript:selectProperty(this);" name="CHASSIS_POWER_110V_acknowledge1" id="CHASSIS_POWER_110V_acknowledge1" value="&#10;    0&#10;  " />
							  </td>
							  <td class="right"></td>
							</tr>
							<tr xmlns="" class="fill">
							  <td class="left" width="3px"></td>
							  <td nowrap="nowrap" class="contents borderright borderbottom">Max Power Conservation Mode</td>
							  <td class="contents borderbottom" colspan="3">
								<input type="checkbox" onClick="javascript:selectProperty(this);" name="CHASSIS_POWER_UPSMode1" id="CHASSIS_POWER_UPSMode1" value="&#10;    0&#10;  " />
							  </td>
							  <td class="right"></td>
							</tr>
							<tr xmlns="">
							  <td class="left" width="3px"></td>
							  <td nowrap="nowrap" class="contents borderright borderbottom">Enable Remote Power Logging</td>
							  <td class="contents borderbottom" colspan="3"><input type="checkbox" onClick="javascript:selectProperty(this);" name="CHASSIS_POWER_remote_logging_enable1" id="CHASSIS_POWER_remote_logging_enable1" value="&#10;    0&#10;  " /> 
							<a href="#" onclick="parent.treelist.f_select_by_url('/cgi-bin/webcgi/interfaces');">Remote SysLog Configuration</a></td>
							  <td class="right"></td>
							</tr>
							<tr xmlns="" class="fill">
							  <td class="left" width="3px"></td>
							  <td nowrap="nowrap" class="contents borderright borderbottom">Remote Power Logging Interval (1-1440)
					  </td>
							  <td class="contents borderbottom" colspan="3"><input type="text" size="4" maxlength="4" onkeypress="return isNumericKey(event)" onchange="this.value = extractNumeric(this)" name="CHASSIS_POWER_remote_logging_interval1" id="CHASSIS_POWER_remote_logging_interval1" value="5" disabled="disabled" />Minutes</td>
							  <td class="right"></td>
							</tr>
							<tr xmlns="">
							  <td class="left" width="3px"></td>
							  <td nowrap="nowrap" class="contents borderright borderbottom">Disable AC Power Recovery</td>
							  <td class="contents borderbottom" colspan="3">
								<input type="checkbox" onClick="javascript:selectProperty(this);" name="CHASSIS_POWER_ac_recovery_enable1" id="CHASSIS_POWER_ac_recovery_enable1" value="&#10;    0&#10;  " />
							  </td>
							  <td class="right"></td>
							</tr>
						  </tbody>
						  <tfoot>
							<tr>
							  <td class="bottomleftbuttons" width="3px"></td>
							  <td class="bottombuttons" colspan="4">
								<div class="button_clear">
								  <a class="page_button_emphasized" href="javascript:formSubmit();">
									<span>Apply</span>
								  </a>
								  <a class="page_button" href="javascript:top.globalnav.f_refresh();">
									<span>Cancel</span>
								  </a>
								</div>
							  </td>
							  <td class="bottomrightbuttons"></td>
							</tr>
						  </tfoot>
						</table>
					  </div>
					  <input xmlns="" id="p1" name="p1" type="hidden" value="psuAnyBlacklisted" />
					  <input xmlns="" id="vpsuAnyBlacklisted1" name="vpsuAnyBlacklisted1" type="hidden" value="0" />
					  <input xmlns="" id="vCountpsuAnyBlacklisted" name="vCountpsuAnyBlacklisted" type="hidden" value="1" />
					  <input xmlns="" id="p2" name="p2" type="hidden" value="acPowerSurplus" />
					  <input xmlns="" id="vacPowerSurplus1" name="vacPowerSurplus1" type="hidden" value="12064 W L41164 BTU/hR" />
					  <input xmlns="" id="vCountacPowerSurplus" name="vCountacPowerSurplus" type="hidden" value="1" />
					  <input xmlns="" id="p3" name="p3" type="hidden" value="acPowerRequired" />
					  <input xmlns="" id="vacPowerRequired1" name="vacPowerRequired1" type="hidden" value="4621 W L15767 BTU/hR" />
					  <input xmlns="" id="vCountacPowerRequired" name="vCountacPowerRequired" type="hidden" value="1" />
					  <input xmlns="" id="p4" name="p4" type="hidden" value="CHASSIS_POWER_110V_used" />
					  <input xmlns="" id="vCHASSIS_POWER_110V_used1" name="vCHASSIS_POWER_110V_used1" type="hidden" value="0" />
					  <input xmlns="" id="vCountCHASSIS_POWER_110V_used" name="vCountCHASSIS_POWER_110V_used" type="hidden" value="1" />
					  <input xmlns="" id="p5" name="p5" type="hidden" value="allPSUsEPPCapable" />
					  <input xmlns="" id="vallPSUsEPPCapable1" name="vallPSUsEPPCapable1" type="hidden" value="0" />
					  <input xmlns="" id="vCountallPSUsEPPCapable" name="vCountallPSUsEPPCapable" type="hidden" value="1" />
					  <input xmlns="" id="p6" name="p6" type="hidden" value="anyPSUsEPPCapable" />
					  <input xmlns="" id="vanyPSUsEPPCapable1" name="vanyPSUsEPPCapable1" type="hidden" value="0" />
					  <input xmlns="" id="vCountanyPSUsEPPCapable" name="vCountanyPSUsEPPCapable" type="hidden" value="1" />
					  <input xmlns="" id="p7" name="p7" type="hidden" value="EPPUpperCap" />
					  <input xmlns="" id="vEPPUpperCap1" name="vEPPUpperCap1" type="hidden" value="5944" />
					  <input xmlns="" id="vCountEPPUpperCap" name="vCountEPPUpperCap" type="hidden" value="1" />
					  <input xmlns="" id="p8" name="p8" type="hidden" value="CHASSIS_type_nebs" />
					  <input xmlns="" id="vCHASSIS_type_nebs1" name="vCHASSIS_type_nebs1" type="hidden" value="0" />
					  <input xmlns="" id="vCountCHASSIS_type_nebs" name="vCountCHASSIS_type_nebs" type="hidden" value="1" />
					  <input xmlns="" id="p9" name="p9" type="hidden" value="CHASSIS_type_freshair" />
					  <input xmlns="" id="vCHASSIS_type_freshair1" name="vCHASSIS_type_freshair1" type="hidden" value="0" />
					  <input xmlns="" id="vCountCHASSIS_type_freshair" name="vCountCHASSIS_type_freshair" type="hidden" value="1" />
					  <input xmlns="" id="p10" name="p10" type="hidden" value="ChassisInputTypeIsDC" />
					  <input xmlns="" id="vChassisInputTypeIsDC1" name="vChassisInputTypeIsDC1" type="hidden" value="" />
					  <input xmlns="" id="vCountChassisInputTypeIsDC" name="vCountChassisInputTypeIsDC" type="hidden" value="1" />
					  <input xmlns="" id="p11" name="p11" type="hidden" value="THERMAL_ecm" />
					  <input xmlns="" id="vTHERMAL_ecm1" name="vTHERMAL_ecm1" type="hidden" value="0" />
					  <input xmlns="" id="vCountTHERMAL_ecm" name="vCountTHERMAL_ecm" type="hidden" value="1" />
					  <input xmlns="" id="p12" name="p12" type="hidden" value="CHASSIS_POWER_SBPMMode" />
					  <input xmlns="" id="vCHASSIS_POWER_SBPMMode1" name="vCHASSIS_POWER_SBPMMode1" type="hidden" value="0" />
					  <input xmlns="" id="vCountCHASSIS_POWER_SBPMMode" name="vCountCHASSIS_POWER_SBPMMode" type="hidden" value="1" />
					  <input xmlns="" id="p13" name="p13" type="hidden" value="acPowerBudget" />
					  <input xmlns="" id="vacPowerBudget1" name="vacPowerBudget1" type="hidden" value="16685" />
					  <input xmlns="" id="vacPowerBudget2" name="vacPowerBudget2" type="hidden" value="56931" />
					  <input xmlns="" id="vacPowerBudget3" name="vacPowerBudget3" type="hidden" value="100" />
					  <input xmlns="" id="vCountacPowerBudget" name="vCountacPowerBudget" type="hidden" value="3" />
					  <input xmlns="" id="p14" name="p14" type="hidden" value="psuRedundancy" />
					  <input xmlns="" id="vpsuRedundancy1" name="vpsuRedundancy1" type="hidden" value="1" />
					  <input xmlns="" id="vCountpsuRedundancy" name="vCountpsuRedundancy" type="hidden" value="1" />
					  <input xmlns="" id="p15" name="p15" type="hidden" value="CHASSIS_POWER_epp_enable" />
					  <input xmlns="" id="vCHASSIS_POWER_epp_enable1" name="vCHASSIS_POWER_epp_enable1" type="hidden" value="0" />
					  <input xmlns="" id="vCountCHASSIS_POWER_epp_enable" name="vCountCHASSIS_POWER_epp_enable" type="hidden" value="1" />
					  <input xmlns="" id="p16" name="p16" type="hidden" value="CHASSIS_POWER_performance_over_redundancy" />
					  <input xmlns="" id="vCHASSIS_POWER_performance_over_redundancy1" name="vCHASSIS_POWER_performance_over_redundancy1" type="hidden" value="0" />
					  <input xmlns="" id="vCountCHASSIS_POWER_performance_over_redundancy" name="vCountCHASSIS_POWER_performance_over_redundancy" type="hidden" value="1" />
					  <input xmlns="" id="p17" name="p17" type="hidden" value="psuDynEng" />
					  <input xmlns="" id="vpsuDynEng1" name="vpsuDynEng1" type="hidden" value="0" />
					  <input xmlns="" id="vCountpsuDynEng" name="vCountpsuDynEng" type="hidden" value="1" />
					  <input xmlns="" id="p18" name="p18" type="hidden" value="CHASSIS_POWER_button_disable" />
					  <input xmlns="" id="vCHASSIS_POWER_button_disable1" name="vCHASSIS_POWER_button_disable1" type="hidden" value="0" />
					  <input xmlns="" id="vCountCHASSIS_POWER_button_disable" name="vCountCHASSIS_POWER_button_disable" type="hidden" value="1" />
					  <input xmlns="" id="p19" name="p19" type="hidden" value="CHASSIS_POWER_110V_acknowledge" />
					  <input xmlns="" id="vCHASSIS_POWER_110V_acknowledge1" name="vCHASSIS_POWER_110V_acknowledge1" type="hidden" value="0" />
					  <input xmlns="" id="vCountCHASSIS_POWER_110V_acknowledge" name="vCountCHASSIS_POWER_110V_acknowledge" type="hidden" value="1" />
					  <input xmlns="" id="p20" name="p20" type="hidden" value="CHASSIS_POWER_UPSMode" />
					  <input xmlns="" id="vCHASSIS_POWER_UPSMode1" name="vCHASSIS_POWER_UPSMode1" type="hidden" value="0" />
					  <input xmlns="" id="vCountCHASSIS_POWER_UPSMode" name="vCountCHASSIS_POWER_UPSMode" type="hidden" value="1" />
					  <input xmlns="" id="p21" name="p21" type="hidden" value="CHASSIS_POWER_remote_logging_enable" />
					  <input xmlns="" id="vCHASSIS_POWER_remote_logging_enable1" name="vCHASSIS_POWER_remote_logging_enable1" type="hidden" value="0" />
					  <input xmlns="" id="vCountCHASSIS_POWER_remote_logging_enable" name="vCountCHASSIS_POWER_remote_logging_enable" type="hidden" value="1" />
					  <input xmlns="" id="p22" name="p22" type="hidden" value="CHASSIS_POWER_remote_logging_interval" />
					  <input xmlns="" id="vCHASSIS_POWER_remote_logging_interval1" name="vCHASSIS_POWER_remote_logging_interval1" type="hidden" value="5" />
					  <input xmlns="" id="vCountCHASSIS_POWER_remote_logging_interval" name="vCountCHASSIS_POWER_remote_logging_interval" type="hidden" value="1" />
					  <input xmlns="" id="p23" name="p23" type="hidden" value="CHASSIS_POWER_ac_recovery_enable" />
					  <input xmlns="" id="vCHASSIS_POWER_ac_recovery_enable1" name="vCHASSIS_POWER_ac_recovery_enable1" type="hidden" value="0" />
					  <input xmlns="" id="vCountCHASSIS_POWER_ac_recovery_enable" name="vCountCHASSIS_POWER_ac_recovery_enable" type="hidden" value="1" />
					  <input xmlns="" id="p24" name="p24" type="hidden" value="ChassisPSUFailure" />
					  <input xmlns="" id="vChassisPSUFailure1" name="vChassisPSUFailure1" type="hidden" value="1" />
					  <input xmlns="" id="vCountChassisPSUFailure" name="vCountChassisPSUFailure" type="hidden" value="1" />
					  <input xmlns="" id="pCount" name="pCount" type="hidden" value="24" />
					</form>
				  </body>
				</html>
			`),
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
		{
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
		{
			Serial:     "51f3dk2_psu_1",
			CapacityKw: 2.7,
			Status:     "OK",
			PowerKw:    0.184,
			PartNumber: "0TJJ3M",
		},
		{
			Serial:     "51f3dk2_psu_2",
			CapacityKw: 2.7,
			Status:     "OK",
			PowerKw:    0.20862,
			PartNumber: "0TJJ3M",
		},
		{
			Serial:     "51f3dk2_psu_5",
			CapacityKw: 2.7,
			Status:     "OK",
			PowerKw:    0.20772000000000002,
			PartNumber: "0TJJ3M",
		},
		{
			Serial:     "51f3dk2_psu_6",
			CapacityKw: 2.7,
			Status:     "OK",
			PowerKw:    0.25278,
			PartNumber: "0TJJ3M",
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
		if psu.Serial != expectedAnswer[pos].Serial ||
			psu.CapacityKw != expectedAnswer[pos].CapacityKw ||
			psu.PowerKw != expectedAnswer[pos].PowerKw ||
			psu.Status != expectedAnswer[pos].Status ||
			psu.PartNumber != expectedAnswer[pos].PartNumber {
			t.Errorf("Expected answer %v: found %v", expectedAnswer[pos], psu)
		}
	}

	tearDown()
}

func TestChassisBmcType(t *testing.T) {
	expectedAnswer := "m1000e"

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

func TestChassisFans(t *testing.T) {
	expectedAnswer := []*devices.Fan{
		{
			Position:   1,
			Status:     "OK",
			CurrentRPM: 2873,
		},
		{
			Status:     "OK",
			Position:   2,
			CurrentRPM: 4089,
		},
		{
			Status:     "OK",
			Position:   3,
			CurrentRPM: 4103,
		},
		{
			Status:     "OK",
			Position:   4,
			CurrentRPM: 2871,
		},
		{
			Status:     "OK",
			Position:   5,
			CurrentRPM: 4076,
		},
		{
			Status:     "OK",
			Position:   6,
			CurrentRPM: 4078,
		},
		{
			Status:     "OK",
			Position:   7,
			CurrentRPM: 2860,
		},
		{
			Status:     "OK",
			Position:   8,
			CurrentRPM: 4068,
		},
		{
			Status:     "OK",
			Position:   9,
			CurrentRPM: 4086,
		},
	}

	chassis, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	fans, err := chassis.Fans()
	if err != nil {
		t.Fatalf("Found errors calling chassis.Fans %v", err)
	}

	if len(fans) != len(expectedAnswer) {
		t.Fatalf("Expected %v fans: found %v fans", len(expectedAnswer), len(fans))
	}

	for _, fan := range fans {
		found := false
		for _, ef := range expectedAnswer {
			if fan.Position == ef.Position {
				found = true
				if fan.CurrentRPM != ef.CurrentRPM || fan.Status != ef.Status {
					t.Errorf("Expected answer %v: found %v", ef, fan)
				}
			}
		}
		if !found {
			t.Errorf("Unable to find a match for %v", fan)
		}
	}

	tearDown()
}

func TestChassisInterface(t *testing.T) {
	chassis, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	_ = devices.Cmc(chassis)
	_ = devices.Configure(chassis)
	_ = devices.CmcSetup(chassis)

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

func TestChassisIsPsuRedundant(t *testing.T) {
	expectedAnswer := true

	chassis, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := chassis.IsPsuRedundant()
	if err != nil {
		t.Fatalf("Found errors calling chassis.Name %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestChassisRedundancyMode(t *testing.T) {
	expectedAnswer := "Grid"

	chassis, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := chassis.PsuRedundancyMode()
	if err != nil {
		t.Fatalf("Found errors calling chassis.Name %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}
