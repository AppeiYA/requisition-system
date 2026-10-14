package adminhttp

type InviteRequest struct {
	Email string `json:"email" binding:"required,email"`
}
