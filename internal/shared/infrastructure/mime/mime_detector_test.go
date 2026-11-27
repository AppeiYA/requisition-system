package mime_test

import (
	"archive/zip"
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math/rand"
	"net/http"
	"testing"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
	attachmentmime "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/mime"
)

func TestDetectFromBytes(t *testing.T) {
	detector := attachmentmime.New()

	tests := []struct {
		name     string
		head     []byte
		declared string
		want     domain.ContentType
	}{
		{"pdf", []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF\n"), "text/plain", domain.ContentTypePDF},
		{"png", encodedImage(t, png.Encode), "application/pdf", domain.ContentTypePNG},
		{"jpeg", encodedImage(t, func(w io.Writer, m image.Image) error { return jpeg.Encode(w, m, nil) }), "image/png", domain.ContentTypeJPEG},
		{"gif", encodedImage(t, func(w io.Writer, m image.Image) error { return gif.Encode(w, m, nil) }), "application/octet-stream", domain.ContentTypeGIF},
		{"docx", ooxmlBytes(t, "word/document.xml", "application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml", false), "application/zip", domain.ContentTypeDOCX},
		{"docx with spoofed pdf declared type", ooxmlBytes(t, "word/document.xml", "application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml", false), "application/pdf", domain.ContentTypeDOCX},
		{"xlsx", ooxmlBytes(t, "xl/workbook.xml", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml", false), "application/octet-stream", domain.ContentTypeXLSX},
		{"csv", []byte("name,role\nalice,admin\nbob,viewer\n"), "text/plain", domain.ContentTypeCSV},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := detector.Detect(tc.head, tc.declared)
			if err != nil {
				t.Fatalf("Detect: %v", err)
			}
			if got != tc.want {
				t.Fatalf("Detect = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDetectWith4096ByteHeadRecognisesOOXML(t *testing.T) {
	detector := attachmentmime.New()

	tests := []struct {
		name     string
		mainPart string
		override string
		want     domain.ContentType
	}{
		{
			name:     "docx",
			mainPart: "word/document.xml",
			override: "application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml",
			want:     domain.ContentTypeDOCX,
		},
		{
			name:     "xlsx",
			mainPart: "xl/workbook.xml",
			override: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml",
			want:     domain.ContentTypeXLSX,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			full := ooxmlBytes(t, tc.mainPart, tc.override, true)
			if len(full) <= 4096 {
				t.Fatalf("fixture is only %d bytes; need a larger zip to test a 4096-byte head", len(full))
			}
			got, err := detector.Detect(full[:4096], "application/octet-stream")
			if err != nil {
				t.Fatalf("Detect with 4096-byte head: %v", err)
			}
			if got != tc.want {
				t.Fatalf("Detect with 4096-byte head = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestUnsupportedBytesReturn415(t *testing.T) {
	detector := attachmentmime.New()

	tests := []struct {
		name     string
		head     []byte
		declared string
	}{
		{"html declared as pdf", []byte("<html><body><script>alert(1)</script></body></html>"), "application/pdf"},
		{"svg rejected", []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`), "image/svg+xml"},
		{"plain zip rejected", plainZipBytes(t), "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		{"plain text declared as csv", []byte("just one line of text"), "text/csv"},
		{"binary", []byte{0x00, 0x01, 0x02, 0x03, 0xff, 0xfe}, "image/png"},
		{"empty", nil, "application/pdf"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := detector.Detect(tc.head, tc.declared)
			var apiErr *customerrors.CustomError
			if !errors.As(err, &apiErr) {
				t.Fatalf("error = %v, want *customerrors.CustomError", err)
			}
			if apiErr.Code != http.StatusUnsupportedMediaType {
				t.Fatalf("error code = %d, want %d", apiErr.Code, http.StatusUnsupportedMediaType)
			}
		})
	}
}

func encodedImage(t *testing.T, encode func(io.Writer, image.Image) error) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 200, G: 100, B: 50, A: 255})
	var buf bytes.Buffer
	if err := encode(&buf, img); err != nil {
		t.Fatalf("encode image: %v", err)
	}
	return buf.Bytes()
}

func ooxmlBytes(t *testing.T, mainPart, override string, withFiller bool) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	contentTypes := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">` +
		`<Default Extension="xml" ContentType="application/xml"/>` +
		`<Override PartName="/` + mainPart + `" ContentType="` + override + `"/>` +
		`</Types>`
	writeZipEntry(t, zw, "[Content_Types].xml", contentTypes)
	writeZipEntry(t, zw, mainPart, `<?xml version="1.0" encoding="UTF-8"?><root/>`)
	if withFiller {
		// Random bytes stay large after deflate, so the zip really is bigger
		// than 4096 bytes and the head really is truncated.
		filler := make([]byte, 8192)
		if _, err := rand.New(rand.NewSource(1)).Read(filler); err != nil {
			t.Fatalf("generate filler: %v", err)
		}
		writeZipEntry(t, zw, "docProps/filler.bin", string(filler))
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
	return buf.Bytes()
}

func plainZipBytes(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	writeZipEntry(t, zw, "readme.txt", "just a plain zip archive")
	if err := zw.Close(); err != nil {
		t.Fatalf("close plain zip writer: %v", err)
	}
	return buf.Bytes()
}

func writeZipEntry(t *testing.T, zw *zip.Writer, name, content string) {
	t.Helper()
	w, err := zw.Create(name)
	if err != nil {
		t.Fatalf("create zip entry %s: %v", name, err)
	}
	if _, err := io.WriteString(w, content); err != nil {
		t.Fatalf("write zip entry %s: %v", name, err)
	}
}
