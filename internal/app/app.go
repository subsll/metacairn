// Package app implements the Metacairn CLI.
package app

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/subsll/metacairn/internal/metadata"
)

// Version is overridden at build time.
var Version = "0.1.0"

// Run executes the CLI and returns an exit code.
func Run(args []string, out, errOut io.Writer) int {
	if len(args) == 0 {
		usage(errOut)
		return 2
	}
	switch args[0] {
	case "set":
		return runSet(args[1:], out, errOut)
	case "get", "inspect":
		return runGet(args[1:], out, errOut)
	case "formats":
		for _, f := range metadata.SupportedFormats() {
			fmt.Fprintln(out, f)
		}
		return 0
	case "version", "--version", "-version":
		fmt.Fprintf(out, "metacairn %s (%s %s/%s)\n", Version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
		return 0
	case "help", "--help", "-h":
		usage(out)
		return 0
	default:
		fmt.Fprintf(errOut, "metacairn: unknown command %q\n", args[0])
		usage(errOut)
		return 2
	}
}

func runSet(args []string, out, errOut io.Writer) int {
	fs := flag.NewFlagSet("metacairn set", flag.ContinueOnError)
	fs.SetOutput(errOut)
	var (
		output, title, author, subject, description, keywords string
		copyright, creator, software                          string
		makeName, model, lens, date, gps, orientation         string
		category, lastModifiedBy, contentStatus               string
		created, modified                                     string
		inPlace, force                                        bool
		sets                                                  kvList
	)
	fs.StringVar(&output, "o", "", "write to this file")
	fs.StringVar(&output, "output", "", "write to this file")
	fs.BoolVar(&inPlace, "in-place", false, "overwrite input (creates a .bak backup)")
	fs.BoolVar(&force, "force", false, "overwrite output/backup if it exists")
	fs.StringVar(&title, "title", "", "title")
	fs.StringVar(&author, "author", "", "author/creator")
	fs.StringVar(&subject, "subject", "", "subject")
	fs.StringVar(&description, "description", "", "description/caption")
	fs.StringVar(&keywords, "keywords", "", "comma-separated keywords")
	fs.StringVar(&copyright, "copyright", "", "copyright")
	fs.StringVar(&creator, "creator", "", "creating application (PDF)")
	fs.StringVar(&software, "software", "", "software / PDF creator")
	fs.StringVar(&makeName, "make", "", "camera make (images)")
	fs.StringVar(&model, "model", "", "camera model (images)")
	fs.StringVar(&lens, "lens", "", "lens model (images)")
	fs.StringVar(&date, "date", "", "capture/creation date")
	fs.StringVar(&gps, "gps", "", "GPS as LAT,LON (images)")
	fs.StringVar(&orientation, "orientation", "", "EXIF orientation 1-8 (images)")
	fs.StringVar(&category, "category", "", "document category")
	fs.StringVar(&lastModifiedBy, "last-modified-by", "", "last modified by")
	fs.StringVar(&contentStatus, "content-status", "", "content status")
	fs.StringVar(&created, "created", "", "creation date")
	fs.StringVar(&modified, "modified", "", "modified date")
	fs.Var(&sets, "set", "key=value metadata field (repeatable)")
	fs.Var(&sets, "field", "alias for --set")
	fs.Usage = func() {
		fmt.Fprintln(errOut, "usage: metacairn set <file> (--output <file> | --in-place) [metadata flags]")
		fs.PrintDefaults()
	}

	pos, err := parseInterspersed(fs, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.SetOutput(out)
			fs.Usage()
			return 0
		}
		fmt.Fprintf(errOut, "metacairn: %v\n", err)
		return 2
	}
	if len(pos) != 1 {
		fmt.Fprintln(errOut, "metacairn: set requires exactly one input file")
		fs.Usage()
		return 2
	}
	input := pos[0]
	if output == "" && !inPlace {
		fmt.Fprintln(errOut, "metacairn: specify --output or --in-place")
		return 2
	}
	if output != "" && inPlace {
		fmt.Fprintln(errOut, "metacairn: use either --output or --in-place, not both")
		return 2
	}
	target := output
	if inPlace {
		target = input
	}
	if !inPlace && samePath(input, target) {
		fmt.Fprintln(errOut, "metacairn: use --in-place to overwrite the input")
		return 2
	}

	fields := metadata.Fields{
		Title: title, Author: author, Subject: subject, Description: description,
		Copyright: copyright, Creator: creator, Software: software,
		Make: makeName, Model: model, LensModel: lens, Date: date, GPS: gps,
		Orientation: orientation, Category: category, LastModifiedBy: lastModifiedBy,
		ContentStatus: contentStatus, Created: created, Modified: modified,
	}
	if keywords != "" {
		if err := fields.Set("keywords", keywords); err != nil {
			fmt.Fprintf(errOut, "metacairn: %v\n", err)
			return 2
		}
	}
	for _, pair := range sets {
		k, v, ok := strings.Cut(pair, "=")
		if !ok {
			fmt.Fprintf(errOut, "metacairn: --set %q must be key=value\n", pair)
			return 2
		}
		if err := fields.Set(k, v); err != nil {
			fmt.Fprintf(errOut, "metacairn: %v\n", err)
			return 2
		}
	}
	if fields.Empty() {
		fmt.Fprintln(errOut, "metacairn: no metadata fields supplied")
		return 2
	}

	info, err := os.Stat(input)
	if err != nil {
		fmt.Fprintf(errOut, "metacairn: %v\n", err)
		return 1
	}
	if !info.Mode().IsRegular() {
		fmt.Fprintf(errOut, "metacairn: %s is not a regular file\n", input)
		return 1
	}
	if !inPlace {
		if _, err := os.Stat(target); err == nil && !force {
			fmt.Fprintf(errOut, "metacairn: %s exists; use --force to overwrite\n", target)
			return 1
		}
	}

	backup := input + ".bak"
	if inPlace {
		if _, err := os.Stat(backup); err == nil && !force {
			fmt.Fprintf(errOut, "metacairn: backup %s exists; use --force to overwrite\n", backup)
			return 1
		}
		if err := copyFile(input, backup, info.Mode().Perm()); err != nil {
			fmt.Fprintf(errOut, "metacairn: create backup: %v\n", err)
			return 1
		}
	}

	tmp, err := os.CreateTemp(filepath.Dir(target), ".metacairn-*"+filepath.Ext(target))
	if err != nil {
		fmt.Fprintf(errOut, "metacairn: %v\n", err)
		return 1
	}
	tmpName := tmp.Name()
	_ = tmp.Close()
	_ = os.Remove(tmpName)
	defer os.Remove(tmpName)

	format, err := metadata.Apply(input, tmpName, fields)
	if err != nil {
		fmt.Fprintf(errOut, "metacairn: %v\n", err)
		return 1
	}
	if err := os.Chmod(tmpName, info.Mode().Perm()); err != nil {
		fmt.Fprintf(errOut, "metacairn: %v\n", err)
		return 1
	}
	if err := os.Rename(tmpName, target); err != nil {
		fmt.Fprintf(errOut, "metacairn: %v\n", err)
		return 1
	}
	fmt.Fprintf(out, "wrote %s (%s)\n", target, format)
	return 0
}

