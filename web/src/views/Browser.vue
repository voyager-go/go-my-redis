<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch, nextTick } from 'vue'
import { NInputGroup, NInput, NButton, NIcon, NScrollbar, NSpace, NInputNumber, NEmpty, NTimeline, NTimelineItem, NTag, NDescriptions, NDescriptionsItem, NModal, NDataTable, useMessage } from 'naive-ui'
import { redisApi } from '../api/redis'
import { connectionState } from '../services/connectionState'
import { SearchOutline, TrashOutline, SaveOutline, TerminalOutline, CloseOutline, CopyOutline, RefreshOutline } from '@vicons/ionicons5'
import '@fontsource/jetbrains-mono'
import { createTerminalService, type TerminalService } from '../services/terminal'
import { useRouter } from 'vue-router'
import 'xterm/css/xterm.css'

const router = useRouter()
const messageHandler = ref()
const searchPattern = ref('*')
const keyTree = ref<string[]>([])
const selectedKey = ref('')
const keyValue = ref('')
const keyType = ref('')
const ttl = ref(-1)
const saving = ref(false)
const deleting = ref(false)
const terminalRef = ref<HTMLElement | null>(null)
const terminalService = ref<TerminalService | null>(null)
const keyTypes = ref<Map<string, string>>(new Map())
const initialTTL = ref(-1)
const showTerminal = ref(false)
const showReconnectModal = ref(false)
const reconnecting = ref(false)
const commandHistory = ref<string[]>([])
const historyIndex = ref(-1)

// 添加确认对话框的状态
const showDeleteConfirm = ref(false)
const showSaveConfirm = ref(false)

// 添加心跳检测相关的变量
let heartbeatInterval: number | null = null
let reconnectAttempts = 0
const MAX_RECONNECT_ATTEMPTS = 3
const HEARTBEAT_INTERVAL = 15000 // 15秒发送一次心跳
const RECONNECT_INTERVAL = 5000 // 5秒重连一次

const message = useMessage()

// 在 script setup 部分添加类型定义
interface ZSetItem {
  member: string
  score: number
}

const loadKeys = async () => {
  try {
    const response = await redisApi.getKeys(searchPattern.value)
    keyTree.value = response.result || []
    selectedKey.value = ''
    keyValue.value = ''
    keyType.value = ''
    ttl.value = -1
    initialTTL.value = -1
    keyTypes.value.clear() // 清空之前的类型缓存

    // 加载前20个键的类型
    const keysToLoad = keyTree.value.slice(0, 20)
    await Promise.all(keysToLoad.map(key => loadKeyType(key)))
  } catch (error: any) {
    messageHandler.value?.error(error.response?.data?.error || '加载键列表失败')
    keyTree.value = []
    keyTypes.value.clear()
    ttl.value = -1
    initialTTL.value = -1
  }
}

// 添加懒加载类型的方法
const loadKeyType = async (key: string) => {
  if (keyTypes.value.has(key)) return // 如果已经加载过类型，就不再重复加载
  
  try {
    const typeResponse = await redisApi.type(key)
    const typeStr = String(typeResponse.result || 'unknown')
    keyTypes.value.set(key, typeStr)
  } catch (error) {
    console.error(`获取键 ${key} 的类型失败:`, error)
    keyTypes.value.set(key, 'unknown')
  }
}

// 修改滚动处理函数，只处理前20个键的类型加载
const handleScroll = async (e: Event) => {
  const target = e.target as HTMLElement
  const scrollTop = target.scrollTop
  const clientHeight = target.clientHeight
  
  // 只处理前20个键
  const visibleKeys = keyTree.value.slice(0, 20).slice(
    Math.floor(scrollTop / 40),
    Math.ceil((scrollTop + clientHeight) / 40)
  )
  
  // 只加载还未加载过类型的键
  const keysToLoad = visibleKeys.filter(key => !keyTypes.value.has(key))
  
  // 批量加载可见键的类型
  await Promise.all(keysToLoad.map(key => loadKeyType(key)))
}

const handleSearch = () => {
  loadKeys()
}

