<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch, nextTick } from 'vue'
import { useMessage, NInputGroup, NInput, NButton, NIcon, NScrollbar, NMenu, NSpace, NText, NInputNumber, NEmpty, NTimeline, NTimelineItem, NTag, NDescriptions, NDescriptionsItem, NModal } from 'naive-ui'
import { redisApi } from '../api/redis'
import { connectionState } from '../services/connectionState'
import { SearchOutline, TrashOutline, SaveOutline, ChevronDownOutline, ChevronUpOutline, TerminalOutline, CloseOutline } from '@vicons/ionicons5'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import 'xterm/css/xterm.css'
import { useRouter } from 'vue-router'

const router = useRouter()

const message = useMessage()
const searchPattern = ref('*')
const keyTree = ref<string[]>([])
const selectedKey = ref('')
const keyValue = ref('')
const keyType = ref('')
const ttl = ref(-1)
const saving = ref(false)
const deleting = ref(false)
const terminalRef = ref<HTMLElement | null>(null)
let terminal: Terminal | null = null
let fitAddon: FitAddon | null = null
const keyTypes = ref<Map<string, string>>(new Map())
const initialTTL = ref(-1)
const showTerminal = ref(false)
const showReconnectModal = ref(false)
const reconnecting = ref(false)

interface CommandResult {
  result: string
  error?: string
}

const loadKeys = async () => {
  try {
    const response = await redisApi.keys(searchPattern.value)
    keyTree.value = response || []
    selectedKey.value = ''
    keyValue.value = ''
    keyType.value = ''
    initialTTL.value = -1
    // 获取所有键的类型
    for (const key of keyTree.value) {
      const type = await redisApi.type(key)
      keyTypes.value.set(key, type)
    }
  } catch (error: any) {
    message.error(error.response?.data?.error || '加载键列表失败')
    keyTree.value = []
  }
}

const handleSearch = () => {
  loadKeys()
}

const handleKeyClick = async (key: string) => {
  try {
    selectedKey.value = key
    const type = await redisApi.type(key)
    keyType.value = type
    keyTypes.value.set(key, type)
    const response = await redisApi.getKey(key)
    keyValue.value = response.data.value
    const keyTtl = await redisApi.ttl(key)
    ttl.value = keyTtl
    initialTTL.value = keyTtl
  } catch (error: any) {
    message.error(error.response?.data?.error || '加载键值失败')
  }
}

const handleSave = async () => {
  if (!selectedKey.value) return
  saving.value = true
  try {
    await redisApi.set(selectedKey.value, keyValue.value)
    if (ttl.value !== initialTTL.value) {
      if (ttl.value > 0 || ttl.value != -1) {
        await redisApi.expire(selectedKey.value, ttl.value)
      }
      initialTTL.value = ttl.value
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
    message.success('删除成功')
    selectedKey.value = ''
    keyValue.value = ''
    keyType.value = ''
    ttl.value = -1
    initialTTL.value = -1
    await loadKeys()
  } catch (error: any) {
    message.error(error.response?.data?.error || '删除失败')
  } finally {
    deleting.value = false
  }
}

const initTerminal = () => {
  if (!terminalRef.value) return

  terminal = new Terminal({
    cursorBlink: true,
    theme: {
      background: '#1e1e1e',  // 统一使用深色背景
      foreground: '#ffffff',  // 更亮的前景色
      cursor: '#ffffff',      // 更亮的光标颜色
      black: '#000000',
      red: '#e06c75',
      green: '#98c379',
      yellow: '#d19a66',
      blue: '#61afef',
      magenta: '#c678dd',
      cyan: '#56b6c2',
      white: '#ffffff',
      brightBlack: '#5c6370',
      brightRed: '#e06c75',
      brightGreen: '#98c379',
      brightYellow: '#d19a66',
      brightBlue: '#61afef',
      brightMagenta: '#c678dd',
      brightCyan: '#56b6c2',
      brightWhite: '#ffffff',
    },
    fontSize: 14,
    fontFamily: 'Menlo, Monaco, "Courier New", monospace',
    cursorStyle: 'block',
    convertEol: true,
  })

  fitAddon = new FitAddon()
  terminal.loadAddon(fitAddon)
  terminal.loadAddon(new WebLinksAddon())

  terminal.open(terminalRef.value)
  
  // 确保终端完全打开后再调整大小
  nextTick(() => {
    fitAddon?.fit()
  })

  let currentCommand = ''

  terminal.onKey(({ key, domEvent }) => {
    if (!terminal) return
    
    const printable = !domEvent.altKey && !domEvent.ctrlKey && !domEvent.metaKey

    if (domEvent.keyCode === 13) { // Enter
      terminal.write('\r\n')
      if (currentCommand.trim()) {
        executeTerminalCommand(currentCommand)
      } else {
        terminal.write('> ')
      }
      currentCommand = ''
    } else if (domEvent.keyCode === 8) { // Backspace
      if (currentCommand.length > 0) {
        currentCommand = currentCommand.slice(0, -1)
        terminal.write('\b \b')
      }
    } else if (printable) {
      currentCommand += key
      terminal.write(key)
    }
  })

  terminal.writeln('Redis CLI - 输入命令后按回车执行')
  terminal.writeln('示例: SET key value / GET key / DEL key')
  terminal.write('\r\n> ')
}

const executeTerminalCommand = async (command: string) => {
  if (!terminal) return
  
  if (!command.trim()) {
    terminal.write('\r\n> ')
    return
  }
  
  try {
    const response = await redisApi.executeCommand(command)
    const result = response.data?.result || response.data
    if (typeof result === 'string') {
      terminal.writeln(result)
    } else if (result === null || result === undefined) {
      terminal.writeln('(nil)')
    } else {
      terminal.writeln(JSON.stringify(result))
    }
  } catch (error: any) {
    terminal.writeln(`错误: ${error.response?.data?.error || '命令执行失败'}`)
  }
  terminal.write('\r\n> ')
}

const handleResize = () => {
  fitAddon?.fit()
}

const getTypeColor = (type: string) => {
  switch (type.toLowerCase()) {
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
      if (!terminal) {
        initTerminal()
      } else {
        fitAddon?.fit()
        terminal?.refresh(0, terminal?.rows - 1)
      }
    }
  })
})

