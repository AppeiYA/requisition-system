// Package mime detects an attachment's content type from its bytes.
package mime

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"io"
	"net/http"
	"strings"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

const (
	maxZipEntries        = 128
	maxContentTypesBytes = 64 << 10
	maxCSVSampleLines    = 10
)

var _ attachmentout.ContentTypeDetector = (*Detector)(nil)

// Detector derives the stored content type from the leading bytes of a file.
// Declared types may never widen the result beyond what the bytes prove.
type Detector struct{}

func New() *Detector {
	return &Detector{}
}

// Detect sniffs media types using http.DetectContentType, inspects zip
// containers for OOXML [Content_Types].xml with a zip-bomb guard, and applies
// a bounded CSV heuristic on text/plain payloads.
func (d *Detector) Detect(head []byte, declared string) (domain.ContentType, error) {
	_ = declared
	if len(head) == 0 {
		return "", customerrors.NewCustomError("unsupported attachment content type", http.StatusUnsupportedMediaType)
	}

	mediaType := strings.ToLower(strings.TrimSpace(strings.SplitN(http.DetectContentType(head), ";", 2)[0]))

	switch mediaType {
	case "application/zip", "application/x-zip-compressed":
		ooxml, ok := detectOOXMLFromZip(head)
		if !ok {
			return "", customerrors.NewCustomError("unsupported attachment content type", http.StatusUnsupportedMediaType)
		}
		mediaType = ooxml
	case "text/plain":
		if isLikelyCSV(head) {
			mediaType = string(domain.ContentTypeCSV)
		} else {
			return "", customerrors.NewCustomError("unsupported attachment content type", http.StatusUnsupportedMediaType)
		}
	}

	contentType, err := domain.NewContentType(mediaType)
	if err != nil {
		return "", customerrors.NewCustomError("unsupported attachment content type", http.StatusUnsupportedMediaType)
	}
	return contentType, nil
}

func detectOOXMLFromZip(head []byte) (string, bool) {
	zr, err := zip.NewReader(bytes.NewReader(head), int64(len(head)))
	if err != nil {
		// When only a truncated leading prefix of a large OOXML zip is
		// available, the central directory is past the head; fall back to
		// checking local-header filenames in the prefix.
		if bytes.HasPrefix(head, []byte("PK\x03\x04")) && len(head) >= 4096 {
			return ooxmlFromZipPrefix(head)
		}
		return "", false
	}

	var hasContentTypes bool
	var hasWord, hasXL bool
	var contentTypesXML []byte

	limit := len(zr.File)
	if limit > maxZipEntries {
		limit = maxZipEntries
	}
	for i := 0; i < limit; i++ {
		f := zr.File[i]
		name := strings.TrimPrefix(f.Name, "/")
		switch {
		case name == "[Content_Types].xml":
			hasContentTypes = true
			rc, openErr := f.Open()
			if openErr == nil {
				contentTypesXML, _ = io.ReadAll(io.LimitReader(rc, maxContentTypesBytes))
				_ = rc.Close()
			}
		case strings.HasPrefix(name, "word/"):
			hasWord = true
		case strings.HasPrefix(name, "xl/"):
			hasXL = true
		}
	}

	if !hasContentTypes {
		return "", false
	}
	if hasWord || bytes.Contains(contentTypesXML, []byte("wordprocessingml.document")) {
		return string(domain.ContentTypeDOCX), true
	}
	if hasXL || bytes.Contains(contentTypesXML, []byte("spreadsheetml.sheet")) {
		return string(domain.ContentTypeXLSX), true
	}
	return "", false
}

func ooxmlFromZipPrefix(head []byte) (string, bool) {
	if !bytes.Contains(head, []byte("[Content_Types].xml")) {
		return "", false
	}
	switch {
	case bytes.Contains(head, []byte("word/")):
		return string(domain.ContentTypeDOCX), true
	case bytes.Contains(head, []byte("xl/")):
		return string(domain.ContentTypeXLSX), true
	default:
		return "", false
	}
}

func isLikelyCSV(head []byte) bool {
	if bytes.IndexByte(head, 0) >= 0 {
		return false
	}
	trimmed := strings.TrimSpace(string(head))
	if trimmed == "" || strings.HasPrefix(trimmed, "<") {
		return false
	}

	lines := make([]string, 0, maxCSVSampleLines)
	for _, line := range strings.Split(trimmed, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		lines = append(lines, line)
		if len(lines) == maxCSVSampleLines {
			break
		}
	}
	if len(lines) < 2 {
		return false
	}

	sample := strings.Join(lines, "\n")
	for _, delim := range []rune{',', ';', '\t'} {
		r := csv.NewReader(strings.NewReader(sample))
		r.Comma = delim
		r.FieldsPerRecord = 0
		records, err := r.ReadAll()
		if err != nil || len(records) < 2 {
			continue
		}
		if len(records[0]) >= 2 {
			return true
		}
	}
	return false
}
