<script setup>
import { ref, onMounted, watch } from 'vue'
import MarkdownIt from 'markdown-it'

const props = defineProps({
  title: String,
  content: String,
  isOpen: Boolean
})

const emit = defineEmits(['close'])

const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true,
})

const renderedContent = ref('')

watch(() => props.content, (newContent) => {
  if (newContent) {
    renderedContent.value = md.render(newContent)
  }
}, { immediate: true })

const close = () => {
  emit('close')
}
</script>

<template>
  <Transition name="fade-scale">
    <div v-if="isOpen" class="doc-viewer-overlay" @click.self="close">
      <div class="doc-modal glass">
        <header class="doc-header">
          <div class="header-info">
            <span class="doc-icon">📖</span>
            <h3>{{ title }}</h3>
          </div>
          <button class="close-btn" @click="close">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2.5">
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
          </button>
        </header>

        <div class="doc-body markdown-body" v-html="renderedContent"></div>
        
        <footer class="doc-footer">
           <p>Documentação Sincronizada em Tempo Real</p>
           <button class="done-btn" @click="close">FECHAR</button>
        </footer>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.doc-viewer-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(10px);
  z-index: 2000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px;
}

.doc-modal {
  width: 100%;
  max-width: 900px;
  max-height: 85vh;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 24px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 30px 60px rgba(0, 0, 0, 0.6);
  animation: modalSpawn 0.4s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}

@keyframes modalSpawn {
  from { opacity: 0; transform: scale(0.95) translateY(20px); }
  to { opacity: 1; transform: scale(1) translateY(0); }
}

.doc-header {
  padding: 20px 30px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(255, 255, 255, 0.02);
}

.header-info { display: flex; align-items: center; gap: 15px; }
.doc-icon { font-size: 1.5rem; }
.header-info h3 { margin: 0; font-size: 1.1rem; color: #fff; letter-spacing: 0.5px; text-transform: uppercase; }

.close-btn {
  background: transparent; border: none; color: #64748b; cursor: pointer;
  padding: 8px; border-radius: 10px; transition: all 0.2s;
}
.close-btn:hover { background: rgba(239, 68, 68, 0.1); color: #ef4444; }

.doc-body {
  flex: 1;
  overflow-y: auto;
  padding: 40px 60px;
  color: #cbd5e1;
  line-height: 1.7;
}

/* Scrollbar estilizada */
.doc-body::-webkit-scrollbar { width: 6px; }
.doc-body::-webkit-scrollbar-track { background: transparent; }
.doc-body::-webkit-scrollbar-thumb { background: rgba(255, 255, 255, 0.1); border-radius: 10px; }

.doc-footer {
  padding: 20px 30px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
  background: rgba(0, 0, 0, 0.2);
}

.doc-footer p { font-size: 0.75rem; color: #475569; margin: 0; }

.done-btn {
  background: rgba(59, 130, 246, 0.1);
  border: 1px solid rgba(59, 130, 246, 0.2);
  color: #60a5fa;
  padding: 8px 24px;
  border-radius: 8px;
  font-weight: 700;
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.2s;
}
.done-btn:hover { background: #3b82f6; color: #fff; }

/* Reutilização dos estilos Markdown de ChatLog se necessário */
:deep(.markdown-body h1) { color: #fff; font-size: 1.8rem; margin-bottom: 2rem; border-bottom: 2px solid rgba(59, 130, 246, 0.3); padding-bottom: 10px; }
:deep(.markdown-body h2) { color: #f1f5f9; font-size: 1.4rem; margin-top: 1.5rem; }
:deep(.markdown-body ul) { padding-left: 20px; list-style-type: none; }
:deep(.markdown-body li) { position: relative; margin-bottom: 8px; }
:deep(.markdown-body li::before) { content: "→"; position: absolute; left: -20px; color: #3b82f6; opacity: 0.6; }
:deep(.markdown-body table) { width: 100%; border-collapse: collapse; margin: 2rem 0; background: rgba(255, 255, 255, 0.02); }
:deep(.markdown-body th, .markdown-body td) { padding: 12px; border: 1px solid rgba(255, 255, 255, 0.1); text-align: left; }
:deep(.markdown-body th) { background: rgba(59, 130, 246, 0.1); color: #fff; }

/* Transições */
.fade-scale-enter-active, .fade-scale-leave-active { transition: all 0.3s ease; }
.fade-scale-enter-from, .fade-scale-leave-to { opacity: 0; transform: scale(1.02); }
</style>