const handleKeyClick = async (key: string) => {
  try {
    selectedKey.value = key
    const typeResponse = await redisApi.type(key)
    // 从响应中提取类型值
    const typeStr = String(typeResponse.result || 'unknown')
    keyType.value = typeStr
    keyTypes.value.set(key, typeStr)
    
    // 根据类型获取数据
    let response
    switch (typeStr) {
      case 'string':
        response = await redisApi.getKey(key)
        keyValue.value = response.result || ''
        break
      case 'list':
        response = await redisApi.getList(key)
        keyValue.value = response || []
        break
      case 'set':
        response = await redisApi.getSet(key)
        keyValue.value = response || []
        break
      case 'hash':
        response = await redisApi.getHash(key)
        keyValue.value = response || {}
        break
      case 'zset':
        response = await redisApi.getZSet(key)
        keyValue.value = response || []
        break
      default:
        keyValue.value = ''
    }

    const ttlResponse = await redisApi.ttl(key)
    // 确保 TTL 是数字类型
    ttl.value = Number(ttlResponse.result)
    initialTTL.value = ttl.value
  } catch (error: any) {
    messageHandler.value?.error(error.response?.data?.error || '加载键值失败')
    keyValue.value = ''
    keyType.value = 'unknown'
    ttl.value = -1
    initialTTL.value = -1
  }
}

const handleSave = async () => {
  if (!selectedKey.value) return
  saving.value = true
  try {
    // 根据类型执行不同的命令
    switch (keyType.value) {
      case 'string':
        // 如果 keyValue.value 是对象，则转换为 JSON 字符串
        let keyValueString = keyValue.value
        if (typeof keyValueString === 'object') {
          keyValueString = JSON.stringify(keyValueString)
        }
        // 如果 keyValue.value 是数组，则转换为 JSON 字符串
        if (Array.isArray(keyValueString)) {
          keyValueString = JSON.stringify(keyValueString)
        }
        // 如果 keyValue.value 是字符串，且两个字符串之间存在空格，则需要将完整内容带上双引号
        if (typeof keyValueString === 'string' && keyValueString.includes(' ')) {
          keyValueString = `\"${keyValueString}\"`
        }
        await redisApi.executeCommand(`SET ${selectedKey.value} ${keyValueString}`)
        break
      case 'list':
        // 先删除原列表
        await redisApi.executeCommand(`DEL ${selectedKey.value}`)
        // 重新添加所有元素
        for (const item of keyValue.value) {
          await redisApi.executeCommand(`RPUSH ${selectedKey.value} ${item}`)
        }
        break
      case 'set':
        // 先删除原集合
        await redisApi.executeCommand(`DEL ${selectedKey.value}`)
        // 重新添加所有元素
        for (const item of keyValue.value) {
          await redisApi.executeCommand(`SADD ${selectedKey.value} ${item}`)
        }
        break
      case 'hash':
        // 先删除原哈希
        await redisApi.executeCommand(`DEL ${selectedKey.value}`)
        // 重新添加所有字段
        for (const [field, value] of Object.entries(keyValue.value)) {
          await redisApi.executeCommand(`HSET ${selectedKey.value} ${field} ${value}`)
        }
        break
      case 'zset':
        // 先删除原有序集合
        await redisApi.executeCommand(`DEL ${selectedKey.value}`)
        // 重新添加所有元素
        if (Array.isArray(keyValue.value)) {
          for (const item of keyValue.value as ZSetItem[]) {
            await redisApi.executeCommand(`ZADD ${selectedKey.value} ${item.score} ${item.member}`)
          }
        }
        break
      default:
        throw new Error('不支持的数据类型')
    }

    // 确保 TTL 是数字类型
    const currentTTL = Number(ttl.value)
    if (currentTTL !== initialTTL.value) {
      if (currentTTL > 0 || currentTTL === -1) {
        await redisApi.executeCommand(`EXPIRE ${selectedKey.value} ${currentTTL}`)
      }
      initialTTL.value = currentTTL
    }
    message.success('保存成功')
  } catch (error: any) {
    message.error(error.response?.data?.error || '保存失败')
  } finally {
    saving.value = false
  }
}

