<template>
  <div class="home-container">
    <n-card title="Go My Redis" class="welcome-card">
      <template #header-extra>
        <n-space>
          <n-button type="primary" @click="$router.push('/connect')">
            <n-icon>
              <component :is="RefreshOutline" />
            </n-icon>
          </n-button>
          <n-button v-if="historyCount > 0" type="info" @click="$router.push('/connect')">
            <n-icon>
              <component :is="TimerOutline" />
            </n-icon>
            <span style="margin-left: 4px">{{ historyCount }}</span>
          </n-button>
        </n-space>
      </template>
      <n-space vertical>
        <n-alert type="info">
          欢迎使用 Go My Redis，这是一个简单的 Redis 管理工具。
          您可以通过此工具连接到 Redis 服务器，并进行数据浏览和管理操作。
        </n-alert>
        <n-list>
          <n-list-item>
            <n-thing title="数据浏览" description="支持查看和编辑各种数据类型（String、List、Hash、Set）">
              <template #avatar>
                <n-icon size="24" color="var(--primary-color)">
                  <ServerOutline />
                </n-icon>
              </template>
            </n-thing>
          </n-list-item>
          <n-list-item>
            <n-thing title="键管理" description="支持键的搜索、删除、重命名等操作">
              <template #avatar>
                <n-icon size="24" color="var(--primary-color)">
                  <KeyOutline />
                </n-icon>
              </template>
            </n-thing>
          </n-list-item>
          <n-list-item>
            <n-thing title="TTL 设置" description="支持设置键的过期时间">
              <template #avatar>
                <n-icon size="24" color="var(--primary-color)">
                  <TimeOutline />
                </n-icon>
              </template>
            </n-thing>
          </n-list-item>
        </n-list>
      </n-space>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ServerOutline, KeyOutline, TimeOutline, RefreshOutline, TimerOutline } from '@vicons/ionicons5'
import { ref } from 'vue'
import { connectionHistory } from '../services/connectionHistory'

const historyCount = ref(connectionHistory.getHistory().length)
</script>

<style scoped>
.home-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: calc(100vh - 112px);
}

.welcome-card {
  width: 100%;
  max-width: 800px;
}

:deep(.n-card-header) {
  text-align: center;
}

:deep(.n-card-header__main) {
  font-size: 24px;
  font-weight: 500;
}

:deep(.n-list-item:not(:last-child)) {
  border-bottom: 1px solid var(--n-border-color);
}
</style> 