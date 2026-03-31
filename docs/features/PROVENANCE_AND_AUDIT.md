# Funcionalidade: Proveniência e Auditoria de Conhecimento 🔎✨🎻

A **Proveniência (Provenance)** no Lumaestro é o mecanismo de rastreabilidade total que permite ao usuário verificar a fonte de cada fato no Grafo de Conhecimento.

## 🛡️ Rastreabilidade na Fonte (Linhagem)
Ao indexar notas ou memórias, o Lumaestro grava o campo `path` (caminho do arquivo) e `content` (fragmento original) no Qdrant. Isso cria um **Elo Inquebrável** entre o fato e o documento.

## 🔍 Interface de Auditoria Total
- **Sidebar de Proveniência**: Implementada no `GraphVisualizer.vue`, este painel lateral de vidro se abre ao clicar em qualquer nó.
- **Auditoria de Nós**: Exibe o ID do nó, o tipo de documento (Chunk, Memory, System) e o trecho de texto fundamentado (Grounding).
- **Acesso à Fonte Nativo**: O botão "Abrir Arquivo Fonte" utiliza o método `OpenFileInEditor` do backend para abrir o arquivo original no sistema do usuário (Obsidian, VSCode ou leitor de PDF).

## 🏙️ Grounded Chat Reasoning
As respostas do chat agora podem ser auditadas visualmente:
- A IA identifica os IDs dos nós de origem.
- O frontend permite que o usuário clique no nó para verificar a veracidade da afirmação contra o texto original.

---
**Status**: Implementado.
**Módulos**: `app.go` (GetNodeDetails), `GraphVisualizer.vue` (Sidebar).
