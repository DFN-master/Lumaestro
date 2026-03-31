# Design UX: Trilhas de Raciocínio e Trajetórias Visuais 🌌✨🎻

O Lumaestro utiliza uma interface de **RAG Transparente**, onde o fluxo de pensamento da IA é visualmente mapeado no Grafo 3D em tempo real.

## 🟩 Trilhas Verde Néon (Highlighting)
Inspirado por visualizações cinemáticas de dados, o sistema destaca os caminhos percorridos durante uma consulta:
- **Evento**: `graph:highlight` (disparado pelo backend no momento da expansão de contexto).
- **Estética**: As arestas (links) entre os nós consultados brilham em **Verde Néon (#4ade80)**.
- **Micro-animações**: Partículas direcionais surgem nessas trilhas, indicando o sentido do fluxo de informação dos documentos para a memória de chat.
- **Efeito de Rastro**: O destaque persiste por 4 segundos antes de se desvanecer, permitindo que o usuário acompanhe o "salto" entre as notas sem poluir o grafo permanentemente.

## 🏙️ Estética de Vidro (Glassmorphism)
Todos os painéis de controle e auditoria seguem o design premium do Lumaestro:
- **Transparência**: Uso de `backdrop-filter: blur(25px)` para manter o grafo 3D visível atrás da interface.
- **Borda de Neônio**: Bordas sutis em azul e roxo que sinalizam o status de atividade da IA.
- **Feedback Pulsante**: Nós ativos e núcleos de conhecimento pulsam em tons de ouro e platina, facilitando o foco em meio a milhares de pontos.

## 🔎 UX de Auditoria (Sidebar)
A **Sidebar de Proveniência** lateral direita desliza de forma fluida (`slide-fade animation`), integrando-se organicamente à exploração semântica sem interromper a visualização do grafo.

---
**Status**: Implementado.
**Módulos**: `GraphVisualizer.vue`, `navigation.go` (Eventos).
