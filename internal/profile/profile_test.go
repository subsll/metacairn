package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDemoProfile(t *testing.T) {
	fields, names, err := Load([]string{"../../examples/profiles/canon-eos-5d-mark-iv.demo.json"})
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "canon-eos-5d-mark-iv-demo" {
		t.Fatalf("names = %v", names)
	}
	if fields.Make != "Canon" || fields.Model != "Canon EOS 5D Mark IV" || fields.ISO != "100" {
		t.Fatalf("fields = %+v", fields)
	}
	if fields.FNumber != "2.8" || fields.ExposureTime != "1/250" || fields.FocalLength != "50" {
		t.Fatalf("exposure fields = %+v", fields)
	}
}

func TestLoadDirectoryAndMerge(t *testing.T) {
	dir := t.TempDir()
	a := `{"name":"a","fields":{"make":"Canon","model":"A"}}`
	b := `{"name":"b","fields":{"model":"B","iso":200}}`
	if err := os.WriteFile(filepath.Join(dir, "a.json"), []byte(a), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.json"), []byte(b), 0o644); err != nil {
		t.Fatal(err)
	}
	fields, names, err := Load([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 {
		t.Fatalf("names = %v", names)
	}
	if fields.Make != "Canon" || fields.Model != "B" || fields.ISO != "200" {
		t.Fatalf("fields = %+v", fields)
	}
}

func TestUnknownField(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte(`{"fields":{"not_a_field":"x"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Load([]string{path}); err == nil {
		t.Fatal("expected unknown field error")
	}
}
