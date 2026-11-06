// Package profile loads reusable metadata profiles.
package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/subsll/metacairn/internal/metadata"
)

// Profile is a reusable set of metadata fields.
type Profile struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Fields      map[string]any `json:"fields"`
}

// Load merges profile files/directories into metadata.
func Load(paths []string) (metadata.Fields, []string, error) {
	var (
		fields metadata.Fields
		names  []string
	)
	for _, path := range paths {
		profiles, err := loadPath(path)
		if err != nil {
			return metadata.Fields{}, nil, err
		}
		for _, p := range profiles {
			if p.Name != "" {
				names = append(names, p.Name)
			}
			if len(p.Fields) == 0 {
				return metadata.Fields{}, nil, fmt.Errorf("%s: profile has no fields", path)
			}
			for key, value := range p.Fields {
				s, err := stringValue(value)
				if err != nil {
					return metadata.Fields{}, nil, fmt.Errorf("%s: field %q: %w", path, key, err)
				}
				if err := fields.Set(key, s); err != nil {
					return metadata.Fields{}, nil, fmt.Errorf("%s: %w", path, err)
				}
			}
		}
	}
	return fields, names, nil
}

func loadPath(path string) ([]Profile, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		p, err := loadFile(path)
		if err != nil {
			return nil, err
		}
		return []Profile{p}, nil
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() || !strings.EqualFold(filepath.Ext(e.Name()), ".json") {
			continue
		}
		files = append(files, filepath.Join(path, e.Name()))
	}
	sort.Strings(files)
	if len(files) == 0 {
		return nil, fmt.Errorf("%s: no .json profiles found", path)
	}
	var out []Profile
	for _, file := range files {
		p, err := loadFile(file)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func loadFile(path string) (Profile, error) {
	f, err := os.Open(path)
	if err != nil {
		return Profile{}, err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	dec.UseNumber()
	var p Profile
	if err := dec.Decode(&p); err != nil {
		return Profile{}, fmt.Errorf("parse %s: %w", path, err)
	}
	if len(p.Fields) == 0 {
		return Profile{}, fmt.Errorf("%s: profile has no fields", path)
	}
	return p, nil
}

func stringValue(v any) (string, error) {
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x), nil
	case json.Number:
		return x.String(), nil
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64), nil
	case bool:
		return strconv.FormatBool(x), nil
	case []any:
		parts := make([]string, 0, len(x))
		for _, item := range x {
			s, err := stringValue(item)
			if err != nil {
				return "", err
			}
			parts = append(parts, s)
		}
		if len(parts) == 2 {
			if _, err := strconv.ParseFloat(parts[0], 64); err == nil {
				if _, err := strconv.ParseFloat(parts[1], 64); err == nil {
					return parts[0] + "," + parts[1], nil
				}
			}
		}
		return strings.Join(parts, ", "), nil
	default:
		return "", fmt.Errorf("unsupported value type %T", v)
	}
}
