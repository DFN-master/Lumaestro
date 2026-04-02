package provider

import (
	"context"
	"fmt"

	"Lumaestro/internal/config"
	"Lumaestro/internal/utils"

	"google.golang.org/genai"
)

// EmbeddingService gerencia a geração de vetores via Gemini com suporte a pool de chaves.
type EmbeddingService struct {
	Client *genai.Client
	ctx    context.Context
}

// NewEmbeddingService inicializa o serviço com a API Key ativa do pool.
func NewEmbeddingService(ctx context.Context, apiKey string) (*EmbeddingService, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("falha ao criar cliente GenAI: %w", err)
	}

	return &EmbeddingService{Client: client, ctx: ctx}, nil
}

// rotateAndRetry tenta rotacionar a chave e recriar o client.
// Retorna true se conseguiu rotacionar, false se não há mais chaves.
func (s *EmbeddingService) rotateAndRetry() bool {
	cfg, err := config.Load()
	if err != nil || cfg == nil || cfg.GeminiKeyCount() <= 1 {
		return false
	}

	newKey := cfg.RotateGeminiKey()
	if newKey == "" {
		return false
	}

	// Recria o client com a nova chave
	newClient, err := genai.NewClient(s.ctx, &genai.ClientConfig{
		APIKey:  newKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		fmt.Printf("[KeyPool] ❌ Falha ao criar client com nova chave: %v\n", err)
		return false
	}

	s.Client = newClient
	fmt.Printf("[KeyPool] ✅ Client recriado com nova chave.\n")
	return true
}

// GenerateEmbedding transforma um texto em um vetor []float32.
// Em caso de erro de quota, rotaciona automaticamente para a próxima chave do pool.
func (s *EmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	contents := []*genai.Content{
		{
			Parts: []*genai.Part{
				{Text: text},
			},
		},
	}

	res, err := s.Client.Models.EmbedContent(ctx, "gemini-embedding-2-preview", contents, nil)
	if err != nil {
		// Tenta rotacionar se for erro de quota
		if utils.IsQuotaError(err) {
			fmt.Printf("[KeyPool] ⚠️ Chave exausta (quota). Tentando próxima...\n")

			cfg, _ := config.Load()
			maxRetries := 0
			if cfg != nil {
				maxRetries = cfg.GeminiKeyCount() - 1
			}

			for i := 0; i < maxRetries; i++ {
				if !s.rotateAndRetry() {
					break
				}
				res, err = s.Client.Models.EmbedContent(ctx, "gemini-embedding-2-preview", contents, nil)
				if err == nil {
					break
				}
				if !utils.IsQuotaError(err) {
					return nil, fmt.Errorf("erro ao gerar embedding (pós-rotação): %w", err)
				}
				fmt.Printf("[KeyPool] ⚠️ Chave #%d também exausta: %s\n", i+2, utils.FormatGenAIError(err))
			}

			if err != nil {
				return nil, fmt.Errorf("todas as chaves do pool foram exaustas: %w", err)
			}
		} else {
			return nil, fmt.Errorf("erro ao gerar embedding: %w", err)
		}
	}

	if len(res.Embeddings) == 0 || res.Embeddings[0] == nil || len(res.Embeddings[0].Values) == 0 {
		return nil, fmt.Errorf("vetor de embedding vazio na resposta")
	}

	return res.Embeddings[0].Values, nil
}

