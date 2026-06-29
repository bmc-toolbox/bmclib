package lenovo

import (
	"context"
	"testing"
)

// Requirement: Read accounts.
func TestUserRead(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	users, err := c.UserRead(context.Background())
	if err != nil {
		t.Fatalf("UserRead: %v", err)
	}
	// Only the named account (USERID) is returned; the empty slot is skipped.
	if len(users) != 1 {
		t.Fatalf("got %d users, want 1", len(users))
	}
	if users[0]["Username"] != "USERID" || users[0]["RoleID"] != "Administrator" {
		t.Errorf("unexpected user: %+v", users[0])
	}
}

// Requirement: Create account — via POST.
func TestUserCreatePost(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	ok, err := c.UserCreate(context.Background(), "ops", "secret123", "Operator")
	if err != nil || !ok {
		t.Fatalf("UserCreate = (%v, %v), want (true, nil)", ok, err)
	}
	if !ts.didPostAccount() {
		t.Error("expected an account POST")
	}
}

// Requirement: Create account — Purley PATCH fallback.
func TestUserCreatePurleyFallback(t *testing.T) {
	ts := newTestServer(t, testServerOpts{rejectAccountPost: true})
	c := ts.openedClient(t)

	ok, err := c.UserCreate(context.Background(), "ops", "secret123", "Operator")
	if err != nil || !ok {
		t.Fatalf("UserCreate (fallback) = (%v, %v), want (true, nil)", ok, err)
	}
	if !ts.didPatchAccount() {
		t.Error("expected a slot PATCH fallback when POST is rejected")
	}
}

// Requirement: Create account — duplicate username errors.
func TestUserCreateDuplicate(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	if _, err := c.UserCreate(context.Background(), "USERID", "secret123", "Administrator"); err == nil {
		t.Fatal("expected an error creating a duplicate username")
	}
}

// Requirement: Create account — missing params error.
func TestUserCreateMissingParams(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	if _, err := c.UserCreate(context.Background(), "ops", "", "Operator"); err == nil {
		t.Fatal("expected an error when the password is empty")
	}
}

// Requirement: Update account — change a user's role.
func TestUserUpdate(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	ok, err := c.UserUpdate(context.Background(), "USERID", "", "Operator")
	if err != nil || !ok {
		t.Fatalf("UserUpdate = (%v, %v), want (true, nil)", ok, err)
	}
	if !ts.didPatchAccount() {
		t.Error("expected the matching account to be PATCHed")
	}
}

// Requirement: Update account — unknown user errors.
func TestUserUpdateUnknown(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	if _, err := c.UserUpdate(context.Background(), "ghost", "x", "Operator"); err == nil {
		t.Fatal("expected an error updating an unknown user")
	}
}

// Requirement: Delete account.
func TestUserDelete(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	ok, err := c.UserDelete(context.Background(), "USERID")
	if err != nil || !ok {
		t.Fatalf("UserDelete = (%v, %v), want (true, nil)", ok, err)
	}
	if !ts.didPatchAccount() {
		t.Error("expected the account slot to be cleared via PATCH")
	}
}

// Requirement: Custom role and account-service management — read roles.
func TestRoles(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	roles, err := c.Roles(context.Background())
	if err != nil {
		t.Fatalf("Roles: %v", err)
	}
	if len(roles) != 2 {
		t.Fatalf("got %d roles, want 2", len(roles))
	}
	var admin *RoleInfo
	for i := range roles {
		if roles[i].ID == "Administrator" {
			admin = &roles[i]
		}
	}
	if admin == nil {
		t.Fatal("Administrator role not found")
	}
	if len(admin.Privileges) == 0 || !admin.Predefined {
		t.Errorf("unexpected Administrator role: %+v", admin)
	}
}

// Requirement: Custom role and account-service management — create a custom role.
func TestRoleCreate(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	err := c.RoleCreate(context.Background(), "CustomOps", []string{"Login", "ConfigureComponents"})
	if err != nil {
		t.Fatalf("RoleCreate: %v", err)
	}
	if !ts.didPostRole() {
		t.Error("expected a POST to the roles collection")
	}
}
