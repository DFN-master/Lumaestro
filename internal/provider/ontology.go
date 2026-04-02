package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"Lumaestro/internal/utils"
)

// Prompter define a interface necessária para realizar consultas ao agente local.
type Prompter interface {
	AskSync(sessionID string, prompt string, images []map[string]string) (string, error)
}

// Triple representa a unidade básica de conhecimento semântico.
type Triple struct {
	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`
}

// Entity representa um nó no grafo do Lumaestro.
type Entity struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Label       string                 `json:"label"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Relation representa uma aresta conectando duas entidades.
type Relation struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Predicate string `json:"predicate"`
}

// OntologyService gerencia a extração de fatos estruturados via ACP (Agente Local).
type OntologyService struct {
	Prompter  Prompter
	SessionID string
	ctx       context.Context
}

// NewOntologyService inicializa o serviço com um Prompter (ex: ACPExecutor) e o ID da sessão.
func NewOntologyService(ctx context.Context, prompter Prompter, sessionID string) *OntologyService {
	return &OntologyService{Prompter: prompter, SessionID: sessionID, ctx: ctx}
}

// ExtractTriples extrai fatos estruturados usando o Agente ACP (Gemini CLI).
func (s *OntologyService) ExtractTriples(ctx context.Context, text string, contextHint string) ([]Triple, error) {
	prompt := fmt.Sprintf(`Extraia triplas semânticas (Sujeito-Predicado-Objeto) do texto abaixo.
Retorne APENAS um ARRAY JSON puro. NÃO use tags de markdown e NÃO use wrappers como {"triples": [...]}.
Exemplo exato do formato esperado:
[
  {"subject": "Lumaestro", "predicate": "uses", "object": "Qdrant"}
]

## DICA DE CONTEXTO GLOBAL:
Use esta informação para resolver pronomes como "ele", "ela", "o projeto", "a empresa": 
> %s

## BLUEPRINT OBRIGATÓRIO:
1. CLASSES: [Person, Project, Task, Concept, Technology, Milestone, Bug, Decision]
2. RELAÇÕES: [is_part_of, works_on, uses, defines, explains, mentions, created, resolved, depends_on]
3. REGRA: Use apenas os termos acima. Atomize os fatos.

Texto:
%s`, contextHint, text)

	response, err := s.Prompter.AskSync(s.SessionID, prompt, nil)
	if err != nil {
		return nil, fmt.Errorf("falha na extração ACP: %w", err)
	}

	return parseTriples(response)
}

// ValidateConflict decide entre informações contraditórias usando o Agente ACP.
func (s *OntologyService) ValidateConflict(ctx context.Context, oldFact, newFact, contextStr string) (string, error) {
	prompt := fmt.Sprintf(`Você é o Agente Validador de Verdade.
Detectamos um conflito no Grafo de Conhecimento.

FATO ANTIGO: %s
FATO NOVO: %s
CONTEXTO RECENTE: %s

Sua tarefa:
Responda APENAS "UPDATE" se o Fato Novo for claramente uma atualização ou correção válida.
Responda APENAS "CONFLICT" se houver dúvida real.

Decisão:`, oldFact, newFact, contextStr)

	response, err := s.Prompter.AskSync(s.SessionID, prompt, nil)
	if err != nil {
		return "CONFLICT", err
	}

	if strings.Contains(strings.ToUpper(response), "UPDATE") {
		return "UPDATE", nil
	}
	return "CONFLICT", nil
}

// ProcessMedia extrai conhecimento de arquivos visuais via Agente ACP.
func (s *OntologyService) ProcessMedia(ctx context.Context, data []byte, mimeType string) (string, []Triple, error) {
	prompt := `Analise este arquivo. Forneça uma descrição detalhada e extraia triplas semânticas (Sujeito, Predicado, Objeto).
	
Formato:
---DESCRICAO---
[Texto]
---TRIPLAS---
[JSON]`

	// Prepara a imagem para o ACP
	images := []map[string]string{
		{
			"type": mimeType,
			"data": utils.EncodeBase64(data),
		},
	}

	response, err := s.Prompter.AskSync(s.SessionID, prompt, images)
	if err != nil {
		return "", nil, fmt.Errorf("erro no processamento multimodal ACP: %w", err)
	}

	parts := strings.Split(response, "---TRIPLAS---")
	description := strings.TrimSpace(strings.TrimPrefix(parts[0], "---DESCRICAO---"))
	
	var triples []Triple
	if len(parts) > 1 {
		triples, _ = parseTriples(parts[1])
	}
	return description, triples, nil
}

func parseTriples(rawJSON string) ([]Triple, error) {
	rawJSON = strings.TrimPrefix(rawJSON, "```json")
	rawJSON = strings.TrimPrefix(rawJSON, "```")
	rawJSON = strings.TrimSuffix(rawJSON, "```")
	rawJSON = strings.TrimSpace(rawJSON)

	var triples []Triple
	err := json.Unmarshal([]byte(rawJSON), &triples)
	return triples, err
}
