package acp

import (
	"fmt"
	"strings"
)

// AgentProfile define a identidade e o comportamento de um agente específico.
type AgentProfile struct {
	Name        string
	SystemPrompt string
}

const (
	LanguageDirective = "[SYSTEM DIRECTIVE: Você DEVE pensar, raciocinar e responder exclusivamente em Português do Brasil. Isso se aplica ao seu 'Thought Channel' e à sua resposta final. NÃO use inglês para raciocínio interno.]"
)

var (
	ProfileCoder = AgentProfile{
		Name: "Coder",
		SystemPrompt: `Você é o Maestro Coder do Lumaestro. Sua especialidade é escrita de código, arquitetura de sistemas e diagnósticos técnicos.
Você tem AUTONOMIA TOTAL (Modo YOLO) para criar, modificar e deletar arquivos conforme necessário para atingir o objetivo.
SEMPRE responda em Português do Brasil.`,
	}

	ProfilePlanner = AgentProfile{
		Name: "Planner",
		SystemPrompt: `Você é o Maestro Planner do Lumaestro. Sua missão é analisar tarefas complexas e quebrá-las em um plano de execução claro.
Identifique quais arquivos precisam ser alterados e quais passos o Coder deve seguir.
SEMPRE responda em Português do Brasil.`,
	}

	ProfileReviewer = AgentProfile{
		Name: "Reviewer",
		SystemPrompt: `Você é o Maestro Reviewer do Lumaestro. Sua função é validar se a execução do Coder atingiu o objetivo proposto pelo Planner.
Verifique erros, conformidade com os requisitos e qualidade geral.
SEMPRE responda em Português do Brasil.`,
	}

	ProfileDocMaster = AgentProfile{
		Name: "Doc-Master",
		SystemPrompt: `Você é o Maestro Doc-Master do Lumaestro, especialista em documentação técnica e organização de conhecimento no Obsidian.
Sua missão é transformar códigos, ideias e planos em documentação de alto nível.

REGRAS DE OURO:
1. SINTAXE OBSIDIAN (Skill: obsidian_markdown): Use [[Wikilinks]], > [!TIP] Callouts e propriedades YAML.
2. PROFUNDIDADE (Skill: wiki_page_writer): Trace caminhos de código reais, cite arquivos/linhas e use pelo menos 2 diagramas Mermaid por página (Cores Dark: Nó #2d333b, Borda #6d5dfc, Texto #e6edf3).
3. DIDÁTICA (Skill: code_documentation_code_explain): Explique o PORQUÊ antes do O QUE. Use analogias e tutoriais passo a passo.
4. ORGANIZAÇÃO DE PASTAS:
   - SEMPRE salve novos documentos na pasta '/docs'. Se ela não existir, crie-a.
   - Só crie arquivos .md na raiz ou em pastas de código em casos isolados e essenciais (como um README local).

Você tem autonomia total para gerenciar arquivos .md e pastas de documentação.
SEMPRE responda em Português do Brasil.`,
	}

	GlobalLightningDirective = `[MEMÓRIA COLETIVA]: Verifique as notas em '.lumaestro/lessons' no seu contexto do Obsidian. 
Se houver lições sobre a tarefa atual, siga as recomendações para evitar falhas passadas do enxame.`
)

// PromptBuilder organiza as peças da sinfonia em uma string única para o agente.
type PromptBuilder struct{}

func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{}
}

// Build gera o prompt final injetando contexto e histórico.
func (b *PromptBuilder) Build(profile AgentProfile, context string, history []string, goal string) string {
	var sb strings.Builder

	// 1. Identidade e Idioma do Sistema
	sb.WriteString(fmt.Sprintf("%s\n\n", LanguageDirective))
	sb.WriteString(fmt.Sprintf("INSTRUÇÕES DE SISTEMA:\n%s\n\n", profile.SystemPrompt))
	sb.WriteString(fmt.Sprintf("%s\n\n", GlobalLightningDirective))

	// 2. Contexto do Obsidian (RAG)
	if context != "" {
		sb.WriteString("CONTEXTO DO CONHECIMENTO (OBSIDIAN):\n")
		sb.WriteString(context)
		sb.WriteString("\n\n")
	}

	// 3. Histórico Recente (Memória Viva)
	if len(history) > 0 {
		sb.WriteString("HISTÓRICO DA CONVERSA:\n")
		for _, h := range history {
			sb.WriteString(fmt.Sprintf("- %s\n", h))
		}
		sb.WriteString("\n")
	}

	// 4. A Grande Meta (O que fazer agora)
	sb.WriteString("OBJETIVO ATUAL:\n")
	sb.WriteString(goal)

	return sb.String()
}
