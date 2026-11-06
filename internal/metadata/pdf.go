package metadata

import (
	"context"
	"fmt"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

func applyPDF(input, output string, f Fields) error {
	if err := pdfFields(f); err != nil {
		return err
	}
	props := make(map[string]string)
	set := func(key, value string) {
		if value != "" {
			props[key] = value
		}
	}
	set("Title", f.Title)
	set("Author", f.Author)
	subject := firstNonEmpty(f.Subject, f.Description)
	set("Subject", subject)
	set("Creator", firstNonEmpty(f.Creator, f.Software))
	if len(props) == 0 {
		return fmt.Errorf("no PDF metadata fields were supplied")
	}
	conf := model.NewStatelessConfiguration()
	if err := api.AddPropertiesFile(context.Background(), input, output, props, conf); err != nil {
		return fmt.Errorf("write PDF properties: %w", err)
	}
	if err := api.ValidateFile(context.Background(), output, conf, nil); err != nil {
		return fmt.Errorf("validate output PDF: %w", err)
	}
	return nil
}

func readPDF(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	conf := model.NewStatelessConfiguration()
	info, err := api.PDFInfo(context.Background(), f, path, nil, false, conf)
	if err != nil {
		return nil, fmt.Errorf("read PDF info: %w", err)
	}
	out := make(map[string]string)
	put := func(k, v string) {
		if v != "" {
			out[k] = v
		}
	}
	put("title", info.Title)
	put("author", info.Author)
	put("subject", info.Subject)
	put("producer", info.Producer)
	put("creator", info.Creator)
	put("creation_date", info.CreationDate)
	put("modification_date", info.ModificationDate)
	if len(info.Keywords) > 0 {
		out["keywords"] = joinKeywords(info.Keywords)
	}
	for k, v := range info.Properties {
		if v != "" {
			out[k] = v
		}
	}
	return out, nil
}
