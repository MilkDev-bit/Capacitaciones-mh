import './assets/main.css'
import 'izitoast/dist/css/iziToast.min.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import * as Sentry from '@sentry/vue'

import App from './App.vue'
import router from './router'
import { vReveal } from './composables/useReveal'

import { VueDatePicker } from '@vuepic/vue-datepicker'
import '@vuepic/vue-datepicker/dist/main.css'

const app = createApp(App)

app.component('VueDatePicker', VueDatePicker)
// Directiva global: la usan la tienda, la ficha de curso y la página de planes.
app.directive('reveal', vReveal)

app.use(createPinia())
app.use(router)

const sentryDsn = import.meta.env.VITE_SENTRY_DSN as string | undefined
if (sentryDsn) {
  Sentry.init({
    app,
    dsn: sentryDsn,
    integrations: [
      Sentry.browserTracingIntegration({ router }),
      Sentry.replayIntegration(),
    ],
    sendDefaultPii: true,
    tracePropagationTargets: [/^\/api/, window.location.origin],
    // Captura el 10% de trazas de rendimiento en producción
    tracesSampleRate: import.meta.env.PROD ? 0.1 : 1.0,
    // Captura el 10% de replays de sesión; 100% al haber un error
    replaysSessionSampleRate: 0.1,
    replaysOnErrorSampleRate: 1.0,
    environment: import.meta.env.MODE,
  })
}

app.mount('#app')
