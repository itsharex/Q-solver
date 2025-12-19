<template>
  <Transition name="overlay-fade">
    <div v-if="visible" class="voice-overlay-wrapper">
      <div class="voice-card" :class="{ 'is-active': !isIdle }">
        <!-- 动态声波/状态指示器 -->
        <div class="indicator-section">
          <div v-if="isIdle" class="idle-state">
            <div class="mic-icon">
              <svg viewBox="0 0 24 24" width="20" height="20" stroke="currentColor" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round" class="feather-mic"><path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"></path><path d="M19 10v2a7 7 0 0 1-14 0v-2"></path><line x1="12" y1="19" x2="12" y2="23"></line><line x1="8" y1="23" x2="16" y2="23"></line></svg>
            </div>
            <div class="breathing-ring"></div>
          </div>
          
          <div v-else class="active-wave">
            <div class="wave-bar" v-for="n in 5" :key="n" :style="{ '--delay': (n * 0.1) + 's' }"></div>
          </div>
        </div>

        <!-- 文本显示区域 -->
        <div class="text-section">
          <span class="main-text" :class="{ 'placeholder': isIdle }">
            {{ text }}
          </span>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  isRecording: Boolean,
  text: {
    type: String,
    default: '正在聆听...'
  }
})

const visible = computed(() => props.isRecording)
const isIdle = computed(() => !props.text || props.text === '正在聆听...')
</script>

<style scoped>
.voice-overlay-wrapper {
  position: fixed;
  top: 100px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 9999;
  pointer-events: none;
  perspective: 1000px;
  width: auto;
  display: flex;
  justify-content: center;
}

.voice-card {
  background: rgba(18, 18, 20, 0.85);
  backdrop-filter: blur(20px) saturate(180%);
  -webkit-backdrop-filter: blur(20px) saturate(180%);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 50px;
  padding: 10px 24px;
  display: flex;
  align-items: center;
  gap: 16px;
  box-shadow: 
    0 10px 30px rgba(0, 0, 0, 0.5),
    0 0 0 1px rgba(255, 255, 255, 0.05) inset;
  min-width: 160px;
  max-width: 85vw;
  transition: all 0.4s cubic-bezier(0.2, 0.8, 0.2, 1);
}

.voice-card.is-active {
  background: rgba(25, 25, 30, 0.95);
  border-color: rgba(76, 175, 80, 0.3);
  box-shadow: 
    0 15px 40px rgba(0, 0, 0, 0.6),
    0 0 0 1px rgba(76, 175, 80, 0.15) inset,
    0 0 20px rgba(76, 175, 80, 0.1);
  padding: 12px 28px; /* Slightly expand */
}

/* --- Indicator Section --- */
.indicator-section {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  position: relative;
  flex-shrink: 0;
}

/* Idle State */
.idle-state {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  color: rgba(255, 255, 255, 0.7);
}

.mic-icon {
  z-index: 2;
  display: flex;
}

.breathing-ring {
  position: absolute;
  width: 100%;
  height: 100%;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.15);
  z-index: 1;
  animation: breathe 2s infinite ease-in-out;
}

@keyframes breathe {
  0%, 100% { transform: scale(1); opacity: 0.4; }
  50% { transform: scale(1.6); opacity: 0; }
}

/* Active Wave */
.active-wave {
  display: flex;
  align-items: center;
  gap: 3px;
  height: 14px;
}

.wave-bar {
  width: 3px;
  background: #4CAF50;
  border-radius: 2px;
  height: 100%;
  animation: wave-dance 1s infinite ease-in-out;
  animation-delay: var(--delay);
  box-shadow: 0 0 6px rgba(76, 175, 80, 0.5);
}

@keyframes wave-dance {
  0%, 100% { height: 20%; opacity: 0.6; }
  50% { height: 100%; opacity: 1; }
}

/* --- Text Section --- */
.text-section {
  display: flex;
  align-items: center;
  overflow: hidden;
}

.main-text {
  font-family: 'Nunito', 'Segoe UI', sans-serif;
  font-size: 15px;
  font-weight: 600;
  color: #fff;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 600px;
  letter-spacing: 0.3px;
  transition: color 0.3s;
  text-shadow: 0 1px 2px rgba(0,0,0,0.5);
}

.main-text.placeholder {
  color: rgba(255, 255, 255, 0.4);
  font-weight: 500;
  font-style: italic;
}

/* --- Transitions --- */
.overlay-fade-enter-active,
.overlay-fade-leave-active {
  transition: all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.overlay-fade-enter-from,
.overlay-fade-leave-to {
  opacity: 0;
  transform: translate(-50%, -10px) scale(0.95);
}
</style>
