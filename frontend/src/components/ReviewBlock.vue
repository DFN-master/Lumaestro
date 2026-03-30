<script setup>
import { useOrchestratorStore } from '../stores/orchestrator';

const props = defineProps({
  review: { type: Object, required: true }
});

const orchestrator = useOrchestratorStore();

const approve = () => orchestrator.submitReview(true);
const reject = () => orchestrator.submitReview(false);
</script>

<template>
  <div v-if="review" class="review-overlay">
    <div class="review-card glass">
      <div class="review-header">
        <span class="warning-icon">⚠️</span>
        <div class="review-titles">
          <h3>PEDIDO DE REVISÃO</h3>
          <span class="review-subtitle">Segurança das Mãos (Backend Security)</span>
        </div>
      </div>
      
      <div class="review-body">
        <p>A IA solicita permissão para executar uma ação sensível:</p>
        <div class="action-details">
          <div class="action-label">{{ review.action }}</div>
          <div class="action-content">{{ review.details }}</div>
        </div>
      </div>

      <div class="review-footer">
        <button @click="reject" class="btn-reject">REJEITAR</button>
        <button @click="approve" class="btn-approve">APROVAR E CONTINUAR</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.review-overlay {
  position: absolute;
  inset: 0;
  background: rgba(15, 23, 42, 0.85);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 24px;
}

.review-card {
  width: 100%;
  max-width: 480px;
  background: #1e293b;
  border: 1px solid rgba(59, 130, 246, 0.3);
  border-radius: 20px;
  overflow: hidden;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.8), 0 0 40px rgba(59, 130, 246, 0.1);
  animation: scaleIn 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

@keyframes scaleIn {
  from { opacity: 0; transform: scale(0.9); }
  to { opacity: 1; transform: scale(1); }
}

.review-header {
  padding: 24px;
  background: rgba(59, 130, 246, 0.1);
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  display: flex;
  gap: 16px;
  align-items: center;
}

.warning-icon { font-size: 2rem; }
.review-titles h3 { margin: 0; font-size: 14px; font-weight: 900; letter-spacing: 2px; color: #60a5fa; }
.review-subtitle { font-size: 10px; color: #94a3b8; text-transform: uppercase; }

.review-body { padding: 24px; }
.review-body p { margin-top: 0; font-size: 14px; color: #cbd5e1; }

.action-details {
  background: #0f172a;
  border-radius: 12px;
  padding: 16px;
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.action-label { font-size: 11px; font-weight: 800; color: #64748b; margin-bottom: 8px; text-transform: uppercase; }
.action-content { font-family: 'JetBrains Mono', monospace; font-size: 13px; color: #e2e8f0; word-break: break-all; }

.review-footer {
  padding: 24px;
  display: flex;
  gap: 16px;
  background: rgba(15, 23, 42, 0.5);
}

.review-footer button {
  flex: 1;
  padding: 12px;
  border-radius: 12px;
  font-weight: 800;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-reject {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.2);
  color: #f87171;
}

.btn-reject:hover { background: #ef4444; color: #fff; }

.btn-approve {
  background: #3b82f6;
  border: 1px solid #3b82f6;
  color: #fff;
  box-shadow: 0 4px 15px rgba(59, 130, 246, 0.3);
}

.btn-approve:hover { transform: translateY(-2px); box-shadow: 0 8px 25px rgba(59, 130, 246, 0.5); }
</style>
