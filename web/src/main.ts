import { createApp } from 'vue'
import { createRouter, createWebHashHistory } from 'vue-router'
import App from './App.vue'
import Rules from './views/Rules.vue'
import Logs from './views/Logs.vue'

const routes = [
  { path: '/', component: Rules },
  { path: '/logs', component: Logs },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

createApp(App).use(router).mount('#app')
