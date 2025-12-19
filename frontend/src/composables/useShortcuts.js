import { reactive, ref } from 'vue'
import { StartRecordingKey } from '../../wailsjs/go/main/App'

export function useShortcuts() {
  const shortcuts = reactive({})
  const tempShortcuts = reactive({})
  const recordingAction = ref(null)
  const recordingText = ref('')

  const shortcutActions = [
    { action: 'solve', label: '一键解题', default: 'Alt+~' },
    { action: 'toggle', label: '隐藏/展示', default: 'Alt+H' },
    { action: 'clickthrough', label: '鼠标穿透', default: 'Alt+T' },
    { action: 'move_up', label: '向上移动', default: 'Alt+↑' },
    { action: 'move_down', label: '向下移动', default: 'Alt+↓' },
    { action: 'move_left', label: '向左移动', default: 'Alt+←' },
    { action: 'move_right', label: '向右移动', default: 'Alt+→' },
    { action: 'scroll_up', label: '向上滚动', default: 'Alt+PgUp' },
    { action: 'scroll_down', label: '向下滚动', default: 'Alt+PgDn' },
  ]

  function recordKey(action) {
    recordingAction.value = action
    recordingText.value = '请按键...'
    StartRecordingKey(action)
  }

  return {
    shortcuts,
    tempShortcuts,
    recordingAction,
    recordingText,
    shortcutActions,
    recordKey
  }
}
