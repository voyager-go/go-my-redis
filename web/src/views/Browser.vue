<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch, nextTick } from 'vue'
import { NInputGroup, NInput, NButton, NIcon, NScrollbar, NSpace, NText, NInputNumber, NEmpty, NTimeline, NTimelineItem, NTag, NDescriptions, NDescriptionsItem, NModal, NDataTable } from 'naive-ui'
import { redisApi } from '../api/redis'
import { connectionState } from '../services/connectionState'
import { SearchOutline, TrashOutline, SaveOutline, TerminalOutline, CloseOutline } from '@vicons/ionicons5'
import '@fontsource/jetbrains-mono'
import { createTerminalService, type TerminalService } from '../services/terminal'
import { useRouter } from 'vue-router'
import MessageHandler from '../components/MessageHandler.vue'
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

const loadKeys = async () => {
  try {
    const response = await redisApi.getKeys(searchPattern.value)
    keyTree.value = response.rdb_keys || []
    selectedKey.value = ''
    keyValue.value = ''
    keyType.value = ''
    ttl.value = -1
    initialTTL.value = -1
    keyTypes.value.clear() // 清空之前的类型缓存
    
    // 获取所有键的类型
    const typePromises = keyTree.value.map(async (key) => {
      try {
        const typeResponse = await redisApi.type(key)
        // 从响应中提取类型值
        const typeStr = String(typeResponse.data || 'unknown')
        keyTypes.value.set(key, typeStr)
      } catch (error) {
        console.error(`获取键 ${key} 的类型失败:`, error)
        keyTypes.value.set(key, 'unknown')
      }
    })
    
    await Promise.all(typePromises)
  } catch (error: any) {
    messageHandler.value?.error(error.response?.data?.error || '加载键列表失败')
    keyTree.value = []
    keyTypes.value.clear()
    ttl.value = -1
    initialTTL.value = -1
  }
}

const handleSearch = () => {
  loadKeys()
}

