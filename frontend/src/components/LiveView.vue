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
      <div v-for="msg in messages" :key="msg.id" class="chat-bubble" :class="[msg.type, { typing: !msg.isComplete }]">
        <div class="bubble-avatar">
          <span v-if="msg.type === 'interviewer'">👤</span>
          <span v-else>🤖</span>
        </div>
        <div class="bubble-content">
          <div class="bubble-header">
            <span class="bubble-role">{{ msg.type === 'interviewer' ? '面试官' : 'AI 助手' }}</span>
            <span class="bubble-time">{{ formatTime(msg.timestamp) }}</span>
          </div>
          <div class="bubble-text" v-html="msg.type === 'ai' ? renderMarkdown(msg.content) : msg.content"></div>
          <div v-if="!msg.isComplete && msg.type === 'ai'" class="typing-indicator">
            <span></span><span></span><span></span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { marked } from 'marked'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import { StartLiveSession, StopLiveSession } from '../../wailsjs/go/main/App'

const status = ref('disconnected') // disconnected, connecting, connected, error
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

// 渲染 Markdown
function renderMarkdown(text) {
  if (!text) return ''
  return marked.parse(text)
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
    case 'error': return `连接失败: ${errorMsg.value}`
    default: return '未知状态'
  }
})

// 自动滚动到底部
function scrollToBottom() {
  nextTick(() => {
    if (chatContainer.value) {
      chatContainer.value.scrollTop = chatContainer.value.scrollHeight
    }
  })
}

// 事件监听
function onLiveStatus(newStatus) {
  status.value = newStatus
}

function onLiveTranscript(text) {
  // 如果没有当前面试官消息，创建一个
  if (!currentInterviewerMsg.value) {
    currentInterviewerMsg.value = createMessage('interviewer')
    messages.value.push(currentInterviewerMsg.value)
  }
  currentInterviewerMsg.value.content += text
  scrollToBottom()
}

function onLiveInterviewerDone() {
  // 标记面试官消息完成
  if (currentInterviewerMsg.value) {
    currentInterviewerMsg.value.isComplete = true
    currentInterviewerMsg.value = null
  }
}

function onLiveAiText(text) {
  // 如果没有当前 AI 消息，创建一个
  if (!currentAiMsg.value) {
    currentAiMsg.value = createMessage('ai')
    messages.value.push(currentAiMsg.value)
  }
  currentAiMsg.value.content += text
  scrollToBottom()
}

function onLiveError(err) {
  status.value = 'error'
  errorMsg.value = err
}

function onLiveDone() {
  // 一轮对话完成，标记 AI 消息完成
  if (currentAiMsg.value) {
    currentAiMsg.value.isComplete = true
    currentAiMsg.value = null
  }
}

onMounted(() => {
  // 注册事件监听
  EventsOn('live:status', onLiveStatus)
  EventsOn('live:transcript', onLiveTranscript)
  EventsOn('live:interviewer-done', onLiveInterviewerDone)
  EventsOn('live:ai-text', onLiveAiText)
  EventsOn('live:error', onLiveError)
  EventsOn('live:done', onLiveDone)

  // 启动 Live 会话
  StartLiveSession()
})

onUnmounted(() => {
  // 停止 Live 会话
  StopLiveSession()

  // 取消事件监听
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
  background: linear-gradient(180deg, rgba(18, 18, 22, 0.98) 0%, rgba(12, 12, 15, 1) 100%);
}

/* 状态栏 */
.status-bar {
  flex-shrink: 0;
  padding: 12px 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.status-indicator {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 14px;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 500;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.status-disconnected {
  background: rgba(255, 100, 100, 0.12);
  color: #ff6b6b;
}

.status-disconnected .status-dot {
  background: #ff6b6b;
}

.status-connecting {
  background: rgba(255, 193, 7, 0.12);
  color: #ffc107;
}

.status-connecting .status-dot {
  background: #ffc107;
  animation: pulse 1.2s infinite;
}

.status-connected {
  background: rgba(76, 175, 80, 0.12);
  color: #4caf50;
}

.status-connected .status-dot {
  background: #4caf50;
}

.status-error {
  background: rgba(255, 100, 100, 0.12);
  color: #ff6b6b;
}

.status-error .status-dot {
  background: #ff6b6b;
}

@keyframes pulse {

  0%,
  100% {
    opacity: 1;
    transform: scale(1);
  }

  50% {
    opacity: 0.5;
    transform: scale(0.85);
  }
}

/* 聊天消息区域 */
.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.chat-messages::-webkit-scrollbar {
  width: 6px;
}

.chat-messages::-webkit-scrollbar-track {
  background: transparent;
}

.chat-messages::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.15);
  border-radius: 3px;
}

.chat-messages::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.25);
}

