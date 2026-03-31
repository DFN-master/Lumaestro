<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { GetConfig, SaveConfig, GetToolsStatus, InstallTool, SetupTool } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime'

const config = ref({
  obsidian_vault_path: '',
  qdrant_url: '',
  gemini_api_key: '',
  use_gemini_api_key: false,
  claude_api_key: '',
  use_claude_api_key: false,
  active_agent: 'gemini',
  auto_start_agents: [],
  agent_language: 'Português do Brasil',
  graph_depth: 1,
  graph_neighbor_limit: 5,
  graph_context_limit: 4000,
  security: {
    allow_read: false,
    allow_write: false,
    allow_create: false,
    allow_delete: false,
    allow_move: false,
    allow_run_commands: false,
    full_machine_access: false
  }
})

// Helpers para auto-start toggles
const isAutoStart = (agent) => {
  return (config.value.auto_start_agents || []).includes(agent)
}

const toggleAutoStart = async (agent) => {
  if (!config.value.auto_start_agents) {
    config.value.auto_start_agents = []
  }
  const idx = config.value.auto_start_agents.indexOf(agent)
  if (idx >= 0) {
    config.value.auto_start_agents.splice(idx, 1)
  } else {
    config.value.auto_start_agents.push(agent)
  }
  // Salva imediatamente para ser persistente
  await SaveConfig(config.value)
}

const status = ref({
  qdrant: false,
  tools: {
    gemini: false,
    claude: false,
    obsidian: false,
    claude_auth: false,
    gemini_auth: false
  }
})

const installLogs = ref([])
const installStatus = ref('')
const logContainer = ref(null)

const scrollToConsole = async () => {
  await nextTick()
  setTimeout(() => {
    const view = document.querySelector('.settings-view')
    if (view) {
      view.scrollTo({
        top: view.scrollHeight,
        behavior: 'smooth'
      })
    }
  }, 100)
}

onMounted(async () => {
  const savedConfig = await GetConfig()
  console.log("Configurações recebidas do Maestro:", savedConfig)
  if (savedConfig) {
    config.value = savedConfig
  }
  
  refreshStatus()

  // Ouvir logs do instalador em tempo real
  EventsOn('installer:log', (log) => {
    installLogs.value.push(log)
    if (logContainer.value) {
      setTimeout(() => {
        logContainer.value.scrollTop = logContainer.value.scrollHeight
      }, 10)
    }
  })
})

const refreshStatus = async () => {
  status.value.tools = await GetToolsStatus()
}

const fixEnv = async () => {
  installLogs.value = []
  installStatus.value = "Iniciando correção de ambiente..."
  scrollToConsole()
  // @ts-ignore
  const res = await window.go.main.App.FixEnvironment()
  installStatus.value = res
  refreshStatus()
}

const save = async () => {
  const res = await SaveConfig(config.value)
  alert(res)
  refreshStatus()
}

const install = async (name) => {
  try {
    installLogs.value = []
    installStatus.value = `Iniciando operação para ${name}...`
    scrollToConsole()

    const res = await InstallTool(name)
    installStatus.value = res ? res : "Operação finalizada."
  } catch (err) {
    installStatus.value = `ERRO Crítico: ${err}`
  }
  refreshStatus()
}

const setup = async (name) => {
  installStatus.value = `Abrindo terminal de configuração para ${name}...`
  scrollToConsole()
  const res = await SetupTool(name)
  installStatus.value = res
}

const getAuthLabel = (agent) => {
  if (config.value[`use_${agent}_api_key`]) {
    return 'CHAVE API ⚡'
  }
  return agent === 'claude' ? 'FAZER LOGIN (OAUTH)' : 'CONFIGURAR LOGIN'
}

const getAuthStyle = (agent) => {
  if (config.value[`use_${agent}_api_key`]) {
    return 'border-color: rgba(245, 158, 11, 0.4); color: #fde68a; background: rgba(245, 158, 11, 0.08);'
  }
  return 'border-color: var(--primary);'
}

