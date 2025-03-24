import { createRouter, createWebHistory } from 'vue-router'
import Home from '../views/Home.vue'
import Connect from '../views/Connect.vue'
import Browser from '../views/Browser.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: Home
    },
    {
      path: '/connect',
      name: 'connect',
      component: Connect
    },
    {
      path: '/browser',
      name: 'browser',
      component: Browser
    }
  ]
})

export default router 