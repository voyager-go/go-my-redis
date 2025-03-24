<template>
  <n-card title="连接配置" :bordered="false" size="large" class="connect-card">
    <n-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-placement="left"
      label-width="80"
      require-mark-placement="right-hanging"
      size="large"
    >
      <n-grid :cols="24" :x-gap="24">
        <n-grid-item :span="16">
          <n-form-item label="主机" path="host">
            <n-input v-model:value="form.host" placeholder="localhost">
              <template #prefix>
                <n-icon><ServerOutline /></n-icon>
              </template>
            </n-input>
          </n-form-item>
        </n-grid-item>
        <n-grid-item :span="8">
          <n-form-item label="端口" path="port">
            <n-input-number
              v-model:value="form.port"
              :min="1"
              :max="65535"
              class="full-width"
            />
          </n-form-item>
        </n-grid-item>
      </n-grid>

      <n-form-item label="密码" path="password">
        <n-input
          v-model:value="form.password"
          type="password"
          show-password-on="click"
          placeholder="可选"
        >
          <template #prefix>
            <n-icon><LockClosedOutline /></n-icon>
          </template>
        </n-input>
      </n-form-item>

      <n-form-item label="数据库" path="db">
        <n-input-number
          v-model:value="form.db"
          :min="0"
          :max="15"
          class="full-width"
        />
      </n-form-item>

      <div class="form-actions">
        <n-space justify="end">
          <n-button
            round
            type="primary"
            size="large"
            :loading="loading"
            @click="handleConnect"
          >
            <template #icon>
              <n-icon><LogInOutline /></n-icon>
            </template>
            连接到服务器
          </n-button>
        </n-space>
      </div>
    </n-form>
  </n-card>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import { ServerOutline, LockClosedOutline, LogInOutline } from '@vicons/ionicons5'
import type { FormInst, FormRules } from 'naive-ui'
import { redisApi } from '../api/redis'
import type { RedisConfig } from '../types/redis'

const router = useRouter()
const message = useMessage()
const formRef = ref<FormInst | null>(null)
const loading = ref(false)

const form = reactive<RedisConfig>({
  host: 'localhost',
  port: 6379,
  password: '',
  db: 0
})

const rules: FormRules = {
  host: [
    { required: true, message: '请输入主机地址', trigger: 'blur' }
  ],
  port: [
    { required: true, message: '请输入端口号', trigger: 'blur' },
    { type: 'number', min: 1, max: 65535, message: '端口号范围：1-65535', trigger: 'blur' }
  ],
  db: [
    { type: 'number', min: 0, max: 15, message: '数据库范围：0-15', trigger: 'blur' }
  ]
}

const handleConnect = () => {
  if (!formRef.value) return

  formRef.value.validate(async (errors) => {
    if (!errors) {
      loading.value = true
      try {
        await redisApi.connect(form)
        message.success('连接成功')
        router.push('/browser')
      } catch (error: any) {
        message.error(error.response?.data?.error || '连接失败')
      } finally {
        loading.value = false
      }
    }
  })
}
</script>

<style scoped>
.connect-card {
  max-width: 600px;
  margin: 0 auto;
}

.full-width {
  width: 100%;
}

.form-actions {
  margin-top: 24px;
}

:deep(.n-card-header) {
  text-align: center;
}

:deep(.n-card__content) {
  padding: 24px !important;
}
</style> 