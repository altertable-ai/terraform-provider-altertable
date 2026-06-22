package client

// Environment models an Altertable environment.
type Environment struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type EnvironmentCreateInput struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type EnvironmentUpdateInput struct {
	Name string `json:"name"`
}

// Catalog models a catalog within an environment.
type Catalog struct {
	ID            string `json:"id"`
	EnvironmentID string `json:"environment_id"`
	Slug          string `json:"slug"`
	Name          string `json:"name"`
}

type CatalogCreateInput struct {
	EnvironmentID string `json:"environment_id"`
	Slug          string `json:"slug"`
	Name          string `json:"name"`
}

type CatalogUpdateInput struct {
	Name string `json:"name"`
}

// User models an Altertable user.
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type UserCreateInput struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type UserUpdateInput struct {
	Name string `json:"name"`
}

// ServiceAccount models a machine identity.
type ServiceAccount struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ServiceAccountCreateInput struct {
	Name string `json:"name"`
}

type ServiceAccountUpdateInput struct {
	Name string `json:"name"`
}

// RoleGrant is a single (role, optional resource) tuple.
type RoleGrant struct {
	Role       string `json:"role"`
	ResourceID string `json:"resource_id,omitempty"`
}

// PrincipalRef identifies the principal a role set belongs to; exactly one field is set.
type PrincipalRef struct {
	UserID           string `json:"user_id,omitempty"`
	ServiceAccountID string `json:"service_account_id,omitempty"`
}

// RoleSet is the complete set of grants for a principal.
type RoleSet struct {
	PrincipalID string      `json:"principal_id"`
	Grants      []RoleGrant `json:"grants"`
}

// Credential models an environment credential. Password is only populated on create.
type Credential struct {
	ID               string `json:"id"`
	ServiceAccountID string `json:"service_account_id"`
	EnvironmentID    string `json:"environment_id"`
	Label            string `json:"label"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	CreatedAt        string `json:"created_at"`
}

type CredentialCreateInput struct {
	ServiceAccountID string `json:"service_account_id"`
	EnvironmentID    string `json:"environment_id"`
	Label            string `json:"label"`
}

type CredentialUpdateInput struct {
	Label string `json:"label"`
}
