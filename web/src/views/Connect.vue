<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage, NCard, NSpace, NForm, NFormItem, NInput, NInputNumber, NButton, NDivider, NEmpty, NList, NListItem, NThing, NIcon, NResult } from 'naive-ui'
import { redisApi } from '../api/redis'
import { connectionHistory, type ConnectionRecord } from '../services/connectionHistory'
import { connectionState } from '../services/connectionState'
import { ServerOutline } from '@vicons/ionicons5'

const router = useRouter()
const message = useMessage()
const loading = ref(false)
const historyRecords = ref<ConnectionRecord[]>([])

const formData = ref({
  host: 'localhost',
  port: 6379,
  username: '',
  password: '',
  db: 0
})

const loadHistory = () => {
  historyRecords.value = connectionHistory.getHistory()
  console.log('加载历史记录:', historyRecords.value)
}

onMounted(() => {
  loadHistory()
})

const handleConnect = async () => {
  loading.value = true
  try {
    await redisApi.connect(formData.value)
    connectionHistory.addRecord(formData.value)
    loadHistory()
    message.success('连接成功')
    router.push('/browser')
  } catch (error: any) {
    message.error(error.response?.data?.error || '连接失败')
  } finally {
    loading.value = false
  }
}

const handleHistoryClick = (record: any) => {
  formData.value = { ...record }
  handleConnect()
}

const handleClearHistory = () => {
  connectionHistory.clearHistory()
  loadHistory()
  message.success('历史记录已清除')
}
</script>

<template>
  <div class="connect-container">
    <n-card title="连接到 Redis" class="connect-card">
      <n-space vertical>
        <n-form
          ref="formRef"
          :model="formData"
          label-placement="left"
          label-width="80"
          require-mark-placement="right-hanging"
        >
          <n-form-item label="主机" path="host">
            <n-input v-model:value="formData.host" placeholder="输入 Redis 主机地址" />
          </n-form-item>
          <n-form-item label="端口" path="port">
            <n-input-number
              v-model:value="formData.port"
              :min="1"
              :max="65535"
              placeholder="输入端口号"
            />
          </n-form-item>
          <n-form-item label="账号" path="username">
            <n-input
              v-model:value="formData.username"
              placeholder="输入账号（可选）"
            />
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
            <n-input-number
              v-model:value="formData.db"
              :min="0"
              :max="15"
              placeholder="选择数据库"
            />
          </n-form-item>
          <div style="margin-top: 24px;">
            <n-button
              type="primary"
              @click="handleConnect"
              :loading="loading"
            >
              连接
            </n-button>
          </div>
        </n-form>

        <n-divider>历史连接</n-divider>
        
        <n-empty v-if="historyRecords.length === 0" description="暂无历史记录" />
        <n-list v-else>
          <n-list-item v-for="record in historyRecords" :key="record.timestamp">
            <n-thing
              :title="`${record.host}:${record.port}`"
              :description="`数据库: ${record.db}`"
              @click="handleHistoryClick(record)"
            >
              <template #avatar>
                <n-icon size="24" color="var(--primary-color)">
                  <ServerOutline />
                </n-icon>
              </template>
            </n-thing>
          </n-list-item>
        </n-list>

        <n-button
          v-if="historyRecords.length > 0"
          type="error"
          ghost
          @click="handleClearHistory"
        >
          清除历史记录
        </n-button>
      </n-space>
    </n-card>
  </div>
</template>

<style scoped>
.connect-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: calc(100vh - 112px);
}

.connect-card {
  width: 100%;
  max-width: 480px;
}

:deep(.n-card-header) {
  text-align: center;
}

:deep(.n-card-header__main) {
  font-size: 24px;
  font-weight: 500;
}

:deep(.n-list-item) {
  cursor: pointer;
  transition: background-color 0.3s;
}

:deep(.n-list-item:hover) {
  background-color: var(--n-item-color-hover);
}
</style> 