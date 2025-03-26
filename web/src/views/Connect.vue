<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage, NCard, NSpace, NForm, NFormItem, NInput, NInputNumber, NButton, NList, NListItem, NThing, NIcon, NCollapseTransition } from 'naive-ui'
import { redisApi } from '../api/redis'
import { ServerOutline, ChevronDownOutline, ChevronUpOutline, TrashOutline } from '@vicons/ionicons5'

const router = useRouter()
const message = useMessage()
const loading = ref(false)

interface ConnectionHistory {
  host: string
  port: number
  username: string
  password: string
  db: number
  sessionName: string
  timestamp: number
}

const formData = ref({
  host: 'localhost',
  port: 6379,
  username: '',
  password: '',
  db: 0,
  sessionName: 'Localhost'
})

const connectionHistory = ref<ConnectionHistory[]>([])

const rules = {
  host: {
    required: true,
    message: '请输入主机地址',
    trigger: 'blur'
  },
  port: {
    required: true,
    message: '请输入端口号',
    trigger: 'blur'
  },
  db: {
    required: true,
    message: '请选择数据库',
    trigger: 'blur'
  }
}

const loadHistory = () => {
  const history = localStorage.getItem('redis_connection_history')
  if (history) {
    try {
      const parsedHistory = JSON.parse(history)
      connectionHistory.value = parsedHistory.slice(0, 10)
    } catch (error) {
      console.error('加载历史记录失败:', error)
      connectionHistory.value = []
    }
  }
}

const saveToHistory = () => {
  const newConnection: ConnectionHistory = {
    ...formData.value,
    timestamp: Date.now()
  }
  
  connectionHistory.value = connectionHistory.value.filter(
    (item: ConnectionHistory) => 
      item.host !== newConnection.host || 
      item.port !== newConnection.port || 
      item.db !== newConnection.db
  )
  
  connectionHistory.value.unshift(newConnection)
  
  connectionHistory.value = connectionHistory.value.slice(0, 10)
  
  localStorage.setItem('redis_connection_history', JSON.stringify(connectionHistory.value))
}

onMounted(() => {
  loadHistory()
})

const handleConnect = async () => {
  loading.value = true
  try {
    const success = await redisApi.connect(formData.value)
    if (success) {
      message.success('连接成功')
      saveToHistory()
      router.push('/browser')
    } else {
      message.error('连接失败')
    }
  } catch (error: any) {
    message.error(error.response?.data?.error || '连接失败')
  } finally {
    loading.value = false
  }
}

// 点击历史会话应该自动连接   
const handleHistoryClick = (history: ConnectionHistory) => {
  formData.value = {
    ...history,
    sessionName: history.sessionName || 'Localhost'
  }
  handleConnect()
}

const handleClearHistory = () => {
  connectionHistory.value = []
  localStorage.removeItem('redis_connection_history')
  message.success('历史记录已清除')
}

const showHistory = ref(false)

const toggleHistory = () => {
  showHistory.value = !showHistory.value
}
</script>

<template>
  <div class="connect-container">
    <div class="wave-background">
      <img src="@/assets/wave.svg" alt="wave" class="wave-svg">
    </div>
    <div class="connect-content">
      <n-card title="Redis 连接配置" :bordered="false">
        <n-form
          ref="formRef"
          :model="formData"
          :rules="rules"
          label-placement="left"
          label-width="100"
          require-mark-placement="right-hanging"
        >
          <n-form-item label="会话名称" path="sessionName">
            <n-input v-model:value="formData.sessionName" placeholder="输入会话名称" />
          </n-form-item>
          
          <n-form-item label="主机地址" path="host">
            <n-input v-model:value="formData.host" placeholder="输入主机地址" />
          </n-form-item>
          
          <n-form-item label="端口" path="port">
            <n-input-number v-model:value="formData.port" :min="1" :max="65535" placeholder="输入端口" />
          </n-form-item>
          
          <n-form-item label="用户名" path="username">
            <n-input v-model:value="formData.username" placeholder="输入用户名（可选）" />
          </n-form-item>
          
          <n-form-item label="密码" path="password">
            <n-input
              v-model:value="formData.password"
              type="password"
              placeholder="输入密码（可选）"
              show-password-on="click"
            />
          </n-form-item>
          
          <n-form-item label="数据库" path="db">
            <n-input-number v-model:value="formData.db" :min="0" :max="15" placeholder="选择数据库" />
          </n-form-item>
          
          <n-form-item>
            <div class="form-footer">
              <n-button
                type="primary"
                @click="handleConnect"
                :loading="loading"
                :disabled="loading"
              >
                连接
              </n-button>
            </div>
          </n-form-item>
        </n-form>
      </n-card>

      <!-- 历史会话列表 -->
      <n-card v-if="connectionHistory.length > 0" title="历史会话" :bordered="false" class="history-card">
        <template #header-extra>
          <n-space>
            <n-button
              circle
              @click="toggleHistory"
            >
              <template #icon>
                <n-icon>
                  <component :is="showHistory ? ChevronDownOutline : ChevronUpOutline" />
                </n-icon>
              </template>
            </n-button>
            <n-button
              circle
              @click="handleClearHistory"
              title="清除历史"
            >
            <template #icon>
                <n-icon>
                  <component :is="TrashOutline" />
                </n-icon>
              </template>
            </n-button>
          </n-space>
        </template>
        
        <n-collapse-transition :show="showHistory">
          <n-list>
            <n-list-item v-for="history in connectionHistory" :key="history.timestamp">
              <n-thing
                :title="history.sessionName"
                :description="`${history.host}:${history.port} (DB: ${history.db})`"
                @click="handleHistoryClick(history)"
              >
                <template #avatar>
                  <n-icon><ServerOutline /></n-icon>
                </template>
              </n-thing>
            </n-list-item>
          </n-list>
        </n-collapse-transition>
      </n-card>
    </div>
  </div>
</template>

<style scoped>
.connect-container {
  position: relative;
  width: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  overflow: hidden;
}

.wave-background {
  position: fixed;
  bottom: -20px;
  left: 0;
  width: 100%;
  z-index: 0;
}

.wave-svg {
  width: 100%;
  height: 100%;
  object-fit: cover;
  opacity: 0.8;
}

.connect-content {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 500px;
  margin: 0 1rem;
  overflow-y: auto;
}

/* 添加动画效果 */
.wave-svg {
  animation: waveFloat 20s ease-in-out infinite;
}

@keyframes waveFloat {
  0% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-10px);
  }
  100% {
    transform: translateY(0);
  }
}

.history-card {
  margin-top: 20px;
}

:deep(.n-list-item) {
  cursor: pointer;
  padding: 8px;
  border-radius: 4px;
  transition: background-color 0.3s;
}

:deep(.n-list-item:hover) {
  background-color: var(--n-item-color-hover);
}

:deep(.n-card-header-extra) {
  display: flex;
  align-items: center;
  gap: 8px;
}

:deep(.n-form-item:last-child) {
  margin-bottom: 0;
}

:deep(.n-form-item:last-child .n-form-item-feedback-wrapper) {
  display: none;
}

:deep(.n-form-item:last-child .n-form-item-blank) {
  display: flex;
  justify-content: flex-end;
  margin-left: 100px;
}

/* 添加全局样式防止页面滚动 */
:global(body) {
  margin: 0;
  padding: 0;
  overflow: hidden;
}
</style> 