package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

var (
	errUnsupportedDocument = errors.New("unsupported document type for text extraction")
	errNoExtractableText   = errors.New("no extractable text found")
)

var pdfLiteralRegex = regexp.MustCompile(`\(([^()]*)\)`)

// OCR runner is a variable to allow deterministic unit tests.
var imageOCRExtractor = runTesseractOCR

func extractDocumentText(contentType, filename string, body []byte) (string, error) {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	switch {
	case ct == "application/pdf" || strings.HasSuffix(strings.ToLower(filename), ".pdf"):
		text, err := extractTextFromPDF(body)
		if err != nil {
			return "", err
		}
		if text == "" {
			return "", errNoExtractableText
		}
		return text, nil
	case strings.HasPrefix(ct, "image/"):
		text, err := imageOCRExtractor(body)
		if err != nil {
			return "", fmt.Errorf("ocr extraction failed: %w", err)
		}
		text = normalizeExtractedText(text)
		if text == "" {
			return "", errNoExtractableText
		}
		return text, nil
	default:
		return "", errUnsupportedDocument
	}
}

// Exported wrappers for worker package use without duplicating logic.
func ExtractDocumentTextForWorker(contentType, filename string, body []byte) (string, error) {
	return extractDocumentText(contentType, filename, body)
}

func BuildSummaryFromExtractedTextForWorker(filename string, sizeBytes int64, extracted string) string {
	return buildSummaryFromExtractedText(filename, sizeBytes, extracted)
}

func NormalizeSummaryExtractionErrorForWorker(err error) string {
	switch {
	case errors.Is(err, errUnsupportedDocument):
		return "summary extraction is only supported for PDF and image documents"
	case errors.Is(err, errNoExtractableText):
		return "could not extract readable text from document"
	default:
		return err.Error()
	}
}

func ErrUnsupportedDocumentForWorker() error { return errUnsupportedDocument }
func ErrNoExtractableTextForWorker() error   { return errNoExtractableText }

func extractTextFromPDF(body []byte) (string, error) {
	if len(body) == 0 {
		return "", errors.New("empty pdf payload")
	}
	if !bytes.HasPrefix(body, []byte("%PDF")) {
		return "", errors.New("invalid pdf header")
	}

	matches := pdfLiteralRegex.FindAllSubmatch(body, -1)
	parts := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		text := decodePDFLiteral(string(m[1]))
		text = normalizeExtractedText(text)
		if text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, " "), nil
}

func decodePDFLiteral(s string) string {
	s = strings.ReplaceAll(s, `\(`, "(")
	s = strings.ReplaceAll(s, `\)`, ")")
	s = strings.ReplaceAll(s, `\\`, `\`)
	s = strings.ReplaceAll(s, `\n`, " ")
	s = strings.ReplaceAll(s, `\r`, " ")
	s = strings.ReplaceAll(s, `\t`, " ")
	return s
}

func normalizeExtractedText(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	lastSpace := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if !lastSpace {
				b.WriteRune(' ')
				lastSpace = true
			}
			continue
		}
		b.WriteRune(r)
		lastSpace = false
	}
	return strings.TrimSpace(b.String())
}

func runTesseractOCR(body []byte) (string, error) {
	if len(body) == 0 {
		return "", errors.New("empty image payload")
	}

	tmpDir, err := os.MkdirTemp("", "hx-ocr-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	inFile := filepath.Join(tmpDir, "input-image")
	if err := os.WriteFile(inFile, body, 0o600); err != nil {
		return "", err
	}

	cmd := exec.Command("tesseract", inFile, "stdout", "-l", "eng")
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("tesseract command failed: %s", msg)
	}
	return string(out), nil
}

func buildSummaryFromExtractedText(filename string, sizeBytes int64, extracted string) string {
	kb := sizeBytes / 1024
	if kb < 1 {
		kb = 1
	}
	preview := extracted
	if len(preview) > 380 {
		preview = preview[:380] + "..."
	}
	return fmt.Sprintf(
		"AI summary (local extraction, non-diagnostic): Document %q (~%d KB). Extracted text highlights: %s",
		filename,
		kb,
		preview,
	)
}