const mcpName = ref('')
const mcpCommand = ref('')
const mcpServers = ref('')
const showMcpList = ref(false)

const generateGeminiMD = async () => {
  if (window.go && window.go.main && window.go.main.App) {
    const res = await window.go.main.App.GenerateGeminiMD()
    alert(res)
  }
}

const addMCPServer = async () => {
  if (!mcpName.value || !mcpCommand.value) {
    alert("Preencha o Nome e o Comando para o MCP")
    return
  }
  installLogs.value = []
  installStatus.value = `Instalando servidor MCP: ${mcpName.value}...`
  scrollToConsole()
  
  if (window.go && window.go.main && window.go.main.App) {
    const res = await window.go.main.App.AddMCPServer(mcpName.value, mcpCommand.value)
    installStatus.value = "Instalação do MCP Finalizada."
    mcpName.value = ''
    mcpCommand.value = ''
    alert("Retorno do Terminal:\n" + res)
  }
}

const listMCPServers = async () => {
  if (window.go && window.go.main && window.go.main.App) {
    const res = await window.go.main.App.ListMCPServers()
    mcpServers.value = res
    showMcpList.value = true
  }
}

const activeTab = ref('geral')
</script>

<template>
  <main class="settings-view animate-fade-up">
    <header class="settings-header">
      <div class="brand-badge">SISTEMA</div>
      <h1 class="gradient-text">Configurações Maestro</h1>
      <p class="subtitle">Gerencie o cérebro e as ferramentas da sua IA.</p>
    </header>

    <div class="tabs-nav">
      <button @click="activeTab = 'geral'" :class="{ 'active': activeTab === 'geral' }" class="tab-btn">GERAL</button>
      <button @click="activeTab = 'agentes'" :class="{ 'active': activeTab === 'agentes' }" class="tab-btn">AGENTES</button>
      <button @click="activeTab = 'seguranca'" :class="{ 'active': activeTab === 'seguranca' }" class="tab-btn">SEGURANÇA</button>
      <button @click="activeTab = 'mcp'" :class="{ 'active': activeTab === 'mcp' }" class="tab-btn">AVANÇADO (MCP)</button>
    </div>

    <div class="content-grid-tabs">
      <!-- ABA GERAL -->
      <section v-if="activeTab === 'geral'" class="glass premium-shadow panel-main animate-fade-in">
        <h2 class="section-title">Configurações Base</h2>
        
        <div class="form-group">
          <label>Idioma Nativo do Agente</label>
          <div class="input-wrapper">
            <select v-model="config.agent_language" class="premium-input premium-select">
              <option value="Português do Brasil">Português (Brasil)</option>
              <option value="English">English</option>
              <option value="Español">Español</option>
              <option value="Français">Français</option>
              <option value="Deutsch">Deutsch</option>
              <option value="Italiano">Italiano</option>
              <option value="日本語 (Japanese)">日本語 (Japonês)</option>
            </select>
            <div class="input-glow"></div>
          </div>
        </div>

        <div class="form-group">
          <label>Caminho do Obsidian Vault</label>
          <div class="input-wrapper">
            <input v-model="config.obsidian_vault_path" type="text" class="premium-input" placeholder="C:\Users\...\Obsidian" />
            <div class="input-glow"></div>
          </div>
        </div>

        <div class="form-group">
          <label>Alcance da Teia (Vizinhos por Nó): <span class="val-highlight">{{ config.graph_neighbor_limit }}</span></label>
          <div class="input-wrapper range-wrapper">
            <input v-model.number="config.graph_neighbor_limit" type="range" min="1" max="25" step="1" class="premium-range" />
            <div class="range-track"></div>
          </div>
          <p class="desc-small">Define quantos vizinhos o Maestro deve buscar para cada nota encontrada no RAG.</p>
        </div>

        <div class="form-group">
          <label>URL do Qdrant Cloud</label>
          <div class="input-wrapper">
            <input v-model="config.qdrant_url" type="text" class="premium-input" placeholder="https://..." />
            <div class="input-glow"></div>
          </div>
        </div>

        <div style="display: flex; gap: 15px; margin-top: 2rem;">
          <button @click="save" class="btn-premium save-btn">
            <span>SALVAR ALTERAÇÕES</span>
            <div class="btn-shimmer"></div>
          </button>
        </div>
      </section>

      <!-- ABA AGENTES -->
      <div v-if="activeTab === 'agentes'" class="agents-layout animate-fade-in">
        <section class="glass premium-shadow panel-main">
          <h2 class="section-title">Autenticação e Chaves</h2>
          
          <div class="form-group">
            <label>Google Gemini API Key</label>
            <div class="input-wrapper">
              <input v-model="config.gemini_api_key" type="password" class="premium-input" placeholder="••••••••••••••••" />
              <div class="input-glow"></div>
            </div>
          </div>

          <div class="form-group toggle-group" style="margin-bottom: 2rem;">
            <label class="toggle-label">
              <input type="checkbox" v-model="config.use_gemini_api_key" class="premium-toggle" />
              <span class="toggle-slider"></span>
              <div class="toggle-text">
                <span class="title">Modo Autônomo API</span>
                <span class="desc">Usar chave em vez da sessão OAuth.</span>
              </div>
            </label>
          </div>

          <div class="form-group">
            <label>Anthropic Claude API Key</label>
            <div class="input-wrapper">
              <input v-model="config.claude_api_key" type="password" class="premium-input" placeholder="••••••••••••••••" :disabled="!config.use_claude_api_key" />
              <div class="input-glow"></div>
            </div>
          </div>

          <div class="form-group toggle-group">
            <label class="toggle-label">
              <input type="checkbox" v-model="config.use_claude_api_key" class="premium-toggle" />
              <span class="toggle-slider"></span>
              <div class="toggle-text">
                <span class="title">Modo Autônomo API</span>
                <span class="desc">Chave API em vez de Login OAuth.</span>
              </div>
            </label>
          </div>

          <button @click="save" class="btn-premium save-btn" style="margin-top: 1rem;">
            <span>SALVAR CHAVES</span>
            <div class="btn-shimmer"></div>
          </button>
        </section>

        <section class="tools-container">
          <h2 class="section-title">Hub de Orquestração</h2>
          <div class="tools-grid">
            <div class="tool-card glass glow-on-hover" :class="{ 'active': status.tools.gemini }">
              <div class="card-header">
                <div class="status-indicator" :class="{ 'online': status.tools.gemini }"></div>
                <h3>Gemini CLI</h3>
              </div>
              <p>IA generativa em tempo real.</p>
              <div v-if="status.tools.gemini" class="autostart-toggle" @click="toggleAutoStart('gemini')">
                <div class="autostart-dot" :class="{ 'on': isAutoStart('gemini') }"></div>
                <span>AUTO-START</span>
              </div>
              <div style="display: flex; gap: 8px;">
                <button @click="install('gemini')" class="tool-btn">SYNC</button>
                <button v-if="status.tools.gemini" @click="setup('gemini')" class="tool-btn" :style="getAuthStyle('gemini')">AUTH</button>
              </div>
            </div>

            <div class="tool-card glass glow-on-hover" :class="{ 'active': status.tools.claude }">
              <div class="card-header">
                <div class="status-indicator" :class="{ 'online': status.tools.claude }"></div>
                <h3>Claude Code</h3>
              </div>
              <p>Engine de codificação.</p>
              <div v-if="status.tools.claude" class="autostart-toggle" @click="toggleAutoStart('claude')">
                <div class="autostart-dot" :class="{ 'on': isAutoStart('claude') }"></div>
                <span>AUTO-START</span>
              </div>
              <div style="display: flex; gap: 8px;">
                <button @click="install('claude')" class="tool-btn">SYNC</button>
                <button v-if="status.tools.claude" @click="setup('claude')" class="tool-btn" :style="getAuthStyle('claude')">AUTH</button>
              </div>
            </div>
          </div>
        </section>
      </div>

      <!-- ABA SEGURANÇA -->
      <section v-if="activeTab === 'seguranca'" class="glass premium-shadow panel-main animate-fade-in" style="border-color: rgba(239, 68, 68, 0.2);">
        <h2 class="section-title" style="color: #ef4444;">🛡️ Segurança da Sinfonia</h2>
        <p class="subtitle" style="font-size: 0.85rem; margin-bottom: 2rem;">Controle as permissões de acesso dos agentes ao seu sistema.</p>

        <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 1.5rem;">
          <div class="form-group toggle-group" v-for="(val, key) in {
            allow_read: 'Permitir Leitura',
            allow_write: 'Permitir Escrita',
            allow_run_commands: 'Executar Comandos',
            full_machine_access: 'Acesso Global'
          }" :key="key">
            <label class="toggle-label" :style="key === 'full_machine_access' ? 'border-color: #7f1d1d' : ''">
              <input type="checkbox" v-model="config.security[key]" class="premium-toggle" />
              <span class="toggle-slider"></span>
              <div class="toggle-text">
                <span class="title">{{ val }}</span>
                <span class="desc">Ativar/Desativar privilégio de {{ val.toLowerCase() }}.</span>
              </div>
            </label>
          </div>
        </div>

        <button @click="save" class="btn-premium save-btn" style="background: linear-gradient(135deg, #450a0a 0%, #1e1b4b 100%); border-color: #ef4444; margin-top: 2rem;">
          <span>CONFIRMAR PERMISSÕES</span>
          <div class="btn-shimmer"></div>
        </button>
      </section>

      <!-- ABA MCP -->
      <section v-if="activeTab === 'mcp'" class="glass premium-shadow panel-main animate-fade-in">
        <h2 class="section-title">Controle MCP</h2>
        <p class="subtitle" style="font-size: 0.85rem; margin-bottom: 2rem;">Adicione novas extensões de contexto via Model Context Protocol.</p>

        <div style="display: flex; flex-direction: column; gap: 2rem;">
          <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 1rem;">
            <div class="form-group" style="margin-bottom: 0;">
              <label>Nome do Servidor</label>
              <div class="input-wrapper">
                <input v-model="mcpName" type="text" class="premium-input" placeholder="ex: github" />
                <div class="input-glow"></div>
              </div>
            </div>
            <div class="form-group" style="margin-bottom: 0;">
              <label>Comando NPX / Bin</label>
              <div class="input-wrapper">
                <input v-model="mcpCommand" type="text" class="premium-input" placeholder="npx -y ..." />
                <div class="input-glow"></div>
              </div>
            </div>
          </div>

          <div style="display: flex; gap: 10px;">
            <button @click="addMCPServer" class="tool-btn" style="flex: 1; border-color: var(--success); color: var(--success);">ADICIONAR SERVIDOR</button>
            <button @click="listMCPServers" class="tool-btn" style="flex: 1;">LISTAR INSTALADOS</button>
          </div>

          <div v-if="showMcpList" class="console-body" style="height: auto; max-height: 250px; padding: 1rem; background: rgba(0,0,0,0.4);">
            <pre style="margin: 0; font-family: 'Fira Code', monospace; font-size: 0.85rem; color: #a1a1aa;">{{ mcpServers }}</pre>
          </div>
        </div>
      </section>
    </div>

    <!-- Console Section -->
    <footer class="console-section" v-show="installStatus !== '' || installLogs.length > 0">
      <div class="console-header glass">
        <div class="header-left">
          <div class="console-icon"></div>
          <span>OPERATIONAL TERMINAL</span>
        </div>
        <div class="pulse-indicator">
          <span>ACTIVE SESSION</span>
          <div class="dot"></div>
        </div>
      </div>
      <div class="console-body" ref="logContainer">
        <div class="scanlines"></div>
        <div v-for="(log, index) in installLogs" :key="index" class="log-entry">
          <span class="timestamp">[{{ new Date().toLocaleTimeString() }}]</span>
          <span class="prompt">LOG_INFO:</span> 
          <span class="message">{{ log }}</span>
        </div>
        <div v-if="installStatus" class="status-entry animate-pulse">
          >> SYSTEM: {{ installStatus }}
        </div>
      </div>
    </footer>
  </main>
