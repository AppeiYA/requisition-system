package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

type TransitionRequisitionUsecase struct {
	transitionRepo requisitionout.RequisitionTransitionRepository
	idGenerator    requisitionout.IDGenerator
	txManager      requisitionout.TransactionManager
}

func NewTransitionRequisitionUsecase(
	transitionRepo requisitionout.RequisitionTransitionRepository,
	idGenerator requisitionout.IDGenerator,
	txManager requisitionout.TransactionManager,
) *TransitionRequisitionUsecase {
	return &TransitionRequisitionUsecase{
		transitionRepo: transitionRepo,
		idGenerator:    idGenerator,
		txManager:      txManager,
	}
}

func (uc *TransitionRequisitionUsecase) Execute(
	ctx context.Context,
	cmd requisitionin.TransitionRequisitionCommand,
) (*domain.Requisition, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !strings.HasPrefix(cmd.RequisitionID, domain.RequisitionPrefixID) {
		return nil, domain.ErrInvalidRequisitionID
	}
	if cmd.ActorID == "" {
		return nil, domain.ErrRequestedIDCannotBeEmpty
	}

	toStatus, err := domain.NewRequisitionStatus(cmd.ToStatus)
	if err != nil {
		return nil, err
	}
	actorRole, err := actorRoleFromIdentity(cmd.ActorRole)
	if err != nil {
		return nil, err
	}
	if uc.txManager == nil {
		return nil, customerrors.NewInternalServerError("requisition transaction manager is not configured")
	}

	historyID, err := uc.idGenerator.GenerateID(domain.StatusHistoryPrefixID)
	if err != nil {
		return nil, err
	}

	var updated *domain.Requisition
	err = uc.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		requisition, err := uc.transitionRepo.GetRequisitionByIDForUpdate(txCtx, cmd.RequisitionID)
		if err != nil {
			return err
		}
		if err := authorizeTransition(requisition, cmd.ActorID, actorRole, toStatus); err != nil {
			return err
		}

		history, err := requisition.ApplyStatusTransition(
			historyID, toStatus, cmd.ActorID, actorRole, cmd.Reason, time.Now().UTC(),
		)
		if err != nil {
			return err
		}
		if err := uc.transitionRepo.UpdateRequisitionStatus(txCtx, requisition, history); err != nil {
			return err
		}
		updated = requisition
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// authorizeTransition keeps the policy narrow: administrators may perform any
// transition the domain allows, while a non-admin requester may only cancel
// their own pending requisition.
func authorizeTransition(
	requisition *domain.Requisition,
	actorID string,
	actorRole domain.RequisitionActorRole,
	toStatus domain.RequisitionStatus,
) error {
	if actorRole == domain.ActorRoleAdmin {
		return nil
	}
	if requisition.RequestedBy() != actorID {
		return ErrRequisitionTransitionForbidden
	}
	if requisition.Status() != domain.RequisitionStatusPending || toStatus != domain.RequisitionStatusCancelled {
		return ErrRequisitionTransitionForbidden
	}
	return nil
}
