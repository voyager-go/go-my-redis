import { createRouter, createWebHistory } from 'vue-router'
import Home from '../views/Home.vue'
import Connect from '../views/Connect.vue'
import Browser from '../views/Browser.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'connect',
      component: Connect
    },
    {
      path: '/home',
      name: 'home',
      component: Home
    },
    {
      path: '/browser',
      name: 'browser',
      component: Browser
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: '/'
    }
  ]
})

export default router 