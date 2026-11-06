package metadata

import "fmt"

// Apply writes f to input and produces output. The input format is detected
// from its contents.
func Apply(input, output string, f Fields) (Format, error) {
	if f.Empty() {
		return "", fmt.Errorf("no metadata fields were supplied")
	}
	format, err := Detect(input)
	if err != nil {
		return "", err
	}
	switch format {
	case FormatJPEG, FormatPNG, FormatTIFF, FormatWebP:
		err = applyImage(input, output, f)
	case FormatPDF:
		err = applyPDF(input, output, f)
	case FormatDOCX, FormatXLSX, FormatPPTX:
		err = applyOOXML(input, output, f)
	default:
		err = fmt.Errorf("unsupported format %s", format)
	}
	return format, err
}

// Inspect reads the supported metadata fields from path.
func Inspect(path string) (Format, map[string]string, error) {
	format, err := Detect(path)
	if err != nil {
		return "", nil, err
	}
	switch format {
	case FormatJPEG, FormatPNG, FormatTIFF, FormatWebP:
		values, err := readImage(path)
		return format, values, err
	case FormatPDF:
		values, err := readPDF(path)
		return format, values, err
	case FormatDOCX, FormatXLSX, FormatPPTX:
		values, err := readOOXML(path)
		return format, values, err
	default:
		return "", nil, fmt.Errorf("unsupported format %s", format)
	}
}

// SupportedFormats returns the formats this build can write.
func SupportedFormats() []Format {
	return []Format{FormatJPEG, FormatPNG, FormatTIFF, FormatWebP, FormatPDF, FormatDOCX, FormatXLSX, FormatPPTX}
}
