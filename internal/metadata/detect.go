package metadata

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
)

// Detect identifies a supported file from its contents.
func Detect(path string) (Format, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	head := make([]byte, 512)
	n, err := io.ReadFull(f, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return "", err
	}
	head = head[:n]

	switch {
	case len(head) >= 2 && head[0] == 0xff && head[1] == 0xd8:
		return FormatJPEG, nil
	case bytes.HasPrefix(head, []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}):
		return FormatPNG, nil
	case bytes.HasPrefix(head, []byte("II*\x00")) || bytes.HasPrefix(head, []byte("MM\x00*")):
		return FormatTIFF, nil
	case len(head) >= 12 && string(head[:4]) == "RIFF" && string(head[8:12]) == "WEBP":
		return FormatWebP, nil
	case bytes.HasPrefix(head, []byte("%PDF-")):
		return FormatPDF, nil
	case bytes.HasPrefix(head, []byte("PK\x03\x04")):
		return detectOOXML(path)
	default:
		return "", fmt.Errorf("unsupported file format (expected JPEG, PNG, TIFF, WebP, PDF, DOCX, XLSX or PPTX)")
	}
}

func detectOOXML(path string) (Format, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return "", fmt.Errorf("open ZIP: %w", err)
	}
	defer zr.Close()
	has := make(map[string]bool)
	for _, f := range zr.File {
		has[f.Name] = true
	}
	switch {
	case has["word/document.xml"]:
		return FormatDOCX, nil
	case has["xl/workbook.xml"]:
		return FormatXLSX, nil
	case has["ppt/presentation.xml"]:
		return FormatPPTX, nil
	default:
		return "", errors.New("ZIP is not a DOCX, XLSX or PPTX package")
	}
}