/* 空状态 */
.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: rgba(255, 255, 255, 0.4);
}

.empty-icon {
  font-size: 48px;
  opacity: 0.6;
}

.empty-text {
  font-size: 16px;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.6);
}

.empty-hint {
  font-size: 13px;
}

/* 聊天气泡 */
.chat-bubble {
  display: flex;
  gap: 12px;
  max-width: 80%;
  animation: bubbleIn 0.3s ease-out;
}

.chat-bubble.interviewer {
  align-self: flex-start;
}

.chat-bubble.ai {
  align-self: flex-end;
  flex-direction: row-reverse;
}

@keyframes bubbleIn {
  from {
    opacity: 0;
    transform: translateY(12px) scale(0.95);
  }

  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

/* 头像 */
.bubble-avatar {
  flex-shrink: 0;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
}

.interviewer .bubble-avatar {
  background: rgba(255, 255, 255, 0.08);
}

.ai .bubble-avatar {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

/* 气泡内容 */
.bubble-content {
  background: rgba(45, 45, 55, 0.85);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 18px;
  padding: 12px 16px;
  min-width: 60px;
}

.interviewer .bubble-content {
  border-radius: 18px 18px 18px 4px;
}

.ai .bubble-content {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.25) 0%, rgba(118, 75, 162, 0.25) 100%);
  border-color: rgba(102, 126, 234, 0.3);
  border-radius: 18px 18px 4px 18px;
}

.bubble-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.bubble-role {
  font-size: 12px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.6);
}

.ai .bubble-role {
  color: rgba(167, 139, 250, 0.9);
}

.bubble-time {
  font-size: 11px;
  color: rgba(255, 255, 255, 0.35);
}

.bubble-text {
  font-size: 14px;
  line-height: 1.6;
  color: rgba(255, 255, 255, 0.9);
  word-wrap: break-word;
  white-space: pre-wrap;
}

.bubble-text :deep(p) {
  margin: 0 0 8px 0;
}

.bubble-text :deep(p:last-child) {
  margin-bottom: 0;
}

.bubble-text :deep(code) {
  background: rgba(0, 0, 0, 0.3);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
}

.bubble-text :deep(pre) {
  background: rgba(0, 0, 0, 0.4);
  padding: 12px;
  border-radius: 8px;
  overflow-x: auto;
  margin: 8px 0;
}

.bubble-text :deep(pre code) {
  background: none;
  padding: 0;
}

/* 打字指示器 */
.typing-indicator {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 8px;
}

.typing-indicator span {
  width: 6px;
  height: 6px;
  background: rgba(167, 139, 250, 0.7);
  border-radius: 50%;
  animation: typingDot 1.4s infinite ease-in-out both;
}

.typing-indicator span:nth-child(1) {
  animation-delay: -0.32s;
}

.typing-indicator span:nth-child(2) {
  animation-delay: -0.16s;
}

.typing-indicator span:nth-child(3) {
  animation-delay: 0s;
}

@keyframes typingDot {

  0%,
  80%,
  100% {
    transform: scale(0.8);
    opacity: 0.5;
  }

  40% {
    transform: scale(1.2);
    opacity: 1;
  }
}

/* 正在输入状态 */
.chat-bubble.typing .bubble-content {
  box-shadow: 0 0 0 2px rgba(167, 139, 250, 0.2);
}

.chat-bubble.interviewer.typing .bubble-content {
  box-shadow: 0 0 0 2px rgba(255, 255, 255, 0.1);
}
</style>
