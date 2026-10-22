package keys

// Conflict descreve duas ações que disputam a mesma tecla no mesmo contexto.
//
// O mesmo par de teclas em contextos diferentes não é um conflito: é um
// override proposital, resolvido pela precedência da pilha.
type Conflict struct {
	Context string
	Key     string
	ActionA string
	ActionB string
}

// Conflicts devolve todos os conflitos entre ações do mesmo contexto.
//
// A verificação cobre o contexto global e todos os contextos definidos, mesmo
// os que não estão ativos no momento. A ordem é determinística: global
// primeiro, depois a ordem de definição; dentro de cada contexto, a ordem dos
// bindings e das teclas. A mesma ação repetindo uma tecla não gera conflito.
//
// Deve ser chamado no boot, antes de o usuário interagir com a aplicação.
func (r *Resolver) Conflicts() []Conflict {
	if r == nil {
		return nil
	}

	var conflicts []Conflict
	if r.global != nil {
		conflicts = append(conflicts, contextConflicts(r.global)...)
	}
	for _, name := range r.order {
		if name == globalContext {
			continue
		}
		if context, exists := r.contexts[name]; exists {
			conflicts = append(conflicts, contextConflicts(context)...)
		}
	}
	return conflicts
}

// contextConflicts compara as teclas de um único contexto.
func contextConflicts(context *Context) []Conflict {
	var conflicts []Conflict
	seen := make(map[string]string)

	for _, binding := range context.Bindings {
		for _, raw := range binding.Keys {
			key := normalizeKey(raw)
			if key == "" {
				continue
			}
			if first, exists := seen[key]; exists {
				if first != binding.Action {
					conflicts = append(conflicts, Conflict{
						Context: context.Name,
						Key:     key,
						ActionA: first,
						ActionB: binding.Action,
					})
				}
				continue
			}
			seen[key] = binding.Action
		}
	}
	return conflicts
}
