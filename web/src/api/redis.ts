import axios from 'axios'
import type { KeyDetail } from '../types/redis'

const api = axios.create({
  baseURL: '/api'
})

export interface ConnectionConfig {
  host: string
  port: number
  username?: string
  password?: string
  db?: number
}

export interface RedisConfig {
  host: string
  port: number
  password: string
  db: number
}

export interface RedisKey {
  key: string
  type: string
  value: any
  ttl: number
}

export interface ZSetMember {
  member: string
  score: number
}

export const redisApi = {
  // 连接管理
  connect: async (config: ConnectionConfig) => {
    const response = await api.post('/connect', config)
    return response.data
  },

  disconnect: async () => {
    const response = await api.post('/disconnect')
    return response.data
  },

  ping: async () => {
    const response = await api.post('/command', { command: 'PING' })
    return response.data
  },

  status: () => api.get('/status'),

  // 键管理
  getKey: async (key: string) => {
    const response = await api.post('/command', { command: `GET ${key}` })
    return response.data
  },
  set: async (key: string, value: any) => {
    const response = await api.post('/key', { key, value })
    return response.data
  },
  del: async (key: string) => {
    const response = await api.post('/command', { command: `DEL ${key}` })
    return response.data
  },
  type: async (key: string) => {
    const response = await api.post('/command', { command: `TYPE ${key}` })
    return response.data
  },
  ttl: async (key: string) => {
    const response = await api.post('/command', { command: `TTL ${key}` })
    return response.data
  },
  expire: async (key: string, seconds: number) => {
    const response = await api.post('/expire', { key, seconds })
    return response.data
  },

  // 命令执行
  executeCommand: async (command: string) => {
    const response = await api.post('/command', { command })
    return response.data
  },

  // List 操作
  getList: async (key: string) => {
    const response = await api.get(`/list/${encodeURIComponent(key)}`)
    return response.data
  },
  getListLength: async (key: string) => {
    const response = await api.get(`/list/${encodeURIComponent(key)}/length`)
    return response.data
  },

  // Set 操作
  getSet: async (key: string) => {
    const response = await api.get(`/set/${encodeURIComponent(key)}`)
    return response.data
  },
  getSetLength: async (key: string) => {
    const response = await api.get(`/set/${encodeURIComponent(key)}/length`)
    return response.data
  },

  // Hash 操作
  getHash: async (key: string) => {
    const response = await api.get(`/hash/${encodeURIComponent(key)}`)
    return response.data
  },
  getHashLength: async (key: string) => {
    const response = await api.get(`/hash/${encodeURIComponent(key)}/length`)
    return response.data
  },

  // ZSet 操作
  getZSet: async (key: string) => {
    const response = await api.get(`/zset/${encodeURIComponent(key)}`)
    return response.data
  },
  getZSetLength: async (key: string) => {
    const response = await api.get(`/zset/${encodeURIComponent(key)}/length`)
    return response.data
  },

  // 键值操作
  getKeys: async (pattern: string = '*') => {
    const response = await api.post('/command', { command: `KEYS ${pattern}` })
    return response.data
  },
  renameKey: (oldKey: string, newKey: string) => api.post('/command', { command: 'RENAME', args: [oldKey, newKey] }),

  // 数据库操作
  flushDb: () => api.post('/flushdb'),
  flushAll: () => api.post('/flushall'),
  info: () => {
    return api.get<string>('/info').then(res => res.data)
  },

  get: (key: string) => {
    return api.get<string>(`/get/${encodeURIComponent(key)}`).then(res => res.data)
  },

  async execute(command: string, ...args: any[]) {
    const response = await api.post('/command', { command, args })
    return response.data
  },

  async getKeyType(key: string) {
    const response = await api.post('/command', { command: 'TYPE', args: [key] })
    return response.data
  },

  async getListItems(key: string, start: number = 0, end: number = -1) {
    const response = await api.post('/command', { command: 'LRANGE', args: [key, start, end] })
    return response.data
  },

  async getSetMembers(key: string) {
    const response = await api.post('/command', { command: 'SMEMBERS', args: [key] })
    return response.data
  },

  async getHashFields(key: string) {
    const response = await api.post('/command', { command: 'HGETALL', args: [key] })
    return response.data
  },

  async getZSetMembers(key: string, start: number = 0, end: number = -1, withScores: boolean = true) {
    const command = withScores ? 'ZRANGE' : 'ZRANGE'
    const args = withScores ? [key, start, end, 'WITHSCORES'] : [key, start, end]
    const response = await api.post('/command', { command, args })
    return response.data
  }
} 