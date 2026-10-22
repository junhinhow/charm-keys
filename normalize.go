package keys

import (
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// keyAliases mapeia grafias alternativas para a forma canônica usada por
// [Resolver.Resolve] e [Binding.Matches].
var keyAliases = map[string]string{
	"escape":    "esc",
	"return":    "enter",
	"spacebar":  "space",
	"del":       "delete",
	"ins":       "insert",
	"bs":        "backspace",
	"pgdn":      "pgdown",
	"pagedown":  "pgdown",
	"page-down": "pgdown",
	"page_down": "pgdown",
	"pageup":    "pgup",
	"page-up":   "pgup",
	"page_up":   "pgup",
}

// namedKeys contém os nomes de teclas especiais que preservamos em minúsculas.
var namedKeys = func() map[string]struct{} {
	names := []string{
		"esc", "enter", "tab", "backspace", "delete", "insert", "home", "end",
		"pgup", "pgdown", "space", "up", "down", "left", "right",
	}
	for i := 1; i <= 63; i++ {
		names = append(names, "f"+strconv.Itoa(i))
	}
	set := make(map[string]struct{}, len(names))
	for _, name := range names {
		set[name] = struct{}{}
	}
	return set
}()

// modifierRank fixa a ordem em que os modificadores aparecem na forma
// canônica: ctrl, alt, shift, meta, hyper, super (a mesma do Bubble Tea).
var modifierRank = map[string]int{
	"ctrl":  0,
	"alt":   1,
	"shift": 2,
	"meta":  3,
	"hyper": 4,
	"super": 5,
}

// normalizeKey converte uma tecla textual para a forma canônica.
//
// Devolve "" quando a entrada está vazia ou não representa uma única tecla.
// A tabela completa está documentada em [Resolver] e no README do módulo.
func normalizeKey(key string) string {
	if key == " " {
		return "space"
	}
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return ""
	}

	mods, final, ok := parseKey(trimmed)
	if !ok || final == "" {
		return ""
	}
	if len(mods) == 0 {
		return final
	}

	sort.SliceStable(mods, func(i, j int) bool {
		return modifierRank[mods[i]] < modifierRank[mods[j]]
	})
	return strings.Join(mods, "+") + "+" + final
}

// parseKey separa modificadores e a tecla final, aceitando as notações
// "ctrl+shift+a" e "C-M-a".
func parseKey(s string) (mods []string, final string, ok bool) {
	if mod, rest, found := emacsModifier(s); found {
		restMods, restFinal, restOK := parseKey(rest)
		if !restOK {
			return nil, "", false
		}
		return appendModifier(restMods, mod), restFinal, true
	}

	if strings.Contains(s, "+") {
		parts := splitOnPlus(s)
		if len(parts) == 1 {
			final := canonicalFinal(parts[0], false)
			return nil, final, final != ""
		}
		for _, part := range parts[:len(parts)-1] {
			mod, found := modifierName(part)
			if !found {
				return nil, "", false
			}
			mods = appendModifier(mods, mod)
		}
		final := canonicalFinal(parts[len(parts)-1], len(mods) > 0)
		if final == "" {
			return nil, "", false
		}
		return mods, final, true
	}

	final = canonicalFinal(s, false)
	return nil, final, final != ""
}

// splitOnPlus divide a tecla em segmentos, preservando um "+" literal no
// final ("ctrl++" vira ["ctrl", "+"] e "+" continua sendo ["+"]).
func splitOnPlus(s string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '+' && i != len(s)-1 {
			parts = append(parts, strings.TrimSpace(s[start:i]))
			start = i + 1
		}
	}
	parts = append(parts, strings.TrimSpace(s[start:]))
	return parts
}

// emacsModifier reconhece o prefixo "C-", "M-", "S-", "A-" ou "H-".
func emacsModifier(s string) (mod, rest string, ok bool) {
	if len(s) < 3 || s[1] != '-' {
		return "", "", false
	}
	switch s[0] {
	case 'C', 'c':
		return "ctrl", s[2:], true
	case 'M', 'm':
		return "meta", s[2:], true
	case 'S', 's':
		return "shift", s[2:], true
	case 'A', 'a':
		return "alt", s[2:], true
	case 'H', 'h':
		return "hyper", s[2:], true
	default:
		return "", "", false
	}
}

// modifierName traduz o nome de um modificador, aceitando abreviações.
func modifierName(part string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(part)) {
	case "ctrl", "control", "c":
		return "ctrl", true
	case "alt", "opt", "option", "a":
		return "alt", true
	case "shift", "s":
		return "shift", true
	case "meta", "m":
		return "meta", true
	case "hyper", "h":
		return "hyper", true
	case "super", "cmd", "command", "win":
		return "super", true
	default:
		return "", false
	}
}

// appendModifier adiciona um modificador sem duplicá-lo.
func appendModifier(mods []string, mod string) []string {
	for _, existing := range mods {
		if existing == mod {
			return mods
		}
	}
	return append(mods, mod)
}

// canonicalFinal normaliza a tecla final. Nomes de teclas especiais viram
// minúsculas; um caractere isolado preserva maiúsculas/minúsculas, exceto
// quando há modificador (aí a caixa não distingue a tecla).
func canonicalFinal(s string, hasMods bool) string {
	if s == "" {
		return ""
	}

	lower := strings.ToLower(s)
	if canonical, found := keyAliases[lower]; found {
		return canonical
	}
	if _, found := namedKeys[lower]; found {
		return lower
	}

	runes := []rune(s)
	if len(runes) == 1 {
		r := runes[0]
		if r == ' ' {
			return "space"
		}
		if hasMods && unicode.IsLetter(r) {
			return string(unicode.ToLower(r))
		}
		return s
	}

	return lower
}
