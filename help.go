package keys

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// helpEntry é uma linha de ajuda já resolvida.
type helpEntry struct {
	keys string
	desc string
}

// line devolve o texto exibido para a entrada.
func (e helpEntry) line() string {
	return e.keys + "  " + e.desc
}

// Help gera o texto de ajuda do contexto ativo mais o global, uma entrada por
// linha.
//
// A ordem é da maior para a menor prioridade: contexto do topo da pilha, depois
// os anteriores e, por último, o global. Dentro de cada contexto vale a ordem
// de definição dos bindings. Uma ação já exibida não se repete e teclas
// sombreadas por um contexto superior não aparecem, então toda tecla exibida
// resolve para a ação ao lado dela.
func (r *Resolver) Help() string {
	entries := r.helpEntries()
	lines := make([]string, len(entries))
	for i, entry := range entries {
		lines[i] = entry.line()
	}
	return strings.Join(lines, "\n")
}

// HelpWidth gera o help em colunas, respeitando a largura máxima em células de
// terminal.
//
// Larguras menores ou iguais a zero devolvem "" (não há linha possível sem
// violar o limite). Nenhuma linha gerada ultrapassa width: o texto é truncado
// com [ansi.Truncate], que preserva sequências ANSI e caracteres largos.
func (r *Resolver) HelpWidth(width int) string {
	if r == nil || width <= 0 {
		return ""
	}

	entries := r.helpEntries()
	if len(entries) == 0 {
		return ""
	}

	labels := make([]string, len(entries))
	maxWidth := 0
	for i, entry := range entries {
		labels[i] = entry.line()
		if current := ansi.StringWidth(labels[i]); current > maxWidth {
			maxWidth = current
		}
	}

	columnWidth := maxWidth + 2
	if columnWidth > width {
		columnWidth = width
	}

	columns := width / columnWidth
	if columns < 1 {
		columns = 1
	}
	if columns > len(entries) {
		columns = len(entries)
	}
	rows := (len(entries) + columns - 1) / columns

	contentWidth := width
	if columns > 1 {
		contentWidth = columnWidth - 2
	}

	lines := make([]string, 0, rows)
	for row := 0; row < rows; row++ {
		var line strings.Builder
		for column := 0; column < columns; column++ {
			index := column*rows + row
			if index >= len(entries) {
				break
			}

			cell := ansi.Truncate(labels[index], contentWidth, "")
			line.WriteString(cell)

			if index+rows < len(entries) {
				padding := columnWidth - ansi.StringWidth(cell)
				if padding > 0 {
					line.WriteString(strings.Repeat(" ", padding))
				}
			}
		}

		trimmed := strings.TrimRight(line.String(), " ")
		lines = append(lines, ansi.Truncate(trimmed, width, ""))
	}
	return strings.Join(lines, "\n")
}

// helpEntries coleta as entradas efetivas na ordem de prioridade, sem repetir
// ações nem teclas sombreadas por contextos superiores.
func (r *Resolver) helpEntries() []helpEntry {
	if r == nil {
		return nil
	}

	// Mapa efetivo: a primeira ocorrência da tecla vence, de cima para baixo.
	effective := make(map[string]string)
	for _, context := range r.priorityContexts() {
		for key, action := range context.index {
			if _, exists := effective[key]; !exists {
				effective[key] = action
			}
		}
	}

	seenActions := make(map[string]struct{})
	var entries []helpEntry

	for _, context := range r.priorityContexts() {
		for _, binding := range context.Bindings {
			if _, exists := seenActions[binding.Action]; exists {
				continue
			}

			seenKeys := make(map[string]struct{})
			var keys []string
			for _, raw := range binding.Keys {
				normalized := normalizeKey(raw)
				if normalized == "" || effective[normalized] != binding.Action {
					continue
				}
				if _, exists := seenKeys[normalized]; exists {
					continue
				}
				seenKeys[normalized] = struct{}{}
				keys = append(keys, raw)
			}
			if len(keys) == 0 {
				continue
			}

			description := binding.Help
			if strings.TrimSpace(description) == "" {
				description = binding.Action
			}
			seenActions[binding.Action] = struct{}{}
			entries = append(entries, helpEntry{
				keys: strings.Join(keys, ", "),
				desc: description,
			})
		}
	}
	return entries
}

// priorityContexts devolve os contextos da maior para a menor prioridade:
// topo da pilha primeiro e o global por último.
func (r *Resolver) priorityContexts() []*Context {
	if r == nil {
		return nil
	}

	contexts := make([]*Context, 0, len(r.stack)+1)
	for i := len(r.stack) - 1; i >= 0; i-- {
		if context, exists := r.contexts[r.stack[i]]; exists {
			contexts = append(contexts, context)
		}
	}
	if r.global != nil {
		contexts = append(contexts, r.global)
	}
	return contexts
}
