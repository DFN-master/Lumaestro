# Arquitetura do Motor Cognitivo Lumaestro

Este documento detalha o funcionamento interno do sistema de RAG e do Grafo 3D.

## Camadas de Conhecimento

### 1. Conhecimento Estático (Vault + Sistema)
Os arquivos `.md` são lidos pelo `crawler.go` e indexados no Qdrant com metadados de `document-type` (chunk, source, system).
- **Vault**: Notas pessoais (Azul Neon).
- **Sistema**: Documentação do projeto (Branco Platina).
- **Mídia**: PDFs e imagens com OCR (Roxo).

### 2. Conhecimento Dinâmico (Chat Memory)
- As conversas no chat são processadas pelo `KnowledgeWeaver`.
- Sinapses (Triplas) são extraídas e salvas na coleção `knowledge_graph`.
- Identificação por `session_id` para separação de contextos.
- Cor: Rosa / Magenta.

### 3. Validação e Consistência (Agente da Verdade)
- Detecção de contradições semânticas pré-ingestão.
- Agente validador do Gemini decide entre:
    - **UPDATE**: Substitui a verdade ativa e marca a anterior como `legacy`.
    - **CONFLICT**: Emite alerta visual e pede confirmação do usuário.

## Configurações Técnicas
- **Embedding**: Gemini Google AI.
- **Banco Vetorial**: Qdrant (Coolify Cloud).
- **Visualização**: Three.js + 3D-Force-Graph.
