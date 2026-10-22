package keys_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/junhinhow/charm-keys"
)

// press monta a mensagem de uma tecla pressionada.
func press(code rune, text string, mod tea.KeyMod) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code, Text: text, Mod: mod}
}

func TestPrecedenciaGlobalTelaModal(t *testing.T) {
	r := keys.New().
		Global(
			keys.Binding{Action: "quit", Keys: []string{"q", "ctrl+c"}, Help: "sair"},
			keys.Binding{Action: "help", Keys: []string{"?"}, Help: "ajuda"},
		).
		Define("list",
			keys.Binding{Action: "open", Keys: []string{"enter"}, Help: "abrir"},
			keys.Binding{Action: "delete", Keys: []string{"d"}, Help: "apagar"},
			keys.Binding{Action: "quit", Keys: []string{"x"}, Help: "sair da lista"},
		).
		Define("modal:confirm",
			keys.Binding{Action: "confirm", Keys: []string{"enter"}, Help: "confirmar"},
			keys.Binding{Action: "cancel", Keys: []string{"esc"}, Help: "cancelar"},
			keys.Binding{Action: "modal-delete", Keys: []string{"d"}, Help: "apagar no modal"},
		).
		Push("list")

	assertAction(t, r, "enter", "open")
	assertAction(t, r, "d", "delete")
	assertAction(t, r, "q", "quit")
	assertAction(t, r, "ctrl+c", "quit")
	assertAction(t, r, "?", "help")

	r.Push("modal:confirm")

	assertAction(t, r, "enter", "confirm")
	assertAction(t, r, "d", "modal-delete")
	assertAction(t, r, "esc", "cancel")
	// Teclas não definidas no modal continuam caindo nos contextos de baixo.
	assertAction(t, r, "x", "quit")
	assertAction(t, r, "?", "help")

	if got := r.Pop(); got != "modal:confirm" {
		t.Fatalf("Pop() = %q, quer %q", got, "modal:confirm")
	}
	assertAction(t, r, "enter", "open")
	assertAction(t, r, "d", "delete")

	if got := r.Pop(); got != "list" {
		t.Fatalf("Pop() = %q, quer %q", got, "list")
	}
	assertNoAction(t, r, "enter")
	assertNoAction(t, r, "d")
	assertAction(t, r, "q", "quit")

	if got := r.Pop(); got != "" {
		t.Fatalf("Pop() no vazio = %q, quer %q", got, "")
	}
	if got := r.Pop(); got != "" {
		t.Fatalf("Pop() repetido no vazio = %q, quer %q", got, "")
	}
}

func TestPushIdempotenteEActive(t *testing.T) {
	r := keys.New()
	if got := r.Active(); len(got) != 0 {
		t.Fatalf("Active() inicial = %v, quer vazio", got)
	}

	r.Global(keys.Binding{Action: "quit", Keys: []string{"q"}})
	r.Push("list").Push("modal").Push("modal").Push("")

	want := []string{"global", "list", "modal"}
	got := r.Active()
	if len(got) != len(want) {
		t.Fatalf("Active() = %v, quer %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Active() = %v, quer %v", got, want)
		}
	}

	// Active devolve uma cópia: alterar o resultado não mexe no Resolver.
	got[0] = "alterado"
	if r.Active()[0] != "global" {
		t.Fatal("Active() expôs o slice interno")
	}

	r.Push("global")
	if got := r.Active(); len(got) != 3 {
		t.Fatalf("Push(\"global\") alterou a pilha: %v", got)
	}

	if got := r.Pop(); got != "modal" {
		t.Fatalf("Pop() = %q, quer %q", got, "modal")
	}
	if got := r.Pop(); got != "list" {
		t.Fatalf("Pop() = %q, quer %q", got, "list")
	}
	if got := r.Pop(); got != "" {
		t.Fatalf("Pop() do global = %q, quer vazio", got)
	}
}

