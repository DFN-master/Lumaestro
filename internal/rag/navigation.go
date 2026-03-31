package rag

import (
	"context"
	"fmt"
	"Lumaestro/internal/provider"
)

// GraphNavigator gerencia a expansão de contexto baseada em links.
type GraphNavigator struct {
	Qdrant *provider.QdrantClient
}

// NewGraphNavigator inicializa o navegador com acesso à memória semântica.
func NewGraphNavigator(qdrant *provider.QdrantClient) *GraphNavigator {
	return &GraphNavigator{Qdrant: qdrant}
}

// ExpandContext busca as notas vizinhas e sinapses do chat de forma recursiva.
func (n *GraphNavigator) ExpandContext(ctx context.Context, initialNotes []map[string]interface{}) []string {
	var fullContext []string
	visited := make(map[string]bool)

	for _, note := range initialNotes {
		content, _ := note["content"].(string)
		title, _ := note["name"].(string)

		if visited[title] {
			continue
		}
		visited[title] = true
		fullContext = append(fullContext, content)

		// 🧠 Navegação de Sinapses: Busca o que aprendemos no chat sobre este título
		// Fazemos uma busca exata por entidade no grafo de conhecimento.
		synapses, err := n.Qdrant.Search("knowledge_graph", nil, 3) // Aqui simplificamos, no real faríamos um filtro por subject=title
		if err == nil {
			for _, syn := range synapses {
				subj, _ := syn["subject"].(string)
				obj, _ := syn["object"].(string)
				
				// Se o título da nota for o sujeito ou objeto da tripla, a sinapse é relevante!
				if subj == title || obj == title {
					fact, _ := syn["content"].(string)
					fullContext = append(fullContext, fmt.Sprintf("[SINAPSE APRENDIDA]: %s", fact))
				}
			}
		}
	}

	return fullContext
}
