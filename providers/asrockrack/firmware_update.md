
 Flashing a BMC firmware seems to be a multi step process


 1. PUT /api/maintenance/flash
       no payload (seems to set the device to be in flash mode or such)
       200 OK - takes about a minute to return

 2. POST /api/maintenance/firmware
      Content-Type: multipart/form-data
      ------WebKitFormBoundaryESKCgdjyLnqUPHBK
		Content-Disposition: form-data; name="fwimage"; filename="E3C246D4I-NL_L0.01.00.ima"
		Content-Type: application/octet-stream
     ------WebKitFormBoundaryESKCgdjyLnqUPHBK--
 .   response - '{"cc": 0}' - successful upload

 3. GET /api/maintenance/firmware/verification
       500 - Bad firmware payload -> invoke reset
       200 - OK
           [ { "id": 1, "current_image_name": "ast2500e", "current_image_version1": "0.01.00", "current_image_version2": "", "new_image_version": "0.03.00", "section_status": 0, "verification_status": 5 } ]
 
 4. If verificaion fails OR firmware update progress is at 100% done - invoke reset
 
      GET /api/maintenance/reset
       200 OK
 
 5. PUT /api/maintenance/firmware/upgrade
      payload {"preserve_config":1,"preserve_network":0,"preserve_user":0,"flash_status":1}
      200 OK
      response - same as payload

 6. GET https://10.230.148.171/api/maintenance/firmware/flash-progress
     { "id": 1, "action": "Flashing...", "progress": "12% done         ", "state": 0 }
     { "id": 1, "action": "Flashing...", "progress": "100% done", "state": 0 }