# charm-keys

Context-scoped keybindings for Bubble Tea: a global → screen → modal priority
stack, conflict detection, and help rendering.

> [Leia em Português](README.pt-br.md)

## Install

```bash
go get github.com/junhinhow/charm-keys
```

## Usage

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
			keys.Binding{Action: "quit", Keys: []string{"q", "ctrl+c"}, Help: "quit"},
			keys.Binding{Action: "help", Keys: []string{"?"}, Help: "help"},
		).
		Define("list",
			keys.Binding{Action: "open", Keys: []string{"enter"}, Help: "open"},
			keys.Binding{Action: "delete", Keys: []string{"d"}, Help: "delete"},
		).
		Define("modal:confirm",
			keys.Binding{Action: "confirm", Keys: []string{"enter"}, Help: "confirm"},
			keys.Binding{Action: "cancel", Keys: []string{"esc"}, Help: "cancel"},
		)

	for _, conflict := range resolver.Conflicts() {
		fmt.Printf("conflict: %s in %s (%s x %s)\n",
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

`Enter` opens the modal and from then on resolves to `confirm`; `esc` closes
it. `d` still resolves to the list action because the modal does not define it.

## API

| Symbol | Description |
|--------|-------------|
| `New() *Resolver` | Creates an empty resolver |
| `(*Resolver).Global(bindings ...Binding) *Resolver` | Defines the always-active base context |
| `(*Resolver).Define(name string, bindings ...Binding) *Resolver` | Registers or replaces a named context |
| `(*Resolver).Push(name string) *Resolver` | Activates a context with the highest priority (idempotent) |
| `(*Resolver).Pop() string` | Deactivates the top context; `""` when empty |
| `(*Resolver).Active() []string` | Active contexts, lowest to highest priority |
| `(*Resolver).Resolve(msg tea.KeyPressMsg) (string, bool)` | Resolves a Bubble Tea key message |
| `(*Resolver).ResolveString(key string) (string, bool)` | Resolves a textual key such as `"C-c"` |
| `(*Resolver).Conflicts() []Conflict` | Reports actions fighting over the same key in the same context |
| `(*Resolver).Help() string` | Effective help, one binding per line |
| `(*Resolver).HelpWidth(width int) string` | Effective help laid out in columns |
| `FromBubbles(key.Binding) Binding` | Converts a `bubbles/key` binding |
| `(Binding).ToBubbles() key.Binding` | Converts back to a `bubbles/key` binding |
| `(Binding).Matches(key string) bool` | Reports whether a textual key belongs to the binding |

Types: `Binding{Action, Keys, Help}`, `Context{Name, Bindings}` and
`Conflict{Context, Key, ActionA, ActionB}`.

## Design notes

### Precedence

Lookup walks the stack from the top down and stops at the global context. A key
defined in a modal never triggers the screen action underneath it, even when
both bind the same key.

`Push` is idempotent and `Pop` on an empty stack returns `""`. The global
context is never popped. `Active` returns `"global"` first when it was defined,
followed by the stack in activation order.

### Key normalization

| Input | Canonical |
|-------|-----------|
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

Modifier order is always `ctrl`, `alt`, `shift`, `meta`, `hyper`, `super`.
Modifier names are case-insensitive, as are special key names. A single
character without modifiers keeps its case (`A` and `a` are different keys);
with a modifier, case is irrelevant.

### Conflicts

`Conflicts` checks each context in isolation: two actions using the same key in
the same context produce a `Conflict`. The same key in different contexts is a
deliberate override and is not reported. The check covers every defined
context, active or not, and the output order is deterministic (global first,
then definition order). Run it at boot.

When a context repeats a key, resolution keeps the first binding and the
conflict is still reported.

### Help

`Help` shows one entry per action, taken from the highest-priority context
that defines it. Actions are deduplicated and keys shadowed by a more specific
context are omitted, so every displayed key really resolves to the action next
to it. An action keeps a single entry even if lower contexts bind extra keys
to it.

`HelpWidth` lays the same entries out in columns using display width (via
`ansi.StringWidth`). No line exceeds `width`; overflowing text is truncated
with `ansi.Truncate`, preserving ANSI sequences and wide characters. Widths
`<= 0` return `""`.

### Limits

- The resolver is safe for concurrent reads after configuration, but
  `Define`, `Global`, `Push` and `Pop` are not synchronized.
- Bindings have no enabled/disabled flag. `FromBubbles` maps a disabled
  `key.Binding` to an empty, ignored `Binding`.
- `Action` cannot survive `ToBubbles` → `FromBubbles` when it differs from
  `Help`: `bubbles/key` has no action field.
- Normalization works on the textual representation produced by Bubble Tea; raw
  byte sequences are not interpreted.
- Empty context names are ignored by `Define` and `Push`. Pushing a context
  before defining it is allowed; it resolves nothing until `Define` is called.

## License

MIT
