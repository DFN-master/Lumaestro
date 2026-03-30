import { defineStore } from 'pinia';
import { ref, watch } from 'vue';
import { EventsOn } from '../../wailsjs/runtime/runtime';

// Helper para chamar funções do Wails com segurança
const safeCall = async (pkg, func, ...args) => {
  try {
    if (window.go && window.go.main && window.go.main.App && window.go.main.App[func]) {
      return await window.go.main.App[func](...args);
    }
    console.warn(`[Wails SafeCall] Função ${func} não encontrada`);
    return null;
  } catch (err) {
    console.error(`[Wails SafeCall] Erro ao chamar ${func}:`, err);
    throw err;
  }
};

export const useOrchestratorStore = defineStore('orchestrator', () => {
  const messages = ref([]);
  const isThinking = ref(false);
  const isTerminalMode = ref(false);
  const activeAgent = ref(null);
  const runningSessions = ref([]);
  
  // Estado para histórico e checkpoints (Sinfonias)
  const sessions = ref([]);
  const currentACPID = ref(null);
  
  // Estado para revisões de segurança pendentes
  const pendingReview = ref(null);

  const initListeners = () => {
    // 1. Logs Estruturados da IA (ACP)
    EventsOn('agent:log', (log) => {
      console.log("[Store] Logs ACP:", log);
      
      if (log.source === 'SYSTEM') {
        messages.value.push({ role: 'assistant', text: log.content, mode: 'system' });
        return;
      }

      // Tratamento de mensagens da IA
      const lastMsg = messages.value[messages.value.length - 1];
      
      // Se for um log de progresso (ex: [client/readFile] ... )
      if (log.content.startsWith('[')) {
          // Poderíamos criar um "Progresso de Ferramenta" aqui
          console.log("[ACP Progress]", log.content);
      }

      if (lastMsg && lastMsg.role === 'assistant' && lastMsg.mode !== 'system' && lastMsg.agent === log.source) {
          lastMsg.text += log.content;
      } else {
          messages.value.push({ 
            role: 'assistant', 
            text: log.content, 
            agent: log.source || 'Maestro' 
          });
      }
      
      isThinking.value = false;
    });

    // 2. Pedidos de Revisão Manual (Security Hands)
    EventsOn('agent:review_request', (review) => {
      console.log("[Store] Pedido de Revisão:", review);
      pendingReview.value = review;
    });

    EventsOn('terminal:started', (info) => {
      const agent = info?.agent;
      if (agent && !runningSessions.value.includes(agent)) runningSessions.value.push(agent);
      activeAgent.value = agent;
      isTerminalMode.value = true;
    });

    EventsOn('terminal:closed', (agent) => {
      runningSessions.value = runningSessions.value.filter(a => a !== agent);
      if (activeAgent.value === agent) {
        activeAgent.value = runningSessions.value[0] || null;
        if (!activeAgent.value) isTerminalMode.value = false;
      }
      isThinking.value = false;
    });

    // 3. Detecção de Erros de Autenticação (Login)
    EventsOn('agent:login_required', async (agent) => {
      console.warn("[Store] Login necessário para:", agent);
      isThinking.value = false;
      messages.value.push({ 
        role: 'assistant', 
        text: `⚠️ O ${agent} precisa de autenticação. Abrindo terminal de login...`, 
        mode: 'system' 
      });
      // Dispara o SetupTool (terminal externo) para o agente
      await safeCall('main', 'SetupTool', agent);
    });

    // 🚀 Sincronização de Sinfonias (Checkpoints): Quando o turno termina, atualizamos a lista lateral
    window.runtime.EventsOn("agent:turn_complete", (agent) => {
      console.log(`[Store] Turno concluído para ${agent}. Atualizando Sinfonias...`);
      fetchSessions(agent);
    });

    // 4. Watcher de Resiliência: Mantém a UI síncrona com a realidade do Backend
    watch(runningSessions, (sessions) => {
      console.log("[Store] Resiliência: Sessões Ativas:", sessions);
      if (sessions.length === 0) {
        console.warn("[Store] Nenhuma sessão ativa. Limpando estados fantasmas.");
        activeAgent.value = null;
        isThinking.value = false;
        isTerminalMode.value = false;
      } else if (activeAgent.value && !sessions.includes(activeAgent.value)) {
        // Se o agente ativo atual morreu, foca no próximo disponível
        activeAgent.value = sessions[0];
      }
    }, { deep: true });
  };

  const ask = async (agent, prompt) => {
    messages.value.push({ role: 'user', text: prompt });
    isThinking.value = true;
    activeAgent.value = agent;

    try {
      await safeCall('main', 'AskAgent', agent, prompt);
    } catch (err) {
      messages.value.push({ role: 'assistant', text: `❌ Erro: ${err}`, mode: 'system' });
      isThinking.value = false;
    }
  };

  const startSession = async (agent) => {
    console.log(`[Store] Iniciando Sessão ACP para: ${agent}`);
    isThinking.value = true;
    isTerminalMode.value = true;
    activeAgent.value = agent;
    
    try {
      await safeCall('main', 'StartAgentSession', agent);
      
      // 🚀 Após iniciar o processo, tentamos buscar o histórico
      await fetchSessions(agent);
      
      // Se houver histórico e não estivermos carregando um específico,
      // sugerimos o último checkpoint encontrado.
      if (sessions.value.length > 0 && !currentACPID.value) {
          const last = sessions.value[0]; 
          currentACPID.value = last.sessionId;
      }
      
    } catch (err) {
      messages.value.push({ role: 'assistant', text: `❌ Falha: ${err}`, mode: 'system' });
      isThinking.value = false;
    }
  };

  const fetchSessions = async (agent) => {
    if (!agent) return;
    try {
      const list = await safeCall('main', 'ListAgentSessions', agent);
      if (list) {
          // Ordenar por data (mais recente primeiro)
          sessions.value = list.sort((a, b) => new Date(b.updatedAt) - new Date(a.updatedAt));
      }
    } catch (err) {
      console.error("[Store] Erro ao buscar sessões:", err);
    }
  };

  const loadSession = async (agent, acpID) => {
    console.log(`[Store] Carregando Sinfonia: ${acpID}`);
    isThinking.value = true;
    currentACPID.value = acpID;
    messages.value = []; // Limpa o chat para receber o novo contexto restaurado
    
    try {
      await safeCall('main', 'LoadAgentSession', agent, acpID);
      await fetchSessions(agent); // Atualiza a lista lateral
    } catch (err) {
      messages.value.push({ role: 'assistant', text: `❌ Erro ao carregar: ${err}`, mode: 'system' });
      isThinking.value = false;
    }
  };

  const newSession = async (agent) => {
    console.log(`[Store] Iniciando nova Sinfonia personalizada...`);
    isThinking.value = true;
    currentACPID.value = null;
    messages.value = [];
    
    try {
      await safeCall('main', 'NewAgentSession', agent);
      await fetchSessions(agent);
    } catch (err) {
      messages.value.push({ role: 'assistant', text: `❌ Erro ao criar: ${err}`, mode: 'system' });
      isThinking.value = false;
    }
  };

  const sendInput = async (agent, text) => {
    console.log(`[Store] Enviando Input ACP (${agent}): ${text}`);
    messages.value.push({ role: 'user', text: text });
    isThinking.value = true; // Feedback visual imediato

    // Timeout de segurança: se a IA hibernar por 25s, destravamos a UI
    const safetyTimeout = setTimeout(() => {
      if (isThinking.value) {
        console.warn("[Store] Timeout de segurança atingido. Destravando UI.");
        isThinking.value = false;
        messages.value.push({ 
          role: 'assistant', 
          text: "⚠️ A Sinfonia está demorando para responder. O processo ainda pode estar ativo no background.", 
          mode: 'system' 
        });
      }
    }, 25000);

    try {
      const resp = await safeCall('main', 'SendAgentInput', agent, text);
      return resp;
    } catch (err) {
      console.error('[Store] Erro ao enviar input:', err);
      isThinking.value = false;
      clearTimeout(safetyTimeout);
    }
  };

  const submitReview = async (approved) => {
    if (!pendingReview.value) return;
    const id = pendingReview.value.id;
    pendingReview.value = null;
    try {
      await safeCall('main', 'SubmitReview', id, approved);
    } catch (err) {
      console.error("Falha ao enviar review:", err);
    }
  };

  const switchAgent = (agent) => {
    activeAgent.value = agent;
  };

  const stopSession = async () => {
    if (!activeAgent.value) return;
    try {
      await safeCall('main', 'StopAgentSession', activeAgent.value);
    } catch (err) {
      console.error("Erro ao fechar sessão:", err);
    }
  };

  const isSidebarOpen = ref(false);
  const toggleSidebar = () => {
    isSidebarOpen.value = !isSidebarOpen.value;
    console.log(`[Store] Histórico ${isSidebarOpen.value ? 'Aberto' : 'Fechado'}`);
  };

  return {
    messages, isThinking, isTerminalMode, activeAgent, runningSessions, pendingReview,
    sessions, currentACPID, isSidebarOpen,
    initListeners, ask, startSession, sendInput, submitReview, switchAgent, stopSession,
    fetchSessions, loadSession, newSession, toggleSidebar
  };
});
