import { ref } from 'vue'

export function useStatus(settings) {
  const statusText = ref('就绪')
  const statusIcon = ref('🟢')
  const isRecording = ref(false)

  function resetStatus() {
    if (!settings.apiKey) {
      statusText.value = '未配置Key'
      statusIcon.value = '⚠️'
      return
    }

    if (settings.voiceListening) {
      statusText.value = '就绪'
      statusIcon.value = '🟢'
    } else {
      statusText.value = '就绪'
      statusIcon.value = '📝'
    }
  }

  return {
    statusText,
    statusIcon,
    isRecording,
    resetStatus
  }
}
