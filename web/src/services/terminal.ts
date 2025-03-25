import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import { WebglAddon } from '@xterm/addon-webgl'
import 'xterm/css/xterm.css'

export interface TerminalService {
  terminal: Terminal | null
  fitAddon: FitAddon | null
  commandHistory: string[]
  historyIndex: number
  currentCommand: string
  cursorPosition: number
  isComposing: boolean
  initTerminal: (container: HTMLElement, onExecuteCommand: (command: string) => Promise<void>) => void
  executeCommand: (command: string) => Promise<void>
  handleResize: () => void
  dispose: () => void
  write: (data: string) => void
  writeln: (data: string) => void
  clear: () => void
  resize: () => void
}

export const createTerminalService = (): TerminalService => {
  let terminal: Terminal | null = null
  let fitAddon: FitAddon | null = null
  let currentCommand = ''
  let cursorPosition = 0
  let isComposing = false
  const commandHistory: string[] = []
  let historyIndex = -1
  let executeCommandCallback: ((command: string) => Promise<void>) | null = null

  const write = (data: string) => {
    terminal?.write(data)
  }

  const writeln = (data: string) => {
    terminal?.writeln(data)
  }

  const clear = () => {
    terminal?.clear()
  }

  const resize = () => {
    fitAddon?.fit()
  }

  const initTerminal = (container: HTMLElement, onExecuteCommand: (command: string) => Promise<void>) => {
    executeCommandCallback = onExecuteCommand
    terminal = new Terminal({
      cursorBlink: true,
      cursorWidth: 2,
      fontSize: 14,
      fontFamily: 'Consolas, "Courier New", "JetBrains Mono", "Microsoft YaHei", "微软雅黑", monospace',
      fontWeight: 'normal',
      lineHeight: 1.2,
      letterSpacing: 0.5,
      theme: {
        background: '#1e1e1e',
        foreground: '#d4d4d4',
        cursor: '#ffffff',
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
      cursorStyle: 'block',
      convertEol: true,
      allowTransparency: true,
      scrollback: 10000,
      rightClickSelectsWord: true,
      macOptionIsMeta: true,
      macOptionClickForcesSelection: true,
      windowsMode: false,
    })

    fitAddon = new FitAddon()
    const webglAddon = new WebglAddon()

    terminal.loadAddon(fitAddon)
    terminal.loadAddon(new WebLinksAddon())
    terminal.loadAddon(webglAddon)

    terminal.open(container)

    // 绑定输入法事件
    container.addEventListener('compositionstart', () => {
      isComposing = true
    })

    container.addEventListener('compositionupdate', (e: CompositionEvent) => {
      if (isComposing && terminal) {
        terminal.write('\b'.repeat(e.data?.length || 0))
      }
    })

    // 监听输入法结束事件，改变了命令字符顺序，需要调整
    container.addEventListener('compositionend', (e: CompositionEvent) => {
      isComposing = false
      if (e.data && terminal) {
        currentCommand = currentCommand.slice(0, cursorPosition) + e.data + currentCommand.slice(cursorPosition)
        cursorPosition += e.data.length
      }
    })

    terminal.onKey((event) => {
      if (event.key === '\u0003' && event.domEvent.ctrlKey) {
        // 清空当前行，保留 > 并回到行首
        terminal?.write('\r\x1b[K> ') 
      }
    })

    // 使用 onData 监听输入，支持中文和常规按键
    terminal.onData((data) => {
      if (!terminal) return

      if (isComposing) return // 输入法组合中，不处理按键

      const charCode = data.charCodeAt(0)
      if (charCode === 13) { // Enter
        terminal.write('\r\n')
        if (currentCommand.trim()) {
          commandHistory.push(currentCommand)
          historyIndex = commandHistory.length
          executeCommand(currentCommand)
        } else {
          terminal.write('> ')
        }
        currentCommand = ''
        cursorPosition = 0
      } else if (charCode === 127) { // Backspace
        if (cursorPosition > 0) {
          currentCommand = currentCommand.slice(0, cursorPosition - 1) + currentCommand.slice(cursorPosition)
          cursorPosition--
          terminal.write('\r\x1b[K> ' + currentCommand)
          if (cursorPosition > 0) {
            terminal.write('\r> ' + currentCommand.slice(0, cursorPosition))
          }
        }
      } else if (data === '\x1b[3~') { // Delete 键
        if (cursorPosition < currentCommand.length) {
          currentCommand = currentCommand.slice(0, cursorPosition) + currentCommand.slice(cursorPosition + 1)
          terminal.write('\r\x1b[K> ' + currentCommand)
          if (cursorPosition > 0) {
            terminal.write('\r> ' + currentCommand.slice(0, cursorPosition))
          }
        }
      } else if (charCode === 27) { // 方向键
        if (data === '\x1b[A') { // Up arrow
          if (historyIndex > 0) {
            historyIndex--
            currentCommand = commandHistory[historyIndex]
            cursorPosition = currentCommand.length
            terminal.write('\r\x1b[K> ' + currentCommand)
          }
        } else if (data === '\x1b[B') { // Down arrow
          if (historyIndex < commandHistory.length - 1) {
            historyIndex++
            currentCommand = commandHistory[historyIndex]
            cursorPosition = currentCommand.length
            terminal.write('\r\x1b[K> ' + currentCommand)
          } else {
            historyIndex = commandHistory.length
            currentCommand = ''
            cursorPosition = 0
            terminal.write('\r\x1b[K> ')
          }
        } else if (data === '\x1b[D') { // Left arrow
          if (cursorPosition > 0) {
            cursorPosition--
            terminal.write('\b')
          }
        } else if (data === '\x1b[C') { // Right arrow
          if (cursorPosition < currentCommand.length) {
            terminal.write(currentCommand[cursorPosition])
            cursorPosition++
          }
        }
      } else { // 普通字符输入
        // 处理粘贴的文本
        if (data.length > 1) {
          // 如果是粘贴的文本，直接插入到当前位置
          currentCommand = currentCommand.slice(0, cursorPosition) + data + currentCommand.slice(cursorPosition)
          cursorPosition += data.length
          terminal.write(data)
        } else {
          // 单个字符输入
          currentCommand = currentCommand.slice(0, cursorPosition) + data + currentCommand.slice(cursorPosition)
          cursorPosition++
          terminal.write(data)
        }
      }
    })

    terminal.writeln('Redis CLI - 输入命令后按回车执行(暂不支持中文输入)')
    terminal.writeln('示例: SET key value / GET key / DEL key')
    terminal.write('\r\n> ')
  }

  const executeCommand = async (command: string) => {

    if (!terminal) return

    if (!command.trim()) {
      terminal.write('\r\n> ')
      return
    }

    // 处理 clear 命令
    if (command.toLowerCase() === 'clear') {
      terminal.clear()
      commandHistory.length = 0
      historyIndex = -1
      terminal.write('> ')
      return
    }

    try {
      if (executeCommandCallback) {
        // 直接传递原始命令，不做任何处理
        await executeCommandCallback(command)
      } else {
        terminal.writeln('命令执行功能未初始化')
      }
    } catch (error: any) {
      terminal.writeln(`错误: ${error.response?.data?.error || '命令执行失败'}`)
    }
  }

  const handleResize = () => {
    fitAddon?.fit()
  }

  const dispose = () => {
    if (terminal) {
      terminal.dispose()
      terminal = null
    }
    if (fitAddon) {
      fitAddon = null
    }
  }

  return {
    terminal,
    fitAddon,
    commandHistory,
    historyIndex,
    currentCommand,
    cursorPosition,
    isComposing,
    initTerminal,
    executeCommand,
    handleResize,
    dispose,
    write,
    writeln,
    clear,
    resize
  }
} 