const handleDelete = async () => {
  if (!selectedKey.value) return
  deleting.value = true
  try {
    await redisApi.del(selectedKey.value)
    messageHandler.value?.success('删除成功')
    selectedKey.value = ''
    keyValue.value = ''
    keyType.value = ''
    ttl.value = -1
    initialTTL.value = -1
    await loadKeys()
  } catch (error: any) {
    messageHandler.value?.error(error.response?.data?.error || '删除失败')
  } finally {
    deleting.value = false
  }
}

// 添加确认对话框的处理函数
const handleDeleteConfirm = () => {
  showDeleteConfirm.value = true
}

const handleSaveConfirm = () => {
  showSaveConfirm.value = true
}

const confirmDelete = async () => {
  showDeleteConfirm.value = false
  await handleDelete()
}

const confirmSave = async () => {
  showSaveConfirm.value = false
  await handleSave()
}

const initTerminal = () => {
  if (terminalRef.value && terminalService.value) {
    try {
      terminalService.value.initTerminal(terminalRef.value, executeTerminalCommand)
    } catch (error) {
      console.error('初始化终端时出错:', error)
    }
  }
}

// 修改 executeTerminalCommand 函数
const executeTerminalCommand = async (command: string) => {
  if (!terminalService.value) return
  
  if (!command.trim()) {
    terminalService.value.write('\r\n> ')
    return
  }

  // 处理 clear 命令
  if (command.toLowerCase() === 'clear') {
    terminalService.value.clear()
    commandHistory.value = []
    historyIndex.value = -1
    terminalService.value.write('> ')
    return
  }
  
  try {
    // 直接传递原始命令给 API
    const response = await redisApi.executeCommand(command)
    const result = response.result || response.data?.result || response.data
    
    // 根据命令类型处理返回值
    const commandLower = command.toLowerCase()
    if (commandLower.startsWith('set')) {
      terminalService.value.writeln('OK')
      // 如果是 SET 命令，提取 key 并自动搜索
      const keyMatch = command.match(/set\s+(\S+)/i)
      if (keyMatch) {
        searchPattern.value = keyMatch[1]
        await loadKeys()
      }
    } else if (result === null || result === undefined) {
      terminalService.value.writeln('(nil)')
    } else if (typeof result === 'string') {
      // 如果是 GET 命令，给字符串值加上双引号
      if (commandLower.startsWith('get')) {
        terminalService.value.writeln(`"${result}"`)
      } else {
        terminalService.value.writeln(result)
      }
    } else if (Array.isArray(result)) {
      terminalService.value.writeln(JSON.stringify(result))
    } else {
      terminalService.value.writeln(JSON.stringify(result))
    }

    // 检查命令是否操作了当前选中的键
    const currentKey = selectedKey.value
    if (currentKey) {
      // 检查命令是否包含当前键
      const keyInCommand = commandLower.includes(currentKey.toLowerCase())
      // 检查是否是修改操作
      const isModifyCommand = commandLower.startsWith('set') || 
                            commandLower.startsWith('del') ||
                            commandLower.startsWith('lpush') ||
                            commandLower.startsWith('rpush') ||
                            commandLower.startsWith('lpop') ||
                            commandLower.startsWith('rpop') ||
                            commandLower.startsWith('hset') ||
                            commandLower.startsWith('hdel') ||
                            commandLower.startsWith('sadd') ||
                            commandLower.startsWith('srem') ||
                            commandLower.startsWith('zadd') ||
                            commandLower.startsWith('zrem')

      if (keyInCommand && isModifyCommand) {
        // 如果是修改操作，刷新当前键的详情
        await handleKeyClick(currentKey)
      }
    }
  } catch (error: any) {
    terminalService.value.writeln(`错误: ${error.response?.data?.error || '命令执行失败'}`)
  }
  terminalService.value.write('> ')
}

const handleResize = () => {
  terminalService.value?.resize()
}

const getTypeColor = (type: string | undefined | null) => {
  if (!type || typeof type !== 'string') return 'default'
  const typeLower = type.toLowerCase()
  switch (typeLower) {
    case 'string':
      return 'success'
    case 'list':
      return 'info'
    case 'hash':
      return 'warning'
    case 'set':
      return 'error'
    case 'zset':
      return 'info'
    default:
      return 'default'
  }
}

