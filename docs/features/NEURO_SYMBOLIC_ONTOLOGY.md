# Inteligência: Ontologia Neuro-Simbólica (Elite Blueprint) 🧠🛡️🎻

A **Ontologia Neuro-Simbólica** no Lumaestro é a união da capacidade de linguagem do LLM com a estrutura lógica rígida dos Grafos de Conhecimento, inspirada no TrustGraph 2.0.

## 🧱 Blueprint Semântico: O Guardião
O motor de extração de triplas (`internal/provider/ontology.go`) foi atualizado de um modelo "livre" para um **Modelo de Blueprint Rígido**.
- **Classes Obrigatórias**: [Person, Project, Task, Concept, Technology, Milestone, Bug, Decision]
- **Relações Obrigatórias**: [is_part_of, works_on, uses, defines, explains, mentions, created, resolved, depends_on]
- **Proibição de Termos**: A IA é proibida de inventar novas classes ou relações. Fatos que não se enquadram na ontologia são filtrados para evitar "alucinação de esquemas".

## 🧬 Disambiguação e Consolidação
O uso de Blueprints permite:
- **Redução de Duplicatas**: Termos como "usuário" e "user" são consolidados em uma única classe `Person`.
- **Prevenção de Ambiguidade**: Perguntas semanticamente diferentes geram caminhos de busca distintos no grafo, aumentando a precisão da resposta.
- **Granding Determinístico**: Cada afirmação gerada no chat agora tem uma âncora em uma tripla ontológica verificável.

---
**Status**: Implementado.
**Serviço**: `OntologyService.ExtractTriples`.
