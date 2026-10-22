package keys_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/junhinhow/charm-keys"
)

// helpResolver monta três níveis: global, tela e modal.
func helpResolver() *keys.Resolver {
	return keys.New().
		Global(
			keys.Binding{Action: "quit", Keys: []string{"q"}, Help: "sair"},
			keys.Binding{Action: "help", Keys: []string{"?"}, Help: "ajuda"},
		).
		Define("list",
			keys.Binding{Action: "open", Keys: []string{"enter"}, Help: "abrir"},
			keys.Binding{Action: "quit", Keys: []string{"x"}, Help: "sair da lista"},
		).
		Define("modal",
			keys.Binding{Action: "confirm", Keys: []string{"enter"}, Help: "confirmar"},
			keys.Binding{Action: "cancel", Keys: []string{"esc"}, Help: "cancelar"},
		).
		Push("list")
}

func TestHelpOrdemEPrecedencia(t *testing.T) {
	r := helpResolver()

	if got, want := r.Help(), "enter  abrir\nx  sair da lista\n?  ajuda"; got != want {
		t.Fatalf("Help() com lista = %q, quer %q", got, want)
	}

	r.Push("modal")
	want := "enter  confirmar\nesc  cancelar\nx  sair da lista\n?  ajuda"
	if got := r.Help(); got != want {
		t.Fatalf("Help() com modal = %q, quer %q", got, want)
	}

	if count := strings.Count(r.Help(), "sair da lista"); count != 1 {
		t.Fatalf("ação repetida no help: %q", r.Help())
	}

	r.Pop()
	if got, want := r.Help(), "enter  abrir\nx  sair da lista\n?  ajuda"; got != want {
		t.Fatalf("Help() depois do Pop = %q, quer %q", got, want)
	}
}

func TestHelpSemGlobal(t *testing.T) {
	r := keys.New()
	if got := r.Help(); got != "" {
		t.Fatalf("Help() vazio = %q, quer vazio", got)
	}
	if got := r.HelpWidth(20); got != "" {
		t.Fatalf("HelpWidth(20) vazio = %q, quer vazio", got)
	}
}

func TestHelpUsaActionQuandoHelpVazio(t *testing.T) {
	r := keys.New().Global(keys.Binding{Action: "quit", Keys: []string{"q"}})
	if got, want := r.Help(), "q  quit"; got != want {
		t.Fatalf("Help() = %q, quer %q", got, want)
	}
}

func TestHelpOmiteTeclasSombreadas(t *testing.T) {
	r := keys.New().
		Define("list", keys.Binding{Action: "back", Keys: []string{"esc", "b"}, Help: "voltar"}).
		Define("modal", keys.Binding{Action: "close", Keys: []string{"esc"}, Help: "fechar"}).
		Push("list").
		Push("modal")

	if got, want := r.Help(), "esc  fechar\nb  voltar"; got != want {
		t.Fatalf("Help() = %q, quer %q", got, want)
	}
}

func TestHelpWidthRespeitaLimite(t *testing.T) {
	r := helpResolver()
	r.Push("modal")

	for _, width := range []int{20, 40} {
		got := r.HelpWidth(width)
		if got == "" {
			t.Fatalf("HelpWidth(%d) vazio", width)
		}
		for _, line := range strings.Split(got, "\n") {
			if lineWidth := ansi.StringWidth(line); lineWidth > width {
				t.Fatalf("HelpWidth(%d) gerou linha com %d células: %q", width, lineWidth, line)
			}
		}
		if second := r.HelpWidth(width); second != got {
			t.Fatalf("HelpWidth(%d) não é estável: %q e %q", width, got, second)
		}
	}
}

func TestHelpWidthCasosDeBorda(t *testing.T) {
	r := helpResolver()
	r.Push("modal")

	if got := r.HelpWidth(0); got != "" {
		t.Fatalf("HelpWidth(0) = %q, quer vazio", got)
	}
	if got := r.HelpWidth(-10); got != "" {
		t.Fatalf("HelpWidth(-10) = %q, quer vazio", got)
	}

	for _, width := range []int{1, 2, 5} {
		got := r.HelpWidth(width)
		for _, line := range strings.Split(got, "\n") {
			if lineWidth := ansi.StringWidth(line); lineWidth > width {
				t.Fatalf("HelpWidth(%d) gerou linha com %d células: %q", width, lineWidth, line)
			}
		}
	}
}

func TestHelpWidthEmColunas(t *testing.T) {
	r := keys.New().Global(
		keys.Binding{Action: "a", Keys: []string{"a"}, Help: "um"},
		keys.Binding{Action: "b", Keys: []string{"b"}, Help: "dois"},
		keys.Binding{Action: "c", Keys: []string{"c"}, Help: "tres"},
	)

	// 40 células cabem as três colunas em uma única linha.
	if got, want := r.HelpWidth(40), "a  um    b  dois  c  tres"; got != want {
		t.Fatalf("HelpWidth(40) = %q, quer %q", got, want)
	}

	// 20 células forçam duas colunas (preenchimento por coluna).
	if got, want := r.HelpWidth(20), "a  um    c  tres\nb  dois"; got != want {
		t.Fatalf("HelpWidth(20) = %q, quer %q", got, want)
	}
}

func TestHelpWidthNaoQuebraCaracteresLargos(t *testing.T) {
	r := keys.New().Global(
		keys.Binding{Action: "ok", Keys: []string{"ç"}, Help: "ação concluída"},
	)

	got := r.HelpWidth(5)
	if ansi.StringWidth(got) > 5 {
		t.Fatalf("HelpWidth(5) = %q tem largura %d", got, ansi.StringWidth(got))
	}
	if strings.ContainsRune(got, '\uFFFD') {
		t.Fatalf("HelpWidth(5) quebrou um caractere: %q", got)
	}
}
