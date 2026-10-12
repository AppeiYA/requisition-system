package identityout

type PasswordHasher interface {
	HashPassword(password string) (string, error)
	ComparePassword(hashedPassword, password string) error
}
