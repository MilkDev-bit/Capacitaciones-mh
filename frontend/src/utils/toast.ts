import iziToast from 'izitoast'

iziToast.settings({
  transitionIn: 'fadeInDown',
  transitionOut: 'fadeOutUp',
  progressBarColor: 'rgba(255,255,255,0.4)',
  layout: 1,
  balloon: false,
})

export const toast = {
  success(message: string, title = 'Listo') {
    iziToast.success({
      title,
      message,
      position: 'topRight',
      timeout: 3000,
      progressBar: true,
      iconColor: '#fff',
      titleColor: '#fff',
      messageColor: 'rgba(255,255,255,0.9)',
      backgroundColor: '#22c55e',
      progressBarColor: 'rgba(255,255,255,0.35)',
    })
  },

  error(message: string, title = 'Error') {
    iziToast.error({
      title,
      message,
      position: 'topRight',
      timeout: 5000,
      progressBar: true,
      iconColor: '#fff',
      titleColor: '#fff',
      messageColor: 'rgba(255,255,255,0.9)',
      backgroundColor: '#ef4444',
      progressBarColor: 'rgba(255,255,255,0.35)',
    })
  },

  info(message: string, title = 'Info') {
    iziToast.info({
      title,
      message,
      position: 'topRight',
      timeout: 3500,
      progressBar: true,
      iconColor: '#fff',
      titleColor: '#fff',
      messageColor: 'rgba(255,255,255,0.9)',
      backgroundColor: '#f97316',
      progressBarColor: 'rgba(255,255,255,0.35)',
    })
  },

  warning(message: string, title = 'Atención') {
    iziToast.warning({
      title,
      message,
      position: 'topRight',
      timeout: 4000,
      progressBar: true,
      iconColor: '#fff',
      titleColor: '#fff',
      messageColor: 'rgba(255,255,255,0.9)',
      backgroundColor: '#f59e0b',
      progressBarColor: 'rgba(255,255,255,0.35)',
    })
  },

  loading(message: string, title = 'Subiendo...'): { close: () => void } {
    let toastEl: HTMLDivElement | null = null
    iziToast.show({
      title,
      message,
      position: 'topRight',
      timeout: false as any,
      close: false,
      progressBar: false,
      backgroundColor: '#1e293b',
      titleColor: '#fff',
      messageColor: 'rgba(255,255,255,0.75)',
      onOpened(_instance: any, toast: HTMLDivElement) {
        toastEl = toast
      },
    })
    return {
      close() {
        if (toastEl) iziToast.hide({ transitionOut: 'fadeOutUp' }, toastEl, 'custom')
      },
    }
  },

  confirm(message: string, title = '¿Estás seguro?'): Promise<boolean> {
    return new Promise((resolve) => {
      iziToast.question({
        timeout: false as any,
        close: false,
        overlay: true,
        zindex: 99999,
        title,
        message,
        position: 'center',
        backgroundColor: '#ffffff',
        titleColor: '#1d1d1f',
        messageColor: '#86868b',
        buttons: [
          [
            '<button style="background:#f97316;color:#fff;border:none;padding:9px 22px;border-radius:9999px;font-weight:700;cursor:pointer;font-size:0.9rem;box-shadow:0 4px 12px rgba(249,115,22,.3)">Confirmar</button>',
            (instance: any, toastEl: any) => {
              instance.hide({ transitionOut: 'fadeOut' }, toastEl, 'button')
              resolve(true)
            },
            true,
          ],
          [
            '<button style="background:rgba(0,0,0,0.06);color:#1d1d1f;border:none;padding:9px 22px;border-radius:9999px;font-weight:600;cursor:pointer;font-size:0.9rem">Cancelar</button>',
            (instance: any, toastEl: any) => {
              instance.hide({ transitionOut: 'fadeOut' }, toastEl, 'button')
              resolve(false)
            },
            false,
          ],
        ],
      })
    })
  },
}
