<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { GetConfig, SaveConfig, GetToolsStatus, InstallTool, SetupTool, AddGeminiAccount, SwitchGeminiAccount, LoginGeminiAccount, AddMCPServer, ListMCPServers } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime'

const config = ref({
  obsidian_vault_path: '',
  qdrant_url: '',
  gemini_api_key: '',
  use_gemini_api_key: false,
  gemini_accounts: [],
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
  if (savedConfig) {
    config.value = savedConfig
  }
  
  refreshStatus()

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
  return 'border-color: #3b82f6;'
}

const mcpName = ref('')
const mcpCommand = ref('')
const mcpServers = ref('')
const showMcpList = ref(false)

const addMCPServer = async () => {
  if (!mcpName.value || !mcpCommand.value) {
    alert("Preencha o Nome e o Comando para o MCP")
    return
  }
  installLogs.value = []
  installStatus.value = `Instalando servidor MCP: ${mcpName.value}...`
  scrollToConsole()
  const res = await AddMCPServer(mcpName.value, mcpCommand.value)
  installStatus.value = "Instalação do MCP Finalizada."
  mcpName.value = ''
  mcpCommand.value = ''
  alert("Retorno do Terminal:\n" + res)
}

const listMCPServers = async () => {
  const res = await ListMCPServers()
  mcpServers.value = res
  showMcpList.value = true
}

// Funções de Multi-Conta
const handleAddAccount = async () => {
  if (!newAccName.value) return
  await AddGeminiAccount(newAccName.value)
  newAccName.value = ''
  const cfg = await GetConfig()
  if (cfg) config.value = cfg
}

const handleLoginAccount = async (name) => {
  await LoginGeminiAccount(name)
}

const handleSwitchAccount = async (name) => {
  await SwitchGeminiAccount(name)
  const cfg = await GetConfig()
  if (cfg) config.value = cfg
}

const activeTab = ref('geral')
const newAccName = ref('')
</script>

<template>
  <main class="settings-view animate-fade-up">
    <div class="settings-header">
      <div class="brand-badge pulse-aura">LUMAESTRO PREMIER</div>
      <h1 class="gradient-text">Orquestração de IAs</h1>
      <p class="subtitle">Configurações globais e gerenciamento de identidades.</p>
    </div>

    <div class="tabs-nav-glass">
      <button v-for="tab in ['geral', 'agentes', 'contas', 'seguranca', 'mcp']" 
              :key="tab"
              @click="activeTab = tab" 
              :class="{ 'active': activeTab === tab }" 
              class="tab-btn-premium">
        {{ tab === 'contas' ? 'CONTAS GEMINI 💎' : tab.toUpperCase() }}
      </button>
    </div>

    <div class="content-viewport">
      <!-- ABA GERAL -->
      <section v-if="activeTab === 'geral'" class="glass-panel animate-slide-up">
        <h2 class="section-title">Base da Sinfonia</h2>
        
        <div class="form-grid">
          <div class="premium-form-group">
            <label>Idioma Nativo do Agente</label>
            <select v-model="config.agent_language" class="maestro-input">
              <option value="Português do Brasil">Português (Brasil)</option>
              <option value="English">English</option>
              <option value="Español">Español</option>
              <option value="Français">Français</option>
              <option value="Deutsch">Deutsch</option>
              <option value="Italiano">Italiano</option>
              <option value="日本語 (Japanese)">日本語 (Japonês)</option>
            </select>
          </div>

          <div class="premium-form-group">
            <label>Caminho do Obsidian Vault</label>
            <input v-model="config.obsidian_vault_path" type="text" class="maestro-input" placeholder="C:\Users\...\Obsidian" />
          </div>
        </div>

        <div class="premium-form-group">
          <label>Alcance da Teia (Vizinhos): <span class="highlight-val">{{ config.graph_neighbor_limit }}</span></label>
          <input v-model.number="config.graph_neighbor_limit" type="range" min="1" max="25" step="1" class="maestro-range" />
        </div>

        <div class="premium-form-group">
          <label>URL do Qdrant Cloud</label>
          <input v-model="config.qdrant_url" type="text" class="maestro-input" placeholder="https://..." />
        </div>

        <button @click="save" class="btn-glow-blue">SALVAR ALTERAÇÕES GERAIS</button>
      </section>

      <!-- ABA AGENTES -->
      <div v-if="activeTab === 'agentes'" class="agents-layout-grid animate-slide-up">
        <section class="glass-panel">
          <h2 class="section-title">Autenticação Legada (API)</h2>
          
          <div class="premium-form-group">
            <label>Gemini API Key</label>
            <input v-model="config.gemini_api_key" type="password" class="maestro-input" placeholder="••••••••" />
          </div>

          <div class="toggle-card-premium">
             <label class="hybrid-toggle">
                <input type="checkbox" v-model="config.use_gemini_api_key" />
                <span class="t-slider"></span>
                <div class="t-info">
                   <p class="t-title">Modo Autônomo API</p>
                   <p class="t-desc">Usar chave em vez de OAuth.</p>
                </div>
             </label>
          </div>

          <div class="premium-form-group">
            <label>Claude API Key</label>
            <input v-model="config.claude_api_key" type="password" class="maestro-input" placeholder="••••••••" :disabled="!config.use_claude_api_key" />
          </div>

          <div class="toggle-card-premium">
             <label class="hybrid-toggle">
                <input type="checkbox" v-model="config.use_claude_api_key" />
                <span class="t-slider"></span>
                <div class="t-info">
                   <p class="t-title">Claude API Mode</p>
                   <p class="t-desc">Habilitar chave direta Anthropic.</p>
                </div>
             </label>
          </div>

          <button @click="save" class="btn-glow-blue" style="margin-top: 1.5rem; width: 100%;">SALVAR CHAVES</button>
        </section>

        <section class="engines-panel">
          <h2 class="section-title">Hub de Motores</h2>
          <div class="engine-cards-stack">
             <div v-for="tool in ['gemini', 'claude']" :key="tool" class="engine-unit glass">
                <div class="unit-head">
                   <span class="unit-icon">{{ tool === 'gemini' ? '⚡' : '🦾' }}</span>
                   <h4>{{ tool.toUpperCase() }}</h4>
                </div>
                <div class="unit-actions">
                   <button @click="install(tool)" class="unit-btn">SYNC</button>
                   <button v-if="status.tools[tool]" @click="setup(tool)" class="unit-btn auth" :style="getAuthStyle(tool)">{{ getAuthLabel(tool) }}</button>
                </div>
                <div class="unit-footer">
                   <label class="mini-toggle">
                      <input type="checkbox" :checked="isAutoStart(tool)" @change="toggleAutoStart(tool)" />
                      <span class="m-slider"></span> AUTO-START
                   </label>
                </div>
             </div>
          </div>
        </section>
      </div>

      <!-- ABA CONTAS GEMINI -->
      <section v-if="activeTab === 'contas'" class="glass-panel animate-slide-up">
         <div class="header-with-action">
            <div>
               <h2 class="section-title">Identidades OAuth</h2>
               <p class="subtitle-maestro">Alterne entre logins Google isolados.</p>
            </div>
            <div class="quick-add">
               <input v-model="newAccName" placeholder="Nome Perfil..." class="maestro-input-compact" @keyup.enter="handleAddAccount"/>
               <button @click="handleAddAccount" class="add-circle">+</button>
            </div>
         </div>

         <div class="accounts-grid-premium">
            <div v-for="acc in config.gemini_accounts" :key="acc.name" class="profile-card" :class="{ 'active-profile': acc.active }">
               <div class="profile-header">
                  <div class="avatar-glow">{{ acc.name[0].toUpperCase() }}</div>
                  <div class="profile-meta">
                     <span class="name">{{ acc.name }}</span>
                     <span class="status-chip">{{ acc.active ? 'ATURA SESSÃO' : 'STANDBY' }}</span>
                  </div>
               </div>
               <div class="profile-actions">
                  <button @click="handleLoginAccount(acc.name)" class="btn-util login">LOGAR 🔑</button>
                  <button v-if="!acc.active" @click="handleSwitchAccount(acc.name)" class="btn-util activate">ATIVAR</button>
               </div>
            </div>
         </div>
      </section>

      <!-- ABA SEGURANÇA (TOTAL RESTORED) -->
      <section v-if="activeTab === 'seguranca'" class="glass-panel animate-slide-up" style="border-color: rgba(239, 68, 68, 0.1);">
         <h2 class="section-title" style="color: #ef4444;">🛡️ Segurança da Sinfonia</h2>
         <div class="security-grid-comprehensive">
             <div v-for="(label, key) in {
                allow_read: 'Permitir Leitura',
                allow_write: 'Permitir Escrita',
                allow_create: 'Criar Arquivos',
                allow_delete: 'Excluir Arquivos',
                allow_move: 'Mover/Renomear',
                allow_run_commands: 'Comandos Shell',
                full_machine_access: 'Acesso Global'
             }" :key="key" class="sec-card glass">
                <div class="sec-info">
                   <h5>{{ label }}</h5>
                   <p>{{ key === 'full_machine_access' ? 'Cuidado: Risco Máximo' : 'Permissão de ' + label.toLowerCase() }}</p>
                </div>
                <input type="checkbox" v-model="config.security[key]" class="maestro-toggle-core" />
             </div>
         </div>
         <button @click="save" class="btn-glow-red" style="margin-top: 2rem;">REVALIDAR PERMISSÕES</button>
      </section>

      <!-- ABA MCP -->
      <section v-if="activeTab === 'mcp'" class="glass-panel animate-slide-up">
        <h2 class="section-title">Model Context Protocol (MCP)</h2>
        <div class="mcp-restored-form">
           <div class="form-grid">
              <input v-model="mcpName" placeholder="Identificador (ex: postgres)" class="maestro-input" />
              <input v-model="mcpCommand" placeholder="Command (ex: npx -y ...)" class="maestro-input" />
           </div>
           <div class="mcp-actions-row">
              <button @click="addMCPServer" class="btn-glow-blue" style="flex: 1;">INSTALAR SERVIDOR</button>
              <button @click="listMCPServers" class="btn-outline" style="flex: 1;">LISTAR REGISTRADOS</button>
           </div>
           <pre v-if="showMcpList" class="mcp-output-box">{{ mcpServers }}</pre>
        </div>
      </section>
    </div>

    <!-- Terminal de Logs (Restored Logic) -->
    <footer class="maestro-terminal-v2" v-show="installStatus !== '' || installLogs.length > 0">
      <div class="t-bar">
         <span class="t-title">SYSTEM_ORCHESTRATOR_OUTPUT</span>
         <div class="t-pulse"><span></span> ACTIVE</div>
      </div>
      <div class="t-contents" ref="logContainer">
        <div v-for="(log, i) in installLogs" :key="i" class="t-entry">> {{ log }}</div>
        <div v-if="installStatus" class="t-status">>> {{ installStatus }}</div>
      </div>
    </footer>
  </main>
