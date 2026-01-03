<template>
  <div class="welcome-container">
    <div class="welcome-content">
      <div class="logo-container">
        <div class="ghost-shape">
          <div class="eyes">
            <div class="eye left"></div>
            <div class="eye right"></div>
          </div>
        </div>
        
        <Transition name="fade-ring" mode="out-in">
          <div v-if="initStatus !== 'ready'" class="loading-ring" key="loading"></div>
          <div v-else-if="showSuccess" class="success-ring" key="success">
            <svg class="checkmark-svg" viewBox="0 0 52 52">
              <circle class="checkmark-circle" cx="26" cy="26" r="25" fill="none"/>
              <path class="checkmark-check" fill="none" d="M14.1 27.2l7.1 7.2 16.7-16.8"/>
            </svg>
          </div>
          <div v-else class="pulse-ring" key="pulse"></div>
        </Transition>
        

      </div>
      
      <div class="welcome-title">
        <span class="glitch" data-text="Q-">Q-SOLVER</span>
        <!-- <span class="solver">SLOVER</span> -->
      </div>
      
      <div class="welcome-desc">
        <Transition name="fade-slide" mode="out-in">
          <div v-if="initStatus === 'ready' && !showSuccess" class="shortcuts-container" key="shortcuts">
            <div class="shortcut-row">
              <div class="key-combo">
                <span class="key">{{ solveShortcut }}</span>
              </div>
              <span class="action">一键解题</span>
            </div>
            <div class="shortcut-row">
              <div class="key-combo">
                <span class="key">{{ toggleShortcut }}</span>
              </div>
              <span class="action">隐藏窗口</span>
            </div>
          </div>
          <div v-else-if="showSuccess" class="success-message" key="success">
            <span class="success-text">系统就绪</span>
          </div>
          <div v-else class="loading-status" key="loading">
            <div class="loading-text">
              {{ initStatus === 'loading-model' ? '正在加载神经网络...' : '系统初始化中...' }}
            </div>
            <div class="loading-dots">
              <span></span><span></span><span></span>
            </div>
          </div>
        </Transition>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'


const props = defineProps({
  solveShortcut: String,
  toggleShortcut: String,
  initStatus: {
    type: String,
    default: 'ready'
  }
})

const showSuccess = ref(false)

watch(() => props.initStatus, (newVal, oldVal) => {
  if (newVal === 'ready' && oldVal !== 'ready') {
    showSuccess.value = true
    setTimeout(() => {
      showSuccess.value = false
    }, 2000)
  }
})
</script>

<style scoped>
.welcome-container {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  width: 100%;
  position: absolute;
  top: 0;
  left: 0;
  z-index: 10;
  color: var(--text-primary);
  font-family: var(--font-family);
  pointer-events: none;
}

.welcome-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-6);
  pointer-events: auto;
  background: var(--bg-card);
  padding: var(--space-10) var(--space-8);
  border-radius: var(--radius-xl);
  border: 1px solid var(--border-subtle);
  backdrop-filter: blur(12px);
  box-shadow: var(--shadow-xl);
  animation: float 6s ease-in-out infinite;
}

/* Logo / Ghost Styling */
.logo-container {
  position: relative;
  width: 80px;
  height: 80px;
  display: flex;
  justify-content: center;
  align-items: center;
  margin-bottom: 10px;
  z-index: 10; /* 确保 Logo 在最上层 */
}

