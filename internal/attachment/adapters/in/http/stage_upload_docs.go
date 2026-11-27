package attachmenthttp

import (
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
)

var _ = response.Response{}

// stageUploadDoc documents the staged upload handler in stage_upload.go.
//
//	@Summary		Stage a file before creating a requisition
//	@Description	Streams a file to temporary storage and records its owner. No requisition or requisition-item ID is required; pass the returned ID in the requisition creation payload.
//	@Tags			attachment
//	@Accept			mpfd
//	@Produce		json
//	@Security		BearerAuth
//	@Param			file	formData	file	true	"File to stage"
//	@Success		201		{object}	response.Response{data=StagedUploadResponse}
//	@Failure		400		{object}	response.ErrorResponse	"Invalid multipart body"
//	@Failure		401		{object}	response.Response		"Missing or invalid access token"
//	@Failure		413		{object}	response.Response		"File exceeds the configured size limit"
//	@Failure		415		{object}	response.Response		"Unsupported content type"
//	@Router			/uploads [post]
func stageUploadDoc() {}
