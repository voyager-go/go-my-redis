import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import router from './router'
import {
  create,
  NConfigProvider,
  NMessageProvider,
  NLayout,
  NLayoutHeader,
  NLayoutContent,
  NLayoutSider,
  NButton,
  NInput,
  NInputNumber,
  NForm,
  NFormItem,
  NCard,
  NSpace,
  NIcon,
  NGrid,
  NGridItem,
  NTabs,
  NTabPane,
  NDataTable,
  NTree,
  NTag,
  NEmpty,
  NPopconfirm,
  NDivider,
  NList,
  NListItem,
  NThing
} from 'naive-ui'

const naive = create({
  components: [
    NConfigProvider,
    NMessageProvider,
    NLayout,
    NLayoutHeader,
    NLayoutContent,
    NLayoutSider,
    NButton,
    NInput,
    NInputNumber,
    NForm,
    NFormItem,
    NCard,
    NSpace,
    NIcon,
    NGrid,
    NGridItem,
    NTabs,
    NTabPane,
    NDataTable,
    NTree,
    NTag,
    NEmpty,
    NPopconfirm,
    NDivider,
    NList,
    NListItem,
    NThing
  ]
})

const app = createApp(App)
app.use(router)
app.use(naive)
app.mount('#app')
