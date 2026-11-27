package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	identitytasks "github.com/AppeiYA/requisition-system/internal/identity/tasks"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
)

// RequestPasswordResetUsecase starts the self-service reset flow. It never
// reveals whether an account exists: unknown, inactive and throttled requests
// all return nil without creating or sending anything.
type RequestPasswordResetUsecase struct {
	userRepo       identityout.UsersRepository
	tokenRepo      identityout.PasswordResetTokenRepository
	queue          identityout.PasswordResetEmailQueue
	tokenGenerator identityout.OpaqueTokenGenerator
	idGenerator    identityout.IDGenerator
	txManager      identityout.TransactionManager
	config         config.PasswordResetConfig
	telemetry      identityout.Telemetry
}

func NewRequestPasswordResetUsecase(
	userRepo identityout.UsersRepository,
	tokenRepo identityout.PasswordResetTokenRepository,
	queue identityout.PasswordResetEmailQueue,
	tokenGenerator identityout.OpaqueTokenGenerator,
	idGenerator identityout.IDGenerator,
	txManager identityout.TransactionManager,
	cfg config.PasswordResetConfig,
	telemetry identityout.Telemetry,
) *RequestPasswordResetUsecase {
	return &RequestPasswordResetUsecase{
		userRepo:       userRepo,
		tokenRepo:      tokenRepo,
		queue:          queue,
		tokenGenerator: tokenGenerator,
		idGenerator:    idGenerator,
		txManager:      txManager,
		config:         cfg,
		telemetry:      telemetry,
	}
}

func (u *RequestPasswordResetUsecase) Execute(ctx context.Context, email string) error {
	return u.telemetry.Measure(ctx, identityout.OperationRequestPasswordReset, func(ctx context.Context) error {
		return u.execute(ctx, email)
	})
}

func (u *RequestPasswordResetUsecase) execute(ctx context.Context, email string) error {
	emailAddr, err := shareddomain.NewEmail(email)
	if err != nil {
		return err
	}

	// The lookup runs for every address. A missing account is indistinguishable
	// from a throttled one at the HTTP boundary.
	user, err := u.userRepo.GetUserByEmail(ctx, emailAddr)
	if err != nil {
		if isNotFoundError(err) {
			u.logOutcome(ctx, "unknown", "")
			return nil
		}
		return err
	}
	if user == nil {
		u.logOutcome(ctx, "unknown", "")
		return nil
	}
	if !user.IsActive() {
		u.logOutcome(ctx, "inactive", user.ID())
		return nil
	}

	var payload *identitytasks.PasswordResetEmailPayload
	outcome := "throttled"
	err = u.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		// The user row is the serialisation point. Every reset request for
		// this account takes it before touching token rows.
		locked, err := u.userRepo.GetUserByIDForUpdate(txCtx, user.ID())
		if err != nil {
			if isNotFoundError(err) {
				outcome = "unknown"
				return nil
			}
			return err
		}
		if locked == nil {
			outcome = "unknown"
			return nil
		}
		if !locked.IsActive() {
			outcome = "inactive"
			return nil
		}

		now := time.Now().UTC()
		latest, err := u.tokenRepo.GetLatestActiveForUser(txCtx, locked.ID(), now)
		if err != nil {
			return err
		}
		if latest != nil && u.config.RequestCooldown > 0 &&
			now.Sub(latest.CreatedAt()) < u.config.RequestCooldown {
			outcome = "throttled"
			return nil
		}

		if u.config.MaxPerHour > 0 {
			count, err := u.tokenRepo.CountCreatedSince(txCtx, locked.ID(), now.Add(-time.Hour))
			if err != nil {
				return err
			}
			if count >= u.config.MaxPerHour {
				outcome = "throttled"
				return nil
			}
		}

		// Expired or superseded rows are made terminal here, before the
		// partial unique index can be violated by the insert below.
		if err := u.tokenRepo.InvalidateActiveForUser(txCtx, locked.ID(), now); err != nil {
			return err
		}

		rawToken, err := u.tokenGenerator.Generate()
		if err != nil {
			return err
		}
		tokenHash := u.tokenGenerator.Hash(rawToken)

		resetID, err := u.idGenerator.GenerateID(domain.PasswordResetTokenIDPrefix)
		if err != nil {
			return err
		}

		ttl := u.config.TokenTTL
		if ttl <= 0 {
			ttl = domain.DefaultPasswordResetTokenTTL
		}
		token, err := domain.NewPasswordResetToken(resetID, locked.ID(), tokenHash, now, ttl)
		if err != nil {
			return err
		}
		if err := u.tokenRepo.Create(txCtx, token); err != nil {
			return err
		}

		outcome = "accepted"
		payload = &identitytasks.PasswordResetEmailPayload{
			ResetID:   resetID,
			To:        locked.Email().Value(),
			Token:     rawToken,
			ExpiresAt: token.ExpiresAt().Format(time.RFC3339),
		}
		return nil
	})
	if err != nil {
		return err
	}

	u.logOutcome(ctx, outcome, user.ID())
	if outcome != "accepted" || payload == nil {
		return nil
	}

	// Enqueue only after commit, so a rolled-back token is never emailed.
	if err := u.queue.EnqueuePasswordResetEmail(ctx, *payload); err != nil {
		// Invalidate by ID, not by user: a concurrent request may already
		// have created a newer token. The HTTP response stays generic.
		if invalidateErr := u.tokenRepo.InvalidateByID(ctx, payload.ResetID, time.Now().UTC()); invalidateErr != nil {
			slog.ErrorContext(ctx, "password_reset.email_enqueue_failed",
				slog.String("event", "password_reset.email_enqueue_failed"),
				slog.String("reset.id", payload.ResetID),
				slog.String("reason", "invalidate_failed"),
			)
		} else {
			slog.WarnContext(ctx, "password_reset.email_enqueue_failed",
				slog.String("event", "password_reset.email_enqueue_failed"),
				slog.String("reset.id", payload.ResetID),
				slog.String("reason", "enqueue_failed"),
			)
		}
		return nil
	}

	slog.InfoContext(ctx, "password_reset.email_enqueued",
		slog.String("event", "password_reset.email_enqueued"),
		slog.String("reset.id", payload.ResetID),
	)
	return nil
}

func (u *RequestPasswordResetUsecase) logOutcome(ctx context.Context, outcome, userID string) {
	attrs := []any{
		slog.String("event", "password_reset.requested"),
		slog.String("outcome", outcome),
	}
	if userID != "" {
		attrs = append(attrs, slog.String("user.id", userID))
	}
	slog.InfoContext(ctx, "password_reset.requested", attrs...)
}
