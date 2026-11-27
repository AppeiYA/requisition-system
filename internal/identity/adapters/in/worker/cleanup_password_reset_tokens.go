package identityworkerin

import (
	"context"

	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	"github.com/hibiken/asynq"
)

// CleanupPasswordResetTokensHandler runs the scheduled retention sweep.
type CleanupPasswordResetTokensHandler struct {
	usecase identityin.CleanupPasswordResetTokensUsecaseInt
}

func NewCleanupPasswordResetTokensHandler(usecase identityin.CleanupPasswordResetTokensUsecaseInt) *CleanupPasswordResetTokensHandler {
	return &CleanupPasswordResetTokensHandler{usecase: usecase}
}

func (h *CleanupPasswordResetTokensHandler) Handle(ctx context.Context, _ *asynq.Task) error {
	_, err := h.usecase.Execute(ctx)
	return err
}
