import { ref } from 'vue'
import { GetBalance } from '../../wailsjs/go/main/App'

export function useBalance(settings, statusText, statusIcon, resetStatus) {
  const balance = ref(null)
  const tempBalance = ref(null)
  const lastRefreshTime = ref(0)
  const isRefreshingBalance = ref(false)

  async function fetchBalance(force = false) {
    if (isRefreshingBalance.value || !settings.apiKey) return
    
    const now = Date.now()
    if (!force && now - lastRefreshTime.value < 5000) return

    isRefreshingBalance.value = true
    try {
      // 后端 GetBalance 改为验证 API 连通性
      await GetBalance(settings.apiKey)
      // 验证成功
      statusText.value = '已连接'
      statusIcon.value = '✅'
      balance.value = 0 // 占位符，不再显示实际金额
      lastRefreshTime.value = Date.now()
      
      // 如果之前的状态是对用户不友好的错误，重置一下
      if (statusText.value === 'Key无效' || statusText.value === '余额不足') {
        // 其实上面已经设置了'已连接'，这里逻辑保留或简化即可
      }
    } catch (e) {
      console.error('Verify API Key error', e)
      const errMsg = e ? e.toString().toLowerCase() : ''
      
      // 根据错误信息判断
      if (errMsg.includes('401') || errMsg.includes('invalid') || errMsg.includes('incorrect')) {
         statusText.value = 'Key无效'
         statusIcon.value = '🚫'
         balance.value = -1
      } else if (errMsg.includes('quota') || errMsg.includes('余额不足')) {
         // 虽然现在不查余额，但如果 list models 报 quota 错（极少见但可能），也归为 Key 无效或资源耗尽
         statusText.value = '资源耗尽'
         statusIcon.value = '💸'
         balance.value = -1
      } else {
         statusText.value = '连接失败'
         statusIcon.value = '❌'
         balance.value = -1
      }
    } finally {
      isRefreshingBalance.value = false
    }
  }

  function refreshBalance() {
    fetchBalance(true)
  }

  return {
    balance,
    tempBalance,
    isRefreshingBalance,
    fetchBalance,
    refreshBalance
  }
}
