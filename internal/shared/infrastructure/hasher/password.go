package hasher

import "golang.org/x/crypto/bcrypt"

type PasswordHasher struct {
	cost int
}

func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{cost: bcrypt.DefaultCost}
}

func NewPasswordHasherWithCost(cost int) *PasswordHasher {
	return &PasswordHasher{cost: cost}
}

func (h *PasswordHasher) HashPassword(password string) (string, error) {
	cost := h.cost
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (h *PasswordHasher) ComparePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (h *PasswordHasher) Hash(password string) (string, error) {
	return h.HashPassword(password)
}

func (h *PasswordHasher) Compare(hashedPassword, password string) error {
	return h.ComparePassword(hashedPassword, password)
}