func runGet(args []string, out, errOut io.Writer) int {
	fs := flag.NewFlagSet("metacairn get", flag.ContinueOnError)
	fs.SetOutput(errOut)
	var asJSON bool
	fs.BoolVar(&asJSON, "json", false, "JSON output")
	fs.Usage = func() {
		fmt.Fprintln(errOut, "usage: metacairn get <file> [--json]")
	}
	pos, err := parseInterspersed(fs, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.SetOutput(out)
			fs.Usage()
			return 0
		}
		fmt.Fprintf(errOut, "metacairn: %v\n", err)
		return 2
	}
	if len(pos) != 1 {
		fmt.Fprintln(errOut, "metacairn: get requires exactly one input file")
		return 2
	}
	format, values, err := metadata.Inspect(pos[0])
	if err != nil {
		fmt.Fprintf(errOut, "metacairn: %v\n", err)
		return 1
	}
	normalized := make(map[string]string, len(values))
	for k, v := range values {
		normalized[normalizeOutputKey(k)] = v
	}
	if asJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		_ = enc.Encode(map[string]any{"file": pos[0], "format": format, "metadata": normalized})
		return 0
	}
	fmt.Fprintf(out, "file:   %s\nformat: %s\n", pos[0], format)
	keys := make([]string, 0, len(normalized))
	for k := range normalized {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(out, "%-15s %s\n", k+":", normalized[k])
	}
	return 0
}

func normalizeOutputKey(k string) string {
	var b strings.Builder
	for i, r := range k {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(r + ('a' - 'A'))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func samePath(a, b string) bool {
	aa, err1 := filepath.Abs(a)
	bb, err2 := filepath.Abs(b)
	return err1 == nil && err2 == nil && strings.EqualFold(aa, bb)
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "metacairn — write EXIF and document metadata from the command line")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "usage:")
	fmt.Fprintln(w, "  metacairn set <file> (--output <file> | --in-place) [flags]")
	fmt.Fprintln(w, "  metacairn get <file> [--json]")
	fmt.Fprintln(w, "  metacairn formats")
	fmt.Fprintln(w, "  metacairn version")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Formats: JPEG, PNG, TIFF, WebP, PDF, DOCX, XLSX, PPTX")
	fmt.Fprintln(w, "Run 'metacairn set -h' for metadata flags.")
}

type kvList []string

func (k *kvList) String() string { return strings.Join(*k, ",") }
func (k *kvList) Set(v string) error {
	*k = append(*k, v)
	return nil
}

// parseInterspersed parses flags anywhere among positionals.
func parseInterspersed(fs *flag.FlagSet, args []string) ([]string, error) {
	var pos []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			pos = append(pos, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			pos = append(pos, arg)
			continue
		}
		name := strings.TrimLeft(arg, "-")
		value := ""
		hasValue := false
		if idx := strings.IndexByte(name, '='); idx >= 0 {
			value = name[idx+1:]
			name = name[:idx]
			hasValue = true
		}
		if name == "help" || name == "h" {
			return nil, flag.ErrHelp
		}
		f := fs.Lookup(name)
		if f == nil {
			return nil, fmt.Errorf("unknown flag -%s", name)
		}
		if hasValue {
			if err := fs.Set(name, value); err != nil {
				return nil, err
			}
			continue
		}
		if bf, ok := f.Value.(interface{ IsBoolFlag() bool }); ok && bf.IsBoolFlag() {
			if err := fs.Set(name, "true"); err != nil {
				return nil, err
			}
			continue
		}
		if i+1 >= len(args) {
			return nil, fmt.Errorf("flag -%s needs a value", name)
		}
		i++
		if err := fs.Set(name, args[i]); err != nil {
			return nil, err
		}
	}
	return pos, nil
}
