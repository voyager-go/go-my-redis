<script setup lang="ts">
import { useMessage } from 'naive-ui'
import { redisApi } from '../api/redis'
import { useRouter } from 'vue-router'

const router = useRouter()
const message = useMessage()

const handleDisconnect = async () => {
  try {
    await redisApi.disconnect()
    message.success('已断开连接')
    router.push('/')
  } catch (error: any) {
    message.error(error.response?.data?.error || '断开连接失败')
  }
}

defineExpose({
  handleDisconnect
})
</script>

<template>
  <div></div>
</template> 