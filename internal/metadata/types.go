// Package metadata writes metadata to images, PDFs and OOXML documents.
package metadata

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// Format is a supported container format.
type Format string

const (
	FormatJPEG Format = "jpeg"
	FormatPNG  Format = "png"
	FormatTIFF Format = "tiff"
	FormatWebP Format = "webp"
	FormatPDF  Format = "pdf"
	FormatDOCX Format = "docx"
	FormatXLSX Format = "xlsx"
	FormatPPTX Format = "pptx"
)

// Fields are the metadata values a caller wants to write. Empty values are
// left unchanged.
type Fields struct {
	Title          string
	Author         string
	Subject        string
	Description    string
	Keywords       []string
	Copyright      string
	Creator        string
	Producer       string
	Software       string
	Make           string
	Model          string
	LensModel      string
	ISO            string
	FNumber        string
	ExposureTime   string
	FocalLength    string
	Date           string
	GPS            string
	Orientation    string
	Category       string
	LastModifiedBy string
	ContentStatus  string
	Created        string
	Modified       string
}

// Empty reports whether no metadata field is set.
func (f Fields) Empty() bool {
	return f.Title == "" && f.Author == "" && f.Subject == "" &&
		f.Description == "" && len(f.Keywords) == 0 && f.Copyright == "" &&
		f.Creator == "" && f.Producer == "" && f.Software == "" &&
		f.Make == "" && f.Model == "" && f.LensModel == "" && f.ISO == "" &&
		f.FNumber == "" && f.ExposureTime == "" && f.FocalLength == "" && f.Date == "" &&
		f.GPS == "" && f.Orientation == "" && f.Category == "" &&
		f.LastModifiedBy == "" && f.ContentStatus == "" && f.Created == "" &&
		f.Modified == ""
}

// Set assigns one user-facing field. Names are case-insensitive and ignore
// spaces, hyphens and underscores.
func (f *Fields) Set(name, value string) error {
	key := normalizeKey(name)
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("field %q has an empty value", name)
	}
	switch key {
	case "title":
		f.Title = value
	case "author", "artist", "creator":
		f.Author = value
	case "subject":
		f.Subject = value
	case "description", "caption", "imagedescription":
		f.Description = value
	case "keywords", "keyword":
		f.Keywords = splitKeywords(value)
	case "copyright", "rights":
		f.Copyright = value
	case "producer":
		f.Producer = value
	case "software":
		f.Software = value
	case "creatorapp", "application":
		f.Creator = value
	case "make", "camera":
		f.Make = value
	case "model", "cameramodel":
		f.Model = value
	case "lens", "lensmodel":
		f.LensModel = value
	case "iso", "isospeed":
		f.ISO = value
	case "fnumber", "aperture", "fstop":
		f.FNumber = value
	case "exposure", "exposuretime", "shutter", "shutterspeed":
		f.ExposureTime = value
	case "focal", "focallength":
		f.FocalLength = value
	case "date", "datetime", "datetimeoriginal", "capturetime":
		f.Date = value
	case "gps", "location", "coordinates":
		f.GPS = value
	case "orientation":
		f.Orientation = value
	case "category":
		f.Category = value
	case "lastmodifiedby":
		f.LastModifiedBy = value
	case "contentstatus":
		f.ContentStatus = value
	case "created", "creationdate":
		f.Created = value
	case "modified", "moddate":
		f.Modified = value
	default:
		return fmt.Errorf("unknown metadata field %q", name)
	}
	return nil
}

func normalizeKey(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	return strings.NewReplacer(" ", "", "-", "", "_", "").Replace(s)
}

func splitKeywords(s string) []string {
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ';' })
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseTime(s string) (time.Time, error) {
	for _, layout := range []string{
		time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05", "2006:01:02 15:04:05", "2006-01-02",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date %q (use RFC3339, YYYY-MM-DD or YYYY:MM:DD HH:MM:SS)", s)
}

