export interface ConnectionRecord {
  host: string
  port: number
  password: string
  db: number
  timestamp: number
}

const STORAGE_KEY = 'redis_connection_history'
const MAX_RECORDS = 10

export const connectionHistory = {
  getHistory(): ConnectionRecord[] {
    try {
      const history = localStorage.getItem(STORAGE_KEY)
      console.log('从 localStorage 读取历史记录:', history)
      return history ? JSON.parse(history) : []
    } catch (error) {
      console.error('读取历史记录失败:', error)
      return []
    }
  },

  addRecord(record: Omit<ConnectionRecord, 'timestamp'>) {
    try {
      const history = this.getHistory()
      const newRecord = {
        ...record,
        timestamp: Date.now()
      }
      
      // 移除重复记录
      const filteredHistory = history.filter(
        h => h.host !== record.host || h.port !== record.port
      )
      
      // 添加新记录并保持最新的10条
      const updatedHistory = [newRecord, ...filteredHistory].slice(0, MAX_RECORDS)
      console.log('保存历史记录:', updatedHistory)
      localStorage.setItem(STORAGE_KEY, JSON.stringify(updatedHistory))
    } catch (error) {
      console.error('保存历史记录失败:', error)
    }
  },

  clearHistory() {
    try {
      localStorage.removeItem(STORAGE_KEY)
      console.log('历史记录已清除')
    } catch (error) {
      console.error('清除历史记录失败:', error)
    }
  }
} 