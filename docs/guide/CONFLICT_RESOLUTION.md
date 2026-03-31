# Guia: Resolução de Conflitos e Saúde do Grafo 🛡️🛡️✨🎻

O Lumaestro agora gerencia a veracidade das informações com base em decisões de auditoria do usuário.

## ⚠️ Detectando Conflitos
O motor de integridade (`AnalyzeGraphHealth`) identifica nós que possuem informações contraditórias.
- No Grafo 3D, estes nós brilharão em **Vermelho Alerta** (`#ef4444`).
- No **Health HUD** (canto superior), o contador de conflitos indicará quantas anomalias precisam de atenção.

## 🏛️ A Filosofia "Ativo vs Legado"
Ao resolver um conflito, você tem duas opções:
1. **Manter a Antiga**: A informação existente é preservada como a verdade ativa.
2. **Aceitar a Nova (Update)**: A informação antiga é rebaixada para `status: legacy`. Ela deixa de ser a verdade primária, mas permanece no grafo como um nó "fantasma" cinza para manter o seu histórico de pensamento.

## 📉 Densidade Semântica
- **Baixa Densidade (< 30%)**: Indica uma base de conhecimento fragmentada, com muitas notas isoladas.
- **Alta Densidade (> 70%)**: Indica um ecossistema de ideias maduro e interconectivo, onde a IA consegue realizar saltos complexos de raciocínio.

---
**Status**: Integrado à UI `GraphVisualizer.vue`. 🌌
