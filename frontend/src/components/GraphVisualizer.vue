<script setup>
import * as d3 from 'd3'
import { nextTick, onMounted, ref, watch } from 'vue'
import { ScanVault } from '../../wailsjs/go/main/App'

const props = defineProps({
  nodes: { type: Array, default: () => [] },
  edges: { type: Array, default: () => [] },
  graphLogs: { type: Array, default: () => [] },
  activeNode: { type: String, default: null } // O Roteador de Luz (Nó atual pensando na AI)
})

const svgRef = ref(null)
const containerRef = ref(null)
const logContainerRef = ref(null)
const getSafeId = (id) => `node-${id.replace(/[^a-zA-Z0-9_-]/g, '_')}`

let simulation = null
let svg = null
let g = null
let linkGroup = null
let nodeGroup = null

// Shadow State para a Fisica Limpa do D3 (Blindado contra Proxies de Vue3)
const localNodesMap = new Map()
const localEdgesMap = new Map()
const localNodesList = []
const localEdgesList = []

// Setup inicial do "Palco" (SVG, Filtros, Zoom e Forças Base) apenas UMA vez!
const mountGraphEnvironment = () => {
  if (!svgRef.value || !containerRef.value) return

  const width = containerRef.value.clientWidth
  const height = containerRef.value.clientHeight

  svg = d3.select(svgRef.value)
    .attr('width', '100%')
    .attr('height', '100%')
    .attr('viewBox', `0 0 ${width} ${height}`)

  svg.selectAll("*").remove() 

  // Efeitos GLOW Dourado
  const defs = svg.append('defs')
  const filter = defs.append('filter').attr('id', 'glow').attr('x', '-50%').attr('y', '-50%').attr('width', '200%').attr('height', '200%')
  filter.append('feGaussianBlur').attr('stdDeviation', '2.5').attr('result', 'coloredBlur')
  const feMerge = filter.append('feMerge')
  feMerge.append('feMergeNode').attr('in', 'coloredBlur')
  feMerge.append('feMergeNode').attr('in', 'SourceGraphic')
  
  // Destaque Glow Ativo (Amarelo Raciocínio)
  const filterActive = defs.append('filter').attr('id', 'glow-active').attr('x', '-50%').attr('y', '-50%').attr('width', '200%').attr('height', '200%')
  filterActive.append('feGaussianBlur').attr('stdDeviation', '5').attr('result', 'coloredBlur')
  const feMergeA = filterActive.append('feMerge')
  feMergeA.append('feMergeNode').attr('in', 'coloredBlur')
  feMergeA.append('feMergeNode').attr('in', 'SourceGraphic')

  g = svg.append('g')

  // Zoom behavior
  const zoom = d3.zoom().scaleExtent([0.1, 4]).on('zoom', (event) => g.attr('transform', event.transform))
  svg.call(zoom)

  // Criar camadas separadas para Nodes sobrepor Links sempre.
  linkGroup = g.append('g').attr('class', 'links')
  nodeGroup = g.append('g').attr('class', 'nodes')

  // Inicializa Físicas Vázias.
  simulation = d3.forceSimulation()
    .force('link', d3.forceLink().id(d => d.id).distance(150))
    .force('charge', d3.forceManyBody().strength(-500))
    .force('center', d3.forceCenter(width / 2, height / 2))
    .force('collision', d3.forceCollide().radius(30))
}