.ghost-shape {
  width: 50px;
  height: 60px;
  background: linear-gradient(180deg, #ffffff 0%, #e0e0e0 100%);
  position: relative;
  z-index: 2;
  /* 使用 drop-shadow 替代 box-shadow，使其跟随 clip-path 形状 */
  filter: drop-shadow(0 0 10px rgba(255, 255, 255, 0.4));
  animation: hover 3s ease-in-out infinite;
  /* 使用 path 精确绘制圆角头部和波浪底部，解决白角问题 */
  clip-path: path("M 0 25 A 25 25 0 0 1 50 25 L 50 60 L 41.5 54 L 33 60 L 25 54 L 17 60 L 8.5 54 L 0 60 Z");
}

.eyes {
  display: flex;
  justify-content: space-between;
  padding: 22px 12px 0;
}

.eye {
  width: 8px;
  height: 10px;
  background: #333;
  border-radius: 50%;
  animation: blink 4s infinite;
}

.pulse-ring {
  position: absolute;
  width: 100%;
  height: 100%;
  border-radius: 50%;
  border: 2px solid rgba(255, 255, 255, 0.2);
  animation: pulse-ring 2s cubic-bezier(0.215, 0.61, 0.355, 1) infinite;
}

/* Typography */
.welcome-title {
  font-size: 32px;
  font-weight: 800;
  letter-spacing: 4px;
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 10px;
}

.glitch {
  color: #fff;
  text-shadow: 2px 2px rgba(255, 0, 255, 0.3), -2px -2px rgba(0, 255, 255, 0.3);
}

.solver {
  color: transparent;
  -webkit-text-stroke: 1px rgba(255, 255, 255, 0.8);
  font-weight: 300;
}

/* Shortcuts */
.welcome-desc {
  display: flex;
  flex-direction: column;
  gap: 12px;
  width: 100%;
  min-width: 240px;
}

.shortcut-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--bg-inset);
  padding: var(--space-3) var(--space-4);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle);
  transition: all var(--transition-fast);
}

.shortcut-row:hover {
  background: var(--bg-card-hover);
  transform: translateX(5px);
  border-color: var(--border-default);
}

.key-combo {
  display: flex;
  gap: var(--space-1);
}

.key {
  background: var(--color-primary-light);
  padding: var(--space-1) var(--space-2);
  border-radius: var(--radius-sm);
  font-family: var(--font-mono);
  font-size: var(--text-xs);
  font-weight: bold;
  color: var(--color-primary);
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.action {
  font-size: var(--text-sm);
  color: var(--text-secondary);
  font-weight: 500;
}

/* Update Log */
.update-log {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%) scale(0.5);
  width: 280px;
  max-height: 300px;
  background: var(--bg-elevated);
  border-radius: var(--radius-md);
  padding: var(--space-4);
  border: 1px solid var(--border-subtle);
  display: flex;
  flex-direction: column;
  opacity: 0;
  z-index: -1;
  pointer-events: none;
  transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
  box-shadow: var(--shadow-xl);
}

/* Hover Logo 时显示公告 */
.welcome-content:hover .update-log,
.update-log:hover {
  opacity: 1;
  pointer-events: auto;
  left: 100%;
  top: 0;
  transform: translate(20px, -10px) scale(1);
}

.log-title {
  font-size: var(--text-xs);
  font-weight: bold;
  color: var(--color-primary);
  margin-bottom: var(--space-2);
  text-transform: uppercase;
  letter-spacing: 1px;
  flex-shrink: 0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

/* 添加一个小箭头指向 Logo */
.update-log::before {
  content: '';
  position: absolute;
  left: -6px;
  top: 25px;
  width: 12px;
  height: 12px;
  background: var(--bg-elevated);
  border-left: 1px solid var(--border-subtle);
  border-bottom: 1px solid var(--border-subtle);
  transform: rotate(45deg);
}

.log-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.log-list li {
  font-size: var(--text-xs);
  color: var(--text-secondary);
  margin-bottom: var(--space-1);
  padding-left: var(--space-3);
  position: relative;
  line-height: 1.4;
}

.log-list li::before {
  content: "•";
  position: absolute;
  left: 0;
  color: var(--color-primary);
}

/* Animations */
@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}

@keyframes hover {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-5px); }
}

@keyframes blink {
  0%, 96%, 100% { transform: scaleY(1); }
  98% { transform: scaleY(0.1); }
}

@keyframes pulse-ring {
  0% { transform: scale(0.8); opacity: 0.5; }
  100% { transform: scale(1.5); opacity: 0; }
}

/* HTML Content Styling */
.html-content {
  font-size: 14px;
  line-height: 1.6;
  color: rgba(255, 255, 255, 0.8);
  text-align: left;
  /* 启用滚动条 */
  overflow-y: auto;
  padding-right: 8px;
  /* 防止长单词撑开容器 */
  word-wrap: break-word;
  word-break: break-word;
}

/* 自定义滚动条 */
.html-content::-webkit-scrollbar {
  width: 4px;
}

.html-content::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 2px;
}