// 监听终端显示状态变化
watch(showTerminal, (value) => {
  nextTick(() => {
    if (value) {
      if (!terminalService.value) {
        initTerminal()
      } else {
        terminalService.value.resize()
      }
    }
  })
})

// 修改心跳检测函数
const sendHeartbeat = async () => {
  try {
    await redisApi.executeCommand('PING')
  } catch (error) {
    console.error('心跳检测失败:', error)
    if (!showReconnectModal.value) {
      showReconnectModal.value = true
    }
  }
}

// 修改重连处理函数
const handleReconnect = async () => {
  reconnecting.value = true
  try {
    const success = await connectionState.reconnect()
    if (success) {
      messageHandler.value?.success('重新连接成功')
      showReconnectModal.value = false
      reconnectAttempts = 0 // 重置重连次数
      await loadKeys()
    } else {
      reconnectAttempts++
      if (reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
        messageHandler.value?.error('重连次数过多，请检查连接配置')
        router.push('/connect')
      } else {
        messageHandler.value?.error(`重新连接失败，${MAX_RECONNECT_ATTEMPTS - reconnectAttempts}次重试机会`)
        // 延迟后自动重试
        setTimeout(handleReconnect, RECONNECT_INTERVAL)
      }
    }
  } catch (error: any) {
    reconnectAttempts++
    if (reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
      messageHandler.value?.error('重连次数过多，请检查连接配置')
      router.push('/connect')
    } else {
      messageHandler.value?.error(`重新连接失败，${MAX_RECONNECT_ATTEMPTS - reconnectAttempts}次重试机会`)
      // 延迟后自动重试
      setTimeout(handleReconnect, RECONNECT_INTERVAL)
    }
  } finally {
    reconnecting.value = false
  }
}

// 修改组件挂载和卸载逻辑
onMounted(() => {
  // 创建终端服务
  terminalService.value = createTerminalService()
  // 初始化终端
  initTerminal()
  // 加载键列表
  loadKeys()
  // 启动心跳检测
  heartbeatInterval = window.setInterval(sendHeartbeat, HEARTBEAT_INTERVAL)
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  if (terminalService.value) {
    try {
      terminalService.value.dispose()
    } catch (error) {
      console.error('清理终端时出错:', error)
    }
  }
  if (heartbeatInterval) {
    clearInterval(heartbeatInterval)
  }
  window.removeEventListener('resize', handleResize)
  if (terminalRef.value) {
    terminalRef.value = null
  }
  terminalService.value = null
})

// 修改表格列定义
const listColumns = [
  { title: '索引', key: 'index', width: 80 },
  { title: '值', key: 'value' }
]

const setColumns = [
  { title: '成员', key: 'value' }
]

const hashColumns = [
  { title: '字段', key: 'field' },
  { title: '值', key: 'value' }
]

const zsetColumns = [
  { title: '成员', key: 'member' },
  { title: '分数', key: 'score', width: 120 }
]

// 在 script setup 部分添加计算属性
const getLengthLabel = computed(() => {
  switch (keyType.value) {
    case 'list':
      return '列表长度'
    case 'set':
      return '集合大小'
    case 'hash':
      return '字段数量'
    case 'zset':
      return '成员数量'
    default:
      return '长度'
  }
})

const getLengthValue = computed(() => {
  if (!keyValue.value) return 0
  switch (keyType.value) {
    case 'list':
      return Array.isArray(keyValue.value) ? keyValue.value.length : 0
    case 'set':
      return Array.isArray(keyValue.value) ? keyValue.value.length : 0
    case 'hash':
      return Object.keys(keyValue.value).length
    case 'zset':
      return Array.isArray(keyValue.value) ? keyValue.value.length : 0
    default:
      return 0
  }
})

// 修改复制方法
const copyKey = async () => {
  try {
    await navigator.clipboard.writeText(selectedKey.value)
    message.success('已复制到剪贴板')
  } catch (error) {
    message.error('复制失败')
  }
}

// 修改刷新方法
const refreshKey = async () => {
  if (selectedKey.value) {
    message.info('正在刷新...')
    await handleKeyClick(selectedKey.value)
    message.success('刷新完成')
  }
}
</script>

