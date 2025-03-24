<template>
  <div class="terminal-container" ref="terminalRef">
    <div class="terminal-header">
      <n-text>Redis 终端</n-text>
      <n-space>
        <n-button quaternary circle size="small" @click="clearTerminal">
          <template #icon>
            <n-icon><TrashOutline /></n-icon>
          </template>
        </n-button>
        <n-button quaternary circle size="small" @click="copyToClipboard">
          <template #icon>
            <n-icon><CopyOutline /></n-icon>
          </template>
        </n-button>
      </n-space>
    </div>
    <div class="terminal" ref="xtermRef"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import { WebLinksAddon } from 'xterm-addon-web-links'
import { TrashOutline, CopyOutline } from '@vicons/ionicons5'
import { useMessage } from 'naive-ui'
import { redisApi } from '../api/redis'
import 'xterm/css/xterm.css'

const message = useMessage()
const terminalRef = ref<HTMLElement>()
const xtermRef = ref<HTMLElement>()
let terminal: Terminal
let fitAddon: FitAddon

const initTerminal = () => {
  if (!xtermRef.value) return

  terminal = new Terminal({
    cursorBlink: true,
    fontSize: 14,
    fontFamily: 'Fira Code, Menlo, Monaco, monospace',
    theme: {
      background: '#1e1e1e',
      foreground: '#ffffff',
      cursor: '#ffffff',
      black: '#000000',
      red: '#cd3131',
      green: '#0dbc79',
      yellow: '#e5e510',
      blue: '#2472c8',
      magenta: '#bc3fbc',
      cyan: '#11a8cd',
      white: '#e5e5e5',
      brightBlack: '#666666',
      brightRed: '#f14c4c',
      brightGreen: '#23d18b',
      brightYellow: '#f5f543',
      brightBlue: '#3b8eea',
      brightMagenta: '#d670d6',
      brightCyan: '#29b8db',
      brightWhite: '#e5e5e5'
    }
  })

  fitAddon = new FitAddon()
  terminal.loadAddon(fitAddon)
  terminal.loadAddon(new WebLinksAddon())

  terminal.open(xtermRef.value)
  fitAddon.fit()

  let commandBuffer = ''
  terminal.writeln('Redis CLI - 输入命令并按回车执行')
  terminal.writeln('输入 help 查看帮助信息\n')
  terminal.write('\x1b[32mredis>\x1b[0m ')

  terminal.onData((data) => {
    // 处理退格键
    if (data === '\u007f') {
      if (commandBuffer.length > 0) {
        commandBuffer = commandBuffer.slice(0, -1)
        terminal.write('\b \b')
      }
      return
    }

    // 处理回车键
    if (data === '\r') {
      terminal.writeln('')
      if (commandBuffer.trim()) {
        executeCommand(commandBuffer.trim())
      } else {
        terminal.write('\x1b[32mredis>\x1b[0m ')
      }
      commandBuffer = ''
      return
    }

    // 处理普通字符输入
    commandBuffer += data
    terminal.write(data)
  })
}

const executeCommand = async (command: string) => {
  try {
    const { data } = await redisApi.executeCommand(command)
    if (data.success) {
      if (data.data !== null) {
        if (typeof data.data === 'object') {
          terminal.writeln('\x1b[36m' + JSON.stringify(data.data, null, 2) + '\x1b[0m')
        } else {
          terminal.writeln('\x1b[36m' + data.data + '\x1b[0m')
        }
      } else {
        terminal.writeln('\x1b[90m(nil)\x1b[0m')
      }
    } else {
      terminal.writeln('\x1b[31mError: ' + (data.error || 'Unknown error') + '\x1b[0m')
    }
  } catch (error: any) {
    terminal.writeln('\x1b[31mError: ' + (error.response?.data?.error || error.message) + '\x1b[0m')
  }
  terminal.write('\x1b[32mredis>\x1b[0m ')
}

const clearTerminal = () => {
  terminal.clear()
  terminal.writeln('Redis CLI - 输入命令并按回车执行')
  terminal.writeln('输入 help 查看帮助信息\n')
  terminal.write('\x1b[32mredis>\x1b[0m ')
}

const copyToClipboard = async () => {
  try {
    await navigator.clipboard.writeText(terminal.buffer.active.getLine(terminal.buffer.active.length - 1)?.translateToString() || '')
    message.success('已复制到剪贴板')
  } catch (error) {
    message.error('复制失败')
  }
}

onMounted(() => {
  initTerminal()
  window.addEventListener('resize', () => fitAddon?.fit())
})

onUnmounted(() => {
  window.removeEventListener('resize', () => fitAddon?.fit())
  terminal?.dispose()
})
</script>

<style scoped>
.terminal-container {
  height: 100%;
  background: #1E1E1E;
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.terminal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  background: #252526;
  border-bottom: 1px solid #333;
}

.terminal {
  flex: 1;
  padding: 8px;
}

:deep(.xterm) {
  padding: 8px;
  height: 100%;
}
</style> 