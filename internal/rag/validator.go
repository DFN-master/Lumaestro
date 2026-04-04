package rag

import (
	"Lumaestro/internal/lightning"
	"fmt"
	"strings"
)

// AgentValidator é o auditor lógico do enxame Lumaestro.
type AgentValidator struct {
	Store  *lightning.DuckDBStore
	Graph  *rag.GraphEngine
}

// NewAgentValidator cria uma nova instância do juiz neural.
func NewAgentValidator(store *lightning.DuckDBStore, graph *rag.GraphEngine) *AgentValidator {
	return &AgentValidator{
		Store: store,
		Graph: graph,
	}
}

// Conflict representa uma contradição lógica detectada entre dois nós.
type Conflict struct {
	SubjectID string
	ObjectID  string
	LogicA    string
	LogicB    string
	SourceA   string // Path ou Nota
	SourceB   string
}

// AuditKnowledge escania o banco de dados em busca de triplas contraditórias.
func (v *AgentValidator) AuditKnowledge() ([]Conflict, error) {
	if v.Store == nil {
		return nil, fmt.Errorf("repositório analítico offline")
	}

	// Busca por 'Predicados Opostos' entre os mesmos sujeitos e objetos
	// Ex: (A) --"define"-- (B)  VS  (A) --"contradiz"-- (B)
	query := `
		SELECT e1.source_id, e1.target_id, e1.relation_type as relA, e2.relation_type as relB
		FROM graph_edges e1
		JOIN graph_edges e2 ON e1.source_id = e2.source_id AND e1.target_id = e2.target_id
		WHERE e1.relation_type < e2.relation_type
		AND (
			(e1.relation_type = 'is' AND e2.relation_type = 'is_not') OR
			(e1.relation_type = 'defines' AND e2.relation_type = 'refutes') OR
			(e1.relation_type = 'mentions' AND e2.relation_type = 'ignores')
		)
	`
	rows, err := v.Store.GetDB().Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conflicts []Conflict
	for rows.Next() {
		var s, t, ra, rb string
		if err := rows.Scan(&s, &t, &ra, &rb); err == nil {
			conflicts = append(conflicts, Conflict{
				SubjectID: s,
				ObjectID:  t,
				LogicA:    ra,
				LogicB:    rb,
			})
		}
	}

	return conflicts, nil
}

// ResolveConflict simula uma decisão do juiz (IA) para limpar o grafo.
func (v *AgentValidator) ResolveConflict(c Conflict) string {
    // [LOGICA FUTURA]: Usar o LLM para decidir com base no timestamp ou autoridade
	return fmt.Sprintf("🛡️ Conflito Detectado: %s em dúvida sobre %s (%s vs %s)", c.SubjectID, c.ObjectID, c.LogicA, c.LogicB)
}
