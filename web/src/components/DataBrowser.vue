<template>
  <div class="browser-container">
    <n-layout has-sider>
      <!-- 左侧面板 -->
      <n-layout-sider
        bordered
        collapse-mode="width"
        :collapsed-width="0"
        :width="320"
        show-trigger="bar"
        content-style="padding: 16px;"
      >
        <div class="sider-content">
          <n-input
            v-model:value="searchPattern"
            placeholder="搜索键..."
            clearable
            @clear="handleSearch"
            @keyup.enter="handleSearch"
          >
            <template #prefix>
              <n-icon><SearchOutline /></n-icon>
            </template>
          </n-input>

          <div class="key-list">
            <n-tree
              block-line
              :data="keyTree"
              :pattern="searchPattern"
              :render-label="renderKeyLabel"
              @update:selected-keys="handleKeySelect"
            />
          </div>
        </div>
      </n-layout-sider>

      <!-- 右侧内容 -->
      <n-layout content-style="padding: 16px;">
        <n-tabs
          type="segment"
          animated
          style="margin-bottom: 16px;"
          v-if="selectedKey"
        >
          <n-tab-pane name="value" tab="数据">
            <n-card :bordered="false">
              <template #header>
                <n-space align="center" justify="space-between">
                  <n-space>
                    <span class="key-name">{{ selectedKey }}</span>
                    <n-tag :type="getTagType(keyType)" size="small">
                      {{ keyType }}
                    </n-tag>
                  </n-space>
                  <n-space>
                    <n-popconfirm
                      @positive-click="handleDelete"
                      positive-text="删除"
                      negative-text="取消"
                    >
                      <template #trigger>
                        <n-button
                          quaternary
                          circle
                          type="error"
                          :disabled="!selectedKey"
                        >
                          <template #icon>
                            <n-icon><TrashOutline /></n-icon>
                          </template>
                        </n-button>
                      </template>
                      确定要删除该键吗？
                    </n-popconfirm>
                    <n-button
                      quaternary
                      circle
                      type="primary"
                      :disabled="!selectedKey"
                      @click="handleSave"
                    >
                      <template #icon>
                        <n-icon><SaveOutline /></n-icon>
                      </template>
                    </n-button>
                  </n-space>
                </n-space>
              </template>

              <n-space vertical size="large">
                <div class="key-info">
                  <n-space vertical :size="16">
                    <n-space justify="space-between">
                      <n-text>键名：{{ selectedKey }}</n-text>
                      <n-text>类型：{{ keyType }}</n-text>
                    </n-space>
                    <n-descriptions bordered :column="1" size="small">
                      <n-descriptions-item label="类型">
                        {{ keyType }}
                      </n-descriptions-item>
                      <n-descriptions-item label="TTL">
                        {{ ttl === -1 ? '永久' : `${ttl}秒` }}
                      </n-descriptions-item>
                      <n-descriptions-item v-if="keyType === 'list'" label="长度">
                        {{ Array.isArray(detailValue) ? detailValue.length : 0 }}
                      </n-descriptions-item>
                      <n-descriptions-item v-if="keyType === 'hash'" label="字段数">
                        {{ typeof detailValue === 'object' ? Object.keys(detailValue).length : 0 }}
                      </n-descriptions-item>
                    </n-descriptions>
                    <n-input-number
                      v-model:value="ttl"
                      :min="-1"
                      :max="2147483647"
                      placeholder="TTL (秒)"
                      class="full-width"
                    >
                      <template #prefix>TTL:</template>
                    </n-input-number>
                  </n-space>
                </div>
                <div class="value-editor">
                  <n-input
                    v-model:value="keyValue"
                    type="textarea"
                    placeholder="值"
                    :autosize="{ minRows: 3, maxRows: 15 }"
                  />
                </div>
              </n-space>
            </n-card>
          </n-tab-pane>

          <n-tab-pane name="terminal" tab="终端">
            <redis-terminal />
          </n-tab-pane>
        </n-tabs>

        <n-empty
          v-else
          description="请选择一个键"
          size="large"
          style="margin-top: 64px;"
        >
          <template #icon>
            <n-icon size="64" depth="3">
              <KeyOutline />
            </n-icon>
          </template>
        </n-empty>
      </n-layout>
    </n-layout>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, h, onMounted } from 'vue'
import { useMessage, NTag, NDescriptions, NDescriptionsItem } from 'naive-ui'
import {
  SearchOutline,
  TrashOutline,
  SaveOutline,
  KeyOutline
} from '@vicons/ionicons5'
import type { TreeOption, DataTableColumns } from 'naive-ui'
import { redisApi } from '../api/redis'
import type { RedisKey, KeyDetail } from '../types/redis'
import RedisTerminal from './RedisTerminal.vue'