</template>

<style scoped>
.settings-view {
  width: 100%;
  max-width: 100%;
  height: 100vh;
  padding: 2rem 3rem;
  overflow-y: auto;
  overflow-x: hidden;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.settings-header {
  margin-bottom: 1rem;
}

.brand-badge {
  display: inline-block;
  background: var(--primary-glow);
  color: var(--primary);
  padding: 4px 12px;
  border-radius: 6px;
  font-size: 0.7rem;
  font-weight: 800;
  letter-spacing: 2px;
  margin-bottom: 1rem;
}

.gradient-text {
  font-size: 2.8rem;
  background: linear-gradient(135deg, #f8fafc 0%, #94a3b8 100%);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
  margin: 0;
}

.subtitle {
  color: var(--text-dim);
  font-size: 1.1rem;
  margin-top: 0.5rem;
}

.content-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.2fr) minmax(0, 1fr);
  gap: 2rem;
  align-items: start;
  width: 100%;
  box-sizing: border-box;
}

.panel-main {
  padding: 2.5rem;
}

.tabs-nav {
  display: flex;
  gap: 10px;
  background: rgba(15, 23, 42, 0.4);
  padding: 6px;
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, 0.05);
  margin-bottom: 1rem;
}

.tab-btn {
  flex: 1;
  padding: 12px;
  border: none;
  background: transparent;
  color: #94a3b8;
  font-weight: 700;
  font-size: 0.8rem;
  letter-spacing: 1px;
  cursor: pointer;
  border-radius: 8px;
  transition: all 0.3s;
}

