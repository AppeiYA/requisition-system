package usecase

import (
	"bufio"
	"bytes"
	"errors"
	"io"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

// sniffSize is how many leading bytes are handed to the content-type detector.
const sniffSize = 4096

// detectUploadBody peeks only the detector's small buffer, then replays those
// bytes in front of the remaining stream. The full body is never held in
// memory and the caller owns closing the original reader.
func detectUploadBody(
	body io.Reader,
	detector attachmentout.ContentTypeDetector,
) (domain.ContentType, io.Reader, error) {
	if body == nil {
		return "", nil, ErrUnsupportedContentType
	}

	buffered := bufio.NewReader(body)
	head, peekErr := buffered.Peek(sniffSize)
	if peekErr != nil && !errors.Is(peekErr, io.EOF) {
		return "", nil, peekErr
	}

	contentType, err := detector.Detect(head, "")
	if err != nil {
		var customErr *customerrors.CustomError
		if errors.As(err, &customErr) {
			return "", nil, err
		}
		return "", nil, ErrUnsupportedContentType
	}
	if !contentType.IsValid() {
		return "", nil, ErrUnsupportedContentType
	}

	if _, err := buffered.Discard(len(head)); err != nil {
		return "", nil, err
	}
	return contentType, io.MultiReader(bytes.NewReader(head), buffered), nil
}

// attachmentSizeLimiter accepts at most max bytes and reports 413 as soon as
// one more byte is read. It never reads past max+1 bytes from the body.
type attachmentSizeLimiter struct {
	reader io.Reader
	max    int64
	read   int64
}

func (l *attachmentSizeLimiter) Read(p []byte) (int, error) {
	remaining := l.max + 1 - l.read
	if remaining <= 0 {
		return 0, ErrAttachmentTooLarge
	}
	if int64(len(p)) > remaining {
		p = p[:remaining]
	}
	n, err := l.reader.Read(p)
	l.read += int64(n)
	if l.read > l.max {
		return n, ErrAttachmentTooLarge
	}
	return n, err
}
