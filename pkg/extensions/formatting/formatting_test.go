package formatting

import "testing"

func TestWhatsApp(t *testing.T) {
	got := Render("# Title\n\n**bold**\n- item\n[link](https://example.com)", ProfileWhatsApp)
	want := "*Title*\n\n*bold*\n• item\nlink (https://example.com)"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