// O Update "Cérebro Vivo": Não limpa, ele dá JOIN em dados que chegam
const updateGraph = () => {
  if (!simulation) return

  // 1. Clonagem e Hidratação dos Shadow Arrays
  props.nodes.forEach(n => {
    if (!localNodesMap.has(n.id)) {
      const cw = containerRef.value?.clientWidth || 500
      const ch = containerRef.value?.clientHeight || 500
      
      const clone = { 
         ...n, 
         x: cw / 2 + (Math.random() - 0.5) * 50,
         y: ch / 2 + (Math.random() - 0.5) * 50
      }
      localNodesMap.set(n.id, clone)
      localNodesList.push(clone)
    }
  })

  // 1.5 Criação de Nós Virtuais (Para links que ainda não existem como arquivos)
  props.edges.forEach(e => {
    const t = e.target.id || e.target
    if (!localNodesMap.has(t)) {
      const cw = containerRef.value?.clientWidth || 500
      const ch = containerRef.value?.clientHeight || 500
      const virtualNode = { 
        id: t, 
        name: t, 
        virtual: true,
        x: cw / 2 + (Math.random() - 0.5) * 50,
        y: ch / 2 + (Math.random() - 0.5) * 50
      }
      localNodesMap.set(t, virtualNode)
      localNodesList.push(virtualNode)
    }
  })

  // Higieniza as arestas (Recria a cada tick)
  localEdgesList.length = 0
  localEdgesMap.clear()

  props.edges.forEach(e => {
    const s = e.source.id || e.source
    const t = e.target.id || e.target
    const key = `${s}-${t}`
    
    if (localNodesMap.has(s) && localNodesMap.has(t)) {
        if (!localEdgesMap.has(key)) {
          const clone = { ...e, source: s, target: t }
          localEdgesMap.set(key, clone)
          localEdgesList.push(clone)
        }
    }
  })

  // 2. UPDATE EDGES (Energia Fluindo)
  const links = linkGroup.selectAll("line").data(localEdgesList, d => `${d.source.id || d.source}-${d.target.id || d.target}`)
  const linksEnter = links.enter()
    .append("line")
    .attr("class", "edge-flow")
    .attr("stroke", "rgba(59, 130, 246, 0.4)")
    .attr("stroke-width", 2)
  links.exit().remove()
  const allLinks = linksEnter.merge(links)

  // 3. UPDATE NODES
  const nodes = nodeGroup.selectAll("g").data(localNodesList, d => d.id)
  const nodesEnter = nodes.enter().append("g")
      .call(d3.drag()
        .on("start", dragstarted)
        .on("drag", dragged)
        .on("end", dragended))

  // Círculos LUMINOSOS
  nodesEnter.append("circle")
    .attr("r", 0) 
    .attr("fill", d => d.virtual ? "rgba(59, 130, 246, 0.2)" : "var(--primary)")
    .attr("stroke", d => d.virtual ? "rgba(59, 130, 246, 0.5)" : "none")
    .attr("stroke-dasharray", d => d.virtual ? "2,2" : "none")
    .attr("filter", d => d.virtual ? "none" : "url(#glow)")
    .attr("class", "node-circle")
    .attr("id", d => getSafeId(d.id))
    .transition().duration(500).attr("r", d => d.virtual ? 4 : 6)

  // Nomes 
  nodesEnter.append("text")
    .text(d => d.name || d.id)
    .attr("x", 12).attr("y", 4)
    .attr("class", "node-label")
    .style("opacity", d => d.virtual ? 0.3 : 1)

  nodes.exit().remove()
  const allNodes = nodesEnter.merge(nodes)

  // 4. REINICIAR GRAVIDADES DESACOLHADAS DE VUE
  simulation.nodes(localNodesList)
  simulation.force("link").links(localEdgesList)
  simulation.alpha(1).restart() // Aumentado para 1 para dar mais "vida" ao movimento

  simulation.on("tick", () => {
    allLinks.attr("x1", d => d.source.x).attr("y1", d => d.source.y)
            .attr("x2", d => d.target.x).attr("y2", d => d.target.y)
    allNodes.attr("transform", d => `translate(${d.x}, ${d.y})`)
  })

  function dragstarted(event, d) { if (!event.active) simulation.alphaTarget(0.3).restart(); d.fx = d.x; d.fy = d.y; }
  function dragged(event, d) { d.fx = event.x; d.fy = event.y; }
  function dragended(event, d) { if (!event.active) simulation.alphaTarget(0); d.fx = null; d.fy = null; }
}

