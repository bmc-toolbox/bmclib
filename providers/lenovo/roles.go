package lenovo

import (
	"context"
	"net/url"
)

// RoleInfo describes an XCC account role.
type RoleInfo struct {
	// ID is the RoleId (e.g. "Administrator", "Operator", "ReadOnly" or a custom
	// role name).
	ID string
	// Privileges are the Redfish privileges assigned to the role.
	Privileges []string
	// Predefined reports whether the role is a built-in (non-removable) role.
	Predefined bool
}

// Roles returns the XCC account roles and their assigned privileges.
//
// This is an XCC-specific provider method (roles are not modelled by a
// bmc.Feature interface).
func (c *Conn) Roles(ctx context.Context) ([]RoleInfo, error) {
	service, err := c.redfishwrapper.AccountService()
	if err != nil {
		return nil, err
	}

	roles, err := service.Roles()
	if err != nil {
		return nil, err
	}

	out := make([]RoleInfo, 0, len(roles))
	for _, role := range roles {
		privileges := make([]string, 0, len(role.AssignedPrivileges))
		for _, p := range role.AssignedPrivileges {
			privileges = append(privileges, string(p))
		}

		out = append(out, RoleInfo{
			ID:         role.RoleID,
			Privileges: privileges,
			Predefined: role.IsPredefined,
		})
	}

	return out, nil
}

// RoleCreate creates a custom XCC role with the given privileges by POSTing to
// the AccountService Roles collection.
//
// This is an XCC-specific provider method.
func (c *Conn) RoleCreate(ctx context.Context, roleID string, privileges []string) error {
	service, err := c.redfishwrapper.AccountService()
	if err != nil {
		return err
	}

	rolesURL, err := url.JoinPath(service.ODataID, "Roles")
	if err != nil {
		return err
	}
	payload := map[string]any{
		"RoleId":             roleID,
		"AssignedPrivileges": privileges,
	}

	return checkResponse(c.redfishwrapper.PostWithHeaders(ctx, rolesURL, payload, nil))
}
