package supermicro

//{"error":{"code":"Base.v1_10_3.GeneralError","Message":"A general error has occurred. See ExtendedInfo for more information.","@Messag
//e.ExtendedInfo":[{"MessageId":"SMC.1.0.OemFirmwareAlreadyInUpdateMode","Severity":"Warning","Resolution":"Please check if there was the next step with respective API to execute
//.","Message":"The BMC firmware update was already in update mode.","MessageArgs":["BMC"],"RelatedProperties":["EnterUpdateMode_StatusCheck"]}]}}
//
//{"Accepted":{"code":"Base.v1_10_3.Accepted","Message":"Successfully Accepted Request. Please see the location header and ExtendedInfo for more information.","@Message.ExtendedInfo":[{"MessageId":"SMC.1.0.OemSimpleupdateAcceptedMessage","Severity":"Ok","Resolution":"No resolution was required.","Message":"Please also check Task Resource /redfish/v1/TaskService/Tasks/1 to see more information.","MessageArgs":["/redfish/v1/TaskService/Tasks/1"],"RelatedProperties":["BmcVerifyAccepted"]}]}}
//2023/10/25 11:33:11 upload taskID: 1
//2023/10/25 11:33:16 retrying in 5 secs..:  id: 1, state: Running, status: OK: firmware uploaded and is currently being verified
//{"Accepted":{"code":"Base.v1_10_3.Accepted","Message":"Successfully Accepted Request. Please see the location header and ExtendedInfo for more information.","@Message.ExtendedInfo":[{"MessageId":"SMC.1.0.OemSimpleupdateAcceptedMessage","Severity":"Ok","Resolution":"No resolution was required.","Message":"Please also check Task Resource /redfish/v1/TaskService/Tasks/2 to see more information.","MessageArgs":["/redfish/v1/TaskService/Tasks/2"],"RelatedProperties":["BmcUpdateAccepted"]}]}
//
//
//
//TODO: test firmwareUploadBMC
