package response

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type bindRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Age      int    `json:"age" binding:"gte=18"`
}

// bindError runs the same binding path the handlers use and returns the raw
// error so ValidationErrors can be exercised against real errors.
func bindError(t *testing.T, body string) error {
	t.Helper()
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	var req bindRequest
	return c.ShouldBindJSON(&req)
}

func TestValidationErrors_Nil(t *testing.T) {
	if got := ValidationErrors(nil); got != nil {
		t.Fatalf("ValidationErrors(nil) = %v, want nil", got)
	}
}

func TestValidationErrors_FieldRules(t *testing.T) {
	err := bindError(t, `{"email":"not-an-email","password":"short","age":10}`)
	if err == nil {
		t.Fatal("expected a binding error, got nil")
	}

	details := ValidationErrors(err)
	if len(details) != 3 {
		t.Fatalf("got %d details, want 3: %+v", len(details), details)
	}

	byField := make(map[string]string, len(details))
	for _, d := range details {
		byField[d.Field] = d.Message
	}

	if byField["email"] != "must be a valid email address" {
		t.Errorf("email message = %q", byField["email"])
	}
	if byField["password"] != "must be at least 8 characters long" {
		t.Errorf("password message = %q", byField["password"])
	}
	if byField["age"] != "must be greater than or equal to 18" {
		t.Errorf("age message = %q", byField["age"])
	}
}

func TestValidationErrors_Required(t *testing.T) {
	err := bindError(t, `{"age":20}`)
	details := ValidationErrors(err)

	sawRequired := false
	for _, d := range details {
		if d.Message == "this field is required" {
			sawRequired = true
		}
	}
	if !sawRequired {
		t.Fatalf("expected a required-field message, got: %+v", details)
	}
}

func TestValidationErrors_MalformedJSON(t *testing.T) {
	err := bindError(t, `{"email":`)
	details := ValidationErrors(err)

	if len(details) != 1 || details[0].Field != "body" {
		t.Fatalf("got %+v, want a single body detail", details)
	}
}

func TestValidationErrors_EmptyBody(t *testing.T) {
	err := bindError(t, "")
	if err == nil {
		t.Fatal("expected a binding error for empty body, got nil")
	}

	details := ValidationErrors(err)
	if len(details) != 1 || details[0].Field != "body" {
		t.Fatalf("got %+v, want a single body detail", details)
	}
}

func TestValidationErrors_WrongType(t *testing.T) {
	err := bindError(t, `{"email":"a@b.com","password":"longenough","age":"twenty"}`)
	details := ValidationErrors(err)

	if len(details) != 1 || details[0].Field != "age" || details[0].Message != "must be a valid int" {
		t.Fatalf("got %+v, want age/int type error", details)
	}
}

func TestValidationErrors_UnknownError(t *testing.T) {
	details := ValidationErrors(io.ErrClosedPipe)

	if len(details) != 1 || details[0].Field != "body" || details[0].Message != "invalid request body" {
		t.Fatalf("got %+v, want generic body detail", details)
	}
}