const checkConnection = async () => {
  const isConnected = await connectionState.checkConnection()
  if (!isConnected) {
    showReconnectModal.value = true
  }
}

const handleReconnect = async () => {
  reconnecting.value = true
  try {
    const success = await connectionState.reconnect()
    if (success) {
      message.success('重新连接成功')
      showReconnectModal.value = false
      await loadKeys()
    } else {
      message.error('重新连接失败')
    }
  } catch (error: any) {
    message.error(error.response?.data?.error || '重新连接失败')
  } finally {
    reconnecting.value = false
  }
}

const handleDisconnect = async () => {
  try {
    await redisApi.disconnect()
    connectionState.clearState()
    message.success('已断开连接')
    router.push('/')
  } catch (error: any) {
    message.error(error.response?.data?.error || '断开连接失败')
  }
}

onMounted(() => {
  checkConnection()
  loadKeys()
  initTerminal()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  // 清理终端
  if (terminal) {
    terminal.dispose()
    terminal = null
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
              :type="getTypeColor(keyTypes.get(key) || '')"
              :content="key"
              :time="keyTypes.get(key)"
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
          <n-descriptions bordered :column="2" size="small" label-placement="left">
            <n-descriptions-item label="类型">
              {{ keyType }}
            </n-descriptions-item>
            <n-descriptions-item label="TTL">
              {{ ttl === -1 ? '永久' : `${ttl}秒` }}
            </n-descriptions-item>
            <n-descriptions-item v-if="keyType === 'list'" label="长度">
              {{ Array.isArray(keyValue) ? keyValue.length : 0 }}
            </n-descriptions-item>
            <n-descriptions-item v-if="keyType === 'hash'" label="字段数">
              {{ typeof keyValue === 'object' ? Object.keys(keyValue).length : 0 }}
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
          <n-input
            v-model:value="keyValue"
            type="textarea"
            placeholder="值"
            :autosize="{ minRows: 3, maxRows: 15 }"
          />

          <!-- 操作按钮 -->
          <n-space justify="end">
            <n-button
              type="error"
              strong
              secondary
              @click="handleDelete"
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
              @click="handleSave"
              :loading="saving"
            >
              <template #icon>
                <n-icon><SaveOutline /></n-icon>
              </template>
              保存
            </n-button>
            <n-button
              type="warning"
              strong
              secondary
              @click="handleDisconnect"
            >
              <template #icon>
                <n-icon><CloseOutline /></n-icon>
              </template>
              断开连接
            </n-button>
          </n-space>
        </n-space>
      </n-card>
      <n-empty v-else description="请选择一个键" />

      <!-- 在 content-area 最后添加 -->
      <div class="cli-trigger" @click="showTerminal = !showTerminal">
        <n-button text size="small">
          <template #icon>
            <n-icon><component :is="showTerminal ? 'CloseOutline' : 'TerminalOutline'" /></n-icon>
          </template>
          {{ showTerminal ? '关闭终端' : '打开终端' }}
        </n-button>
      </div>

      <!-- 终端区域 -->
      <div v-show="showTerminal" class="terminal-area">
        <div ref="terminalRef" class="terminal-container"></div>
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
  background-color: #1e1e1e;  /* 确保容器也使用相同的背景色 */
}

:deep(.xterm) {
  padding: 0;
  height: 100%;
}

:deep(.xterm-viewport) {
  overflow-y: auto !important;
  background-color: #1e1e1e !important;  /* 强制使用深色背景 */
}

:deep(.xterm-screen) {
  text-align: left;
}
</style>