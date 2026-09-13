package formatting

import (
	"regexp"
	"strings"
)

type Profile string

const (
	ProfileMarkdown Profile = "markdown"
	ProfileWhatsApp Profile = "whatsapp"
	ProfileTelegram Profile = "telegram"
	ProfilePlain    Profile = "plain"
)

var linkRE = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
var headingRE = regexp.MustCompile(`(?m)^#{1,6}\s+(.+)$`)

// Render converts canonical Markdown only at a channel boundary.
func Render(input string, profile Profile) string {
	switch profile {
	case ProfileWhatsApp:
		return whatsapp(input)
	case ProfilePlain:
		return plain(input)
	default:
		return input
	}
}

func whatsapp(input string) string {
	input = renderTables(input, ProfileWhatsApp)
	input = linkRE.ReplaceAllString(input, "$1 ($2)")
	input = headingRE.ReplaceAllString(input, "*$1*")
	input = strings.ReplaceAll(input, "**", "*")
	lines := strings.Split(input, "\n")
	for i, line := range lines {
		trim := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trim, "- ") || strings.HasPrefix(trim, "* ") {
			indent := line[:len(line)-len(trim)]
			lines[i] = indent + "• " + trim[2:]
		}
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func plain(input string) string {
	input = renderTables(input, ProfilePlain)
	input = linkRE.ReplaceAllString(input, "$1 ($2)")
	input = headingRE.ReplaceAllString(input, "$1")
	input = strings.ReplaceAll(input, "**", "")
	return strings.TrimSpace(input)
}
