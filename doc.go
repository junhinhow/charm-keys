// Package keys oferece keybindings com escopo por contexto, detecção de
// conflito e geração de ajuda para aplicações Bubble Tea.
//
// As teclas são resolvidas contra uma pilha de contextos: o topo (por exemplo,
// um modal) tem prioridade sobre as telas anteriores e sobre o contexto global.
// Assim, uma tecla definida no modal não dispara a ação da tela de baixo.
//
//	r := keys.New().
//		Global(keys.Binding{Action: "quit", Keys: []string{"q", "ctrl+c"}, Help: "sair"}).
//		Define("list", keys.Binding{Action: "open", Keys: []string{"enter"}, Help: "abrir"}).
//		Define("modal:confirm", keys.Binding{Action: "confirm", Keys: []string{"enter"}, Help: "confirmar"}).
//		Push("list")
//
//	action, ok := r.ResolveString("enter") // "open", true
//	r.Push("modal:confirm")
//	action, ok = r.ResolveString("enter") // "confirm", true
//
// # Normalização
//
// As teclas são comparadas por uma forma canônica. A tabela abaixo resume as
// equivalências (a caixa dos nomes de modificadores e de teclas especiais não
// importa):
//
//	entrada                         canônica
//	ctrl+c, C-c, ctrl+C, Control+C  ctrl+c
//	esc, escape, Esc                esc
//	enter, return, Enter            enter
//	space, spacebar, " "            space
//	del, delete                     delete
//	ins, insert                     insert
//	pgup, pageup, page-up           pgup
//	pgdown, pagedown, pgdn, page-down  pgdown
//	bs, backspace                   backspace
//	C-A-x, ctrl+alt+x, alt+ctrl+x   ctrl+alt+x
//	C-M-x, ctrl+meta+x              ctrl+meta+x
//	ç                               ç
//
// A ordem canônica dos modificadores é ctrl, alt, shift, meta, hyper, super.
// Um caractere isolado preserva a caixa quando não há modificador ("A" e "a"
// são teclas diferentes); com modificador, a caixa é irrelevante.
//
// # Conflitos
//
// Duas ações que usam a mesma tecla no mesmo contexto geram um [Conflict].
// A mesma tecla em contextos diferentes é um override proposital e não é
// reportada. Chame [Resolver.Conflicts] no boot para validar a configuração.
//
// # Limitações
//
// A normalização cobre as grafias textuais do Bubble Tea; sequências de bytes
// cruas não são interpretadas. O Resolver não é seguro para mutações
// concorrentes (Define, Global, Push, Pop) e não possui noção de bindings
// desabilitados — para isso, use [FromBubbles] (bindings desabilitados viram
// um Binding vazio, ignorado).
package keys