// ==========================================
// 🌟 EFEITO RAG TRACKER (PULO DA SINAPSE) 🌟
// ==========================================
watch(() => props.activeNode, (newId) => {
  if (!nodeGroup) return
  
  // Apaga as notas anteriores (Reseta p/ Azul Padrão)
  nodeGroup.selectAll("circle")
    .transition().duration(200)
    .attr("fill", "var(--primary)").attr("r", 6).style("filter", "url(#glow)")

  if (newId) {
    // ⚡ ALPHA BOOST: Pulso de energia física na simulação ao ativar nota (Feel TrustGraph)
    if (simulation) simulation.alpha(0.8).restart()

    // ⚡ PULSO DE ENERGIA (Sinapses): Brilho nos caminhos conectados
    if (linkGroup) {
      linkGroup.selectAll(".edge-flow")
        .classed("edge-active", d => {
          const sourceId = d.source.id || d.source
          const targetId = d.target.id || d.target
          return sourceId === newId || targetId === newId
        })
      
      // Remove o pulso após um tempo (raciocínio passou)
      setTimeout(() => {
        if (linkGroup) {
          linkGroup.selectAll(".edge-active").classed("edge-active", false)
        }
      }, 3000)
    }

    // Acende a Bola Lida pela IA (Amarelo Ouro + Gigante + Pulsação)
    const safeId = getSafeId(newId)
    // Seletor mais robusto usando atributo ID exato (evita falhas com caracteres especiais do CSS)
    const target = nodeGroup.select(`[id="${safeId}"]`)
    
    target.transition().duration(400)
      .attr("fill", "#fcd34d")
      .attr("r", 15)
      .style("filter", "drop-shadow(0 0 15px #fcd34d)")
      .on("end", function() {
          d3.select(this).classed("pulse-ring", true)
      })

    // Decaimento natural após o "pensamento" passar (opcional, mas charmoso)
    target.transition().delay(5000).duration(2000)
      .attr("fill", "var(--primary)")
      .attr("r", 6)
      .style("filter", "url(#glow)")
      .on("start", function() {
          d3.select(this).classed("pulse-ring", false)
      })
  }
})

// Reatividade
watch(() => [props.nodes, props.edges], () => {
  updateGraph()
}, { deep: true })

// Auto-Scroll Raciocínio (Logs descendo na telinha lateral)
watch(() => props.graphLogs, () => {
  nextTick(() => {
    if (logContainerRef.value) {
      logContainerRef.value.scrollTop = logContainerRef.value.scrollHeight
    }
  })
}, { deep: true })

onMounted(() => {
  mountGraphEnvironment()
  updateGraph()
})

const resetZoom = () => svg.transition().duration(750).call(d3.zoom().transform, d3.zoomIdentity)

const scanning = ref(false)
const triggerScan = async () => {
  if (scanning.value) return
  scanning.value = true
  try {
    await ScanVault()
  } catch (error) {
    console.error("Erro no Scan:", error)
  } finally {
    setTimeout(() => { scanning.value = false }, 1000)
  }
}
</script>

<template>
  <div class="graph-wrapper animate-fade-in" ref="containerRef">
    <svg ref="svgRef" class="main-svg"></svg>
    
    <!-- Controles & Console de Logs (Painel de Pensamento Vidrado) -->
    <div class="graph-ui glass">
      <div class="ui-header">
        <span class="pulse"></span>
        <h3>Conhecimento Obsidian</h3>
      </div>
      
      <div class="ui-actions">
        <div style="display: flex; gap: 8px;">
          <button @click="resetZoom" class="action-btn" title="Centralizar">🎯 <span>RESET</span></button>
          <button @click="triggerScan" class="action-btn" :class="{'scanning-btn': scanning}" title="Forçar Index"><span v-if="!scanning">🔄</span><span v-else class="spin">⏳</span><span>SCAN</span></button>
        </div>
        <div class="stat-item">
          <span class="val">{{ nodes.length }}</span>
          <span class="lab">NOTAS</span>
        </div>
      </div>

      <!-- O CONSOLE VIVO DO RACIOCÍNIO IA -->
      <div class="graph-logs-console" ref="logContainerRef" v-if="graphLogs.length > 0">
        <div v-for="(log, idx) in graphLogs" :key="idx" class="log-entry">
          <span class="log-text">{{ log }}</span>
        </div>
      </div>
    </div>

    <!-- Background Imersivo -->
    <div class="graph-bg"></div>
  </div>
</template>

<style scoped>
.graph-wrapper {
  width: 100%;
  height: 100vh;
  background: var(--bg-dark);
  position: relative;
  overflow: hidden;
}

.main-svg {
  position: relative;
  z-index: 2;
  cursor: grab;
}

.main-svg:active { cursor: grabbing; }

