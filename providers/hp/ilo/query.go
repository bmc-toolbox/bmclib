package ilo

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
)

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
