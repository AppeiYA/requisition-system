package postgres_test

import (
	"context"
	"testing"

	departmentpostgres "github.com/AppeiYA/requisition-system/internal/department/adapters/out/postgres"
	departmentdomain "github.com/AppeiYA/requisition-system/internal/department/domain"
	appdb "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	redisinfra "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/redis"
	"github.com/AppeiYA/requisition-system/tests/testutil"
	"github.com/stretchr/testify/require"
)

func TestDepartmentRepositorySaveAndGet(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	ctx := context.Background()
	require.NoError(t, suite.Reset(ctx))

	redisResource := suite.SetupRedis(t)
	repository := departmentpostgres.NewDepartmentRepository(
		appdb.NewRepository(suite.AppDB),
		redisinfra.NewCache(redisResource.Client),
	)

	name, err := departmentdomain.NewDepartmentName("Finance")
	require.NoError(t, err)
	description, err := departmentdomain.NewDepartmentDescription("Finance department")
	require.NoError(t, err)
	code, err := departmentdomain.NewDepartmentCode("FIN")
	require.NoError(t, err)
	department, err := departmentdomain.NewDepartment("dpt-finance", name, description, code)
	require.NoError(t, err)

	require.NoError(t, repository.SaveDepartment(ctx, department))

	found, err := repository.GetDepartmentByID(ctx, department.ID())
	require.NoError(t, err)
	require.Equal(t, department.ID(), found.ID())
	require.Equal(t, "FIN", found.Code().Value())
	require.Equal(t, "Finance", found.Name().Value())
}
