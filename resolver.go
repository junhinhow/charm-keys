package keys

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// globalContext é o nome reservado do contexto base, sempre ativo.
const globalContext = "global"

// Resolver resolve teclas considerando uma pilha de contextos.
//
// A precedência é do topo da pilha para baixo: modal > tela > ... > global.
// Uma tecla encontrada em um contexto de maior prioridade nunca chega aos
// contextos inferiores.
//
// O Resolver não é seguro para uso concorrente durante Define, Global, Push
// ou Pop. Depois de montado, Resolve, ResolveString, Conflicts, Help e
// HelpWidth podem ser chamados de várias goroutines.
type Resolver struct {
	global   *Context
	contexts map[string]*Context
	order    []string
	stack    []string
}

// New cria um Resolver vazio, sem contexto global e sem contextos ativos.
func New() *Resolver {
	return &Resolver{contexts: make(map[string]*Context)}
}

// Global define o contexto base, que fica sempre ativo e tem a menor
// prioridade. Chamar Global de novo substitui os bindings anteriores.
//
// O nome "global" é reservado: Define("global", ...) tem o mesmo efeito.
func (r *Resolver) Global(bindings ...Binding) *Resolver {
	if r == nil {
		return r
	}
	r.global = newContext(globalContext, bindings)
	return r
}

// Define registra ou atualiza um contexto nomeado.
//
// Nomes vazios (ou apenas espaços) são ignorados. Definir um contexto que já
// existe substitui seus bindings, mantendo a posição original na ordem de
// definição. Bindings com Action vazia ou sem nenhuma tecla válida são
// descartados.
func (r *Resolver) Define(name string, bindings ...Binding) *Resolver {
	if r == nil {
		return r
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return r
	}
	if name == globalContext {
		return r.Global(bindings...)
	}
	if _, exists := r.contexts[name]; !exists {
		r.order = append(r.order, name)
	}
	r.contexts[name] = newContext(name, bindings)
	return r
}

// Push ativa um contexto, dando a ele a maior prioridade.
//
// É idempotente: empilhar um contexto já ativo não altera a pilha. Um contexto
// ainda não definido pode ser empilhado e passa a resolver teclas assim que
// Define for chamado. O nome "global" é ignorado, pois o contexto base já está
// sempre ativo.
func (r *Resolver) Push(name string) *Resolver {
	if r == nil {
		return r
	}
	name = strings.TrimSpace(name)
	if name == "" || name == globalContext {
		return r
	}
	for _, active := range r.stack {
		if active == name {
			return r
		}
	}
	r.stack = append(r.stack, name)
	return r
}

// Pop desativa o contexto do topo e devolve seu nome.
//
// Com a pilha vazia devolve "" e não entra em panic. O contexto global não é
// afetado.
func (r *Resolver) Pop() string {
	if r == nil || len(r.stack) == 0 {
		return ""
	}
	top := r.stack[len(r.stack)-1]
	r.stack = r.stack[:len(r.stack)-1]
	return top
}

// Active devolve os contextos ativos, do menor para o maior prioridade.
//
// "global" aparece primeiro quando foi definido; em seguida vêm os contextos
// empilhados na ordem em que foram ativados. O slice devolvido é uma cópia.
func (r *Resolver) Active() []string {
	if r == nil {
		return nil
	}
	var active []string
	if r.global != nil {
		active = append(active, globalContext)
	}
	return append(active, r.stack...)
}

// Resolve devolve a ação correspondente à tecla e se houve match.
//
// A busca percorre a pilha do topo para baixo e termina no contexto global. A
// representação textual de [tea.KeyPressMsg.String] é tentada primeiro; se não
// houver match, a forma de [tea.KeyPressMsg.Keystroke] (com os modificadores)
// é tentada como fallback.
func (r *Resolver) Resolve(msg tea.KeyPressMsg) (action string, ok bool) {
	if r == nil {
		return "", false
	}
	if action, found := r.resolveNormalized(normalizeKey(msg.String())); found {
		return action, true
	}

	keystroke := msg.Keystroke()
	if keystroke == msg.String() {
		return "", false
	}
	return r.resolveNormalized(normalizeKey(keystroke))
}

// ResolveString aceita a representação textual da tecla ("ctrl+c", "C-c",
// "esc", "enter", "a") e devolve a ação e se houve match.
func (r *Resolver) ResolveString(key string) (action string, ok bool) {
	if r == nil {
		return "", false
	}
	return r.resolveNormalized(normalizeKey(key))
}

// resolveNormalized procura uma tecla já normalizada na pilha de contextos.
func (r *Resolver) resolveNormalized(key string) (string, bool) {
	if key == "" {
		return "", false
	}
	for i := len(r.stack) - 1; i >= 0; i-- {
		context, exists := r.contexts[r.stack[i]]
		if !exists {
			continue
		}
		if action, found := context.index[key]; found {
			return action, true
		}
	}
	if r.global != nil {
		if action, found := r.global.index[key]; found {
			return action, true
		}
	}
	return "", false
}

// newContext copia os bindings, normaliza as teclas e monta o índice de
// resolução. A primeira ação registrada para uma tecla vence.
func newContext(name string, bindings []Binding) *Context {
	context := &Context{
		Name:  name,
		index: make(map[string]string),
	}
	for _, binding := range bindings {
		if strings.TrimSpace(binding.Action) == "" {
			continue
		}

		keys := make([]string, 0, len(binding.Keys))
		for _, raw := range binding.Keys {
			normalized := normalizeKey(raw)
			if normalized == "" {
				continue
			}
			keys = append(keys, raw)
			if _, exists := context.index[normalized]; !exists {
				context.index[normalized] = binding.Action
			}
		}
		if len(keys) == 0 {
			continue
		}

		context.Bindings = append(context.Bindings, Binding{
			Action: binding.Action,
			Keys:   keys,
			Help:   binding.Help,
		})
	}
	return context
}
