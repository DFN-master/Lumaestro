# Arquitetura: Context-Flow RAG (TrustGraph Core) 🚀🎻🔍

O **Context-Flow** é o motor de navegação do Lumaestro que evolui a busca vetorial simples para uma exploração de grafos semânticos em tempo real.

## 🚀 Motor de Navegação N-Hop
Diferente de RAGs tradicionais que apenas buscam notas similares (K-Nearest Neighbors), o Context-Flow realiza uma **Expansão de Vizinhança (N-Hop Traversal)**.
- **Nucleação**: Identifica os nós mestre (Sóis do Conhecimento) via similaridade vetorial.
- **Exploração**: Recupera todos os documentos conectados a esses núcleos até uma profundidade configurável (`graph_depth`).
- **Orquestração**: Os dados são consolidados em um sub-grafo que sustenta o prompt final, eliminando lacunas de contexto entre notas relacionadas.

## ⚡ Otimização: Batch Fetch (GetPoints)
Para garantir alta performance, o Lumaestro utiliza buscadores em lote:
- O método `GetPoints` no cliente Qdrant permite recuperar múltiplos IDs em uma única chamada HTTP.
- Isso reduz a latência da expansão de 1-Hop para milissegundos, mesmo em vaults com milhares de links.

## ⚙️ Controle Dinâmico: Alcance da Teia
A variável `graph_neighbor_limit` (configurável via UI) permite ao usuário ditar a "Curiosidade" da IA:
- **Low (1-3)**: Foco extremo no assunto direto.
- **High (15-25)**: Pesquisa profunda que conecta temas distantes no Vault do Obsidian.

---
**Status**: Implementado e Ativo.
**Base Tecnológica**: Qdrant, Golang, IA Generativa (Gemini 2.0).
