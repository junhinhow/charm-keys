# charm-keys

Keybindings com escopo por contexto para Bubble Tea: pilha de prioridade
global → tela → modal, detecção de conflito e geração de ajuda.

> [Read in English](README.md)

## Instalação

```bash
go get github.com/junhinhow/charm-keys
```

## Uso

```go
package main

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/junhinhow/charm-keys"
)

type model struct {
	keys *keys.Resolver
}

func newModel() model {
	resolver := keys.New().
		Global(
			keys.Binding{Action: "quit", Keys: []string{"q", "ctrl+c"}, Help: "sair"},
			keys.Binding{Action: "help", Keys: []string{"?"}, Help: "ajuda"},
		).
		Define("list",
			keys.Binding{Action: "open", Keys: []string{"enter"}, Help: "abrir"},
			keys.Binding{Action: "delete", Keys: []string{"d"}, Help: "apagar"},
		).
		Define("modal:confirm",
			keys.Binding{Action: "confirm", Keys: []string{"enter"}, Help: "confirmar"},
			keys.Binding{Action: "cancel", Keys: []string{"esc"}, Help: "cancelar"},
		)

	for _, conflict := range resolver.Conflicts() {
		fmt.Printf("conflito: %s em %s (%s x %s)\n",
			conflict.Key, conflict.Context, conflict.ActionA, conflict.ActionB)
	}

	return model{keys: resolver.Push("list")}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if press, ok := msg.(tea.KeyPressMsg); ok {
		action, matched := m.keys.Resolve(press)
		if matched {
			switch action {
			case "quit":
				return m, tea.Quit
			case "open":
				m.keys.Push("modal:confirm")
			case "confirm", "cancel":
				m.keys.Pop()
			}
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	return tea.NewView(m.keys.Help())
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Println(err)
	}
}
```

`Enter` abre o modal e, a partir daí, resolve para `confirm`; `esc` o fecha.
`d` continua resolvendo para a ação da lista porque o modal não a define.

## API

| Símbolo | Descrição |
|---------|-----------|
| `New() *Resolver` | Cria um resolver vazio |
| `(*Resolver).Global(bindings ...Binding) *Resolver` | Define o contexto base sempre ativo |
| `(*Resolver).Define(name string, bindings ...Binding) *Resolver` | Registra ou substitui um contexto nomeado |
| `(*Resolver).Push(name string) *Resolver` | Ativa um contexto com a maior prioridade (idempotente) |
| `(*Resolver).Pop() string` | Desativa o contexto do topo; `""` quando vazio |
| `(*Resolver).Active() []string` | Contextos ativos, do menor para o maior prioridade |
| `(*Resolver).Resolve(msg tea.KeyPressMsg) (string, bool)` | Resolve uma mensagem de tecla do Bubble Tea |
| `(*Resolver).ResolveString(key string) (string, bool)` | Resolve uma tecla textual como `"C-c"` |
| `(*Resolver).Conflicts() []Conflict` | Reporta ações disputando a mesma tecla no mesmo contexto |
| `(*Resolver).Help() string` | Ajuda efetiva, um binding por linha |
| `(*Resolver).HelpWidth(width int) string` | Ajuda efetiva distribuída em colunas |
| `FromBubbles(key.Binding) Binding` | Converte um binding de `bubbles/key` |
| `(Binding).ToBubbles() key.Binding` | Converte de volta para `bubbles/key` |
| `(Binding).Matches(key string) bool` | Informa se uma tecla textual pertence ao binding |

Tipos: `Binding{Action, Keys, Help}`, `Context{Name, Bindings}` e
`Conflict{Context, Key, ActionA, ActionB}`.

## Notas de design

### Precedência

A busca percorre a pilha do topo para baixo e termina no contexto global. Uma
tecla definida no modal nunca dispara a ação da tela de baixo, mesmo quando os
dois contextos usam a mesma tecla.

`Push` é idempotente e `Pop` com a pilha vazia devolve `""`. O contexto global
nunca é removido. `Active` devolve `"global"` primeiro quando ele foi definido,
seguido pela pilha na ordem de ativação.

### Normalização de teclas

| Entrada | Canônica |
|---------|----------|
| `ctrl+c`, `C-c`, `ctrl+C`, `Control+C` | `ctrl+c` |
| `esc`, `escape`, `Esc` | `esc` |
| `enter`, `return`, `Enter` | `enter` |
| `space`, `spacebar`, `" "` | `space` |
| `delete`, `del` | `delete` |
| `insert`, `ins` | `insert` |
| `backspace`, `bs` | `backspace` |
| `pgup`, `pageup`, `page-up` | `pgup` |
| `pgdown`, `pagedown`, `pgdn`, `page-down` | `pgdown` |
| `C-A-x`, `ctrl+alt+x`, `alt+ctrl+x` | `ctrl+alt+x` |
| `C-M-x`, `ctrl+meta+x` | `ctrl+meta+x` |
| `ç` | `ç` |

A ordem dos modificadores é sempre `ctrl`, `alt`, `shift`, `meta`, `hyper`,
`super`. Nomes de modificadores e de teclas especiais não diferenciam
maiúsculas de minúsculas. Um caractere isolado sem modificador preserva a
caixa (`A` e `a` são teclas diferentes); com modificador, a caixa é
irrelevante.

### Conflitos

`Conflicts` verifica cada contexto isoladamente: duas ações usando a mesma
tecla no mesmo contexto geram um `Conflict`. A mesma tecla em contextos
diferentes é um override proposital e não é reportada. A verificação cobre
todos os contextos definidos, ativos ou não, e a ordem da saída é determinística
(global primeiro, depois a ordem de definição). Chame no boot.

Quando um contexto repete uma tecla, a resolução mantém o primeiro binding e o
conflito ainda é reportado.

### Ajuda

`Help` mostra uma entrada por ação, tirada do contexto de maior prioridade que
a define. Ações são deduplicadas e teclas sombreadas por um contexto mais
específico são omitidas, então toda tecla exibida realmente resolve para a ação
ao lado dela. Uma ação mantém uma única entrada mesmo que contextos inferiores
associem outras teclas a ela.

`HelpWidth` distribui as mesmas entradas em colunas usando largura de exibição
(via `ansi.StringWidth`). Nenhuma linha ultrapassa `width`; o texto excedente é
truncado com `ansi.Truncate`, preservando sequências ANSI e caracteres largos.
Larguras `<= 0` devolvem `""`.

### Limitações

- O resolver é seguro para leituras concorrentes depois de configurado, mas
  `Define`, `Global`, `Push` e `Pop` não são sincronizados.
- Bindings não têm flag de habilitado/desabilitado. `FromBubbles` converte um
  `key.Binding` desabilitado em um `Binding` vazio, que é ignorado.
- `Action` não sobrevive a `ToBubbles` → `FromBubbles` quando difere de `Help`:
  `bubbles/key` não tem campo de ação.
- A normalização trabalha sobre a representação textual produzida pelo Bubble
  Tea; sequências cruas de bytes não são interpretadas.
- Nomes de contexto vazios são ignorados por `Define` e `Push`. É permitido
  empilhar um contexto antes de defini-lo; ele não resolve nada até `Define`.

## Licença

MIT
