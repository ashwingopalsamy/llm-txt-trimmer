package compact

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Mode string

const (
	Safe    Mode = "safe"
	Compact Mode = "compact"
	Dense   Mode = "dense"
)

type Stats struct {
	BeforeBytes      int
	AfterBytes       int
	BeforeRunes      int
	AfterRunes       int
	BeforeWhitespace int
	AfterWhitespace  int
	BeforeLines      int
	AfterLines       int
}

func (s Stats) SavedBytes() int { return s.BeforeBytes - s.AfterBytes }

func (s Stats) SavedPercent() float64 {
	if s.BeforeBytes == 0 {
		return 0
	}
	return float64(s.SavedBytes()) * 100 / float64(s.BeforeBytes)
}

func (s Stats) TokenProxyBefore() int { return tokenProxy(s.BeforeRunes) }
func (s Stats) TokenProxyAfter() int  { return tokenProxy(s.AfterRunes) }

func tokenProxy(runes int) int {
	if runes == 0 {
		return 0
	}
	return (runes + 3) / 4
}

func Transform(input string, mode Mode) (string, Stats, error) {
	if mode != Safe && mode != Compact && mode != Dense {
		return "", Stats{}, fmt.Errorf("unknown mode %q", mode)
	}

	normalized := strings.ReplaceAll(strings.ReplaceAll(input, "\r\n", "\n"), "\r", "\n")
	lines := strings.Split(normalized, "\n")

	out := make([]string, 0, len(lines))
	inFence := false
	fenceChar := byte(0)
	fenceLen := 0
	inFrontMatter := false
	frontMatterSeen := false

	var paragraph []string
	flushParagraph := func() {
		if len(paragraph) == 0 {
			return
		}
		if mode == Safe {
			out = append(out, paragraph...)
		} else {
			out = append(out, joinParagraph(paragraph, mode == Dense))
		}
		paragraph = paragraph[:0]
	}

	appendBlank := func() {
		if len(out) == 0 || out[len(out)-1] == "" {
			return
		}
		if mode != Dense {
			out = append(out, "")
		}
	}

	for i, raw := range lines {
		line := raw

		if !frontMatterSeen && i == 0 && strings.TrimSpace(line) == "---" {
			flushParagraph()
			inFrontMatter = true
			frontMatterSeen = true
			out = append(out, line)
			continue
		}
		if inFrontMatter {
			out = append(out, line)
			if strings.TrimSpace(line) == "---" || strings.TrimSpace(line) == "..." {
				inFrontMatter = false
			}
			continue
		}

		if marker, n, ok := fenceMarker(line); ok {
			flushParagraph()
			if !inFence {
				inFence = true
				fenceChar, fenceLen = marker, n
			} else if marker == fenceChar && n >= fenceLen {
				inFence = false
				fenceChar, fenceLen = 0, 0
			}
			out = append(out, line)
			continue
		}
		if inFence {
			out = append(out, line)
			continue
		}

		if strings.TrimSpace(line) == "" {
			flushParagraph()
			appendBlank()
			continue
		}

		if mode == Safe {
			paragraph = append(paragraph, line)
			continue
		}

		if isStructural(line) {
			flushParagraph()
			out = append(out, normalizeStructural(line, mode == Dense))
			continue
		}

		paragraph = append(paragraph, line)
	}
	flushParagraph()

	for len(out) > 0 && out[0] == "" {
		out = out[1:]
	}
	for len(out) > 0 && out[len(out)-1] == "" {
		out = out[:len(out)-1]
	}

	result := strings.Join(out, "\n")
	stats := Stats{
		BeforeBytes:      len(input),
		AfterBytes:       len(result),
		BeforeRunes:      utf8.RuneCountInString(input),
		AfterRunes:       utf8.RuneCountInString(result),
		BeforeWhitespace: countWhitespace(input),
		AfterWhitespace:  countWhitespace(result),
		BeforeLines:      lineCount(input),
		AfterLines:       lineCount(result),
	}
	return result, stats, nil
}

func joinParagraph(lines []string, dense bool) string {
	parts := make([]string, 0, len(lines))
	for _, line := range lines {
		t := strings.TrimSpace(line)
		if dense {
			t = strings.Join(strings.Fields(t), " ")
		}
		if t != "" {
			parts = append(parts, t)
		}
	}
	return strings.Join(parts, " ")
}

func normalizeStructural(line string, dense bool) string {
	if !dense {
		return strings.TrimRight(line, " \t")
	}
	prefixLen := len(line) - len(strings.TrimLeft(line, " \t"))
	prefix := line[:prefixLen]
	body := strings.TrimSpace(line[prefixLen:])
	return prefix + strings.Join(strings.Fields(body), " ")
}

func fenceMarker(line string) (byte, int, bool) {
	s := strings.TrimLeft(line, " ")
	if len(line)-len(s) > 3 || len(s) < 3 {
		return 0, 0, false
	}
	c := s[0]
	if c != '`' && c != '~' {
		return 0, 0, false
	}
	n := 0
	for n < len(s) && s[n] == c {
		n++
	}
	if n < 3 {
		return 0, 0, false
	}
	return c, n, true
}

func isStructural(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}

	leading := len(line) - len(strings.TrimLeft(line, " \t"))
	if leading >= 4 || strings.HasPrefix(line, "\t") {
		return true
	}

	if strings.HasPrefix(trimmed, "#") ||
		strings.HasPrefix(trimmed, ">") ||
		strings.HasPrefix(trimmed, "|") ||
		isHorizontalRule(trimmed) ||
		isListItem(trimmed) ||
		isDefinitionLike(trimmed) {
		return true
	}

	if strings.HasPrefix(trimmed, "<") && strings.HasSuffix(trimmed, ">") {
		return true
	}
	return false
}

func isHorizontalRule(s string) bool {
	compact := strings.ReplaceAll(strings.ReplaceAll(s, " ", ""), "\t", "")
	if len(compact) < 3 {
		return false
	}
	c := compact[0]
	if c != '-' && c != '*' && c != '_' {
		return false
	}
	for i := 1; i < len(compact); i++ {
		if compact[i] != c {
			return false
		}
	}
	return true
}

func isListItem(s string) bool {
	if len(s) >= 2 && (s[0] == '-' || s[0] == '*' || s[0] == '+') && unicode.IsSpace(rune(s[1])) {
		return true
	}

	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i == 0 || i+1 >= len(s) {
		return false
	}
	return (s[i] == '.' || s[i] == ')') && unicode.IsSpace(rune(s[i+1]))
}

func isDefinitionLike(s string) bool {
	idx := strings.IndexByte(s, ':')
	return idx > 0 && idx <= 40 && !strings.ContainsAny(s[:idx], ".!?")
}

func countWhitespace(s string) int {
	n := 0
	for _, r := range s {
		if unicode.IsSpace(r) {
			n++
		}
	}
	return n
}

func lineCount(s string) int {
	if s == "" {
		return 0
	}
	return strings.Count(strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\r", "\n"), "\n") + 1
}
