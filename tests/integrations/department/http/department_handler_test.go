package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	departmenthttp "github.com/AppeiYA/requisition-system/internal/department/adapters/in/http"
	departmentpostgres "github.com/AppeiYA/requisition-system/internal/department/adapters/out/postgres"
	departmentin "github.com/AppeiYA/requisition-system/internal/department/ports/in"
	departmentusecase "github.com/AppeiYA/requisition-system/internal/department/usecase"
	appdb "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/generator"
	redisinfra "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/redis"
	"github.com/AppeiYA/requisition-system/tests/testutil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestDepartmentCreateHandler(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	redisResource := suite.SetupRedis(t)

	baseRepository := appdb.NewRepository(suite.AppDB)
	cache := redisinfra.NewCache(redisResource.Client)
	departmentRepository := departmentpostgres.NewDepartmentRepository(baseRepository, cache)

	createUsecase := departmentusecase.NewCreateDepartmentUsecase(
		departmentRepository,
		&generator.IDGenerator{},
	)
	getByIDUsecase := departmentusecase.NewGetDepartmentByIDUsecase(departmentRepository)
	getDepartmentsUsecase := departmentusecase.NewGetDepartmentsUsecase(departmentRepository)
	ports := departmentin.NewDepartmentPortIn(nil, createUsecase, getByIDUsecase, getDepartmentsUsecase)
	handler := departmenthttp.NewDepartmentHandler(ports)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler.Register(router.Group("/api/v1"), func(c *gin.Context) { c.Next() })

	tests := []struct {
		name           string
		body           map[string]string
		expectedStatus int
	}{
		{
			name: "success",
			body: map[string]string{
				"name":        "Engineering",
				"description": "Engineering department",
				"code":        "ENG",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing required fields",
			body: map[string]string{
				"name": "Missing Data",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.body)
			require.NoError(t, err)
			request := httptest.NewRequest(http.MethodPost, "/api/v1/departments", bytes.NewReader(body))
			request.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			require.Equal(t, tt.expectedStatus, recorder.Code, recorder.Body.String())
		})
	}
}