</template>

<style scoped>
/* CSS PREMIER CONSOLIDATED & RESTORED */
.settings-view { padding: 4rem; color: #f8fafc; background: #020617; min-height: 100vh; font-family: 'Inter', sans-serif; overflow-y: auto; }
.gradient-text { font-size: 3.5rem; font-weight: 800; background: linear-gradient(135deg, #fff 0%, #64748b 100%); -webkit-background-clip: text; color: transparent; margin-bottom: 2rem; }
.brand-badge { background: rgba(59, 130, 246, 0.1); color: #60a5fa; padding: 4px 12px; border-radius: 6px; font-size: 0.7rem; font-weight: 800; display: inline-block; margin-bottom: 1rem; }

.tabs-nav-glass { display: flex; gap: 8px; margin-bottom: 3rem; background: rgba(255,255,255,0.02); padding: 8px; border-radius: 12px; width: fit-content; }
.tab-btn-premium { background: none; border: none; padding: 12px 24px; color: #64748b; font-weight: 700; cursor: pointer; border-radius: 8px; transition: 0.3s; }
.tab-btn-premium.active { color: #fff; background: rgba(255,255,255,0.05); box-shadow: 0 4px 15px rgba(0,0,0,0.2); }

.glass-panel { background: rgba(255,255,255,0.02); border: 1px solid rgba(255,255,255,0.05); border-radius: 24px; padding: 3rem; backdrop-filter: blur(20px); }
.section-title { font-size: 1rem; text-transform: uppercase; color: #3b82f6; letter-spacing: 3px; margin-bottom: 2rem; }
.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 2rem; margin-bottom: 2rem; }

.maestro-input { background: #000; border: 1px solid #1e293b; color: #fff; padding: 16px; border-radius: 12px; width: 100%; }
.maestro-range { width: 100%; height: 6px; -webkit-appearance: none; background: #1e293b; border-radius: 10px; margin: 15px 0; }
.maestro-range::-webkit-slider-thumb { -webkit-appearance: none; width: 20px; height: 20px; background: #3b82f6; border-radius: 50%; box-shadow: 0 0 10px #3b82f6; }

.btn-glow-blue { background: #3b82f6; border: none; color: #fff; font-weight: 800; padding: 16px 32px; border-radius: 12px; cursor: pointer; transition: 0.3s; }
.btn-glow-blue:hover { transform: translateY(-2px); box-shadow: 0 5px 20px rgba(59, 130, 246, 0.4); }

.security-grid-comprehensive { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 1rem; }
.sec-card { padding: 1.5rem; display: flex; justify-content: space-between; align-items: center; background: rgba(0,0,0,0.2); border-radius: 16px; border: 1px solid #1e293b; }
.sec-info h5 { margin: 0; font-size: 0.9rem; }
.sec-info p { margin: 4px 0 0; font-size: 0.7rem; color: #64748b; }

.engine-cards-stack { display: flex; flex-direction: column; gap: 1rem; }
.engine-unit { padding: 1.5rem; display: flex; flex-direction: column; gap: 1rem; border-radius: 16px; border: 1px solid #1e293b; }
.unit-head { display: flex; align-items: center; gap: 12px; }
.unit-actions { display: flex; gap: 8px; }
.unit-btn { background: #000; border: 1px solid #1e293b; color: #fff; padding: 8px 16px; border-radius: 8px; font-weight: 700; font-size: 0.7rem; cursor: pointer; flex: 1; }

.maestro-terminal-v2 { position: fixed; bottom: 2rem; right: 2rem; width: 450px; background: #000; border: 1px solid #1e293b; border-radius: 16px; box-shadow: 0 20px 40px #000; }
.t-bar { padding: 12px; background: #050510; display: flex; justify-content: space-between; font-size: 0.6rem; color: #475569; font-weight: 800; border-bottom: 1px solid #1e293b; }
.t-contents { padding: 1.5rem; font-family: 'Fira Code', monospace; font-size: 0.75rem; max-height: 250px; overflow-y: auto; color: #94a3b8; }
.t-entry { margin-bottom: 4px; }
.t-status { color: #3b82f6; font-weight: 800; margin-top: 1rem; }

.animate-slide-up { animation: slideUp 0.6s cubic-bezier(0.16, 1, 0.3, 1); }
@keyframes slideUp { from { opacity: 0; transform: translateY(30px); } to { opacity: 1; transform: translateY(0); } }

.mcp-output-box { background: #000; padding: 1.5rem; border-radius: 12px; margin-top: 1rem; overflow: auto; color: #4ade80; font-size: 0.8rem; border: 1px solid #1e293b; }
</style>
