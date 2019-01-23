package idrac9

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"reflect"
	"strings"

	log "github.com/sirupsen/logrus"

	bmclibErrors "github.com/bmc-toolbox/bmclib/errors"
)

//diffs two BiosSettings and returns a BiosSettings with the difference.
//!! Note this assumes the struct fields in BiosSettings are all strings !!
func diffBiosSettings(new *BiosSettings, current *BiosSettings) (diff *BiosSettings, err error) {

	//the struct that holds the changes
	diff = &BiosSettings{}

	struct1V := reflect.ValueOf(new).Elem()
	struct2V := reflect.ValueOf(current).Elem()
	diffV := reflect.ValueOf(diff).Elem()

	typeOfStruct := struct1V.Type()

	for i := 0; i < struct1V.NumField(); i++ {
		//
		struct1fieldValue, ok := struct1V.Field(i).Interface().(string)
		if !ok {
			return diff, fmt.Errorf("BiosSettings struct field expected to be of type string: %v",
				typeOfStruct.Field(i).Name)
		}

		struct2fieldValue := struct2V.Field(i).Interface().(string)
		if !ok {
			return diff, fmt.Errorf("BiosSettings struct field expected to be of type string: %v",
				typeOfStruct.Field(i).Name)
		}

		if struct1fieldValue != struct2fieldValue {
			diffV.Field(i).Set(struct1V.Field(i))
		}
	}

	return diff, err
}

func (i *IDrac9) getBiosSettings() (biosSettings *BiosSettings, err error) {

	endpoint := "redfish/v1/Systems/System.Embedded.1/Bios"

	oData := Odata{}
	_, response, err := i.queryRedfish("GET", endpoint, nil)
	if err != nil {
		return biosSettings, err
	}

	err = json.Unmarshal(response, &oData)
	if err != nil {
		return biosSettings, err
	}

	return oData.Attributes, err

}

//returns the bios settings pending reboot
func (i *IDrac9) biosSettingsPendingReboot() (pendingBiosSettings *BiosSettings, err error) {

	endpoint := "redfish/v1/Systems/System.Embedded.1/Bios/Settings"

	oData := Odata{}
	_, response, err := i.queryRedfish("GET", endpoint, nil)
	if err != nil {
		return pendingBiosSettings, err
	}

	err = json.Unmarshal(response, &oData)
	if err != nil {
		return pendingBiosSettings, err
	}

	if oData.Attributes == nil {
		return pendingBiosSettings, err
	}

	return oData.Attributes, err

}

// PATCHs Bios settings, queues setting to be applied at next boot.
func (i *IDrac9) setBiosSettings(biosSettings *BiosSettings) (err error) {

	biosSettingsURI := "redfish/v1/Systems/System.Embedded.1/Bios/Settings"
	idracPayload := make(map[string]*BiosSettings)
	idracPayload["Attributes"] = biosSettings

	payload, err := json.Marshal(idracPayload)
	if err != nil {
		msg := fmt.Sprintf("Error marshalling biosAttributes payload: %s", err)
		return errors.New(msg)
	}

	//PATCH bios settings
	statusCode, _, err := i.queryRedfish("PATCH", biosSettingsURI, payload)
	if err != nil || statusCode != 200 {
		msg := fmt.Sprintf("PATCH request to set Bios config, returned code: %d", statusCode)
		return errors.New(msg)
	}

	//Queue config to be set at next boot.
	return i.queueJobs(biosSettingsURI)
}

func (i *IDrac9) queueJobs(jobURI string) (err error) {

	endpoint := "redfish/v1/Managers/iDRAC.Embedded.1/Jobs"

	if !strings.HasPrefix(jobURI, "/") {
		jobURI = fmt.Sprintf("/%s", jobURI)
	}

	//Queue this setting to be applied at the next boot.
	targetSetting := TargetSettingsURI{
		TargetSettingsURI: jobURI,
	}

	payload, err := json.Marshal(targetSetting)
	if err != nil {
		msg := fmt.Sprintf("Error marshalling job queue payload for uri: %s, error: %s", jobURI, err)
		return errors.New(msg)
	}

	statusCode, _, err := i.queryRedfish("POST", endpoint, payload)
	if err != nil || statusCode != 200 {
		msg := fmt.Sprintf("POST request to queue job, returned code: %d", statusCode)
		return errors.New(msg)
	}

	return err
}

// Given a Job ID, purge it from the job queue
func (i *IDrac9) purgeJob(jobID string) (err error) {

	if !strings.Contains(jobID, "JID") {
		return errors.New("Invalid Job ID given, Job IDs should be prefixed with JID_")
	}

	endpoint := fmt.Sprintf("%s/%s", "redfish/v1/Managers/iDRAC.Embedded.1/Jobs", jobID)

	statusCode, _, err := i.queryRedfish("DELETE", endpoint, nil)
	if err != nil || statusCode != 200 {
		msg := fmt.Sprintf("DELETE request to purge job, returned code: %d", statusCode)
		return errors.New(msg)
	}

	return err
}