<template>
  <div class="browser-container">
    <!-- 左侧面板 -->
    <div class="left-panel">
      <div class="search-area">
        <n-input-group>
          <n-input
            v-model:value="searchPattern"
            placeholder="搜索键名 (支持 * 通配符)"
            @keyup.enter="handleSearch"
          />
          <n-button type="primary" ghost @click="handleSearch">
            <n-icon><SearchOutline /></n-icon>
          </n-button>
        </n-input-group>
      </div>
      <div class="key-list">
        <n-scrollbar 
          style="max-height: calc(100vh - 180px)"
          @scroll="handleScroll"
        >
          <n-timeline item-placement="left">
            <n-timeline-item
              v-for="key in keyTree"
              :key="key"
              :type="getTypeColor(keyTypes.get(key))"
              :content="key"
              :time="keyTypes.get(key) || 'unknown'"
              :class="{ 'timeline-item-selected': key === selectedKey }"
              @click="handleKeyClick(key)"
            />
          </n-timeline>
        </n-scrollbar>
      </div>
    </div>

    <!-- 右侧面板 -->
    <div class="content-area">
      <n-card v-if="selectedKey" :bordered="false">
        <template #header>
          <n-space align="center" justify="space-between">
            <n-space>
              <span class="key-name">{{ selectedKey }}</span>
              <n-space align="center">
                <n-tag :type="getTypeColor(keyType)" size="small" style="height: 24px; line-height: 24px; padding: 0 8px;">
                  {{ keyType }}
                </n-tag>
                <!-- 居中 -->
                <n-button text size="tiny" @click="copyKey" style="height: 24px; width: 24px; padding: 0; display: flex; align-items: center; justify-content: center;">
                  <n-icon size="20"><CopyOutline /></n-icon>
                </n-button>
                <n-button text size="tiny" @click="refreshKey" style="height: 24px; width: 24px; padding: 0; display: flex; align-items: center; justify-content: center;">
                  <n-icon size="20"><RefreshOutline /></n-icon>
                </n-button>
              </n-space>
            </n-space>
          </n-space>
        </template>

        <n-space vertical size="large">
            <n-scrollbar style="max-height: calc(100vh - 180px)">
                <!-- 修改描述列表为左右布局 -->
          <n-descriptions v-if="keyType !== 'string'" bordered :column="2" size="small" label-placement="left">
            <n-descriptions-item label="类型">
              {{ keyType }}
            </n-descriptions-item>
            <n-descriptions-item :label="getLengthLabel">
              {{ getLengthValue }}
            </n-descriptions-item>
          </n-descriptions>

          <!-- TTL 设置 -->
          <n-input-number
            v-model:value="ttl"
            :min="-1"
            :max="2147483647"
            placeholder="TTL (秒)"
            style="width: 200px"
          >
            <template #prefix>TTL:</template>
          </n-input-number>

          <!-- 值编辑区域 -->
          <template v-if="keyType === 'string'">
            <n-input
              style="margin-top: 10px;"
              v-model:value="keyValue"
              type="textarea"
              placeholder="值"
              :autosize="{ minRows: 3, maxRows: 15 }"
            />
          </template>

          <!-- List 类型显示 -->
          <template v-if="keyType === 'list' && Array.isArray(keyValue)">
            <n-scrollbar style="max-height: calc(100vh - 180px)">
              <n-data-table
              :columns="listColumns"
              :data="keyValue.map((item, index) => ({ index, value: item }))"
              :pagination="{ pageSize: 10 }"
              :bordered="false"
              striped
            />
            </n-scrollbar>
          </template>

          <!-- Set 类型显示 -->
          <template v-if="keyType === 'set' && Array.isArray(keyValue)">
            <n-scrollbar style="max-height: calc(100vh - 180px)">
            <n-data-table
              :columns="setColumns"
              :data="keyValue.map(item => ({ value: item }))"
              :pagination="{ pageSize: 10 }"
              :bordered="false"
              striped
            />
            </n-scrollbar>
          </template>

          <!-- Hash 类型显示 -->
          <template v-if="keyType === 'hash' && keyValue">
           <n-scrollbar style="max-height: calc(100vh - 180px)">
            <n-data-table
              :columns="hashColumns"
              :data="Object.entries(keyValue).map(([field, value]) => ({ field, value }))"
              :pagination="{ pageSize: 10 }"
              :bordered="false"
              striped
            />
           </n-scrollbar>
          </template>

          <!-- ZSet 类型显示 -->
          <template v-if="keyType === 'zset' && Array.isArray(keyValue)">
            <n-data-table
              :columns="zsetColumns"
              :data="keyValue.map(item => ({ member: item.member, score: item.score }))"
              :pagination="{ pageSize: 10 }"
              :bordered="false"
              striped
            />
          </template>

          <!-- 操作按钮 margin-top: 10px -->
          <n-space justify="end" style="margin-top: 10px;">
            <n-button
              type="error"
              strong
              secondary
              @click="handleDeleteConfirm"
              :loading="deleting"
            >
              <template #icon>
                <n-icon><TrashOutline /></n-icon>
              </template>
              删除
            </n-button>
            <n-button
              type="success"
              strong
              secondary
              @click="handleSaveConfirm"
              :loading="saving"
            >
              <template #icon>
                <n-icon><SaveOutline /></n-icon>
              </template>
              保存
            </n-button>
          </n-space>
            </n-scrollbar>
        </n-space>
      </n-card>
      <n-empty v-else description="请选择一个键" />

      <!-- 在 content-area 最后添加 -->
      <div class="cli-trigger" @click="showTerminal = !showTerminal">
        <n-button text size="small">
          <template #icon>
            <n-icon><component :is="showTerminal ? CloseOutline : TerminalOutline" /></n-icon>
          </template>
          {{ showTerminal ? '关闭终端' : '打开终端' }}
        </n-button>
      </div>

      <!-- 终端区域 -->
      <div v-show="showTerminal" class="terminal-area">
        <div 
            ref="terminalRef" 
            class="terminal-container"
            tabindex="0"
          ></div>
      </div>
    </div>

    <!-- 重连提示模态框 -->
    <n-modal
      v-model:show="showReconnectModal"
      preset="dialog"
      type="warning"
      title="连接已断开"
      content="是否尝试重新连接？"
      positive-text="重新连接"
      negative-text="取消"
      :positive-button-loading="reconnecting"
      @positive-click="handleReconnect"
    />

    <!-- 添加确认对话框 -->
    <n-modal
      v-model:show="showDeleteConfirm"
      preset="dialog"
      type="warning"
      title="确认删除"
      content="确定要删除这个键吗？此操作不可恢复。"
      positive-text="确认删除"
      negative-text="取消"
      :positive-button-loading="deleting"
      @positive-click="confirmDelete"
    />

    <n-modal
      v-model:show="showSaveConfirm"
      preset="dialog"
      type="info"
      title="确认保存"
      content="确定要保存对当前键的修改吗？"
      positive-text="确认保存"
      negative-text="取消"
      :positive-button-loading="saving"
      @positive-click="confirmSave"
    />
  </div>
