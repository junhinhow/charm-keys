package keys_test

import (
	"reflect"
	"testing"

	"charm.land/bubbles/v2/key"
	"github.com/junhinhow/charm-keys"
)

func TestFromBubbles(t *testing.T) {
	source := key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q/ctrl+c", "quit"),
	)

	got := keys.FromBubbles(source)
	want := keys.Binding{
		Action: "quit",
		Keys:   []string{"q", "ctrl+c"},
		Help:   "quit",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FromBubbles() = %#v, quer %#v", got, want)
	}
}

func TestFromBubblesDesabilitado(t *testing.T) {
	source := key.NewBinding(key.WithHelp("x", "sem teclas"))
	if source.Enabled() {
		t.Fatal("binding de teste deveria estar desabilitado")
	}

	if got := keys.FromBubbles(source); got.Action != "" || len(got.Keys) != 0 {
		t.Fatalf("FromBubbles(desabilitado) = %#v, quer Binding vazio", got)
	}
}

func TestToBubbles(t *testing.T) {
	binding := keys.Binding{
		Action: "quit",
		Keys:   []string{"q", "ctrl+c"},
		Help:   "sair",
	}

	got := binding.ToBubbles()
	if !got.Enabled() {
		t.Fatal("ToBubbles() gerou binding desabilitado")
	}
	if want := []string{"q", "ctrl+c"}; !reflect.DeepEqual(got.Keys(), want) {
		t.Fatalf("Keys() = %v, quer %v", got.Keys(), want)
	}
	if got.Help().Key != "q/ctrl+c" {
		t.Fatalf("Help().Key = %q, quer %q", got.Help().Key, "q/ctrl+c")
	}
	if got.Help().Desc != "sair" {
		t.Fatalf("Help().Desc = %q, quer %q", got.Help().Desc, "sair")
	}

	roundTrip := keys.FromBubbles(got)
	// O bubbles não carrega Action: a descrição vira a ação na volta.
	if roundTrip.Action != "sair" {
		t.Fatalf("Action na ida e volta = %q, quer %q", roundTrip.Action, "sair")
	}
	if !reflect.DeepEqual(roundTrip.Keys, binding.Keys) {
		t.Fatalf("Keys na ida e volta = %v, quer %v", roundTrip.Keys, binding.Keys)
	}
	if roundTrip.Help != binding.Help {
		t.Fatalf("Help na ida e volta = %q, quer %q", roundTrip.Help, binding.Help)
	}
}

func TestToBubblesUsaActionComoDescricao(t *testing.T) {
	binding := keys.Binding{Action: "quit", Keys: []string{"q"}}
	if got := binding.ToBubbles().Help().Desc; got != "quit" {
		t.Fatalf("Help().Desc = %q, quer %q", got, "quit")
	}
}

func TestBindingMatches(t *testing.T) {
	binding := keys.Binding{Action: "quit", Keys: []string{"q", "ctrl+c", "ç"}}

	cases := []struct {
		key  string
		want bool
	}{
		{"q", true},
		{"ctrl+c", true},
		{"C-c", true},
		{"ctrl+C", true},
		{"ç", true},
		{"x", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := binding.Matches(tc.key); got != tc.want {
			t.Fatalf("Matches(%q) = %v, quer %v", tc.key, got, tc.want)
		}
	}

	empty := keys.Binding{Action: "nada"}
	if empty.Matches("") {
		t.Fatal("binding vazio casou com tecla vazia")
	}
}
