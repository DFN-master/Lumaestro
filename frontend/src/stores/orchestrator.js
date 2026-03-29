import { defineStore } from 'pinia';
import { ref } from 'vue';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { 
  AskAgent, 
  ScanVault, 
  StartAgentSession, 
  StopAgentSession, 
  SendAgentInput 
} from '../../wailsjs/go/main/App';

export const useOrchestratorStore = defineStore('orchestrator', () => {
  // --- Estados Reativos (State) ---
  const messages = ref([]);
  const isThinking = ref(false);
  const isTerminalMode = ref(false);
  const isRealPTY = ref(false);
  const activeAgent = ref(null);
  const outputBuffer = ref("");

  // --- Inicialização de Eventos (Ouvir o Backend Go) ---
  const initListeners = () => {
    // 1. Escuta logs da IA (Streaming)
    EventsOn('agent:log', (log) => {
      console.log("[Store] Recebido agent:log:", log);
      
      // Filtra logs de sistema (Scan, etc.) do Chat normal
      if (log.source === 'CRAWLER') {
         messages.value.push({ role: 'assistant', text: log.content, mode: 'system' });
         return;
      }

      // Se for log do agente, tenta acumular na última bolha ou criar nova
      if (log.role === 'assistant') {
        const lastMsg = messages.value[messages.value.length - 1];
        if (lastMsg && lastMsg.role === 'assistant' && lastMsg.mode !== 'system') {
           lastMsg.text += log.content;
        } else {
           messages.value.push({ role: 'assistant', text: log.content, agent: log.agent || activeAgent.value });
        }
      }
    });

    // 2. Escuta status da Sessão PTY
    EventsOn('terminal:started', (info) => {
      isRealPTY.value = !!info?.isRealPTY;
      console.log("[Store] PTY Iniciado. Real:", isRealPTY.value);
    });

    // 3. Escuta logs brutos de execução do terminal
    EventsOn('execution:log', (log) => {
       if (log.source === 'SYSTEM') {
         messages.value.push({ role: 'assistant', text: `⚙️ ${log.content}`, mode: 'system' });
       }
    });
  };

  // --- Ações (Actions) ---
  
  // Enviar comando para IA (Modo One-Shot/Resumo)
  const ask = async (prompt) => {
    isThinking.value = true;
    messages.value.push({ role: 'user', text: prompt });
    
    try {
      const response = await AskAgent(prompt);
      // O AskAgent do Go retorna apenas "Orquestrando...", os logs reais vêm via Evento
      console.log("[Store] Resposta inicial:", response);
    } catch (err) {
      messages.value.push({ role: 'assistant', text: `❌ Erro: ${err}`, mode: 'system' });
      isThinking.value = false;
    }
  };

  // Iniciar Sessão Interativa (Terminal Real)
  const startSession = async (agent) => {
    isThinking.value = true;
    isTerminalMode.value = true;
    activeAgent.value = agent;
    
    messages.value.push({ 
      role: 'user', 
      text: `/cmd ${agent}`,
      mode: 'system' 
    });

    try {
      const result = await StartAgentSession(agent);
      console.log("[Store] Sessão Iniciada:", result);
    } catch (err) {
      messages.value.push({ role: 'assistant', text: `❌ Falha ao iniciar sessão PTY: ${err}`, mode: 'system' });
      isTerminalMode.value = false;
      isThinking.value = false;
    }
  };

  // Parar Sessão
  const stopSession = async () => {
    await StopAgentSession();
    isTerminalMode.value = false;
    isRealPTY.value = false;
    isThinking.value = false;
    activeAgent.value = null;
    
    messages.value.push({ role: 'assistant', text: "Sessão encerrada.", mode: 'system' });
  };

  // Enviar Input para o PTY ativo
  const sendInput = async (text) => {
    if (!isTerminalMode.value) return;
    return await SendAgentInput(text);
  };

  // Escanear Vault Obsidian
  const runScan = async () => {
     messages.value.push({ role: 'assistant', text: "Iniciando indexação do Vault...", mode: 'system' });
     await ScanVault();
  };

  return {
    messages,
    isThinking,
    isTerminalMode,
    isRealPTY,
    activeAgent,
    initListeners,
    ask,
    startSession,
    stopSession,
    sendInput,
    runScan
  };
});