/* Node Styling */
:deep(.node-label) {
  font-family: 'Outfit', sans-serif;
  font-size: 10px;
  fill: rgba(255, 255, 255, 0.4);
  pointer-events: none;
  font-weight: 500;
  letter-spacing: 0.5px;
  transition: opacity 0.3s, fill 0.3s;
}

:deep(g:hover .node-label) {
  fill: white;
  font-size: 12px;
  opacity: 1;
}

:deep(.node-circle) {
  transition: r 0.3s, fill 0.3s;
}

/* UI Panel */
.graph-ui {
  position: absolute;
  top: 2rem;
  left: 2rem;
  z-index: 10;
  padding: 1.2rem;
  border-radius: 20px;
  min-width: 280px;
  width: max-content;
  border: 1px solid rgba(255, 255, 255, 0.05);
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.ui-header {
  display: flex;
  align-items: center;
  gap: 10px;
}

.ui-header h3 {
  font-size: 0.75rem;
  font-weight: 800;
  letter-spacing: 2px;
  text-transform: uppercase;
  color: var(--primary);
  margin: 0;
}

.pulse {
  width: 6px;
  height: 6px;
  background: var(--primary);
  border-radius: 50%;
  box-shadow: 0 0 8px var(--primary);
  display: inline-block;
}

.ui-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1.5rem;
}

.action-btn {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: white;
  padding: 8px 12px;
  border-radius: 10px;
  font-size: 0.6rem;
  font-weight: 800;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all 0.3s;
}

.action-btn:hover {
  background: var(--primary);
  border-color: var(--primary);
  transform: translateY(-2px);
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
}

.val {
  font-size: 1.2rem;
  font-weight: 900;
  color: white;
  line-height: 1;
}

.lab {
  font-size: 0.55rem;
  font-weight: 800;
  color: var(--text-dim);
  letter-spacing: 1px;
}

/* Background Imersivo */
.graph-bg {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  background: radial-gradient(circle at center, rgba(59, 130, 246, 0.05) 0%, transparent 70%);
  pointer-events: none;
  z-index: 1;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.animate-fade-in {
  animation: fadeIn 1s ease-out;
}

@keyframes spinFast {
  100% { transform: rotate(360deg); }
}

.spin {
  display: inline-block;
  animation: spinFast 1s linear infinite;
}

.scanning-btn {
  opacity: 0.7;
  pointer-events: none;
  border-color: var(--primary);
}

/* 🧠 Efeitos do Raciocínio (Cérebro Artificial Vivo) */
.edge-flow {
  stroke-dasharray: 4 4;
  animation: dashFlow 2s linear infinite;
}

@keyframes dashFlow {
  to { stroke-dashoffset: -20; }
}

/* ⚡ Pulso de Sinapse (Energia nos Caminhos) */
.edge-active {
  stroke: #fcd34d !important;
  stroke-width: 3 !important;
  stroke-opacity: 1 !important;
  stroke-dasharray: 8 4 !important;
  animation: dashFlow 0.5s linear infinite !important;
  filter: drop-shadow(0 0 5px #fcd34d);
  transition: stroke 0.3s, stroke-width 0.3s;
}

/* ⚙️ Console Visual Lateral */
.graph-logs-console {
  margin-top: 15px;
  max-height: 180px;
  overflow-y: auto;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
  padding-top: 10px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  scroll-behavior: smooth;
}

.graph-logs-console::-webkit-scrollbar { width: 4px; }
.graph-logs-console::-webkit-scrollbar-thumb { background: rgba(59, 130, 246, 0.5); border-radius: 4px; }

.log-entry {
  font-family: Consolas, 'Fira Code', monospace;
  font-size: 0.6rem;
  color: rgba(255,255,255,0.6);
  border-left: 2px solid rgba(59, 130, 246, 0.5);
  padding-left: 6px;
  line-height: 1.4;
  word-break: break-all;
}

/* 🌀 Pulsação de Nó Ativo */
.pulse-ring {
  animation: pulse-ring 1.5s cubic-bezier(0.215, 0.61, 0.355, 1) infinite;
}

@keyframes pulse-ring {
  0% { transform: scale(1); opacity: 1; }
  50% { transform: scale(1.5); opacity: 0.5; }
  100% { transform: scale(1); opacity: 1; }
}
</style>
