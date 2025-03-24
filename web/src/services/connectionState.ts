import { redisApi } from '../api/redis'

interface ConnectionState {
  host: string
  port: number
  password: string
  db: number
  isConnected: boolean
}

const STORAGE_KEY = 'redis_connection_state'

export const connectionState = {
  getState(): ConnectionState | null {
    try {
      const state = localStorage.getItem(STORAGE_KEY)
      return state ? JSON.parse(state) : null
    } catch (error) {
      console.error('读取连接状态失败:', error)
      return null
    }
  },

  setState(state: ConnectionState) {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(state))
    } catch (error) {
      console.error('保存连接状态失败:', error)
    }
  },

  clearState() {
    try {
      localStorage.removeItem(STORAGE_KEY)
    } catch (error) {
      console.error('清除连接状态失败:', error)
    }
  },

  async checkConnection(): Promise<boolean> {
    try {
      await redisApi.ping()
      return true
    } catch (error) {
      return false
    }
  },

  async reconnect(): Promise<boolean> {
    const state = this.getState()
    if (!state) return false

    try {
      await redisApi.connect({
        host: state.host,
        port: state.port,
        password: state.password,
        db: state.db
      })
      return true
    } catch (error) {
      return false
    }
  }
} 