package metadata

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
)

const corePath = "docProps/core.xml"

type coreDoc struct {
	Title          string `xml:"title"`
	Subject        string `xml:"subject"`
	Creator        string `xml:"creator"`
	Keywords       string `xml:"keywords"`
	Description    string `xml:"description"`
	LastModifiedBy string `xml:"lastModifiedBy"`
	Revision       string `xml:"revision"`
	Created        string `xml:"created"`
	Modified       string `xml:"modified"`
	Category       string `xml:"category"`
	ContentStatus  string `xml:"contentStatus"`
	Identifier     string `xml:"identifier"`
	Language       string `xml:"language"`
	Version        string `xml:"version"`
}

func applyOOXML(input, output string, f Fields) error {
	if err := ooxmlFields(f); err != nil {
		return err
	}
	for _, value := range []string{f.Date, f.Created, f.Modified} {
		if value == "" {
			continue
		}
		if _, err := parseTime(value); err != nil {
			return err
		}
	}
	zr, err := zip.OpenReader(input)
	if err != nil {
		return fmt.Errorf("open OOXML package: %w", err)
	}
	defer zr.Close()
	var core *zip.File
	for _, zf := range zr.File {
		if zf.Name == corePath {
			core = zf
			break
		}
	}
	if core == nil {
		return fmt.Errorf("OOXML package has no %s", corePath)
	}
	doc, err := readCore(core)
	if err != nil {
		return err
	}
	applyCore(&doc, f)
	data, err := marshalCore(doc)
	if err != nil {
		return err
	}

	out, err := os.Create(output)
	if err != nil {
		return err
	}
	zw := zip.NewWriter(out)
	for _, zf := range zr.File {
		if zf.Name == corePath {
			continue
		}
		hdr := zf.FileHeader
		w, err := zw.CreateHeader(&hdr)
		if err != nil {
			zw.Close()
			out.Close()
			return err
		}
		rc, err := zf.Open()
		if err != nil {
			zw.Close()
			out.Close()
			return err
		}
		_, copyErr := io.Copy(w, rc)
		rc.Close()
		if copyErr != nil {
			zw.Close()
			out.Close()
			return copyErr
		}
	}
	w, err := zw.Create(corePath)
	if err != nil {
		zw.Close()
		out.Close()
		return err
	}
	if _, err := w.Write(data); err != nil {
		zw.Close()
		out.Close()
		return err
	}
	closeErr := zw.Close()
	if closeErr == nil {
		closeErr = out.Close()
	} else {
		_ = out.Close()
	}
	if closeErr != nil {
		return closeErr
	}
	return validateOOXML(output)
}

func readOOXML(path string) (map[string]string, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("open OOXML package: %w", err)
	}
	defer zr.Close()
	for _, zf := range zr.File {
		if zf.Name != corePath {
			continue
		}
		doc, err := readCore(zf)
		if err != nil {
			return nil, err
		}
		out := make(map[string]string)
		put := func(k, v string) {
			if v != "" {
				out[k] = v
			}
		}
		put("title", doc.Title)
		put("subject", doc.Subject)
		put("author", doc.Creator)
		put("keywords", doc.Keywords)
		put("description", doc.Description)
		put("last_modified_by", doc.LastModifiedBy)
		put("category", doc.Category)
		put("content_status", doc.ContentStatus)
		put("created", doc.Created)
		put("modified", doc.Modified)
		return out, nil
	}
	return nil, fmt.Errorf("OOXML package has no %s", corePath)
}

func readCore(zf *zip.File) (coreDoc, error) {
	rc, err := zf.Open()
	if err != nil {
		return coreDoc{}, err
	}
	defer rc.Close()
	data, err := io.ReadAll(io.LimitReader(rc, 4<<20))
	if err != nil {
		return coreDoc{}, err
	}
	var doc coreDoc
	if err := xml.Unmarshal(data, &doc); err != nil {
		return coreDoc{}, fmt.Errorf("parse %s: %w", corePath, err)
	}
	return doc, nil
}

func applyCore(doc *coreDoc, f Fields) {
	if f.Title != "" {
		doc.Title = f.Title
	}
	if f.Subject != "" {
		doc.Subject = f.Subject
	}
	if f.Author != "" {
		doc.Creator = f.Author
	}
	if len(f.Keywords) > 0 {
		doc.Keywords = joinKeywords(f.Keywords)
	}
	if f.Description != "" {
		doc.Description = f.Description
	}
	if f.LastModifiedBy != "" {
		doc.LastModifiedBy = f.LastModifiedBy
	}
	if f.Category != "" {
		doc.Category = f.Category
	}
	if f.ContentStatus != "" {
		doc.ContentStatus = f.ContentStatus
	}
	if f.Date != "" {
		if t, err := parseTime(f.Date); err == nil {
			v := t.UTC().Format("2006-01-02T15:04:05Z")
			doc.Created = v
			doc.Modified = v
		}
	}
	if f.Created != "" {
		if t, err := parseTime(f.Created); err == nil {
			doc.Created = t.UTC().Format("2006-01-02T15:04:05Z")
		}
	}
	if f.Modified != "" {
		if t, err := parseTime(f.Modified); err == nil {
			doc.Modified = t.UTC().Format("2006-01-02T15:04:05Z")
		}
	}
}

func marshalCore(doc coreDoc) ([]byte, error) {
	var b bytes.Buffer
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:dcmitype="http://purl.org/dc/dcmitype/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">`)
	writeCoreElem(&b, "dc:title", doc.Title)
	writeCoreElem(&b, "dc:subject", doc.Subject)
	writeCoreElem(&b, "dc:creator", doc.Creator)
	writeCoreElem(&b, "cp:keywords", doc.Keywords)
	writeCoreElem(&b, "dc:description", doc.Description)
	writeCoreElem(&b, "cp:lastModifiedBy", doc.LastModifiedBy)
	writeCoreElem(&b, "cp:revision", doc.Revision)
	if doc.Created != "" {
		fmt.Fprintf(&b, `<dcterms:created xsi:type="dcterms:W3CDTF">%s</dcterms:created>`, escapeXML(doc.Created))
	}
	if doc.Modified != "" {
		fmt.Fprintf(&b, `<dcterms:modified xsi:type="dcterms:W3CDTF">%s</dcterms:modified>`, escapeXML(doc.Modified))
	}
	writeCoreElem(&b, "cp:category", doc.Category)
	writeCoreElem(&b, "cp:contentStatus", doc.ContentStatus)
	writeCoreElem(&b, "dc:identifier", doc.Identifier)
	writeCoreElem(&b, "dc:language", doc.Language)
	writeCoreElem(&b, "cp:version", doc.Version)
	b.WriteString(`</cp:coreProperties>`)
	return b.Bytes(), nil
}

func writeCoreElem(b *bytes.Buffer, name, value string) {
	if value == "" {
		return
	}
	fmt.Fprintf(b, "<%s>%s</%s>", name, escapeXML(value), name)
}

func escapeXML(s string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}

func validateOOXML(path string) error {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("validate OOXML: %w", err)
	}
	defer zr.Close()
	for _, zf := range zr.File {
		if zf.Name == corePath {
			if _, err := readCore(zf); err != nil {
				return fmt.Errorf("validate OOXML: %w", err)
			}
			return nil
		}
	}
	return fmt.Errorf("validate OOXML: %s missing", corePath)
}
