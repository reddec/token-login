import './style.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import { client } from './api/client.gen'

client.setConfig({ throwOnError: true })

const app = createApp(App)

app.use(createPinia())
app.use(router)

app.mount('#app')