const handleKeyClick = async (key: string) => {
  try {
    selectedKey.value = key
    const typeResponse = await redisApi.type(key)
    // 从响应中提取类型值
    const typeStr = String(typeResponse.data || 'unknown')
    keyType.value = typeStr
    keyTypes.value.set(key, typeStr)
    
    // 根据类型获取数据
    let response
    switch (typeStr) {
      case 'string':
        response = await redisApi.getKey(key)
        keyValue.value = response.value || ''
        break
      case 'list':
        response = await redisApi.getList(key)
        keyValue.value = response.data
        break
      case 'set':
        response = await redisApi.getSet(key)
        keyValue.value = response.data
        break
      case 'hash':
        response = await redisApi.getHash(key)
        keyValue.value = response.data
        break
      case 'zset':
        response = await redisApi.getZSet(key)
        keyValue.value = response.data
        break
      default:
        keyValue.value = ''
    }

    const ttlResponse = await redisApi.ttl(key)
    // 确保 TTL 是数字类型
    ttl.value = Number(ttlResponse.data)
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
    await redisApi.set(selectedKey.value, keyValue.value)
    // 确保 TTL 是数字类型
    const currentTTL = Number(ttl.value)
    if (currentTTL !== initialTTL.value) {
      if (currentTTL > 0 || currentTTL === -1) {
        await redisApi.expire(selectedKey.value, currentTTL)
      }
      initialTTL.value = currentTTL
    }
    messageHandler.value?.success('保存成功')
  } catch (error: any) {
    messageHandler.value?.error(error.response?.data?.error || '保存失败')
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
  if (!terminalRef.value) return
  terminalService.value = createTerminalService()
  terminalService.value.initTerminal(terminalRef.value, executeTerminalCommand)
}

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
    const response = await redisApi.executeCommand(command)
    const result = response.data?.result || response.data
    if (typeof result === 'string') {
      terminalService.value.writeln(result)
    } else if (result === null || result === undefined) {
      terminalService.value.writeln('(nil)')
    } else {
      terminalService.value.writeln(JSON.stringify(result))
    }

    // 检查命令是否操作了当前选中的键
    const commandLower = command.toLowerCase()
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

const checkConnection = async () => {
  try {
    const isConnected = await connectionState.checkConnection()
    if (!isConnected) {
      messageHandler.value?.error('连接已断开')
      router.push('/connect')
    }
  } catch (error) {
    messageHandler.value?.error('连接检查失败')
    router.push('/connect')
  }
}

const handleDisconnect = async () => {
  const result = await connectionState.disconnect()
  if (result.success) {
    messageHandler.value?.success(result.message)
    router.push('/connect')
  } else {
    messageHandler.value?.error(result.message)
    router.push('/connect')
  }
}

const handleReconnect = async () => {
  reconnecting.value = true
  try {
    const success = await connectionState.reconnect()
    if (success) {
      messageHandler.value?.success('重新连接成功')
      showReconnectModal.value = false
      await loadKeys()
    } else {
      messageHandler.value?.error('重新连接失败')
    }
  } catch (error: any) {
    messageHandler.value?.error(error.response?.data?.error || '重新连接失败')
  } finally {
    reconnecting.value = false
  }
}

// 定期检查连接状态
let checkInterval: number | null = null

onMounted(() => {
  // 初始化终端
  initTerminal()
  // 加载键列表
  loadKeys()
  // 开始定期检查连接状态，每60秒检查一次
  checkInterval = window.setInterval(checkConnection, 60000)
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  // 清理定时器
  if (checkInterval) {
    clearInterval(checkInterval)
  }
  // 清理终端
  terminalService.value?.dispose()
  window.removeEventListener('resize', handleResize)
  if (terminalRef.value) {
    terminalRef.value.removeEventListener('paste', () => {});
    terminalRef.value.removeEventListener('compositionstart', () => {});
    terminalRef.value.removeEventListener('compositionupdate', () => {});
    terminalRef.value.removeEventListener('compositionend', () => {});
  }
});

// 在 script setup 部分添加表格列定义
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
        <n-scrollbar style="max-height: calc(100vh - 180px)">
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
              <n-tag :type="getTypeColor(keyType)" size="small">
                {{ keyType }}
              </n-tag>
            </n-space>
          </n-space>
        </template>

        <n-space vertical size="large">
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
            class="full-width"
          >
            <template #prefix>TTL:</template>
          </n-input-number>

          <!-- 值编辑区域 -->
          <template v-if="keyType === 'string'">
            <n-input
              v-model:value="keyValue"
              type="textarea"
              placeholder="值"
              :autosize="{ minRows: 3, maxRows: 15 }"
            />
          </template>

          <!-- List 类型显示 -->
          <template v-if="keyType === 'list'">
            <n-data-table
              :columns="listColumns"
              :data="Array.isArray(keyValue) ? keyValue.map((item, index) => ({ index, value: String(item) })) : []"
              :pagination="{ pageSize: 10 }"
              :bordered="false"
              striped
            />
          </template>

          <!-- Set 类型显示 -->
          <template v-if="keyType === 'set'">
            <n-data-table
              :columns="setColumns"
              :data="Array.isArray(keyValue) ? keyValue.map(item => ({ value: String(item) })) : []"
              :pagination="{ pageSize: 10 }"
              :bordered="false"
              striped
            />
          </template>

          <!-- Hash 类型显示 -->
          <template v-if="keyType === 'hash'">
            <n-data-table
              :columns="hashColumns"
              :data="Object.entries(keyValue || {}).map(([field, value]) => ({ field, value: String(value) }))"
              :pagination="{ pageSize: 10 }"
              :bordered="false"
              striped
            />
          </template>

          <!-- ZSet 类型显示 -->
          <template v-if="keyType === 'zset'">
            <n-data-table
              :columns="zsetColumns"
              :data="Array.isArray(keyValue) ? keyValue.map(item => ({ member: String(item.member), score: item.score })) : []"
              :pagination="{ pageSize: 10 }"
              :bordered="false"
              striped
            />
          </template>

          <!-- 操作按钮 -->
          <n-space justify="end">
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
}

:deep(.n-timeline) {
  padding-left: 0;
}

:deep(.n-timeline-item) {
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
  right: 16px;
  z-index: 1000;
}

.terminal-area {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  height: 300px;
  background-color: #1e1e1e;  /* 直接使用固定的深色背景 */
  border-top: 1px solid var(--n-border-color);
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
</style>
