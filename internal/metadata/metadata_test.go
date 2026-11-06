package metadata

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

func testImage(t *testing.T, path, ext string) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 32, 24))
	for y := 0; y < 24; y++ {
		for x := 0; x < 32; x++ {
			img.Set(x, y, color.RGBA{uint8(x * 7), uint8(y * 9), 120, 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	switch ext {
	case ".jpg":
		if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 90}); err != nil {
			t.Fatal(err)
		}
	case ".png":
		if err := png.Encode(f, img); err != nil {
			t.Fatal(err)
		}
	default:
		t.Fatalf("unknown ext %s", ext)
	}
}

func TestImageRoundTrip(t *testing.T) {
	for _, ext := range []string{".jpg", ".png"} {
		t.Run(ext, func(t *testing.T) {
			dir := t.TempDir()
			in := filepath.Join(dir, "in"+ext)
			out := filepath.Join(dir, "out"+ext)
			testImage(t, in, ext)
			fields := Fields{
				Title: "A test image", Author: "Metacairn Test", Copyright: "CC0",
				Keywords: []string{"one", "two"}, Date: "2026-11-05T12:00:00Z",
				GPS: "52.3700,4.8900", Make: "Example", Model: "Camera", LensModel: "Lens",
			}
			if _, err := Apply(in, out, fields); err != nil {
				t.Fatal(err)
			}
			format, values, err := Inspect(out)
			if err != nil {
				t.Fatal(err)
			}
			if format != FormatJPEG && format != FormatPNG {
				t.Fatalf("format = %s", format)
			}
			if values["author"] != "Metacairn Test" || values["copyright"] != "CC0" {
				t.Fatalf("values = %+v", values)
			}
			if values["description"] != "A test image" || values["make"] != "Example" {
				t.Fatalf("values = %+v", values)
			}
			if !strings.Contains(values["gps"], "52.370000") {
				t.Fatalf("gps = %q", values["gps"])
			}
			f, err := os.Open(out)
			if err != nil {
				t.Fatal(err)
			}
			_, _, err = image.Decode(f)
			f.Close()
			if err != nil {
				t.Fatalf("output image does not decode: %v", err)
			}
		})
	}
}

func TestOOXMLRoundTrip(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.docx")
	out := filepath.Join(dir, "out.docx")
	writeDOCX(t, in)
	fields := Fields{
		Title: "Title", Author: "Author", Subject: "Subject",
		Description: "Description", Keywords: []string{"a", "b"},
		Category: "Research", LastModifiedBy: "Editor",
	}
	if _, err := Apply(in, out, fields); err != nil {
		t.Fatal(err)
	}
	format, values, err := Inspect(out)
	if err != nil {
		t.Fatal(err)
	}
	if format != FormatDOCX {
		t.Fatalf("format = %s", format)
	}
	if values["title"] != "Title" || values["author"] != "Author" || values["keywords"] != "a, b" {
		t.Fatalf("values = %+v", values)
	}
	if values["category"] != "Research" || values["last_modified_by"] != "Editor" {
		t.Fatalf("values = %+v", values)
	}
}

func TestPDFRoundTrip(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.pdf")
	out := filepath.Join(dir, "out.pdf")
	writePDF(t, in)
	fields := Fields{Title: "PDF title", Author: "PDF author", Subject: "PDF subject", Creator: "Metacairn"}
	if _, err := Apply(in, out, fields); err != nil {
		t.Fatal(err)
	}
	format, values, err := Inspect(out)
	if err != nil {
		t.Fatal(err)
	}
	if format != FormatPDF {
		t.Fatalf("format = %s", format)
	}
	if values["title"] != "PDF title" || values["author"] != "PDF author" {
		t.Fatalf("values = %+v", values)
	}
}

func TestPDFRejectsKeywords(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.pdf")
	writePDF(t, in)
	_, err := Apply(in, filepath.Join(dir, "out.pdf"), Fields{Keywords: []string{"x"}})
	if err == nil {
		t.Fatal("expected unsupported PDF field error")
	}
}

func writeDOCX(t *testing.T, path string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	files := map[string]string{
		"[Content_Types].xml": `<?xml version="1.0" encoding="UTF-8"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/><Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/></Types>`,
		"_rels/.rels":         `<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/><Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/></Relationships>`,
		"word/document.xml":   `<?xml version="1.0" encoding="UTF-8"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body/></w:document>`,
		"docProps/core.xml":   `<?xml version="1.0" encoding="UTF-8"?><cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/"><dc:title>Old</dc:title></cp:coreProperties>`,
	}
	for name, body := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func writePDF(t *testing.T, path string) {
	t.Helper()
	var b bytes.Buffer
	b.WriteString("%PDF-1.4\n")
	offsets := make([]int, 4)
	offsets[1] = b.Len()
	b.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")
	offsets[2] = b.Len()
	b.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")
	offsets[3] = b.Len()
	b.WriteString("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] >>\nendobj\n")
	xref := b.Len()
	fmt.Fprintf(&b, "xref\n0 4\n0000000000 65535 f \n")
	for i := 1; i <= 3; i++ {
		fmt.Fprintf(&b, "%010d 00000 n \n", offsets[i])
	}
	fmt.Fprintf(&b, "trailer\n<< /Size 4 /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", xref)
	if err := os.WriteFile(path, b.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := api.ValidateFile(context.Background(), path, model.NewStatelessConfiguration(), nil); err != nil {
		t.Fatalf("test PDF is invalid: %v", err)
	}
}
