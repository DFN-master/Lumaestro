import { defineStore } from 'pinia';
import { ref } from 'vue';
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
  const isRealPTY = ref(false);
  const activeAgent = ref(null);
  const runningSessions = ref([]);

  const initListeners = () => {
    EventsOn('agent:log', (log) => {
      console.log("[Store] Evento agent:log:", log);
      
      if (log.source === 'CRAWLER') {
         messages.value.push({ role: 'assistant', text: log.content, mode: 'system' });
         return;
      }

      // Se for resposta da IA, anexa ou cria nova bolha
      if (log.role === 'assistant') {
        const lastMsg = messages.value[messages.value.length - 1];
        if (lastMsg && lastMsg.role === 'assistant' && lastMsg.mode !== 'system') {
           lastMsg.text += log.content;
        } else {
           messages.value.push({ 
             role: 'assistant', 
             text: log.content, 
             agent: log.agent || activeAgent.value || 'Maestro' 
           });
        }
      }
      isThinking.value = false; // IA respondeu, para de pensar
    });

    EventsOn('terminal:started', (info) => {
      isRealPTY.value = !!info?.isRealPTY;
      const agent = info?.agent;
      if (agent && !runningSessions.value.includes(agent)) runningSessions.value.push(agent);
      if (!activeAgent.value && agent) {
         activeAgent.value = agent;
         isTerminalMode.value = true;
      }
    });

    EventsOn('terminal:closed', (agent) => {
      runningSessions.value = runningSessions.value.filter(a => a !== agent);
      if (activeAgent.value === agent) {
        if (runningSessions.value.length > 0) activeAgent.value = runningSessions.value[0];
        else { activeAgent.value = null; isTerminalMode.value = false; }
      }
      isThinking.value = false;
      messages.value.push({ role: 'assistant', text: `Sessão ${agent} encerrada.`, mode: 'system' });
    });

    // Escuta a saída bruta do terminal e injeta no fluxo visual (removendo ANSI)
    EventsOn('terminal:output', (payload) => {
      if (!payload || !payload.data) return;
      
      const binary = atob(payload.data);
      const bytes = new Uint8Array(binary.length);
      for (let i = 0; i < binary.length; i++) {
        bytes[i] = binary.charCodeAt(i);
      }
      const textChunk = new TextDecoder('utf-8').decode(bytes);
      
      // Expressão Regular agressiva para limpar ANSI escape codes, cores, cursores e apagar backspaces
      let cleanText = textChunk
        .replace(/[\u001b\u009b][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-9A-ORZcf-nqry=><]/g, '')
        .replace(/\x1B\[[0-9;]*[a-zA-Z]/g, '')
        .replace(/\r/g, '') // remove carriage returns
        .replace(/\x08/g, '') // remove backspaces físicos temporariamente pra não poluir (spinners)
        .replace(/\x1b/g, ''); // remove escapes isolados

      if (!cleanText) return;

      const lastMsg = messages.value[messages.value.length - 1];
      if (lastMsg && lastMsg.role === 'assistant' && lastMsg.mode === 'terminal_echo' && lastMsg.agent === payload.agent) {
        lastMsg.text += cleanText;
      } else {
        messages.value.push({ 
          role: 'assistant', 
          text: cleanText, 
          agent: payload.agent || activeAgent.value,
          mode: 'terminal_echo'
        });
      }
    });
  };

  const ask = async (agent, prompt) => {
    console.log("[Store] Enviando Ask para:", agent, prompt);
    // 1. Adiciona a mensagem do usuário IMEDIATAMENTE na tela
    messages.value.push({ role: 'user', text: prompt });
    isThinking.value = true;
    
    // Atualiza o agente ativo local para refletir na UI o badge correto
    activeAgent.value = agent;

    try {
      await safeCall('main', 'AskAgent', agent, prompt);
    } catch (err) {
      messages.value.push({ role: 'assistant', text: `❌ Erro de conexão: ${err}`, mode: 'system' });
      isThinking.value = false;
    }
  };

  const startSession = async (agent) => {
    isThinking.value = true;
    isTerminalMode.value = true;
    activeAgent.value = agent;
    messages.value.push({ role: 'user', text: `/cmd ${agent}`, mode: 'system' });

    try {
      await safeCall('main', 'StartAgentSession', agent);
    } catch (err) {
      messages.value.push({ role: 'assistant', text: `❌ Falha PTY: ${err}`, mode: 'system' });
      isThinking.value = false;
    }
  };

  const sendInput = async (agent, text) => {
    // Registra a mensagem no log local para o chat visual:
    messages.value.push({ role: 'user', text: text });
    
    console.log("[Store] Enviando Input Terminal:", agent, text);
    return await safeCall('main', 'SendAgentInput', agent, text);
  };

  const runScan = async () => {
     messages.value.push({ role: 'assistant', text: "Iniciando indexação...", mode: 'system' });
     await safeCall('main', 'ScanVault');
  };

  return {
    messages, isThinking, isTerminalMode, isRealPTY, activeAgent, runningSessions,
    initListeners, ask, startSession, runScan, sendInput
  };
});