.tab-btn:hover {
  color: #f8fafc;
  background: rgba(255, 255, 255, 0.03);
}

.tab-btn.active {
  background: var(--primary);
  color: white;
  box-shadow: 0 4px 15px rgba(59, 130, 246, 0.3);
}

.content-grid-tabs {
  width: 100%;
}

.section-title {
  font-size: 0.9rem;
  text-transform: uppercase;
  letter-spacing: 3px;
  color: var(--primary);
  margin-bottom: 2rem;
  display: flex;
  align-items: center;
  gap: 10px;
}

.agents-layout {
  display: grid;
  grid-template-columns: 1.2fr 1fr;
  gap: 2rem;
}

.animate-fade-in {
  animation: fadeIn 0.4s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.section-title::after {
  content: '';
  flex: 1;
  height: 1px;
  background: linear-gradient(90deg, var(--primary-glow), transparent);
}

.form-group {
  margin-bottom: 2rem;
}

label {
  display: block;
  font-size: 0.85rem;
  font-weight: 600;
  color: #f8fafc;
  margin-bottom: 0.75rem;
  letter-spacing: 0.5px;
}

.input-wrapper {
  position: relative;
}

.premium-input {
  width: 100%;
  padding: 14px 18px;
  font-size: 1rem;
  border-radius: 12px;
  font-family: inherit;
  background: rgba(2, 6, 23, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: white;
  transition: all 0.3s;
}

.val-highlight {
  color: var(--primary);
  font-weight: 800;
  margin-left: 10px;
}

.desc-small {
  font-size: 0.7rem;
  color: var(--text-dim);
  margin-top: 5px;
}

.premium-range {
  width: 100%;
  height: 6px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 3px;
  outline: none;
  cursor: pointer;
  -webkit-appearance: none;
}

.premium-range::-webkit-slider-thumb {
  -webkit-appearance: none;
  width: 18px;
  height: 18px;
  background: var(--primary);
  border-radius: 50%;
  cursor: pointer;
  box-shadow: 0 0 10px var(--primary-glow);
  transition: all 0.2s;
}

.premium-range::-webkit-slider-thumb:hover {
  transform: scale(1.2);
  box-shadow: 0 0 20px var(--primary-glow);
}

.input-glow {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  border-radius: 12px;
  pointer-events: none;
  opacity: 0;
  transition: opacity 0.3s;
  box-shadow: 0 0 15px var(--primary-glow);
}

.premium-input:focus + .input-glow {
  opacity: 1;
}

.premium-input:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  background: rgba(0, 0, 0, 0.2);
}

.premium-select {
  appearance: none;
  background-color: rgba(2, 6, 23, 0.5);
  color: #f8fafc;
  border: 1px solid rgba(255, 255, 255, 0.1);
  cursor: pointer;
  background-image: url("data:image/svg+xml;charset=UTF-8,%3csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%233b82f6' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3e%3cpolyline points='6 9 12 15 18 9'%3e%3c/polyline%3e%3c/svg%3e");
  background-repeat: no-repeat;
  background-position: right 1.25rem center;
  background-size: 1.2em;
  padding-right: 3rem;
  transition: all 0.3s;
}

.premium-select:hover {
  background-color: rgba(15, 23, 42, 0.8);
  border-color: rgba(59, 130, 246, 0.4);
}

.premium-select:focus {
  outline: none;
  border-color: var(--primary);
  box-shadow: 0 0 15px rgba(59, 130, 246, 0.2);
}

.premium-select option {
  background-color: #0f172a;
  color: #f8fafc;
  padding: 12px;
  font-size: 0.95rem;
}

/* Custom Premium Toggle */
.toggle-group {
  margin-top: -0.5rem;
}

.toggle-label {
  display: flex;
  align-items: center;
  gap: 15px;
  cursor: pointer;
  padding: 12px 16px;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 12px;
  transition: all 0.3s;
}

.toggle-label:hover {
  background: rgba(255, 255, 255, 0.04);
  border-color: rgba(59, 130, 246, 0.3);
}

.premium-toggle {
  appearance: none;
  width: 44px;
  height: 24px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 20px;
  position: relative;
  outline: none;
  cursor: pointer;
  transition: 0.3s;
  flex-shrink: 0;
}

.premium-toggle::before {
  content: '';
  position: absolute;
  top: 3px;
  left: 3px;
  width: 18px;
  height: 18px;
  background: #f8fafc;
  border-radius: 50%;
  transition: 0.3s cubic-bezier(0.4, 0.0, 0.2, 1);
  box-shadow: 0 2px 4px rgba(0,0,0,0.3);
}

.premium-toggle:checked {
  background: var(--primary);
  box-shadow: 0 0 12px rgba(59, 130, 246, 0.4);
}

.premium-toggle:checked::before {
  left: 23px;
}

.toggle-text {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.toggle-text .title {
  color: #f8fafc;
  font-weight: 700;
  font-size: 0.85rem;
  line-height: 1.2;
}

.toggle-text .desc {
  color: #94a3b8;
  font-size: 0.70rem;
  line-height: 1.3;
}

.save-btn {
  width: 100%;
  position: relative;
  overflow: hidden;
  margin-top: 1rem;
}

/* Tools Grid */
.tools-grid {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.tool-card {
  padding: 1.5rem;
  transition: all 0.4s cubic-bezier(0.16, 1, 0.3, 1);
}

.tool-card:hover {
  transform: scale(1.02) translateX(5px);
  background: rgba(255, 255, 255, 0.04);
}

.tool-card.active {
  border-color: var(--primary);
  background: rgba(59, 130, 246, 0.05);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 0.5rem;
}

.card-header h3 {
  font-size: 1.1rem;
  margin: 0;
}

.status-indicator {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #334155;
}

.status-indicator.online {
  background: var(--success);
  box-shadow: 0 0 12px var(--success);
}

/* Auto-Start Toggle */
.autostart-toggle {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 14px;
  margin-bottom: 1rem;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.3s;
  font-size: 0.78rem;
  font-weight: 600;
  color: #94a3b8;
  user-select: none;
}

.autostart-toggle:hover {
  background: rgba(59, 130, 246, 0.06);
  border-color: rgba(59, 130, 246, 0.25);
  color: #e2e8f0;
}

.autostart-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #334155;
  transition: all 0.3s;
  flex-shrink: 0;
}

.autostart-dot.on {
  background: var(--success);
  box-shadow: 0 0 10px var(--success);
}

.tool-card p {
  color: var(--text-dim);
  font-size: 0.85rem;
  margin: 0.5rem 0 1.25rem 0;
}

.tool-btn {
  background: transparent;
  border: 1px solid var(--border-color);
  color: #f8fafc;
  padding: 8px 16px;
  border-radius: 8px;
  font-size: 0.75rem;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.3s;
  white-space: nowrap;
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.tool-btn:hover {
  border-color: var(--primary);
  background: var(--primary-glow);
}

/* Console Styling */
.console-section {
  margin-top: 2rem;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.console-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  font-size: 0.7rem;
  font-weight: 800;
  letter-spacing: 1px;
  border-radius: 12px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--primary);
}

.console-icon {
  width: 8px;
  height: 8px;
  background: var(--primary);
  clip-path: polygon(0% 0%, 100% 50%, 0% 100%);
}

.pulse-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--success);
}

