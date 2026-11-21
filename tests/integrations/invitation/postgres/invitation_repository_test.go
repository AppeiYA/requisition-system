package postgres_test

import (
	"context"
	"testing"

	invitationpostgres "github.com/AppeiYA/requisition-system/internal/invitation/adapters/out/postgres"
	invitationdomain "github.com/AppeiYA/requisition-system/internal/invitation/domain"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	appdb "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	"github.com/AppeiYA/requisition-system/tests/testutil"
	"github.com/stretchr/testify/require"
)

func TestInvitationRepositorySaveAndGetByTokenHash(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	ctx := context.Background()
	require.NoError(t, suite.ResetDatabase(ctx))

	_, err := suite.DB.Pool.Exec(ctx, `
		INSERT INTO users (
			id, first_name, last_name, email, password_hash, status, role
		) VALUES (
			'usr-admin', 'Admin', 'User', 'admin@example.com',
			'integration-password-hash', 'active', 'admin'
		);
		INSERT INTO departments (id, name, description, code, status)
		VALUES ('dpt-repository', 'Repository', 'Repository department', 'REP', 'active');
	`)
	require.NoError(t, err)

	repository := invitationpostgres.NewInvitationRepository(appdb.NewRepository(suite.AppDB))
	email, err := shareddomain.NewEmail("invitation.repo@example.com")
	require.NoError(t, err)
	firstName, err := shareddomain.NewUsername("Invitation", shareddomain.FirstName)
	require.NoError(t, err)
	lastName, err := shareddomain.NewUsername("Repository", shareddomain.LastName)
	require.NoError(t, err)
	phone, err := shareddomain.NewOptionalPhoneNumber("")
	require.NoError(t, err)

	invitation, err := invitationdomain.NewInvitation(
		"inv-repository",
		firstName,
		lastName,
		phone,
		"",
		"dpt-repository",
		"Analyst",
		email,
		invitationdomain.HashInvitationToken("repository-token"),
		"usr-admin",
	)
	require.NoError(t, err)

	require.NoError(t, repository.SaveInvitation(ctx, invitation))
	found, err := repository.GetInvitationByTokenHash(
		ctx,
		invitationdomain.HashInvitationToken("repository-token"),
	)
	require.NoError(t, err)
	require.Equal(t, invitation.ID(), found.ID())
	require.Equal(t, "invitation.repo@example.com", found.Email().Value())
}
