import iziToast from 'izitoast'

export const toast = {
  success(message: string, title = '') {
    iziToast.success({
      title,
      message,
      position: 'topRight',
      timeout: 3000,
      progressBar: true,
    })
  },

  error(message: string, title = 'Error') {
    iziToast.error({
      title,
      message,
      position: 'topRight',
      timeout: 5000,
      progressBar: true,
    })
  },

  info(message: string, title = '') {
    iziToast.info({
      title,
      message,
      position: 'topRight',
      timeout: 3500,
      progressBar: true,
    })
  },

  warning(message: string, title = 'Atención') {
    iziToast.warning({
      title,
      message,
      position: 'topRight',
      timeout: 4000,
      progressBar: true,
    })
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
        buttons: [
          [
            '<button style="background:#ef4444;color:#fff;border:none;padding:8px 20px;border-radius:6px;font-weight:700;cursor:pointer">Sí, eliminar</button>',
            (instance: any, toastEl: any) => {
              instance.hide({ transitionOut: 'fadeOut' }, toastEl, 'button')
              resolve(true)
            },
            true,
          ],
          [
            '<button style="background:#e5e7eb;color:#111;border:none;padding:8px 20px;border-radius:6px;font-weight:600;cursor:pointer">Cancelar</button>',
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
