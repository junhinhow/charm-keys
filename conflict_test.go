package keys_test

import (
	"reflect"
	"testing"

	"github.com/junhinhow/charm-keys"
)

func TestConflictsNoMesmoContexto(t *testing.T) {
	r := keys.New().Global(
		keys.Binding{Action: "save", Keys: []string{"s"}},
		keys.Binding{Action: "search", Keys: []string{"s"}},
	)

	want := []keys.Conflict{{
		Context: "global",
		Key:     "s",
		ActionA: "save",
		ActionB: "search",
	}}
	if got := r.Conflicts(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Conflicts() = %#v, quer %#v", got, want)
	}
}

func TestConflictsDetectaFormasEquivalentes(t *testing.T) {
	r := keys.New().Define("list",
		keys.Binding{Action: "back", Keys: []string{"esc"}},
		keys.Binding{Action: "cancel", Keys: []string{"escape"}},
	)

	want := []keys.Conflict{{
		Context: "list",
		Key:     "esc",
		ActionA: "back",
		ActionB: "cancel",
	}}
	if got := r.Conflicts(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Conflicts() = %#v, quer %#v", got, want)
	}
}

func TestConflictsIgnoraMesmaAcao(t *testing.T) {
	r := keys.New().Global(
		keys.Binding{Action: "quit", Keys: []string{"q", "q", "ctrl+c"}},
		keys.Binding{Action: "quit", Keys: []string{"q"}},
	)

	if got := r.Conflicts(); len(got) != 0 {
		t.Fatalf("Conflicts() = %#v, quer vazio", got)
	}
}

func TestContextosDiferentesNaoSaoConflito(t *testing.T) {
	r := keys.New().
		Global(keys.Binding{Action: "global-save", Keys: []string{"s"}}).
		Define("list", keys.Binding{Action: "list-save", Keys: []string{"s"}}).
		Define("modal", keys.Binding{Action: "modal-save", Keys: []string{"s"}})

	if got := r.Conflicts(); len(got) != 0 {
		t.Fatalf("Conflicts() entre contextos = %#v, quer vazio", got)
	}
}

func TestConflictsIncluiContextoInativo(t *testing.T) {
	r := keys.New().Define("fora-da-pilha",
		keys.Binding{Action: "a", Keys: []string{"x"}},
		keys.Binding{Action: "b", Keys: []string{"x"}},
	)

	if got := r.Conflicts(); len(got) != 1 {
		t.Fatalf("Conflicts() = %#v, quer 1 conflito", got)
	}
}

func TestConflictsOrdemDeterministica(t *testing.T) {
	r := keys.New().
		Global(
			keys.Binding{Action: "g1", Keys: []string{"q"}},
			keys.Binding{Action: "g2", Keys: []string{"q"}},
		).
		Define("z-first",
			keys.Binding{Action: "z1", Keys: []string{"a"}},
			keys.Binding{Action: "z2", Keys: []string{"a"}},
		).
		Define("a-second",
			keys.Binding{Action: "a1", Keys: []string{"b"}},
			keys.Binding{Action: "a2", Keys: []string{"b"}},
			keys.Binding{Action: "a3", Keys: []string{"b"}},
		)

	want := []keys.Conflict{
		{Context: "global", Key: "q", ActionA: "g1", ActionB: "g2"},
		{Context: "z-first", Key: "a", ActionA: "z1", ActionB: "z2"},
		{Context: "a-second", Key: "b", ActionA: "a1", ActionB: "a2"},
		{Context: "a-second", Key: "b", ActionA: "a1", ActionB: "a3"},
	}
	got := r.Conflicts()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Conflicts() = %#v, quer %#v", got, want)
	}
	if second := r.Conflicts(); !reflect.DeepEqual(got, second) {
		t.Fatalf("Conflicts() não é estável: %#v e %#v", got, second)
	}
}