</template>

<style>
/* 添加全局样式 */
body {
  margin: 0;
  padding: 0;
  overflow: hidden;
}
</style>

<style scoped>
.browser-container {
  display: flex;
  gap: 24px;
  height: calc(100vh - 64px);
  padding: 16px;
  box-sizing: border-box;
  overflow: hidden;
  position: fixed;
  top: 64px;
  left: 0;
  right: 0;
  bottom: 0;
}

.left-panel {
  width: 320px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  height: 100%;
  overflow: hidden;
}

.search-area {
  width: 100%;
}

:deep(.n-input-group) {
  display: flex !important;
  flex-direction: row !important;
}

:deep(.n-input-group .n-input) {
  flex: 1;
}

:deep(.n-input-group .n-button) {
  margin-left: -1px;
  width: 40px;
  padding: 0;
  height: 34px;
}

.key-list {
  flex: 1;
  border: 1px solid var(--n-border-color);
  border-radius: 8px;
  overflow: hidden;
  background-color: var(--n-color);
  padding: 16px;
  position: relative;
}

:deep(.n-timeline) {
  padding-left: 0;
}

:deep(.n-timeline-item) {
  height: 40px; /* 固定每个键的高度 */
  display: flex;
  align-items: center;
  cursor: pointer;
  padding: 8px 0;
  transition: all 0.3s ease;
}

