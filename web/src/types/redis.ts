export interface RedisConfig {
  host: string
  port: number
  password: string
  db: number
}

export interface RedisKey {
  key: string
  type: string
  ttl: number
  size: number
}

export interface KeyDetail {
  type: string
  value: any
  ttl: number
}

export interface CommandResult {
  success: boolean
  data: any
  error?: string
}

export interface RedisApi {
  executeCommand: (command: string) => Promise<string>
} 