func parseGPS(s string) (float64, float64, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid GPS %q (use LAT,LON)", s)
	}
	lat, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil || lat < -90 || lat > 90 {
		return 0, 0, fmt.Errorf("invalid latitude in %q", s)
	}
	lon, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil || lon < -180 || lon > 180 {
		return 0, 0, fmt.Errorf("invalid longitude in %q", s)
	}
	return lat, lon, nil
}

func parseISO(s string) (uint, error) {
	n, err := strconv.ParseUint(strings.TrimSpace(s), 10, 32)
	if err != nil || n == 0 {
		return 0, fmt.Errorf("invalid ISO %q", s)
	}
	return uint(n), nil
}

func parseAperture(s string) (float64, error) {
	v, err := strconv.ParseFloat(strings.TrimPrefix(strings.TrimSpace(s), "f/"), 64)
	if err != nil || v <= 0 {
		return 0, fmt.Errorf("invalid aperture %q (use e.g. 2.8)", s)
	}
	return v, nil
}

func parseFocalLength(s string) (float64, error) {
	v, err := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(s), "mm"), 64)
	if err != nil || v <= 0 {
		return 0, fmt.Errorf("invalid focal length %q (use e.g. 50)", s)
	}
	return v, nil
}

func parseExposure(s string) (uint32, uint32, error) {
	s = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(s), "s"))
	if strings.Contains(s, "/") {
		parts := strings.SplitN(s, "/", 2)
		num, err1 := strconv.ParseUint(strings.TrimSpace(parts[0]), 10, 32)
		den, err2 := strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 32)
		if err1 != nil || err2 != nil || num == 0 || den == 0 {
			return 0, 0, fmt.Errorf("invalid exposure %q (use e.g. 1/250)", s)
		}
		return uint32(num), uint32(den), nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil || v <= 0 {
		return 0, 0, fmt.Errorf("invalid exposure %q (use e.g. 1/250)", s)
	}
	const scale = 1000000
	num := uint32(math.Round(v * scale))
	if num == 0 {
		num = 1
	}
	g := gcd(num, scale)
	return num / g, scale / g, nil
}

func gcd(a, b uint32) uint32 {
	for b != 0 {
		a, b = b, a%b
	}
	if a == 0 {
		return 1
	}
	return a
}

func parseOrientation(s string) (uint16, error) {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n < 1 || n > 8 {
		return 0, fmt.Errorf("invalid orientation %q (use 1-8)", s)
	}
	return uint16(n), nil
}

func joinKeywords(k []string) string { return strings.Join(k, ", ") }

func imageFields(f Fields) error {
	if f.Creator != "" || f.Producer != "" || f.Software != "" || f.Category != "" ||
		f.LastModifiedBy != "" || f.ContentStatus != "" || f.Created != "" || f.Modified != "" {
		return fmt.Errorf("field not supported for image formats (images support title, author, description, keywords, copyright, make, model, lens, ISO, aperture, exposure, focal length, date, gps and orientation)")
	}
	return nil
}

func pdfFields(f Fields) error {
	if f.Keywords != nil || f.Producer != "" || f.Date != "" || f.Created != "" ||
		f.Modified != "" || f.Make != "" || f.Model != "" || f.LensModel != "" ||
		f.ISO != "" || f.FNumber != "" || f.ExposureTime != "" || f.FocalLength != "" ||
		f.GPS != "" || f.Orientation != "" || f.Category != "" ||
		f.LastModifiedBy != "" || f.ContentStatus != "" {
		return fmt.Errorf("field not supported for PDF (PDF supports title, author, subject, description, creator and software)")
	}
	return nil
}

func ooxmlFields(f Fields) error {
	if f.Make != "" || f.Model != "" || f.LensModel != "" || f.ISO != "" ||
		f.FNumber != "" || f.ExposureTime != "" || f.FocalLength != "" || f.GPS != "" ||
		f.Orientation != "" || f.Creator != "" || f.Producer != "" || f.Software != "" {
		return fmt.Errorf("field not supported for OOXML documents")
	}
	return nil
}
