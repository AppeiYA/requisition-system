package adminhttp

import (
	"time"

	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
)

type InviteRequest struct {
	Email        string `json:"email" binding:"required,email"`
	FirstName    string `json:"first_name" binding:"required"`
	LastName     string `json:"last_name" binding:"required"`
	Phone        string `json:"phone" binding:"omitempty"`
	StaffID      string `json:"staff_id" binding:"omitempty"`
	DepartmentID string `json:"department_id" binding:"required"`
	JobTitle     string `json:"job_title" binding:"required,max=60"`
}

type UpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active suspended inactive"`
	Reason string `json:"reason" binding:"omitempty,max=500"`
}

type UpdateUserRoleRequest struct {
	Role   string `json:"role" binding:"required,oneof=admin user"`
	Reason string `json:"reason" binding:"omitempty,max=500"`
}

type UserSummaryResponse struct {
	ID               string     `json:"id"`
	FirstName        string     `json:"first_name"`
	LastName         string     `json:"last_name"`
	Email            string     `json:"email"`
	StaffID          string     `json:"staff_id"`
	Phone            string     `json:"phone"`
	Role             string     `json:"role"`
	Status           string     `json:"status"`
	DepartmentID     string     `json:"department_id"`
	DepartmentName   string     `json:"department_name"`
	JobTitle         string     `json:"job_title"`
	MembershipStatus string     `json:"membership_status"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        *time.Time `json:"updated_at"`
}

type UserDetailResponse struct {
	UserSummaryResponse
	RequisitionCounts    map[string]int `json:"requisition_counts"`
	TotalEstimatedAmount float64        `json:"total_estimated_amount"`
	LastRequisitionAt    *time.Time     `json:"last_requisition_at"`
}

func toUserSummaryResponse(user adminout.UserSummary) UserSummaryResponse {
	return UserSummaryResponse{
		ID: user.ID, FirstName: user.FirstName, LastName: user.LastName, Email: user.Email,
		StaffID: user.StaffID, Phone: user.Phone, Role: user.Role, Status: user.Status,
		DepartmentID: user.DepartmentID, DepartmentName: user.DepartmentName, JobTitle: user.JobTitle,
		MembershipStatus: user.MembershipStatus, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt,
	}
}

func toUserSummaryResponses(users []adminout.UserSummary) []UserSummaryResponse {
	responses := make([]UserSummaryResponse, len(users))
	for i, user := range users {
		responses[i] = toUserSummaryResponse(user)
	}
	return responses
}

func toUserDetailResponse(user adminout.UserDetail) UserDetailResponse {
	return UserDetailResponse{
		UserSummaryResponse: toUserSummaryResponse(user.UserSummary),
		RequisitionCounts:   user.RequisitionCounts, TotalEstimatedAmount: user.TotalEstimatedAmount,
		LastRequisitionAt: user.LastRequisitionAt,
	}
}

type InvitationSummaryResponse struct {
	ID               string     `json:"id"`
	FirstName        string     `json:"first_name"`
	LastName         string     `json:"last_name"`
	Email            string     `json:"email"`
	DepartmentID     string     `json:"department_id"`
	DepartmentName   string     `json:"department_name"`
	JobTitle         string     `json:"job_title"`
	Status           string     `json:"status"`
	InvitedBy        string     `json:"invited_by"`
	InviterName      string     `json:"inviter_name"`
	ExpiresAt        time.Time  `json:"expires_at"`
	AcceptedAt       *time.Time `json:"accepted_at"`
	RevokedAt        *time.Time `json:"revoked_at"`
	DeliveredAt      *time.Time `json:"delivered_at"`
	DeliveryFailedAt *time.Time `json:"delivery_failed_at"`
	CreatedAt        time.Time  `json:"created_at"`
}

type InvitationDetailResponse struct {
	InvitationSummaryResponse
}

// ResendInvitationResponse deliberately omits the rotated raw token: it is
// delivered by email only and must never be echoed back to an administrator.
type ResendInvitationResponse struct {
	ID        string    `json:"id"`
	ExpiresAt time.Time `json:"expires_at"`
}

func toInvitationSummaryResponse(invitation adminout.InvitationSummary) InvitationSummaryResponse {
	return InvitationSummaryResponse{
		ID:               invitation.ID,
		FirstName:        invitation.FirstName,
		LastName:         invitation.LastName,
		Email:            invitation.Email,
		DepartmentID:     invitation.DepartmentID,
		DepartmentName:   invitation.DepartmentName,
		JobTitle:         invitation.JobTitle,
		Status:           invitation.Status,
		InvitedBy:        invitation.InvitedBy,
		InviterName:      invitation.InviterName,
		ExpiresAt:        invitation.ExpiresAt,
		AcceptedAt:       invitation.AcceptedAt,
		RevokedAt:        invitation.RevokedAt,
		DeliveredAt:      invitation.DeliveredAt,
		DeliveryFailedAt: invitation.DeliveryFailedAt,
		CreatedAt:        invitation.CreatedAt,
	}
}

func toInvitationSummaryResponses(invitations []adminout.InvitationSummary) []InvitationSummaryResponse {
	responses := make([]InvitationSummaryResponse, len(invitations))
	for i, invitation := range invitations {
		responses[i] = toInvitationSummaryResponse(invitation)
	}
	return responses
}

func toInvitationDetailResponse(invitation adminout.InvitationSummary) InvitationDetailResponse {
	return InvitationDetailResponse{InvitationSummaryResponse: toInvitationSummaryResponse(invitation)}
}
