package formatting

// Options controls which canonical source formats are accepted by the
// renderer. Markdown is the current canonical format; plaintext is opt-in.
type Options struct {
	Markdown  bool `json:"markdown"`
	Plaintext bool `json:"plaintext"`
}

var DefaultOptions = Options{Markdown: true, Plaintext: false}

func RenderWithOptions(input string, profile Profile, options Options) string {
	if !options.Markdown && !options.Plaintext {
		return input
	}
	return Render(input, profile)
}
