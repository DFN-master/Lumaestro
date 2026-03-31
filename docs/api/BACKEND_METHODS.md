# API Backend: Referência RPC do Lumaestro 🛠️🎻✨

Este guia detalha as chamadas RPC (`app.go`) que sustentam as novas funcionalidades de auditoria e integridade do sistema.

## 🔎 Proveniência e Auditoria
### `GetNodeDetails(name string)`
- **Descrição**: Recupera o payload completo do Qdrant para um nó específico.
- **Retorno**: Um mapa contendo `path`, `content`, `observed_at` e `status`.
- **Uso**: Preenchimento da Sidebar de auditoria.

### `OpenFileInEditor(path string)`
- **Descrição**: Comando nativo que abre o arquivo original no sistema operacional (Obsidian, VSCode, PDF Reader).
- **Parâmetro**: Caminho absoluto (`path`) do arquivo.

## 🛡️ Integridade e Verdade
### `AnalyzeGraphHealth()`
- **Descrição**: Escaneia a base de conhecimento em busca de anomalias semânticas e calcula a densidade de conexões.
- **Retorno**: Estatísticas de saúde (Density, Conflicts Count).

### `ResolveConflict(decision, subject, predicate, oldId, newValue, sessionID)`
- **Descrição**: Aplica a decisão de verdade do usuário.
- **Lógica**: Se a decisão for "new", o fato antigo é marcado com `status: legacy` e a nova informação torna-se a verdade ativa no grafo.

---
**Base Tecnológica**: Wails V2, Go 1.21+, Qdrant API.
