package client

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

type RoleGrant struct {
	Role       string `json:"role"`
	ResourceID string `json:"resource_id,omitempty"`
}

// Exactly one of UserID or ServiceAccountID is set.
type PrincipalRef struct {
	UserID           string `json:"user_id,omitempty"`
	ServiceAccountID string `json:"service_account_id,omitempty"`
}

type RoleSet struct {
	PrincipalID string      `json:"principal_id"`
	Grants      []RoleGrant `json:"grants"`
}

// Password is only populated on create.
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
