package app

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func makePNG(t *testing.T, path string) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			img.Set(x, y, color.RGBA{uint8(x * 5), uint8(y * 5), 90, 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSetAndGet(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.png")
	out := filepath.Join(dir, "out.png")
	makePNG(t, in)
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"set", in, "--output", out, "--author", "Tester", "--description", "Test image", "--keywords", "one,two"}, &stdout, &stderr); code != 0 {
		t.Fatalf("set code=%d stderr=%s", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"get", out, "--json"}, &stdout, &stderr); code != 0 {
		t.Fatalf("get code=%d stderr=%s", code, stderr.String())
	}
	var report struct {
		Format   string            `json:"format"`
		Metadata map[string]string `json:"metadata"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Format != "png" || report.Metadata["author"] != "Tester" || report.Metadata["description"] != "Test image" {
		t.Fatalf("report = %+v", report)
	}
}

func TestInPlaceBackup(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.png")
	makePNG(t, in)
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"set", in, "--in-place", "--author", "Tester"}, &stdout, &stderr); code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(in + ".bak"); err != nil {
		t.Fatalf("backup missing: %v", err)
	}
}

func TestNoFields(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.png")
	makePNG(t, in)
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"set", in, "--in-place"}, &stdout, &stderr); code != 2 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
}
