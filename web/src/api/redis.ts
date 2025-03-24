import axios from 'axios'
import type { KeyDetail, CommandResult } from '../types/redis'

export interface RedisConfig {
  host: string
  port: number
  password: string
  db: number
}

const api = axios.create({
  baseURL: '/api'
})

export const redisApi = {
  // 连接管理
  connect: (config: RedisConfig) => {
    return api.post('/connect', config)
  },

  disconnect: () => {
    return api.post('/disconnect')
  },

  status: () => api.get('/status'),

  // 键值操作
  getKeys: (pattern: string) => api.get<string[]>(`/keys?pattern=${encodeURIComponent(pattern)}`),
  getKey: (key: string) => api.get<KeyDetail>(`/key/${encodeURIComponent(key)}`),
  setKey: (key: string, detail: KeyDetail) => api.put(`/key/${encodeURIComponent(key)}`, detail),
  deleteKey: (key: string) => api.delete(`/key/${encodeURIComponent(key)}`),
  renameKey: (oldKey: string, newKey: string) => api.post('/rename', { oldKey, newKey }),

  // 命令执行
  executeCommand: (command: string) => api.post<CommandResult>('/command', { command }),

  // 数据库操作
  flushDb: () => api.post('/flushdb'),
  flushAll: () => api.post('/flushall'),
  info: () => {
    return api.get<string>('/info').then(res => res.data)
  },

  keys: (pattern: string) => {
    return api.get<{rdb_keys: string[]}>(`/keys?pattern=${encodeURIComponent(pattern)}`).then(res => res.data.rdb_keys)
  },

  get: (key: string) => {
    return api.get<string>(`/get/${encodeURIComponent(key)}`).then(res => res.data)
  },

  set: (key: string, value: string) => {
    return api.post('/key', { key, value })
  },

  del: (key: string) => {
    return api.delete(`/key/${encodeURIComponent(key)}`)
  },

  type: (key: string) => {
    return api.get<string>(`/type/${encodeURIComponent(key)}`).then(res => res.data)
  },

  ttl: (key: string) => {
    return api.get<number>(`/ttl/${encodeURIComponent(key)}`).then(res => res.data)
  },

  expire: (key: string, seconds: number) => {
    return api.post('/expire', { key, seconds })
  }
} 