import axios from 'axios'
import type { KeyDetail } from '../types/redis'

const api = axios.create({
  baseURL: '/api'
})

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
  connect: async (config: { host: string; port: number; password?: string }) => {
    const response = await api.post('/connect', config)
    return response.data
  },

  disconnect: async () => {
    const response = await api.post('/disconnect')
    return response.data
  },

  ping: () => api.post('/command', { command: 'PING' }),

  status: () => api.get('/status'),

  // 键管理
  getKey: async (key: string) => {
    const response = await api.get(`/key/${encodeURIComponent(key)}`)
    return response.data
  },
  set: (key: string, value: any) => api.post('/key', { key, value }),
  del: (key: string) => api.delete(`/key/${encodeURIComponent(key)}`),
  type: (key: string) => api.get(`/type/${encodeURIComponent(key)}`),
  ttl: (key: string) => api.get(`/ttl/${encodeURIComponent(key)}`),
  expire: (key: string, seconds: number) => api.post('/expire', { key, seconds }),

  // 命令执行
  executeCommand: async (command: string) => {
    const response = await api.post('/command', { command })
    return response.data
  },

  // List 操作
  getList: (key: string) => api.get(`/list/${encodeURIComponent(key)}`),
  getListLength: (key: string) => api.get(`/list/${encodeURIComponent(key)}/length`),

  // Set 操作
  getSet: (key: string) => api.get(`/set/${encodeURIComponent(key)}`),
  getSetLength: (key: string) => api.get(`/set/${encodeURIComponent(key)}/length`),

  // Hash 操作
  getHash: (key: string) => api.get(`/hash/${encodeURIComponent(key)}`),
  getHashLength: (key: string) => api.get(`/hash/${encodeURIComponent(key)}/length`),

  // ZSet 操作
  getZSet: (key: string) => api.get(`/zset/${encodeURIComponent(key)}`),
  getZSetLength: (key: string) => api.get(`/zset/${encodeURIComponent(key)}/length`),

  // 键值操作
  getKeys: async (pattern: string = '*') => {
    const response = await api.get('/keys', { params: { pattern } })
    return response.data
  },
  renameKey: (oldKey: string, newKey: string) => api.post('/rename', { oldKey, newKey }),

  // 数据库操作
  flushDb: () => api.post('/flushdb'),
  flushAll: () => api.post('/flushall'),
  info: () => {
    return api.get<string>('/info').then(res => res.data)
  },

  get: (key: string) => {
    return api.get<string>(`/get/${encodeURIComponent(key)}`).then(res => res.data)
  }
} 