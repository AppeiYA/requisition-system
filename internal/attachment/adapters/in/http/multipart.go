package attachmenthttp

import (
	"errors"
	"io"
	"net/http"
	"strings"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

const maxFormFieldBytes = 1 << 10

type parsedUploadMultipart struct {
	RefName  string
	RefID    string
	FileName string
	Body     io.ReadCloser
}

// parseAttachmentMultipart streams the request with MultipartReader and
// enforces maxSizeBytes via http.MaxBytesReader. When requireParent is true,
// ref_name and ref_id must precede the file part so the body is never buffered.
func parseAttachmentMultipart(c *gin.Context, maxSizeBytes int64, requireParent bool) (*parsedUploadMultipart, bool) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSizeBytes)

	mr, err := c.Request.MultipartReader()
	if err != nil {
		if isAttachmentTooLarge(err) {
			response.ErrorFrom(c, customerrors.NewCustomError(
				"attachment exceeds the maximum allowed size",
				http.StatusRequestEntityTooLarge,
			))
			c.Abort()
			return nil, false
		}
		response.ErrorWithDetails(
			c,
			"validation_error",
			"invalid multipart upload",
			http.StatusBadRequest,
			response.ValidationErrors(err),
		)
		c.Abort()
		return nil, false
	}

	var refName, refID string
	for {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			if isAttachmentTooLarge(err) {
				response.ErrorFrom(c, customerrors.NewCustomError(
					"attachment exceeds the maximum allowed size",
					http.StatusRequestEntityTooLarge,
				))
				c.Abort()
				return nil, false
			}
			response.ErrorWithDetails(
				c,
				"validation_error",
				"invalid multipart upload",
				http.StatusBadRequest,
				response.ValidationErrors(err),
			)
			c.Abort()
			return nil, false
		}

		fieldName := part.FormName()
		if requireParent && (fieldName == "ref_name" || fieldName == "ref_id") {
			valBytes, readErr := io.ReadAll(io.LimitReader(part, maxFormFieldBytes))
			_ = part.Close()
			if readErr != nil {
				if isAttachmentTooLarge(readErr) {
					response.ErrorFrom(c, customerrors.NewCustomError(
						"attachment exceeds the maximum allowed size",
						http.StatusRequestEntityTooLarge,
					))
					c.Abort()
					return nil, false
				}
				response.ErrorWithDetails(
					c,
					"validation_error",
					"invalid multipart upload",
					http.StatusBadRequest,
					response.ValidationErrors(readErr),
				)
				c.Abort()
				return nil, false
			}
			val := strings.TrimSpace(string(valBytes))
			if fieldName == "ref_name" {
				refName = val
			} else {
				refID = val
			}
			continue
		}

		if fieldName == "file" {
			if requireParent && (refName == "" || refID == "") {
				_ = part.Close()
				details := make([]response.FieldError, 0, 2)
				if refName == "" {
					details = append(details, response.FieldError{Field: "ref_name", Message: "this field is required before file"})
				}
				if refID == "" {
					details = append(details, response.FieldError{Field: "ref_id", Message: "this field is required before file"})
				}
				response.ErrorWithDetails(
					c,
					"validation_error",
					"ref_name and ref_id must precede the file part",
					http.StatusBadRequest,
					details,
				)
				c.Abort()
				return nil, false
			}
			return &parsedUploadMultipart{
				RefName:  refName,
				RefID:    refID,
				FileName: part.FileName(),
				Body:     part,
			}, true
		}

		_ = part.Close()
	}

	response.ErrorWithDetails(
		c,
		"validation_error",
		"file is required",
		http.StatusBadRequest,
		[]response.FieldError{{Field: "file", Message: "this field is required"}},
	)
	c.Abort()
	return nil, false
}

func isAttachmentTooLarge(err error) bool {
	if err == nil {
		return false
	}

	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "http: request body too large")
}
