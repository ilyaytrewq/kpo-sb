package analisys_functions

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
	"unicode/utf8"

	"net/http"
	"os"
	"os/exec"

	"github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-analisys/internal/clients/embedding"
)

type File struct {
	Name        *string
	Size        int64
	ContentType *string
	Data        []byte
}

func SplitFileIntoChunks(file File) ([]embedding.TextChunk, error) {
	if len(file.Data) == 0 {
		return nil, fmt.Errorf("empty file data")
	}

	ct := ""
	if file.ContentType != nil {
		ct = *file.ContentType
	}
	if ct == "" || ct == "application/octet-stream" {
		// если не прислали ct — попробуем угадать по байтам
		ct = http.DetectContentType(file.Data)
	}

	var rawText string

	if ct == "application/pdf" {
		t, err := extractTextFromPDF(file.Data)
		if err != nil {
			return nil, err
		}
		rawText = t
	} else {
		if ct != "" && !strings.HasPrefix(ct, "text/") && ct != "application/octet-stream" {
			return nil, fmt.Errorf("unsupported content type: %s", ct)
		}
		if !utf8.Valid(file.Data) {
			return nil, fmt.Errorf("file data is not valid utf-8")
		}
		rawText = string(file.Data)
	}

	text := normalizeText(rawText)
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("empty text after normalization")
	}

	const (
		maxRunes     = 6000
		overlapRunes = 1000
		minRunes     = 20
	)

	units := splitUnits(text, maxRunes)
	chunkTexts := buildChunksWithOverlap(units, maxRunes, overlapRunes)

	sum := sha1.Sum(file.Data)
	base := hex.EncodeToString(sum[:8])

	res := make([]embedding.TextChunk, 0, len(chunkTexts))
	for i, ch := range chunkTexts {
		ch = strings.TrimSpace(ch)
		if ch == "" {
			continue
		}
		if utf8.RuneCountInString(ch) < minRunes {
			continue
		}
		res = append(res, embedding.TextChunk{
			ChunkId:    fmt.Sprintf("%s:%06d", base, i),
			ChunkIndex: i,
			Text:       ch,
		})
	}

	if len(res) == 0 {
		return nil, fmt.Errorf("no chunks produced")
	}
	return res, nil
}

func normalizeText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	s = strings.Join(lines, "\n")

	for strings.Contains(s, "\n\n\n") {
		s = strings.ReplaceAll(s, "\n\n\n", "\n\n")
	}
	return strings.TrimSpace(s)
}

func splitUnits(s string, maxRunes int) []string {
	var out []string

	paras := splitNonEmpty(s, "\n\n")
	for _, p := range paras {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if utf8.RuneCountInString(p) <= maxRunes {
			out = append(out, p)
			continue
		}

		lines := splitNonEmpty(p, "\n")
		for _, ln := range lines {
			ln = strings.TrimSpace(ln)
			if ln == "" {
				continue
			}
			if utf8.RuneCountInString(ln) <= maxRunes {
				out = append(out, ln)
				continue
			}

			words := splitNonEmpty(ln, " ")
			var cur strings.Builder
			curLen := 0

			flush := func() {
				t := strings.TrimSpace(cur.String())
				if t != "" {
					out = append(out, t)
				}
				cur.Reset()
				curLen = 0
			}

			for _, w := range words {
				if w == "" {
					continue
				}
				wLen := utf8.RuneCountInString(w)
				if wLen > maxRunes {
					if curLen > 0 {
						flush()
					}
					rs := []rune(w)
					for i := 0; i < len(rs); i += maxRunes {
						j := i + maxRunes
						if j > len(rs) {
							j = len(rs)
						}
						out = append(out, string(rs[i:j]))
					}
					continue
				}

				addLen := wLen
				if curLen > 0 {
					addLen++
				}
				if curLen+addLen > maxRunes {
					flush()
				}
				if curLen > 0 {
					cur.WriteByte(' ')
				}
				cur.WriteString(w)
				curLen += addLen
			}

			if curLen > 0 {
				flush()
			}
		}
	}

	return out
}

func buildChunksWithOverlap(units []string, maxRunes, overlapRunes int) []string {
	var chunks []string

	var cur strings.Builder
	curLen := 0

	flush := func() string {
		t := strings.TrimSpace(cur.String())
		cur.Reset()
		curLen = 0
		return t
	}

	for _, u := range units {
		u = strings.TrimSpace(u)
		if u == "" {
			continue
		}
		uLen := utf8.RuneCountInString(u)
		if uLen > maxRunes {
			rs := []rune(u)
			for i := 0; i < len(rs); i += maxRunes {
				j := i + maxRunes
				if j > len(rs) {
					j = len(rs)
				}
				part := strings.TrimSpace(string(rs[i:j]))
				if part != "" {
					chunks = append(chunks, part)
				}
			}
			continue
		}

		sepLen := 0
		if curLen > 0 {
			sepLen = 2
		}
		if curLen+sepLen+uLen <= maxRunes {
			if curLen > 0 {
				cur.WriteString("\n\n")
			}
			cur.WriteString(u)
			curLen += sepLen + uLen
			continue
		}

		prev := flush()
		if prev != "" {
			chunks = append(chunks, prev)
		}

		carry := tailRunes(prev, overlapRunes)
		carryLen := utf8.RuneCountInString(carry)
		if carryLen+uLen > maxRunes {
			need := maxRunes - uLen
			if need < 0 {
				need = 0
			}
			carry = tailRunes(prev, need)
			carryLen = utf8.RuneCountInString(carry)
		}

		if carryLen > 0 {
			cur.WriteString(carry)
			curLen = carryLen
		}
		if curLen > 0 {
			cur.WriteString("\n\n")
			curLen += 2
		}
		cur.WriteString(u)
		curLen += uLen
	}

	last := flush()
	if last != "" {
		chunks = append(chunks, last)
	}

	return chunks
}

func splitNonEmpty(s, sep string) []string {
	raw := strings.Split(s, sep)
	out := make([]string, 0, len(raw))
	for _, x := range raw {
		x = strings.TrimSpace(x)
		if x != "" {
			out = append(out, x)
		}
	}
	return out
}

func tailRunes(s string, n int) string {
	if n <= 0 || s == "" {
		return ""
	}
	rs := []rune(s)
	if len(rs) <= n {
		return string(rs)
	}
	return string(rs[len(rs)-n:])
}

func extractTextFromPDF(data []byte) (string, error) {
	pdfF, err := os.CreateTemp("", "apl-*.pdf")
	if err != nil {
		return "", err
	}
	pdfName := pdfF.Name()
	defer os.Remove(pdfName)

	if _, err := pdfF.Write(data); err != nil {
		pdfF.Close()
		return "", err
	}
	if err := pdfF.Close(); err != nil {
		return "", err
	}

	txtF, err := os.CreateTemp("", "apl-*.txt")
	if err != nil {
		return "", err
	}
	txtName := txtF.Name()
	txtF.Close()
	defer os.Remove(txtName)

	cmd := exec.Command("pdftotext", "-layout", pdfName, txtName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if execErr, ok := err.(*exec.Error); ok && execErr.Err == exec.ErrNotFound {
			return "", fmt.Errorf("pdftotext not found: install poppler-utils")
		}
		return "", fmt.Errorf("pdftotext failed: %w: %s", err, string(out))
	}

	b, err := os.ReadFile(txtName)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
