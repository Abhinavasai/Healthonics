package handlers

import (
	"errors"
	"strings"
	"testing"
)

func TestExtractTextFromPDF_Success(t *testing.T) {
	pdf := []byte("%PDF-1.4\n1 0 obj\n<<>>\nstream\nBT (Patient has mild cough) Tj ET\nendstream\nendobj\n%%EOF")
	got, err := extractTextFromPDF(pdf)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(got, "Patient has mild cough") {
		t.Fatalf("expected extracted text, got %q", got)
	}
}

func TestExtractTextFromPDF_InvalidHeader(t *testing.T) {
	_, err := extractTextFromPDF([]byte("not-a-pdf"))
	if err == nil {
		t.Fatal("expected error for invalid pdf header")
	}
}

func TestExtractDocumentText_UnsupportedType(t *testing.T) {
	_, err := extractDocumentText("text/csv", "a.csv", []byte("x,y"))
	if !errors.Is(err, errUnsupportedDocument) {
		t.Fatalf("expected unsupported type error, got %v", err)
	}
}

func TestExtractDocumentText_ImageUsesOCRHook(t *testing.T) {
	orig := imageOCRExtractor
	t.Cleanup(func() { imageOCRExtractor = orig })
	imageOCRExtractor = func(_ []byte) (string, error) {
		return " BP 110 over 70 ", nil
	}

	got, err := extractDocumentText("image/png", "scan.png", []byte{1, 2, 3})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != "BP 110 over 70" {
		t.Fatalf("unexpected normalized OCR text: %q", got)
	}
}

func TestExtractDocumentText_ImageOCRFailure(t *testing.T) {
	orig := imageOCRExtractor
	t.Cleanup(func() { imageOCRExtractor = orig })
	imageOCRExtractor = func(_ []byte) (string, error) {
		return "", errors.New("ocr engine unavailable")
	}

	_, err := extractDocumentText("image/jpeg", "scan.jpg", []byte{1, 2, 3})
	if err == nil || !strings.Contains(err.Error(), "ocr extraction failed") {
		t.Fatalf("expected wrapped ocr failure, got %v", err)
	}
}

func TestBuildSummaryFromExtractedText(t *testing.T) {
	out := buildSummaryFromExtractedText("lab.pdf", 2048, "Hemoglobin normal")
	if !strings.Contains(out, "lab.pdf") || !strings.Contains(out, "Hemoglobin normal") {
		t.Fatalf("unexpected summary output: %s", out)
	}
}

func TestNormalizeExtractedText(t *testing.T) {
	if got := normalizeExtractedText("a\t b\n c\r\n"); got != "a b c" {
		t.Fatalf("unexpected normalized text: %q", got)
	}
}

func TestBuildSummaryPreviewTruncation(t *testing.T) {
	long := strings.Repeat("x", 500)
	out := buildSummaryFromExtractedText("doc.pdf", 15*1024, long)
	if !strings.Contains(out, "...") {
		t.Fatalf("expected truncated preview, got %q", out)
	}
}