func TestDefineComNomeVazioIgnorado(t *testing.T) {
	r := keys.New().
		Define("", keys.Binding{Action: "quit", Keys: []string{"q"}}).
		Define("   ", keys.Binding{Action: "quit", Keys: []string{"q"}}).
		Push("")

	assertNoAction(t, r, "q")
	if got := r.Active(); len(got) != 0 {
		t.Fatalf("Active() = %v, quer vazio", got)
	}
	if got := r.Conflicts(); len(got) != 0 {
		t.Fatalf("Conflicts() = %v, quer vazio", got)
	}
}

func TestDefineGlobalPeloNomeReservado(t *testing.T) {
	r := keys.New().
		Define("global", keys.Binding{Action: "quit", Keys: []string{"q"}})

	if got := r.Active(); len(got) != 1 || got[0] != "global" {
		t.Fatalf("Active() = %v, quer [global]", got)
	}
	assertAction(t, r, "q", "quit")
}

func TestPushAntesDeDefine(t *testing.T) {
	r := keys.New().Push("list")
	if got := r.Active(); len(got) != 1 || got[0] != "list" {
		t.Fatalf("Active() = %v, quer [list]", got)
	}
	assertNoAction(t, r, "q")

	r.Define("list", keys.Binding{Action: "quit", Keys: []string{"q"}})
	assertAction(t, r, "q", "quit")
}

func TestDefineAtualizaContexto(t *testing.T) {
	r := keys.New().
		Define("list", keys.Binding{Action: "old", Keys: []string{"q"}}).
		Push("list").
		Define("list", keys.Binding{Action: "new", Keys: []string{"w"}})

	assertNoAction(t, r, "q")
	assertAction(t, r, "w", "new")
	if got := r.Active(); len(got) != 1 || got[0] != "list" {
		t.Fatalf("Active() depois de redefinir = %v, quer [list]", got)
	}
}

func TestBindingsInvalidosSaoDescartados(t *testing.T) {
	r := keys.New().Define("list",
		keys.Binding{Action: "", Keys: []string{"q"}},
		keys.Binding{Action: "sem-teclas"},
		keys.Binding{Action: "vazia", Keys: []string{"", "   "}},
	)

	if got := r.Help(); got != "" {
		t.Fatalf("Help() = %q, quer vazio", got)
	}
	assertNoAction(t, r, "q")
}

func TestNormalizacaoDeTeclas(t *testing.T) {
	r := keys.New().Global(
		keys.Binding{Action: "copy", Keys: []string{"ctrl+c"}},
		keys.Binding{Action: "back", Keys: []string{"esc"}},
		keys.Binding{Action: "submit", Keys: []string{"enter"}},
		keys.Binding{Action: "toggle", Keys: []string{"space"}},
		keys.Binding{Action: "remove", Keys: []string{"delete"}},
		keys.Binding{Action: "page-down", Keys: []string{"pgdown"}},
		keys.Binding{Action: "combo", Keys: []string{"ctrl+alt+x"}},
		keys.Binding{Action: "cedilla", Keys: []string{"ç"}},
		keys.Binding{Action: "upper", Keys: []string{"A"}},
		keys.Binding{Action: "lower", Keys: []string{"a"}},
	)

	cases := []struct {
		key    string
		action string
	}{
		{"ctrl+c", "copy"},
		{"C-c", "copy"},
		{"ctrl+C", "copy"},
		{"Control+C", "copy"},
		{"CTRL+C", "copy"},
		{"esc", "back"},
		{"escape", "back"},
		{"Esc", "back"},
		{"ESCAPE", "back"},
		{"enter", "submit"},
		{"return", "submit"},
		{"Enter", "submit"},
		{"RETURN", "submit"},
		{"space", "toggle"},
		{"spacebar", "toggle"},
		{" ", "toggle"},
		{"del", "remove"},
		{"delete", "remove"},
		{"pgdown", "page-down"},
		{"pagedown", "page-down"},
		{"pgdn", "page-down"},
		{"page-down", "page-down"},
		{"PageDown", "page-down"},
		{"ctrl+alt+x", "combo"},
		{"alt+ctrl+x", "combo"},
		{"C-A-x", "combo"},
		{"CTRL+ALT+X", "combo"},
		{"ç", "cedilla"},
		{"A", "upper"},
		{"a", "lower"},
	}

	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			assertAction(t, r, tc.key, tc.action)
		})
	}

	for _, invalid := range []string{"", "   ", "ctrl+q", "ctrl+shift", "foo+bar", "ctrl"} {
		t.Run("sem-match-"+invalid, func(t *testing.T) {
			assertNoAction(t, r, invalid)
		})
	}
}

