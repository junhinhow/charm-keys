package keys_test

import (
	"fmt"

	"github.com/junhinhow/charm-keys"
)

func Example() {
	resolver := keys.New().
		Global(keys.Binding{
			Action: "quit",
			Keys:   []string{"q", "ctrl+c"},
			Help:   "sair",
		}).
		Define("list", keys.Binding{
			Action: "open",
			Keys:   []string{"enter"},
			Help:   "abrir",
		}).
		Define("modal:confirm", keys.Binding{
			Action: "confirm",
			Keys:   []string{"enter"},
			Help:   "confirmar",
		}).
		Push("list")

	action, _ := resolver.ResolveString("enter")
	fmt.Println(action)

	resolver.Push("modal:confirm")
	action, _ = resolver.ResolveString("enter")
	fmt.Println(action)

	resolver.Pop()
	fmt.Println(resolver.Active())
	fmt.Println(resolver.Help())

	// Output:
	// open
	// confirm
	// [global list]
	// enter  abrir
	// q, ctrl+c  sair
}

func ExampleResolver_HelpWidth() {
	resolver := keys.New().Global(
		keys.Binding{Action: "quit", Keys: []string{"q"}, Help: "sair"},
		keys.Binding{Action: "help", Keys: []string{"?"}, Help: "ajuda"},
	)

	fmt.Println(resolver.HelpWidth(10))

	// Output:
	// q  sair
	// ?  ajuda
}
