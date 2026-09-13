package formatting

import "testing"

func TestWhatsApp(t *testing.T) {
	got := Render("# Title\n\n**bold**\n- item\n[link](https://example.com)", ProfileWhatsApp)
	want := "*Title*\n\n*bold*\n• item\nlink (https://example.com)"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestWhatsAppTable(t *testing.T) {
	got := Render("| Nombre | Estado |\n|---|---|\n| API | activo |", ProfileWhatsApp)
	want := "• *Nombre:* API · *Estado:* activo"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
