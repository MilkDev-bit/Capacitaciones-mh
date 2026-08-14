import { describe, it, expect, beforeEach } from 'vitest'
import { ref, nextTick, defineComponent, h } from 'vue'
import { mount } from '@vue/test-utils'
import { useScrollLock } from './useScrollLock'

/**
 * useScrollLock usa watch y onUnmounted, así que necesita vivir dentro de un
 * componente montado: fuera de una instancia, onUnmounted no se registra y la
 * prueba pasaría sin verificar la parte que más importa.
 */
function montarCon(activo: ReturnType<typeof ref<boolean>>) {
  return mount(defineComponent({
    setup() {
      useScrollLock(activo as any)
      return () => h('div')
    },
  }))
}

describe('useScrollLock', () => {
  beforeEach(() => {
    document.body.style.overflow = ''
  })

  it('bloquea y libera al cambiar la bandera', async () => {
    const abierto = ref(false)
    montarCon(abierto)
    expect(document.body.style.overflow).toBe('')

    abierto.value = true
    await nextTick()
    expect(document.body.style.overflow).toBe('hidden')

    abierto.value = false
    await nextTick()
    expect(document.body.style.overflow).toBe('')
  })

  it('bloquea de inmediato si ya estaba abierto al montar', async () => {
    const abierto = ref(true)
    montarCon(abierto)
    // immediate: true — si no, abrir el cajón y navegar dejaría la página
    // desplazable por detrás hasta el siguiente cambio de la bandera.
    expect(document.body.style.overflow).toBe('hidden')
  })

  it('restaura el valor previo en vez de vaciarlo', async () => {
    // Otra capa ya había bloqueado el scroll antes.
    document.body.style.overflow = 'hidden'

    const abierto = ref(false)
    montarCon(abierto)
    abierto.value = true
    await nextTick()
    abierto.value = false
    await nextTick()

    // Cerrar ESTA capa no debe desbloquear la de la otra: si se asignara
    // cadena vacía, el modal que seguía abierto se quedaría con la página
    // desplazándose por detrás.
    expect(document.body.style.overflow).toBe('hidden')
  })

  it('libera al desmontar con la capa abierta', async () => {
    const abierto = ref(false)
    const wrapper = montarCon(abierto)

    abierto.value = true
    await nextTick()
    expect(document.body.style.overflow).toBe('hidden')

    // Caso real: el usuario navega tocando un enlace del propio cajón, así que
    // el layout se desmonta sin que nadie cierre la bandera. Sin onUnmounted,
    // la página siguiente quedaría sin scroll y sin forma de recuperarlo.
    wrapper.unmount()
    expect(document.body.style.overflow).toBe('')
  })

  it('no vuelve a guardar el estado si ya estaba bloqueado por esta capa', async () => {
    const abierto = ref(false)
    montarCon(abierto)

    abierto.value = true
    await nextTick()
    // Un segundo disparo de la misma capa no debe capturar "hidden" como
    // valor previo: al liberar dejaría el scroll bloqueado para siempre.
    abierto.value = false
    await nextTick()
    abierto.value = true
    await nextTick()
    abierto.value = false
    await nextTick()

    expect(document.body.style.overflow).toBe('')
  })
})
