package rag

import (
	"fmt"
	"math"
	"sync"

	"gonum.org/v1/gonum/graph/network"
	"gonum.org/v1/gonum/graph/simple"
)

// Node representa um ponto no conhecimento (Nota, Agente, Task)
type Node struct {
	ID   string
	Name string
	Type string // "note", "agent", "source"
}

// Edge representa uma conexão semântica
type Edge struct {
	Source string
	Target string
	Weight float64
	Label  string
}

// GraphEngine é o cérebro relacional nativo do Lumaestro (V20).
// Utiliza a Gonum para matemática pesada e Adjacência nativa para velocidade de navegação.
type GraphEngine struct {
	mu           sync.RWMutex
	nodes        map[string]*Node
	adj          map[string]map[string]float64 // Lista de Adjacência em RAM com pesos
	pageRank     map[string]float64
	
	// Gonum integration for heavy math
	gonumGraph   *simple.DirectedGraph
	nodeToGonum  map[string]int64
	gonumToNode  map[int64]string
	
	// 🏷️ Predicados (Rótulos) - Armazena a relação semântica entre pares
	labels       map[string]string
}

// NewGraphEngine inicializa o motor de elite.
func NewGraphEngine() *GraphEngine {
	return &GraphEngine{
		nodes:       make(map[string]*Node),
		adj:         make(map[string]map[string]float64),
		pageRank:    make(map[string]float64),
		gonumGraph:  simple.NewDirectedGraph(0, math.Inf(1)),
		nodeToGonum: make(map[string]int64),
		gonumToNode: make(map[int64]string),
		labels:      make(map[string]string),
	}
}

// AddNode insere ou atualiza um nó no grafo.
func (ge *GraphEngine) AddNode(id, name, docType string) {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	if _, exists := ge.nodes[id]; !exists {
		ge.nodes[id] = &Node{ID: id, Name: name, Type: docType}
		
		// Map for Gonum
		gID := int64(len(ge.nodeToGonum))
		ge.nodeToGonum[id] = gID
		ge.gonumToNode[gID] = id
		ge.gonumGraph.AddNode(simple.Node(gID))
	}
}

// AddEdge cria uma sinapse com peso e rótulo (Explicação semântica).
func (ge *GraphEngine) AddEdge(sourceID, targetID string, weight float64, label string) {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	// Garante que os nós existem (Sem recursão de trava)
	if _, ok := ge.nodes[sourceID]; !ok {
		gID := int64(len(ge.nodeToGonum))
		ge.nodeToGonum[sourceID] = gID
		ge.gonumToNode[gID] = sourceID
		ge.nodes[sourceID] = &Node{ID: sourceID, Name: sourceID, Type: "unknown"}
		ge.gonumGraph.AddNode(simple.Node(gID))
	}
	if _, ok := ge.nodes[targetID]; !ok {
		gID := int64(len(ge.nodeToGonum))
		ge.nodeToGonum[targetID] = gID
		ge.gonumToNode[gID] = targetID
		ge.nodes[targetID] = &Node{ID: targetID, Name: targetID, Type: "unknown"}
		ge.gonumGraph.AddNode(simple.Node(gID))
	}

	// Adjacência Interna com Suporte a Predicados (Rótulos)
	if _, ok := ge.adj[sourceID]; !ok {
		ge.adj[sourceID] = make(map[string]float64)
	}
	ge.adj[sourceID][targetID] = weight
	ge.labels[fmt.Sprintf("%s-%s", sourceID, targetID)] = label

	// Aresta Gonum (Gonum não suporta labels em Edges simples, mas mantemos o mapeamento interno)
	u := ge.nodeToGonum[sourceID]
	v := ge.nodeToGonum[targetID]
	ge.gonumGraph.SetEdge(simple.Edge{
		F: simple.Node(u),
		T: simple.Node(v),
		W: weight,
	})
}

// ComputePageRank calcula a autoridade das notas (Algoritmo de Elite Google).
func (ge *GraphEngine) ComputePageRank() {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	if len(ge.nodes) == 0 {
		return
	}

	// PageRank (Power Iteration via Gonum)
	results := network.PageRank(ge.gonumGraph, 0.85, 1e-6)
	
	for gID, rank := range results {
		id := ge.gonumToNode[int64(gID)]
		ge.pageRank[id] = rank
	}
	
	fmt.Printf("[GraphEngine] 🧠 PageRank recalculado para %d nós\n", len(ge.pageRank))
}

// GetRank retorna o peso de autoridade de um nó.
func (ge *GraphEngine) GetRank(id string) float64 {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	return ge.pageRank[id]
}

// BFS realiza uma busca por camadas para expansão de contexto e iluminação de mapa.
func (ge *GraphEngine) BFS(startID string, maxDepth int) []string {
	ge.mu.RLock()
	defer ge.mu.RUnlock()

	visited := make(map[string]bool)
	queue := []string{startID}
	depth := make(map[string]int)
	
	var result []string
	visited[startID] = true
	depth[startID] = 0

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if depth[curr] > maxDepth {
			continue
		}

		result = append(result, curr)

		for neighbor := range ge.adj[curr] {
			if !visited[neighbor] {
				visited[neighbor] = true
				depth[neighbor] = depth[curr] + 1
				queue = append(queue, neighbor)
			}
		}
	}

	return result
}

// Clear limpa o motor em RAM.
func (ge *GraphEngine) Clear() {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	
	ge.nodes = make(map[string]*Node)
	ge.adj = make(map[string]map[string]float64)
	ge.pageRank = make(map[string]float64)
	ge.gonumGraph = simple.NewDirectedGraph(0, math.Inf(1))
	ge.nodeToGonum = make(map[string]int64)
	ge.gonumToNode = make(map[int64]string)
}

// Prune remove nós cuja autoridade (PageRank) esteja abaixo do nível de sobrevivência.
func (ge *GraphEngine) Prune(threshold float64) []string {
	ge.mu.Lock()
	defer ge.mu.Unlock()

	var removed []string
	for id, rank := range ge.pageRank {
		if rank < threshold {
			// Não removemos nós do tipo 'system' ou 'source' (Markdown direto) a menos que explicitamente solicitado.
			// Focamos em podar 'memory' (Chat) e 'unknown' (pontas soltas).
			node, ok := ge.nodes[id]
			if !ok || (node.Type != "memory" && node.Type != "unknown") {
				continue
			}

			// Remove do Mapa de Nós
			delete(ge.nodes, id)
			// Remove das Adjacências (Entrada e Saída)
			delete(ge.adj, id)
			for source := range ge.adj {
				delete(ge.adj[source], id)
			}
			// Remove do PageRank
			delete(ge.pageRank, id)
			
			// Remove do Gonum
			if gID, ok := ge.nodeToGonum[id]; ok {
				ge.gonumGraph.RemoveNode(simple.Node(gID))
				delete(ge.gonumToNode, gID)
				delete(ge.nodeToGonum, id)
			}

			removed = append(removed, id)
		}
	}

	if len(removed) > 0 {
		fmt.Printf("[GraphEngine] 🧹 Poda Neural: %d nós irrelevantes removidos do Córtex.\n", len(removed))
	}
	return removed
}