const message = useMessage()
const searchPattern = ref('*')
const keyTree = ref<TreeOption[]>([])
const selectedKey = ref<string>('')
const keyType = ref<string>('')
const ttl = ref<number>(0)
const detailValue = ref<any>(null)
const keyValue = ref<string>('')

// 表格列定义
const listColumns: DataTableColumns = [
  { title: '索引', key: 'index', width: 80 },
  { title: '值', key: 'value', ellipsis: true }
]

const hashColumns: DataTableColumns = [
  { title: '字段', key: 'field', width: 200 },
  { title: '值', key: 'value', ellipsis: true }
]

const setColumns: DataTableColumns = [
  { title: '值', key: 'value', ellipsis: true }
]

// 表格数据计算属性
const listTableData = computed(() => {
  if (detailValue.value.type !== 'list' || !Array.isArray(detailValue.value.value)) return []
  return detailValue.value.value.map((value: any, index: number) => ({
    index,
    value
  }))
})

const hashTableData = computed(() => {
  if (detailValue.value.type !== 'hash' || typeof detailValue.value.value !== 'object') return []
  return Object.entries(detailValue.value.value as Record<string, any>).map(([field, value]) => ({
    field,
    value
  }))
})

const setTableData = computed(() => {
  if (detailValue.value.type !== 'set' || !Array.isArray(detailValue.value.value)) return []
  return detailValue.value.value.map((value: any) => ({ value }))
})

// 获取标签类型
const getTagType = (type: string) => {
  const types: Record<string, string> = {
    'string': 'default',
    'list': 'success',
    'hash': 'warning',
    'set': 'info',
    'zset': 'error'
  }
  return types[type] || 'default'
}

// 渲染键标签
const renderKeyLabel = (info: { option: TreeOption }) => {
  const { option } = info
  return h(
    'div',
    {
      class: 'key-label'
    },
    [
      h('span', { class: 'key-text' }, option.label as string),
      option.type && h(
        NTag,
        {
          size: 'tiny',
          type: getTagType(option.type as string),
          style: { marginLeft: '8px' }
        },
        { default: () => option.type }
      )
    ]
  )
}

// 处理搜索
const handleSearch = async () => {
  try {
    const { data: keys } = await redisApi.getKeys(searchPattern.value)
    keyTree.value = keys.map(key => ({
      key,
      label: key,
      type: null
    }))
  } catch (error: any) {
    message.error(error.response?.data?.error || '获取键列表失败')
  }
}

// 处理键选择
const handleKeySelect = async (keys: string[]) => {
  const key = keys[0]
  if (!key) {
    selectedKey.value = ''
    return
  }

  try {
    selectedKey.value = key
    const type = await redisApi.type(key)
    keyType.value = type
    const response = await redisApi.getKey(key)
    
    // 保存详细信息用于展示
    detailValue.value = response.value
    
    // 设置可编辑的值
    if (type === 'string') {
      keyValue.value = response.value
    } else if (type === 'list' || type === 'set' || type === 'zset') {
      keyValue.value = JSON.stringify(response.value, null, 2)
    } else if (type === 'hash') {
      keyValue.value = JSON.stringify(response.value, null, 2)
    }
    
    const keyTtl = await redisApi.ttl(key)
    ttl.value = keyTtl
  } catch (error: any) {
    message.error(error.response?.data?.error || '获取键详情失败')
  }
}

// 处理保存
const handleSave = async () => {
  if (!selectedKey.value) return

  try {
    await redisApi.setKey(selectedKey.value, {
      ...detailValue.value,
      key: selectedKey.value
    })
    message.success('保存成功')
  } catch (error: any) {
    message.error(error.response?.data?.error || '保存失败')
  }
}

// 处理删除
const handleDelete = async () => {
  if (!selectedKey.value) return

  try {
    await redisApi.deleteKey(selectedKey.value)
    message.success('删除成功')
    selectedKey.value = ''
    await handleSearch()
  } catch (error: any) {
    message.error(error.response?.data?.error || '删除失败')
  }
}

// 初始加载
onMounted(() => {
  handleSearch()
})
</script>

<style scoped>
.browser-container {
  height: 100%;
}

.sider-content {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.key-list {
  flex: 1;
  overflow: auto;
}

.key-label {
  display: flex;
  align-items: center;
  gap: 8px;
}

.key-text {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.key-name {
  font-size: 16px;
  font-weight: 500;
}

.key-info {
  padding: 16px;
  border: 1px solid var(--n-border-color);
  border-radius: 3px;
  background-color: var(--n-card-color);
}

.full-width {
  width: 100%;
}

:deep(.n-descriptions-table-wrapper) {
  margin: 8px 0;
}

:deep(.n-input-number) {
  width: 100%;
}
</style> 