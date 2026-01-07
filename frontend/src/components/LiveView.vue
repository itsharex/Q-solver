<template>
  <div class="live-view">
    <!-- 状态栏 -->
    <div class="status-bar">
      <div class="status-indicator" :class="statusClass">
        <span class="status-dot"></span>
        <span class="status-text">{{ statusText }}</span>
      </div>
    </div>

    <!-- 聊天消息区域 -->
    <div class="chat-messages" ref="chatContainer">
      <!-- 空状态 -->
      <div v-if="messages.length === 0" class="empty-state">
        <div class="empty-icon">🎙️</div>
        <div class="empty-text">等待面试开始...</div>
        <div class="empty-hint">面试官的语音会自动转录显示在左侧</div>
      </div>

      <!-- 消息气泡列表 -->
      <div v-for="msg in messages" :key="msg.id" class="message-row" :class="msg.type">
        <div class="message-bubble" :class="{ typing: !msg.isComplete }">
          <div class="message-meta">
            <span class="message-role">{{ msg.type === 'interviewer' ? '👤 面试官' : '🤖 AI 助手' }}</span>
            <span class="message-time">{{ formatTime(msg.timestamp) }}</span>
          </div>
          <div class="message-text" v-html="msg.type === 'ai' ? renderMarkdown(msg.content) : escapeHtml(msg.content)">
          </div>
        </div>
      </div>

      <!-- 底部占位确保最后消息可见 -->
      <div class="scroll-anchor"></div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { marked } from 'marked'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import { StartLiveSession, StopLiveSession } from '../../wailsjs/go/main/App'

const status = ref('disconnected')
const errorMsg = ref('')
const chatContainer = ref(null)

// 消息列表
const messages = ref([])
const currentInterviewerMsg = ref(null)
const currentAiMsg = ref(null)

// 生成唯一 ID
function generateId() {
  return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
}

// 创建新消息
function createMessage(type) {
  return {
    id: generateId(),
    type,
    content: '',
    timestamp: Date.now(),
    isComplete: false
  }
}

// 格式化时间
function formatTime(timestamp) {
  const date = new Date(timestamp)
  return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

// HTML 转义
function escapeHtml(text) {
  if (!text) return ''
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/\n/g, '<br>')
}

// 渲染 Markdown
function renderMarkdown(text) {
  if (!text) return ''
  const trimmed = text.replace(/\n+$/, '')
  return marked.parse(trimmed)
}

const statusClass = computed(() => ({
  'status-disconnected': status.value === 'disconnected',
  'status-connecting': status.value === 'connecting',
  'status-connected': status.value === 'connected',
  'status-error': status.value === 'error',
}))

const statusText = computed(() => {
  switch (status.value) {
    case 'disconnected': return '未连接'
    case 'connecting': return '连接中...'
    case 'connected': return '已连接'
    case 'error': return `错误: ${errorMsg.value}`
    default: return '未知状态'
  }
})

// 滚动到底部 - 简化版本，直接滚动
function scrollToBottom() {
  setTimeout(() => {
    if (chatContainer.value) {
      chatContainer.value.scrollTop = chatContainer.value.scrollHeight
    }
  }, 10)
}

// 监听消息变化自动滚动
watch(messages, () => {
  scrollToBottom()
}, { deep: true })

// 事件监听
function onLiveStatus(newStatus) {
  status.value = newStatus
}

function onLiveTranscript(text) {
  if (!currentInterviewerMsg.value) {
    currentInterviewerMsg.value = createMessage('interviewer')
    messages.value.push(currentInterviewerMsg.value)
  }
  currentInterviewerMsg.value.content += text
}

function onLiveInterviewerDone() {
  if (currentInterviewerMsg.value) {
    currentInterviewerMsg.value.isComplete = true
    currentInterviewerMsg.value = null
  }
}

function onLiveAiText(text) {
  if (!currentAiMsg.value) {
    currentAiMsg.value = createMessage('ai')
    messages.value.push(currentAiMsg.value)
  }
  currentAiMsg.value.content += text
}

function onLiveError(err) {
  status.value = 'error'
  errorMsg.value = err
}

function onLiveDone() {
  if (currentAiMsg.value) {
    currentAiMsg.value.isComplete = true
    currentAiMsg.value = null
  }
}

