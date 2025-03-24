<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { redisApi } from '../api/redis'

const router = useRouter()
const message = useMessage()
const loading = ref(false)

const formData = ref({
  host: 'localhost',
  port: 6379,
  password: '',
  db: 0
})

const handleConnect = async () => {
  loading.value = true
  try {
    await redisApi.connect(formData.value)
    message.success('连接成功')
    router.push('/browser')
  } catch (error: any) {
    message.error(error.response?.data?.error || '连接失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="connect-container">
    <n-card title="连接到 Redis" class="connect-card">
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
            block
            @click="handleConnect"
            :loading="loading"
          >
            连接
          </n-button>
        </div>
      </n-form>
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
</style> 