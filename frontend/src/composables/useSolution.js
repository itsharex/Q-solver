import { ref, reactive, nextTick } from 'vue'
import { marked } from 'marked'

export function useSolution(settings) {
  const renderedContent = ref('')
  const history = ref([])
  const activeHistoryIndex = ref(0)
  const isLoading = ref(false)
  const isAppending = ref(false)
  const shouldOverwriteHistory = ref(false)
  let streamBuffer = ''

  const errorState = reactive({
    show: false,
    icon: '⚠️',
    title: '出错了',
    desc: '发生了一个未知错误',
    rawError: '',
    showDetails: false
  })

  function renderMarkdown(md) {
    if (!md) return ''
    return marked.parse(md)
  }

  function selectHistory(idx) {
    const item = history.value[idx]
    if (item) {
      renderedContent.value = renderMarkdown(item.full)
      activeHistoryIndex.value = idx
    }
  }

  function handleStreamStart() {
    if (settings.keepContext && history.value.length > 0 && !shouldOverwriteHistory.value) {
      const separator = '\n\n---\n\n'
      streamBuffer = history.value[0].full + separator
      activeHistoryIndex.value = 0
    } else {
      streamBuffer = ''
      renderedContent.value = ''
      
      if (shouldOverwriteHistory.value && history.value.length > 0) {
        history.value[0] = {
          summary: '正在思考...', 
          full: '', 
          time: new Date().toLocaleTimeString() 
        }
        shouldOverwriteHistory.value = false
      } else {
        history.value.unshift({ 
          summary: '正在思考...', 
          full: '', 
          time: new Date().toLocaleTimeString() 
        })
      }
      activeHistoryIndex.value = 0
    }
  }

  function handleStreamChunk(token) {
    if (isLoading.value) isLoading.value = false
    if (isAppending.value) isAppending.value = false
    
    streamBuffer += token
    renderedContent.value = renderMarkdown(streamBuffer)
    
    if (history.value.length > 0) {
      history.value[0].full = streamBuffer
      history.value[0].summary = streamBuffer.substring(0, 30).replace(/\n/g, ' ') + '...'
    }

    nextTick(() => {
      const contentDiv = document.getElementById('content')
      if (contentDiv) {
        contentDiv.scrollTop = contentDiv.scrollHeight
      }
    })
  }

  function handleSolution(data) {
    isLoading.value = false
    
    if (settings.keepContext && history.value.length > 0) {
      renderedContent.value = renderMarkdown(history.value[0].full)
    } else {
      renderedContent.value = renderMarkdown(data)
      if (history.value.length > 0) {
        history.value[0].full = data
        history.value[0].summary = data.substring(0, 30).replace(/\n/g, ' ') + '...'
      }
    }
  }

  function setStreamBuffer(val) {
    streamBuffer = val
  }

  return {
    renderedContent,
    history,
    activeHistoryIndex,
    isLoading,
    isAppending,
    shouldOverwriteHistory,
    errorState,
    renderMarkdown,
    selectHistory,
    handleStreamStart,
    handleStreamChunk,
    handleSolution,
    setStreamBuffer
  }
}
