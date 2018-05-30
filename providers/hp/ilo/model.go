package ilo

type Users struct {
	UsersInfo []UserInfo `json:"users"`
}

// Add/Modify/Delete a user account
// POST
// https://10.193.251.48/json/user_info
type UserInfo struct {
	Id               int    `json:"id,int,omitempty"`
	LoginName        string `json:"login_name,omitempty"`
	UserName         string `json:"user_name,omitempty"`
	Password         string `json:"password,omitempty"`
	RemoteConsPriv   int    `json:"remote_cons_priv,omitempty"`
	VirtualMediaPriv int    `json:"virtual_media_priv,omitempty"`
	ResetPriv        int    `json:"reset_priv,omitempty"`
	ConfigPriv       int    `json:"config_priv,omitempty"`
	UserPriv         int    `json:"user_priv,omitempty"`
	LoginPriv        int    `json:"login_priv,omitempty"`
	Method           string `json:"method"` //mod_user, add_user, del_user
	UserId           int    `json:"user_id,int,omitempty"`
	SessionKey       string `json:"session_key,omitempty"`
}