func TestResolveComKeyPressMsg(t *testing.T) {
	r := keys.New().Global(
		keys.Binding{Action: "quit", Keys: []string{"ctrl+c"}},
		keys.Binding{Action: "submit", Keys: []string{"enter"}},
		keys.Binding{Action: "cancel", Keys: []string{"esc"}},
		keys.Binding{Action: "cedilla", Keys: []string{"ç"}},
	)

	if action, ok := r.Resolve(press('c', "", tea.ModCtrl)); !ok || action != "quit" {
		t.Fatalf("Resolve(ctrl+c) = %q, %v; quer quit, true", action, ok)
	}
	if action, ok := r.Resolve(press(tea.KeyEnter, "", 0)); !ok || action != "submit" {
		t.Fatalf("Resolve(enter) = %q, %v; quer submit, true", action, ok)
	}
	if action, ok := r.Resolve(press(tea.KeyEsc, "", 0)); !ok || action != "cancel" {
		t.Fatalf("Resolve(esc) = %q, %v; quer cancel, true", action, ok)
	}
	if action, ok := r.Resolve(press('ç', "ç", 0)); !ok || action != "cedilla" {
		t.Fatalf("Resolve(ç) = %q, %v; quer cedilla, true", action, ok)
	}
	if action, ok := r.Resolve(press('z', "z", 0)); ok || action != "" {
		t.Fatalf("Resolve(z) = %q, %v; quer vazio, false", action, ok)
	}
	if action, ok := r.Resolve(tea.KeyPressMsg{}); ok || action != "" {
		t.Fatalf("Resolve(tecla vazia) = %q, %v; quer vazio, false", action, ok)
	}

	// Mensagem construída com Text preenchido: String() devolve "c", então o
	// fallback para Keystroke() precisa encontrar "ctrl+c".
	if action, ok := r.Resolve(press('c', "c", tea.ModCtrl)); !ok || action != "quit" {
		t.Fatalf("Resolve(ctrl+c com Text) = %q, %v; quer quit, true", action, ok)
	}
}

func TestResolveComNilResolver(t *testing.T) {
	var r *keys.Resolver

	assertNoAction(t, r, "q")
	if got := r.Pop(); got != "" {
		t.Fatalf("Pop() = %q, quer vazio", got)
	}
	if got := r.Active(); got != nil {
		t.Fatalf("Active() = %v, quer nil", got)
	}
	if got := r.Help(); got != "" {
		t.Fatalf("Help() = %q, quer vazio", got)
	}
	if got := r.HelpWidth(20); got != "" {
		t.Fatalf("HelpWidth(20) = %q, quer vazio", got)
	}
	if got := r.Conflicts(); got != nil {
		t.Fatalf("Conflicts() = %v, quer nil", got)
	}
}

func assertAction(t *testing.T, r *keys.Resolver, key, want string) {
	t.Helper()
	action, ok := r.ResolveString(key)
	if !ok {
		t.Fatalf("ResolveString(%q) não encontrou ação; quer %q", key, want)
	}
	if action != want {
		t.Fatalf("ResolveString(%q) = %q, quer %q", key, action, want)
	}
}

func assertNoAction(t *testing.T, r *keys.Resolver, key string) {
	t.Helper()
	action, ok := r.ResolveString(key)
	if ok {
		t.Fatalf("ResolveString(%q) = %q, true; quer false", key, action)
	}
}
