package rime

import (
	"strings"

	"github.com/gaboolic/moqi-ime/imecore"
)

const (
	dictionaryNoteSeparator = "\u000c"
	dictionaryRowSeparator = "\u000d"
	reverseLookupMarker    = "\u000b"
)

func isDictionaryLookupComment(comment string) bool {
	if comment == "" {
		return false
	}
	return strings.Contains(comment, dictionaryNoteSeparator+dictionaryRowSeparator) ||
		strings.Contains(comment, reverseLookupMarker) ||
		strings.Contains(comment, dictionaryRowSeparator+"1,")
}

func primaryDictionaryJyutping(comment string) string {
	rows := splitDictionaryRows(comment)
	for _, row := range rows {
		fields := splitCSVRow(row)
		if len(fields) > 2 && strings.TrimSpace(fields[2]) != "" {
			return strings.TrimSpace(fields[2])
		}
	}
	return ""
}

func splitDictionaryRows(comment string) []string {
	start := 0
	if idx := strings.Index(comment, dictionaryNoteSeparator+dictionaryRowSeparator); idx >= 0 {
		start = idx + len(dictionaryNoteSeparator+dictionaryRowSeparator)
	} else if idx := strings.Index(comment, dictionaryRowSeparator); idx >= 0 {
		start = idx + 1
	}
	body := comment[start:]
	if body == "" {
		return nil
	}
	parts := strings.Split(body, dictionaryRowSeparator)
	rows := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		rows = append(rows, part)
	}
	return rows
}

func splitCSVRow(row string) []string {
	fields := make([]string, 0, 8)
	var current strings.Builder
	inQuotes := false
	for _, r := range row {
		switch r {
		case '"':
			inQuotes = !inQuotes
		case ',':
			if inQuotes {
				current.WriteRune(r)
			} else {
				fields = append(fields, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	fields = append(fields, current.String())
	return fields
}

func candidateEntryFromItem(candidate candidateItem, inputCode string) imecore.CandidateEntry {
	entry := imecore.CandidateEntry{
		Text: candidate.Text,
	}
	comment := candidate.Comment
	if isDictionaryLookupComment(comment) {
		entry.RawLookupComment = comment
		if jyutping := primaryDictionaryJyutping(comment); jyutping != "" {
			entry.Jyutping = jyutping
		}
	} else if comment != "" {
		entry.Comment = comment
	}
	if inputCode != "" {
		entry.InputCode = inputCode
	}
	return entry
}