:deep(.n-timeline-item-content) {
  text-align: left;
  color: var(--n-text-color);
}

:deep(.n-timeline-item:hover) {
  .n-timeline-item-content {
    color: var(--n-primary-color);
    background: transparent;
    transition: color 0.3s ease;
  }
}

:deep(.n-timeline-item-time) {
  font-size: 12px;
  color: var(--n-text-color-3);
}

.timeline-item-selected :deep(.n-timeline-item-content) {
  color: var(--n-primary-color);
  font-weight: 500;
}

:deep(.n-timeline-item-line) {
  width: 2px;
}

:deep(.n-timeline-item-circle) {
  width: 8px;
  height: 8px;
}

:deep(.n-scrollbar-rail) {
  width: 6px !important;
}

:deep(.n-scrollbar-rail.n-scrollbar-rail--vertical) {
  right: 4px !important;
}

.content-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 16px;
  height: 100%;
  overflow: hidden;
  position: relative;
}

.key-name {
  font-size: 16px;
  font-weight: 500;
}

.full-width {
  width: 100%;
}

:deep(.n-descriptions) {
  .n-descriptions-table-wrapper {
    margin: 0;
  }
}

:deep(.n-descriptions .n-descriptions-table) {
  border-radius: 3px;
  background-color: var(--n-color);
}

:deep(.n-descriptions .n-descriptions-table-header) {
  width: 100px;
  background-color: var(--n-color);
  border-right: 1px solid var(--n-border-color);
  color: var(--n-text-color-2);
  font-weight: 500;
}

:deep(.n-descriptions .n-descriptions-table-content) {
  background-color: var(--n-color);
  color: var(--n-text-color);
}

:deep(.n-descriptions-bordered) {
  border: 1px solid var(--n-border-color);
}

:deep(.n-descriptions-bordered .n-descriptions-table-row:not(:last-child)) {
  border-bottom: 1px solid var(--n-border-color);
}

.cli-trigger {
  position: fixed;
  bottom: 16px;
  left: 16px;
  z-index: 1000;
  background-color: var(--n-color);
  border-radius: 4px;
  padding: 4px 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.terminal-area {
  position: fixed;
  right: 0;
  bottom: 0;
  width: calc(100% - 344px); /* 减去左侧面板宽度(320px)和间距(24px) */
  height: 300px;
  background-color: #1e1e1e;
  border-top: 1px solid var(--n-border-color);
  border-left: 1px solid var(--n-border-color);
  z-index: 999;
}

.terminal-container {
  width: 100%;
  height: 100%;
  padding: 8px 16px;
  box-sizing: border-box;
  background-color: #1e1e1e;
  outline: none;
}

.terminal-container:focus {
  outline: none;
}


:deep(.xterm-viewport) {
  overflow-y: auto !important;
  background-color: #1e1e1e !important;
}

:deep(.xterm-screen) {
  text-align: left;
}

:deep(.xterm-rows) {
  font-family: "JetBrains Mono", "Microsoft YaHei", "微软雅黑", monospace !important;
  font-size: 14px;
  line-height: 1.2;
  letter-spacing: 0;
}

:deep(.xterm-char) {
  width: auto !important;
}

:deep(.xterm) {
  padding: 0;
  height: 100%;
}

.details-panel {
  flex: 1;
  padding: 16px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 100%;
}

.details-content {
  width: 100%;
  max-width: 800px;
  margin: 0 auto;
  text-align: center;
}

.details-content pre {
  white-space: pre-wrap;
  word-break: break-all;
  text-align: left;
  background: #f5f5f5;
  padding: 16px;
  border-radius: 4px;
  margin: 0;
}

.details-content .n-empty {
  margin: 0 auto;
}

/* 添加空状态提示的样式 */
.content-area :deep(.n-empty) {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  margin: 0;
}

:deep(.n-tag) {
  height: 24px;
  line-height: 24px;
  padding: 0 8px;
}

:deep(.n-button.n-button--text) {
  height: 24px;
  width: 24px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  outline: none !important;
}

:deep(.n-button.n-button--text:focus) {
  outline: none !important;
  box-shadow: none !important;
}

:deep(.n-icon) {
  font-size: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
}

:deep(.n-input-number) {
  width: 200px;
}
</style>
