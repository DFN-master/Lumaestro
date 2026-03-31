package rag

import (
	"context"
	"fmt"
	"hash/fnv"
	"time"

	"Lumaestro/internal/provider"
)

// KnowledgeWeaver é o "Tecelão de Conhecimento" que transforma conversas em sinapses.
type KnowledgeWeaver struct {
	Ontology  *provider.OntologyService
	Qdrant    *provider.QdrantClient
	Embedder  *provider.EmbeddingService
}

// NewKnowledgeWeaver inicializa o tecelão.
func NewKnowledgeWeaver(ontology *provider.OntologyService, qdrant *provider.QdrantClient, embedder *provider.EmbeddingService) *KnowledgeWeaver {
	return &KnowledgeWeaver{
		Ontology: ontology,
		Qdrant:   qdrant,
		Embedder: embedder,
	}
}

// WeaveChatKnowledge analisa o texto do chat, extrai fatos e os integra ao grafo.
func (w *KnowledgeWeaver) WeaveChatKnowledge(ctx context.Context, chatText string) error {
	// 1. Extração de Triplas (Sinapses)
	triples, err := w.Ontology.ExtractTriples(ctx, chatText)
	if err != nil {
		return fmt.Errorf("falha ao extrair sinapses: %w", err)
	}

	if len(triples) == 0 {
		return nil // Nada de relevante para aprender aqui
	}

	for _, t := range triples {
		// 2. Gerar Embedding para o fato (para busca semântica futura)
		factText := fmt.Sprintf("%s %s %s", t.Subject, t.Predicate, t.Object)
		vector, _ := w.Embedder.GenerateEmbedding(ctx, factText)

		// 3. Salvar no Qdrant (Coleção knowledge_graph)
		// Geramos um ID determinístico baseado na tripla para evitar duplicatas exatas
		h := fnv.New64a()
		h.Write([]byte(factText))
		id := h.Sum64()

		payload := map[string]interface{}{
			"subject":   t.Subject,
			"predicate": t.Predicate,
			"object":    t.Object,
			"source":    "chat_memory",
			"timestamp": time.Now().Format(time.RFC3339),
			"content":   factText,
		}

		err := w.Qdrant.UpsertPoint("knowledge_graph", id, vector, payload)
		if err != nil {
			fmt.Printf("[KnowledgeWeaver] Erro ao salvar sinapse: %v\n", err)
			continue
		}
		
		fmt.Printf("[KnowledgeWeaver] Sinapse Criada: %s\n", factText)
	}

	return nil
}
