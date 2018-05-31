package ilo

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
)

func (i *Ilo) queryDirectoryGroups() (directoryGroups []DirectoryGroups, err error) {

	endpoint := "json/directory_groups"

	payload, err := i.get(endpoint)
	if err != nil {
		log.WithFields(log.Fields{
			"IP":       i.ip,
			"Model":    i.ModelId(),
			"endpoint": endpoint,
			"step":     funcName(),
			"Error":    err,
		}).Warn("GET request failed.")
		return directoryGroups, err
	}

	var directoryGroupAccts DirectoryGroupAccts
	//fmt.Printf("--> %+v\n", userinfo["users"])
	err = json.Unmarshal(payload, &directoryGroupAccts)
	if err != nil {
		log.WithFields(log.Fields{
			"IP":    i.ip,
			"step":  funcName(),
			"Model": i.ModelId(),
			"Error": err,
		}).Warn("Unable to unmarshal payload.")
		return directoryGroups, err
	}

	return directoryGroupAccts.Groups, err
}

func (i *Ilo) queryUsers() (usersInfo []UserInfo, err error) {

	endpoint := "json/user_info"

	payload, err := i.get(endpoint)
	if err != nil {
		log.WithFields(log.Fields{
			"IP":       i.ip,
			"Model":    i.ModelId(),
			"endpoint": endpoint,
			"step":     funcName(),
			"Error":    err,
		}).Warn("GET request failed.")
		return usersInfo, err
	}

	var users Users
	//fmt.Printf("--> %+v\n", userinfo["users"])
	err = json.Unmarshal(payload, &users)
	if err != nil {
		log.WithFields(log.Fields{
			"step":     "queryUserInfo",
			"resource": "User",
			"IP":       i.ip,
			"Model":    i.ModelId(),
			"Error":    err,
		}).Warn("Unable to unmarshal payload.")
		return usersInfo, err
	}

	return users.UsersInfo, err
}

func (i *Ilo) queryNetworkSntp() (networkSntp NetworkSntp, err error) {

	endpoint := "json/network_sntp/interface/0"

	payload, err := i.get(endpoint)
	if err != nil {
		log.WithFields(log.Fields{
			"IP":       i.ip,
			"Model":    i.ModelId(),
			"endpoint": endpoint,
			"step":     funcName(),
			"Error":    err,
		}).Warn("GET request failed.")
		return networkSntp, err
	}

	err = json.Unmarshal(payload, &networkSntp)
	if err != nil {
		log.WithFields(log.Fields{
			"IP":    i.ip,
			"step":  funcName(),
			"Model": i.ModelId(),
			"Error": err,
		}).Warn("Unable to unmarshal payload.")
		return networkSntp, err
	}

	return networkSntp, err
}
