package identityout

// OpaqueTokenGenerator produces cryptographically secure opaque tokens and
// derives the deterministic hash under which a token is persisted and looked
// up. Refresh tokens and password reset tokens share the same contract.
type OpaqueTokenGenerator interface {
	// Generate returns a new cryptographically secure, URL-safe opaque token.
	Generate() (string, error)

	// Hash returns the storage hash for a raw token. It must be deterministic
	// and must not be reversible.
	Hash(token string) string
}
