package text_extract

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"baliance.com/gooxml/document"
	pdf "github.com/ledongthuc/pdf"
)

var (
	ErrTooLarge          = errors.New("file is too large")
	ErrUnsupportedFormat = errors.New("unsupported file type for text extraction")
	ErrEmptyText         = errors.New("extracted text is empty")
)

type ExtractOptions struct {
	MaxBytes int64
}

func ExtractTextFromMultipart(ctx context.Context, fh *multipart.FileHeader, opt ExtractOptions) (string, error) {
	if opt.MaxBytes <= 0 {
		opt.MaxBytes = 20 << 20
	}

	f, err := fh.Open()
	if err != nil {
		return "", fmt.Errorf("open upload: %w", err)
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	tmpPath, err := saveToTempFile(ctx, f, ext, opt.MaxBytes)
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpPath)

	var txt string
	switch ext {
	case ".txt":
		txt, err = extractTXT(tmpPath, opt.MaxBytes)
	case ".pdf":
		txt, err = extractPDF(tmpPath)
	case ".docx":
		txt, err = extractDOCX(tmpPath)
	default:
		return "", ErrUnsupportedFormat
	}
	if err != nil {
		return "", err
	}

	txt = normalizeText(txt)
	if len(strings.TrimSpace(txt)) == 0 {
		return "", ErrEmptyText
	}
	return txt, nil
}

func saveToTempFile(ctx context.Context, r io.Reader, ext string, maxBytes int64) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	tf, err := os.CreateTemp("", "upload-*"+ext)
	if err != nil {
		return "", fmt.Errorf("create temp: %w", err)
	}
	defer func() {
		_ = tf.Close()
		if err != nil {
			_ = os.Remove(tf.Name())
		}
	}()

	lr := io.LimitReader(r, maxBytes+1)
	n, copyErr := io.Copy(tf, lr)
	if copyErr != nil {
		err = fmt.Errorf("save temp: %w", copyErr)
		return "", err
	}
	if n > maxBytes {
		err = ErrTooLarge
		return "", err
	}

	if syncErr := tf.Sync(); syncErr != nil {
		err = fmt.Errorf("sync temp: %w", syncErr)
		return "", err
	}
	return tf.Name(), nil
}

func extractTXT(path string, maxBytes int64) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read txt: %w", err)
	}
	if int64(len(b)) > maxBytes {
		return "", ErrTooLarge
	}
	return string(b), nil
}

func extractDOCX(path string) (string, error) {
	doc, err := document.Open(path)
	if err != nil {
		return "", fmt.Errorf("open docx: %w", err)
	}

	var sb strings.Builder

	paraText := func(p document.Paragraph) string {
		var b strings.Builder
		for _, r := range p.Runs() {
			b.WriteString(r.Text())
		}
		return b.String()
	}

	for _, p := range doc.Paragraphs() {
		t := strings.TrimSpace(paraText(p))
		if t == "" {
			continue
		}
		sb.WriteString(t)
		sb.WriteString("\n")
	}

	for _, tbl := range doc.Tables() {
		for _, row := range tbl.Rows() {
			for _, cell := range row.Cells() {
				var cellSb strings.Builder
				for _, p := range cell.Paragraphs() {
					t := strings.TrimSpace(paraText(p))
					if t == "" {
						continue
					}
					if cellSb.Len() > 0 {
						cellSb.WriteString(" ")
					}
					cellSb.WriteString(t)
				}
				ct := strings.TrimSpace(cellSb.String())
				if ct != "" {
					sb.WriteString(ct)
					sb.WriteString(" | ")
				}
			}
			sb.WriteString("\n")
		}
	}

	return sb.String(), nil
}

func extractPDF(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", fmt.Errorf("open pdf: %w", err)
	}
	defer f.Close()

	var sb strings.Builder
	total := r.NumPage()
	for i := 1; i <= total; i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}
		txt, err2 := p.GetPlainText(nil)
		if err2 != nil {
			continue
		}
		sb.WriteString(txt)
		sb.WriteString("\n")
	}
	return sb.String(), nil
}

func normalizeText(s string) string {
	s = strings.ReplaceAll(s, "\u0000", " ")
	s = strings.Map(func(r rune) rune {
		if r == '\r' || r == '\t' {
			return ' '
		}
		return r
	}, s)

	var out bytes.Buffer
	out.Grow(len(s))

	space := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if !space {
				out.WriteByte(' ')
				space = true
			}
			continue
		}
		space = false
		out.WriteRune(r)
	}
	return strings.TrimSpace(out.String())
}
