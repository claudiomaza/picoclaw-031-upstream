package formatting

import (
	"regexp"
	"strings"
)

var tableSepRE = regexp.MustCompile(`^\s*\|?\s*:?-{3,}:?\s*(\|\s*:?-{3,}:?\s*)+\|?\s*$`)

func renderTables(input string, profile Profile) string {
	lines := strings.Split(input, "\n")
	var out []string
	for i := 0; i < len(lines); {
		if i+1 >= len(lines) || !strings.Contains(lines[i], "|") || !tableSepRE.MatchString(lines[i+1]) {
			out = append(out, lines[i])
			i++
			continue
		}
		headers := tableCells(lines[i])
		i += 2
		for i < len(lines) && strings.Contains(lines[i], "|") && strings.TrimSpace(lines[i]) != "" {
			cells := tableCells(lines[i])
			i++
			parts := make([]string, 0, len(headers))
			for j, h := range headers {
				if j >= len(cells) || strings.TrimSpace(cells[j]) == "" {
					continue
				}
				if profile == ProfileWhatsApp {
					parts = append(parts, "*"+h+":* "+cells[j])
				} else {
					parts = append(parts, h+": "+cells[j])
				}
			}
			if len(parts) > 0 {
				out = append(out, "• "+strings.Join(parts, " · "))
			}
		}
	}
	return strings.Join(out, "\n")
}

func tableCells(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	parts := strings.Split(line, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
