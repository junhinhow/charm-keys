package keys

import (
	"strings"

	"charm.land/bubbles/v2/key"
)

// Binding associa teclas a uma ação nomeada.
//
// Action é o identificador estável da ação (por exemplo, "quit"). Keys aceita
// qualquer grafia da tabela de normalização (por exemplo, "ctrl+c", "C-c" ou
// "ctrl+C"). Help é o texto exibido por [Resolver.Help].
type Binding struct {
	Action string
	Keys   []string
	Help   string
}

// Context é um grupo nomeado de bindings (por exemplo, "global", "list" ou
// "modal:confirm").
type Context struct {
	Name     string
	Bindings []Binding

	// index mapeia tecla normalizada para ação, preservando a primeira
	// definição em caso de conflito.
	index map[string]string
}

// Matches informa se uma tecla textual corresponde ao binding.
//
// A comparação usa a mesma normalização de [Resolver.ResolveString], então
// "ctrl+c", "C-c" e "ctrl+C" são equivalentes. Uma tecla vazia nunca casa e um
// binding sem teclas nunca casa.
func (b Binding) Matches(key string) bool {
	normalized := normalizeKey(key)
	if normalized == "" {
		return false
	}
	for _, candidate := range b.Keys {
		if normalizeKey(candidate) == normalized {
			return true
		}
	}
	return false
}

// FromBubbles converte um key.Binding do bubbles.
//
// Como key.Binding não possui um identificador de ação, o texto de ajuda
// (Help().Desc) é usado como Action; se estiver vazio, o rótulo de teclas
// (Help().Key) assume esse papel. Bindings desabilitados ou sem teclas viram
// um Binding vazio, que o Resolver ignora.
//
// A conversão de ida e volta por [Binding.ToBubbles] não preserva Action
// quando Help e Action são diferentes: o bubbles só carrega um texto de ajuda.
func FromBubbles(b key.Binding) Binding {
	if !b.Enabled() {
		return Binding{}
	}

	help := b.Help()
	action := strings.TrimSpace(help.Desc)
	if action == "" {
		action = strings.TrimSpace(help.Key)
	}

	return Binding{
		Action: action,
		Keys:   append([]string(nil), b.Keys()...),
		Help:   help.Desc,
	}
}

// ToBubbles converte o binding para um key.Binding do bubbles, para reutilizar
// componentes que esperam esse tipo.
//
// O rótulo de ajuda é formado pelas teclas separadas por "/" e a descrição é
// [Binding.Help] (ou [Binding.Action] quando Help está vazio). O bubbles não
// tem conceito de Action: converter de volta com [FromBubbles] usa a descrição
// como Action.
func (b Binding) ToBubbles() key.Binding {
	description := b.Help
	if description == "" {
		description = b.Action
	}
	return key.NewBinding(
		key.WithKeys(b.Keys...),
		key.WithHelp(strings.Join(b.Keys, "/"), description),
	)
}
