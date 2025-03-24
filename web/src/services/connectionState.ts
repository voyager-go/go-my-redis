import { ref } from 'vue'
import { redisApi } from '../api/redis'

const isConnected = ref(false)
const connectionConfig = ref<{
  host: string
  port: number
  password?: string
} | null>(null)

export const connectionState = {
  isConnected,
  connectionConfig,

  setConnected(connected: boolean) {
    isConnected.value = connected
  },

  setConfig(config: { host: string; port: number; password?: string }) {
    connectionConfig.value = config
  },

  clearState() {
    isConnected.value = false
    connectionConfig.value = null
  },

  async checkConnection() {
    try {
      const response = await redisApi.ping()
      isConnected.value = response.data === 'PONG'
      return isConnected.value
    } catch (error) {
      isConnected.value = false
      return false
    }
  },

  async reconnect() {
    if (!connectionConfig.value) return false
    try {
      await redisApi.connect(connectionConfig.value)
      isConnected.value = true
      return true
    } catch (error) {
      isConnected.value = false
      return false
    }
  },

  async disconnect() {
    try {
      await redisApi.disconnect()
      isConnected.value = false
      connectionConfig.value = null
      return { success: true, message: '已断开连接' }
    } catch (error: any) {
      let errorMessage = '断开连接失败'
      if (error && typeof error === 'object') {
        if (error.response?.data?.error) {
          errorMessage = error.response.data.error
        } else if (error.message) {
          errorMessage = error.message
        }
      }
      isConnected.value = false
      connectionConfig.value = null
      return { success: false, message: errorMessage }
    }
  }
}
