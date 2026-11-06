package metadata

import (
	"fmt"
	"io"
	"os"

	"github.com/FlavioCFOliveira/GoMetadata"
	"github.com/FlavioCFOliveira/GoMetadata/format"
)

func applyImage(input, output string, f Fields) error {
	if err := imageFields(f); err != nil {
		return err
	}
	in, err := os.Open(input)
	if err != nil {
		return err
	}
	defer in.Close()
	m, err := gometadata.Read(in)
	if err != nil {
		return fmt.Errorf("read image metadata: %w", err)
	}
	if !format.SupportsWrite(m.Format()) {
		return fmt.Errorf("image format %s does not support metadata writing", m.Format())
	}
	if caption := firstNonEmpty(f.Description, f.Subject, f.Title); caption != "" {
		m.SetCaption(caption)
	}
	if f.Author != "" {
		m.SetCreator(f.Author)
	}
	if f.Copyright != "" {
		m.SetCopyright(f.Copyright)
	}
	if len(f.Keywords) > 0 {
		m.SetKeywords(f.Keywords)
	}
	if f.Make != "" {
		m.SetMake(f.Make)
	}
	if f.Model != "" {
		m.SetCameraModel(f.Model)
	}
	if f.LensModel != "" {
		m.SetLensModel(f.LensModel)
	}
	if f.ISO != "" {
		v, err := parseISO(f.ISO)
		if err != nil {
			return err
		}
		m.SetISO(v)
	}
	if f.FNumber != "" {
		v, err := parseAperture(f.FNumber)
		if err != nil {
			return err
		}
		m.SetFNumber(v)
	}
	if f.ExposureTime != "" {
		num, den, err := parseExposure(f.ExposureTime)
		if err != nil {
			return err
		}
		m.SetExposureTime(num, den)
	}
	if f.FocalLength != "" {
		v, err := parseFocalLength(f.FocalLength)
		if err != nil {
			return err
		}
		m.SetFocalLength(v)
	}
	if f.Date != "" {
		t, err := parseTime(f.Date)
		if err != nil {
			return err
		}
		m.SetDateTimeOriginal(t)
	}
	if f.GPS != "" {
		lat, lon, err := parseGPS(f.GPS)
		if err != nil {
			return err
		}
		m.SetGPS(lat, lon)
	}
	if f.Orientation != "" {
		v, err := parseOrientation(f.Orientation)
		if err != nil {
			return err
		}
		m.SetOrientation(v)
	}
	if err := m.Validate(); err != nil {
		return fmt.Errorf("metadata validation: %w", err)
	}
	if _, err := in.Seek(0, io.SeekStart); err != nil {
		return err
	}
	out, err := os.Create(output)
	if err != nil {
		return err
	}
	writeErr := gometadata.Write(in, out, m, gometadata.PreserveUnknownSegments(true))
	closeErr := out.Close()
	if writeErr != nil {
		return fmt.Errorf("write image: %w", writeErr)
	}
	if closeErr != nil {
		return closeErr
	}
	if _, err := gometadata.ReadFile(output); err != nil {
		return fmt.Errorf("validate output image: %w", err)
	}
	return nil
}

func readImage(path string) (map[string]string, error) {
	m, err := gometadata.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read image metadata: %w", err)
	}
	out := make(map[string]string)
	if v := m.Caption(); v != "" {
		out["description"] = v
	}
	if v := m.Creator(); v != "" {
		out["author"] = v
	}
	if v := m.Copyright(); v != "" {
		out["copyright"] = v
	}
	if v := m.Keywords(); len(v) > 0 {
		out["keywords"] = joinKeywords(v)
	}
	if v := m.Make(); v != "" {
		out["make"] = v
	}
	if v := m.CameraModel(); v != "" {
		out["model"] = v
	}
	if v := m.LensModel(); v != "" {
		out["lens"] = v
	}
	if v, ok := m.ISO(); ok {
		out["iso"] = fmt.Sprintf("%d", v)
	}
	if v, ok := m.FNumber(); ok {
		out["aperture"] = fmt.Sprintf("%.1f", v)
	}
	if num, den, ok := m.ExposureTime(); ok {
		out["exposure"] = fmt.Sprintf("%d/%d", num, den)
	}
	if v, ok := m.FocalLength(); ok {
		out["focal_length"] = fmt.Sprintf("%.0f", v)
	}
	if t, ok := m.DateTimeOriginal(); ok {
		out["date"] = t.UTC().Format("2006-01-02T15:04:05Z")
	}
	if lat, lon, ok := m.GPS(); ok {
		out["gps"] = fmt.Sprintf("%.6f,%.6f", lat, lon)
	}
	if v, ok := m.Orientation(); ok {
		out["orientation"] = fmt.Sprintf("%d", v)
	}
	if v := m.Software(); v != "" {
		out["software"] = v
	}
	return out, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
