<template>
  <div class="error-container">
    <div class="error-icon-large">{{ errorState.icon }}</div>
    <div class="error-title-large">{{ errorState.title }}</div>
    <div class="error-desc">{{ errorState.desc }}</div>
    
    <!-- 极简模式：只保留快捷键提示 -->
    <div class="error-shortcut-hint" style="margin-top: 20px;">
      按下 <span class="shortcut-tag">{{ solveShortcut }}</span> 重试
    </div>
  </div>
</template>

<script setup>
defineProps({
  errorState: Object,
  solveShortcut: String
})
</script>

<style scoped>
.error-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  text-align: center;
  padding: 20px;
  animation: fadeInSimple 0.3s ease-out;
}

.error-icon-large {
  font-size: 64px;
  margin-bottom: 16px;
  filter: drop-shadow(0 4px 8px rgba(0,0,0,0.3));
}

.error-title-large {
  font-size: 24px;
  font-weight: 700;
  color: #fff;
  margin-bottom: 8px;
  text-shadow: 0 2px 4px rgba(0,0,0,0.5);
}

.error-desc {
  font-size: 14px;
  color: #ccc;
  margin-bottom: 24px;
  max-width: 80%;
  line-height: 1.5;
}

.error-shortcut-hint {
  font-size: 13px;
  color: #aaa;
  margin-bottom: 20px;
  animation: pulse-opacity 2s infinite;
}

.shortcut-tag {
  background: rgba(255, 255, 255, 0.15);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: monospace;
  font-weight: bold;
  color: #fff;
  border: 1px solid rgba(255, 255, 255, 0.2);
}

@keyframes pulse-opacity {
  0% { opacity: 0.6; }
  50% { opacity: 1; }
  100% { opacity: 0.6; }
}

@keyframes fadeInSimple {
  from { opacity: 0; }
  to { opacity: 1; }
}
</style>
