package identityhttp

import (
	"time"

	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
)

// User represents a user in the identity service.
type User struct {
	ID        string     `json:"id"`
	Email     string     `json:"email"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

func ToJSON(u *identityin.UserResponse) *User {
	return &User{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
