package profile

// Profile represents a named Git identity.
type Profile struct {
	// Name is the profile identifier (e.g., "work", "personal").
	Name string

	// UserName is the git user.name value.
	UserName string

	// UserEmail is the git user.email value.
	UserEmail string

	// SigningKey is an optional GPG/SSH signing key ID.
	SigningKey string

	// SSHKeyPath is an optional path to the SSH private key.
	SSHKeyPath string
}