// Purges any jobs related to Bios configuration
func (i *IDrac9) purgeJobsForBiosSettings() (err error) {

	//get current job ids
	jobIDs, err := i.getJobIds()
	if err != nil {
		return err
	}

	//check if any jobs are queued for bios configuration
	if len(jobIDs) > 0 {
		err = i.purgeJobsByType(jobIDs, "BIOSConfiguration")
		if err != nil {
			return err
		}
	}

	return err
}

//Purges jobs of the given type - if they are in the "Scheduled" state
func (i *IDrac9) purgeJobsByType(jobIDs []string, jobType string) (err error) {

	for _, jobID := range jobIDs {
		jState, jType, err := i.getJob(jobID)
		if err != nil {
			return err
		}

		if jType == jobType && jState == "Scheduled" {
			return i.purgeJob(jobID)
		}

		return fmt.Errorf(fmt.Sprintf("Job not in Scheduled state cannot be purged, state: %s, id: %s", jState, jobID))
	}

	return err
}

//Returns the job state, Type for the given Job id
func (i *IDrac9) getJob(jobID string) (jobState string, jobType string, err error) {

	endpoint := fmt.Sprintf("%s/%s", "redfish/v1/Managers/iDRAC.Embedded.1/Jobs/", jobID)

	oData := Odata{}
	_, response, err := i.queryRedfish("GET", endpoint, nil)
	if err != nil {
		return jobState, jobType, err
	}

	err = json.Unmarshal(response, &oData)
	if err != nil {
		return jobState, jobType, err
	}

	return oData.JobState, oData.JobType, err

}

// Returns Job ids
func (i *IDrac9) getJobIds() (jobs []string, err error) {

	endpoint := "redfish/v1/Managers/iDRAC.Embedded.1/Jobs"

	oData := Odata{}
	_, response, err := i.queryRedfish("GET", endpoint, nil)
	if err != nil {
		return jobs, err
	}

	err = json.Unmarshal(response, &oData)
	if err != nil {
		return jobs, err
	}

	//No jobs present.
	if oData.MembersCount < 1 {
		return jobs, err
	}

	//[{"@odata.id":"/redfish/v1/Managers/iDRAC.Embedded.1/Jobs/JID_367624308519"}]
	for _, m := range oData.Members {
		for _, v := range m {
			//extract the Job id from the string
			tokens := strings.Split(v, "/")
			jobID := tokens[len(tokens)-1]
			jobs = append(jobs, jobID)
		}
	}

	return jobs, err
}

func isRequestMethodValid(method string) (valid bool) {
	validMethods := []string{"GET", "POST", "PATCH", "DELETE"}
	for _, m := range validMethods {
		if method == m {
			return true
		}
	}

	return valid
}

// GET data
func (i *IDrac9) queryRedfish(method string, endpoint string, payload []byte) (statusCode int, response []byte, err error) {

	if !isRequestMethodValid(method) {
		return statusCode, response, fmt.Errorf("Invalid request method: %v", method)
	}

	bmcURL := fmt.Sprintf("https://%s", i.ip)

	req, err := http.NewRequest(method, fmt.Sprintf("%s/%s", bmcURL, endpoint), bytes.NewReader(payload))
	if err != nil {
		return statusCode, response, err
	}

	req.SetBasicAuth(i.username, i.password)
	req.Header.Set("Content-Type", "application/json")

	if log.GetLevel() == log.DebugLevel {
		dump, err := httputil.DumpRequestOut(req, true)
		if err == nil {
			log.Println(fmt.Sprintf("[Request] %s/%s", bmcURL, endpoint))
			log.Println(">>>>>>>>>>>>>>>")
			log.Printf("%s\n\n", dump)
			log.Println(">>>>>>>>>>>>>>>")
		}
	}

	resp, err := i.httpClient.Do(req)
	if err != nil {
		return statusCode, response, err
	}
	defer resp.Body.Close()

	if log.GetLevel() == log.DebugLevel {
		dump, err := httputil.DumpResponse(resp, true)
		if err == nil {
			log.Println("[Response]")
			log.Println("<<<<<<<<<<<<<<")
			log.Printf("%s\n\n", dump)
			log.Println("<<<<<<<<<<<<<<")
		}
	}

	if resp.StatusCode == 401 {
		return resp.StatusCode, response, bmclibErrors.Err401Redfish
	}

	response, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return statusCode, response, err
	}

	return resp.StatusCode, response, err
}