.pulse-indicator .dot {
  width: 6px;
  height: 6px;
  background: var(--success);
  border-radius: 50%;
  animation: pulse 1s infinite;
}

.console-body {
  background: #020617;
  border: 1px solid var(--border-color);
  border-radius: 16px;
  height: 250px;
  overflow-y: auto;
  padding: 1.5rem;
  position: relative;
  font-family: 'Fira Code', monospace;
  font-size: 0.85rem;
  line-height: 1.6;
}

.scanlines {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  background: linear-gradient(rgba(18, 16, 16, 0) 50%, rgba(0, 0, 0, 0.25) 50%), linear-gradient(90deg, rgba(255, 0, 0, 0.06), rgba(0, 255, 0, 0.02), rgba(0, 0, 255, 0.06));
  background-size: 100% 2px, 3px 100%;
  pointer-events: none;
  z-index: 10;
}

.log-entry {
  display: flex;
  gap: 12px;
  margin-bottom: 0.5rem;
}

.timestamp { color: #475569; }
.prompt { color: var(--primary); font-weight: bold; }
.message { color: #e2e8f0; }

.status-entry {
  color: var(--primary);
  font-weight: bold;
  margin-top: 1rem;
  border-top: 1px solid var(--border-color);
  padding-top: 0.5rem;
}

/* Destaque quando o console está processando */
.console-section {
  transition: all 0.5s ease;
  border-radius: 16px;
  overflow: hidden;
}

.console-section:has(.animate-pulse) {
  box-shadow: 0 0 30px rgba(59, 130, 246, 0.2);
  transform: translateY(-5px);
  border: 1px solid rgba(59, 130, 246, 0.3);
}

@keyframes pulse {
  0% { transform: scale(1); opacity: 1; }
  50% { transform: scale(1.5); opacity: 0.6; }
  100% { transform: scale(1); opacity: 1; }
}

@keyframes pulse-anim {
  0% { opacity: 0.5; }
  50% { opacity: 1; }
  100% { opacity: 0.5; }
}

.animate-pulse {
  animation: pulse-anim 2s infinite ease-in-out;
}
</style>
