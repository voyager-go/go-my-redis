<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { darkTheme, lightTheme } from 'naive-ui'
import {
  LogoReddit as LogoRedis,
  SunnyOutline,
  MoonOutline,
  LogOutOutline
} from '@vicons/ionicons5'
import MessageHandler from './components/MessageHandler.vue'

const router = useRouter()
const isDark = ref(true)
const isConnected = computed(() => router.currentRoute.value.path !== '/')
const theme = computed(() => isDark.value ? darkTheme : lightTheme)
const messageHandler = ref()

const toggleTheme = () => {
  isDark.value = !isDark.value
}

const handleDisconnect = () => {
  messageHandler.value?.handleDisconnect()
}
</script>

<template>
  <n-config-provider :theme="theme">
    <n-message-provider>
      <message-handler ref="messageHandler" />
      <n-layout position="absolute">
        <n-layout-header bordered style="height: 64px; padding: 16px 24px;">
          <div class="header-content">
            <div class="logo">
              <n-icon size="24" color="var(--primary-color)">
                <LogoRedis />
              </n-icon>
              <h1>Go My Redis</h1>
            </div>
            <n-space>
              <n-button text @click="toggleTheme">
                <template #icon>
                  <n-icon>
                    <component :is="isDark ? SunnyOutline : MoonOutline" />
                  </n-icon>
                </template>
              </n-button>
              <n-button text @click="handleDisconnect" v-if="isConnected">
                <template #icon>
                  <n-icon><LogOutOutline /></n-icon>
                </template>
                断开连接
              </n-button>
            </n-space>
          </div>
        </n-layout-header>

        <n-layout-content content-style="padding: 24px;">
          <router-view v-slot="{ Component }">
            <transition name="fade" mode="out-in">
              <component :is="Component" />
            </transition>
          </router-view>
        </n-layout-content>
      </n-layout>
    </n-message-provider>
  </n-config-provider>
</template>

<style>
html, body {
  margin: 0;
  padding: 0;
  height: 100vh;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen,
    Ubuntu, Cantarell, 'Fira Sans', 'Droid Sans', 'Helvetica Neue', sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

#app {
  height: 100vh;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 100%;
}

.logo {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logo h1 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
