<template>
  <div class="loading-container">
    <div class="loader-content">
      <div class="orb-container">
        <div class="orb"></div>
        <div class="ring ring-1"></div>
        <div class="ring ring-2"></div>
      </div>
      <div class="loading-text">
        <span class="text">深度思考中</span>
        <span class="timer">{{ formattedTime }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'

const startTime = ref(Date.now())
const currentTime = ref(Date.now())
let timerInterval = null

const formattedTime = computed(() => {
  const diff = currentTime.value - startTime.value
  const seconds = Math.floor(diff / 1000)
  const ms = Math.floor((diff % 1000) / 10)
  return `${seconds}.${ms.toString().padStart(2, '0')}s`
})

onMounted(() => {
  timerInterval = setInterval(() => {
    currentTime.value = Date.now()
  }, 30)
})

onUnmounted(() => {
  if (timerInterval) clearInterval(timerInterval)
})
</script>

<style scoped>
.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  width: 100%;
  background: transparent;
  animation: fadeIn 0.5s ease-out;
}

.loader-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 24px;
}

.orb-container {
  position: relative;
  width: 80px;
  height: 80px;
  display: flex;
  justify-content: center;
  align-items: center;
}

.orb {
  width: 40px;
  height: 40px;
  background: radial-gradient(circle at 30% 30%, #e0f7fa, #4dd0e1);
  border-radius: 50%;
  box-shadow: 0 0 25px rgba(77, 208, 225, 0.8);
  animation: breathe 4s ease-in-out infinite;
  z-index: 2;
}

.ring {
  position: absolute;
  border-radius: 50%;
  border: 2px solid transparent;
  border-top-color: rgba(77, 208, 225, 0.6);
  border-right-color: rgba(77, 208, 225, 0.3);
}

.ring-1 {
  width: 60px;
  height: 60px;
  animation: spin 3s cubic-bezier(0.55, 0.055, 0.675, 0.19) infinite;
}

.ring-2 {
  width: 80px;
  height: 80px;
  border-top-color: rgba(77, 208, 225, 0.3);
  border-left-color: rgba(77, 208, 225, 0.6);
  animation: spin 4s cubic-bezier(0.55, 0.055, 0.675, 0.19) infinite reverse;
}

.loading-text {
  display: flex;
  align-items: center;
  gap: 12px;
  font-family: 'Segoe UI', sans-serif;
  font-size: 15px;
  color: rgba(224, 247, 250, 0.9);
  letter-spacing: 1px;
  text-transform: uppercase;
  text-shadow: 0 0 15px rgba(77, 208, 225, 0.5);
}

.timer {
  font-family: 'Consolas', monospace;
  font-weight: bold;
  color: #80deea;
  background: rgba(0, 188, 212, 0.1);
  padding: 2px 8px;
  border-radius: 4px;
  border: 1px solid rgba(0, 188, 212, 0.3);
}

@keyframes breathe {
  0%, 100% { transform: scale(0.95); opacity: 0.8; box-shadow: 0 0 20px rgba(77, 208, 225, 0.5); }
  50% { transform: scale(1.05); opacity: 1; box-shadow: 0 0 40px rgba(77, 208, 225, 0.9); }
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
