package neural

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
)

// WeightRegistry contém o estado aprendido da rede neural local.
type WeightRegistry struct {
	Weights      map[string]float32 `json:"weights"`       // ID do No -> Peso Neural
	Exploration  bool               `json:"exploration"`   // Se true, ignora os pesos (Modo Exploração)
	LearningRate float32            `json:"learning_rate"` // Taxa de ajuste (eta)
}

// Ranker gerencia o ciclo de vida do aprendizado ativo.
type Ranker struct {
	registry     *WeightRegistry
	filePath     string
	mu           sync.RWMutex
	maxWeight    float32
	minWeight    float32
}

// NewRanker inicializa o motor neural com persistência em disco.
func NewRanker() *Ranker {
	r := &Ranker{
		filePath:  ".context/neural_weights.json",
		maxWeight: 10.0,
		minWeight: 0.1,
		registry: &WeightRegistry{
			Weights:      make(map[string]float32),
			Exploration:  false,
			LearningRate: 0.05, // Reforço de 5% por clique
		},
	}
	r.load()
	return r
}

func (r *Ranker) load() {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := os.ReadFile(r.filePath)
	if err == nil {
		json.Unmarshal(data, r.registry)
	}

	// Garantir inicialização se o arquivo estiver vazio
	if r.registry.Weights == nil {
		r.registry.Weights = make(map[string]float32)
	}
	if r.registry.LearningRate == 0 {
		r.registry.LearningRate = 0.05
	}
}

func (r *Ranker) save() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	os.MkdirAll(filepath.Dir(r.filePath), 0755)
	data, _ := json.MarshalIndent(r.registry, "", "  ")
	os.WriteFile(r.filePath, data, 0644)
}

// Reinforce aplica um passo de treinamento positivo (reforço sináptico).
func (r *Ranker) Reinforce(nodeID string) {
	r.mu.Lock()
	
	// Inicializa peso base se não existir (1.0 = neutro)
	current, ok := r.registry.Weights[nodeID]
	if !ok {
		current = 1.0
	}

	// Regra de Aprendizado: Delta Rule simplificada
	// Novo peso cresce em direção ao MaxWeight de forma assintótica
	delta := r.registry.LearningRate * (r.maxWeight - current)
	r.registry.Weights[nodeID] = current + delta
	
	r.mu.Unlock()
	r.save()

	fmt.Printf("[Neural] 🧠 Reforço Sináptico em '%s': %.2f -> %.2f\n", nodeID, current, r.registry.Weights[nodeID])
}

// AdjustScore aplica a ativação aprendida ao score original do motor RAG.
func (r *Ranker) AdjustScore(nodeID string, originalScore float32) float32 {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.registry.Exploration {
		return originalScore
	}

	weight, ok := r.registry.Weights[nodeID]
	if !ok {
		return originalScore
	}

	// Combinação Neural: O peso aprendido escala o score original
	// Usamos log para evitar explosão de escala
	multiplier := float32(math.Log1p(float64(weight)))
	return originalScore * multiplier
}

// GetWeight retorna o peso atual para visualização no frontend.
func (r *Ranker) GetWeight(nodeID string) float32 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	if w, ok := r.registry.Weights[nodeID]; ok {
		return w
	}
	return 1.0
}

// SetExplorationMode ativa/desativa a influência da rede neural.
func (r *Ranker) SetExplorationMode(enabled bool) {
	r.mu.Lock()
	r.registry.Exploration = enabled
	r.mu.Unlock()
	r.save()
}

// IsExplorationMode retorna o estado atual.
func (r *Ranker) IsExplorationMode() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.registry.Exploration
}