.html-content::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 2px;
}

.html-content::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.3);
}

.html-content :deep(img) {
  /* 限制图片大小，防止撑坏布局 */
  max-width: 100%;
  height: auto;
  border-radius: 4px;
  margin: 8px 0;
}

.html-content :deep(p) {
  margin: 8px 0;
}

.html-content :deep(ul), .html-content :deep(ol) {
  padding-left: 20px;
  margin: 8px 0;
}

.html-content :deep(li) {
  margin-bottom: 4px;
}

.html-content :deep(a) {
  color: #64b5f6;
  text-decoration: none;
  border-bottom: 1px dashed rgba(100, 181, 246, 0.5);
  transition: all 0.2s;
}

.html-content :deep(a:hover) {
  color: #90caf9;
  border-bottom-style: solid;
}

.html-content :deep(strong), .html-content :deep(b) {
  color: #fff;
  font-weight: 600;
}

.html-content :deep(code) {
  background: rgba(255, 255, 255, 0.1);
  padding: 2px 4px;
  border-radius: 4px;
  font-family: monospace;
  color: #e0e0e0;
}

/* Loading Styles */
.loading-ring {
  position: absolute;
  width: 140%;
  height: 140%;
  border-radius: 50%;
  border: 3px solid transparent;
  border-top-color: var(--color-primary);
  border-right-color: rgba(16, 185, 129, 0.5);
  animation: spin 1s linear infinite;
}

.loading-status {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-3);
  width: 100%;
}

.loading-text {
  color: var(--text-secondary);
  font-size: var(--text-sm);
  font-weight: 500;
  letter-spacing: 0.5px;
}

.loading-dots {
  display: flex;
  gap: var(--space-1);
}

.loading-dots span {
  width: 4px;
  height: 4px;
  background-color: var(--color-primary);
  border-radius: var(--radius-full);
  animation: wave 1.5s infinite ease-in-out both;
}

.loading-dots span:nth-child(1) { animation-delay: -0.32s; }
.loading-dots span:nth-child(2) { animation-delay: -0.16s; }

/* Success Styles */
.success-ring {
  position: absolute;
  width: 140%;
  height: 140%;
  display: flex;
  justify-content: center;
  align-items: center;
}

.checkmark-svg {
  width: 50px;
  height: 50px;
  border-radius: 50%;
  display: block;
  stroke-width: 3;
  stroke: var(--color-primary);
  stroke-miterlimit: 10;
  box-shadow: inset 0px 0px 0px var(--color-primary);
  animation: fill .4s ease-in-out .4s forwards, scale .3s ease-in-out .9s both;
}

.checkmark-circle {
  stroke-dasharray: 166;
  stroke-dashoffset: 166;
  stroke-width: 2;
  stroke-miterlimit: 10;
  stroke: var(--color-primary);
  fill: none;
  animation: stroke 0.6s cubic-bezier(0.65, 0, 0.45, 1) forwards;
}

.checkmark-check {
  transform-origin: 50% 50%;
  stroke-dasharray: 48;
  stroke-dashoffset: 48;
  animation: stroke 0.3s cubic-bezier(0.65, 0, 0.45, 1) 0.8s forwards;
}

.success-message {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 60px;
}

.success-text {
  color: var(--color-primary);
  font-size: var(--text-base);
  font-weight: 600;
  letter-spacing: 1px;
  animation: fadeIn 0.5s ease;
}

.shortcuts-container {
  display: flex;
  flex-direction: column;
  gap: 12px;
  width: 100%;
}

/* Transitions */
.fade-ring-enter-active,
.fade-ring-leave-active {
  transition: opacity 0.5s ease, transform 0.5s ease;
}

.fade-ring-enter-from,
.fade-ring-leave-to {
  opacity: 0;
  transform: scale(0.8);
}

.fade-slide-enter-active,
.fade-slide-leave-active {
  transition: all 0.4s ease;
}

.fade-slide-enter-from {
  opacity: 0;
  transform: translateY(10px);
}

.fade-slide-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

@keyframes stroke {
  100% {
    stroke-dashoffset: 0;
  }
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

@keyframes wave {
  0%, 80%, 100% { transform: translateY(0); }
  40% { transform: translateY(-5px); }
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(5px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
