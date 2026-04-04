package rag

import (
	"testing"
)

func TestGraphEngine_PageRank(t *testing.T) {
	ge := NewGraphEngine()

	// Criar um grafo de teste: A -> B -> C -> A (Ciclo)
	ge.AddNode("A", "Nota A", "note")
	ge.AddNode("B", "Nota B", "note")
	ge.AddNode("C", "Nota C", "note")

	ge.AddEdge("A", "B", 1.0, "test")
	ge.AddEdge("B", "C", 1.0, "test")
	ge.AddEdge("C", "A", 1.0, "test")

	ge.ComputePageRank()

	rankA := ge.GetRank("A")
	rankB := ge.GetRank("B")
	rankC := ge.GetRank("C")

	if rankA <= 0 || rankB <= 0 || rankC <= 0 {
		t.Errorf("PageRank deve ser positivo, obtido A: %v, B: %v, C: %v", rankA, rankB, rankC)
	}

	// Em um ciclo perfeito, os ranks devem ser aproximadamente iguais
	if mathAbs(rankA-rankB) > 0.01 {
		t.Errorf("Ranks do ciclo devem ser iguais, obtido A: %v, B: %v", rankA, rankB)
	}
}

func TestGraphEngine_BFS(t *testing.T) {
	ge := NewGraphEngine()

	ge.AddNode("A", "Nota A", "note")
	ge.AddNode("B", "Nota B", "note")
	ge.AddNode("C", "Nota C", "note")
	ge.AddNode("D", "Nota D", "note")

	ge.AddEdge("A", "B", 1.0, "test")
	ge.AddEdge("B", "C", 1.0, "test")
	ge.AddEdge("C", "D", 1.0, "test")

	// BFS de A com profundidade 2 deve pegar A, B e C (mas não D)
	res := ge.BFS("A", 2)
	
	foundD := false
	for _, id := range res {
		if id == "D" {
			foundD = true
		}
	}

	if foundD {
		t.Errorf("BFS com profundidade 2 não deveria encontrar D")
	}

	if len(res) != 3 {
		t.Errorf("BFS deveria encontrar 3 nós, encontrou %d", len(res))
	}
}

func mathAbs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