onMounted(() => {
  EventsOn('live:status', onLiveStatus)
  EventsOn('live:transcript', onLiveTranscript)
  EventsOn('live:interviewer-done', onLiveInterviewerDone)
  EventsOn('live:ai-text', onLiveAiText)
  EventsOn('live:error', onLiveError)
  EventsOn('live:done', onLiveDone)
  StartLiveSession()
})

onUnmounted(() => {
  StopLiveSession()
  EventsOff('live:status')
  EventsOff('live:transcript')
  EventsOff('live:interviewer-done')
  EventsOff('live:ai-text')
  EventsOff('live:error')
  EventsOff('live:done')
})
</script>

<style scoped>
.live-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  pointer-events: auto;
}

/* 状态栏 */
.status-bar {
  flex-shrink: 0;
  padding: 10px 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.status-indicator {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 12px;
  border-radius: 16px;
  font-size: 12px;
  font-weight: 500;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.status-disconnected {
  background: rgba(239, 68, 68, 0.15);
  color: #f87171;
}

.status-disconnected .status-dot {
  background: #f87171;
}

.status-connecting {
  background: rgba(251, 191, 36, 0.15);
  color: #fbbf24;
}

.status-connecting .status-dot {
  background: #fbbf24;
  animation: pulse 1s infinite;
}

.status-connected {
  background: rgba(34, 197, 94, 0.15);
  color: #22c55e;
}

.status-connected .status-dot {
  background: #22c55e;
}

.status-error {
  background: rgba(239, 68, 68, 0.15);
  color: #f87171;
}

.status-error .status-dot {
  background: #f87171;
}

@keyframes pulse {

  0%,
  100% {
    opacity: 1;
  }

  50% {
    opacity: 0.4;
  }
}

/* 聊天区域 */
.chat-messages {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 12px 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-height: 0;
  pointer-events: auto;
}

.chat-messages::-webkit-scrollbar {
  width: 4px;
}

.chat-messages::-webkit-scrollbar-track {
  background: transparent;
}

.chat-messages::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 2px;
}

.scroll-anchor {
  height: 20px;
  flex-shrink: 0;
}

/* 空状态 */
.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: rgba(255, 255, 255, 0.4);
}

.empty-icon {
  font-size: 40px;
  opacity: 0.5;
}

.empty-text {
  font-size: 15px;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.5);
}

.empty-hint {
  font-size: 12px;
}

/* 消息行 */
.message-row {
  display: flex;
  animation: slideIn 0.2s ease-out;
}

.message-row.interviewer {
  justify-content: flex-start;
}

.message-row.ai {
  justify-content: flex-end;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(8px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 消息气泡 */
.message-bubble {
  max-width: 85%;
  padding: 10px 14px;
  border-radius: 16px;
  position: relative;
}

.interviewer .message-bubble {
  background: rgba(55, 55, 65, 0.9);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 16px 16px 16px 4px;
}

.ai .message-bubble {
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.9) 0%, rgba(139, 92, 246, 0.9) 100%);
  border-radius: 16px 16px 4px 16px;
}

.message-bubble.typing {
  box-shadow: 0 0 0 2px rgba(139, 92, 246, 0.3);
}

/* 消息元信息 */
.message-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.message-role {
  font-size: 11px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.7);
}

.ai .message-role {
  color: rgba(255, 255, 255, 0.9);
}

.message-time {
  font-size: 10px;
  color: rgba(255, 255, 255, 0.4);
}

/* 消息文本 */
.message-text {
  font-size: 14px;
  line-height: 1.5;
  color: rgba(255, 255, 255, 0.95);
  word-wrap: break-word;
}

.message-text :deep(p) {
  margin: 0 0 6px 0;
}

.message-text :deep(p:last-child) {
  margin-bottom: 0;
}

.message-text :deep(code) {
  background: rgba(0, 0, 0, 0.25);
  padding: 1px 5px;
  border-radius: 4px;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
}

.message-text :deep(pre) {
  background: rgba(0, 0, 0, 0.3);
  padding: 10px;
  border-radius: 8px;
  overflow-x: auto;
  margin: 6px 0;
}

.message-text :deep(pre code) {
  background: none;
  padding: 0;
}

.message-text :deep(ul),
.message-text :deep(ol) {
  margin: 4px 0;
  padding-left: 20px;
}

.message-text :deep(li) {
  margin: 2px 0;
}
</style>
