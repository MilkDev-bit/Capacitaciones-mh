import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '../api'

export interface Plan {
  id: string
  codigo: string
  nombre: string
  descripcion?: string
  /** 'individual' = una persona · 'asientos' = N lugares para un equipo */
  modalidad: string
  /** 'mes' | 'anio' */
  intervalo: string
  precio_centavos: number
  moneda: string
  dias_prueba?: number
  activo: boolean
}

export interface MiSuscripcion {
  id: string
  plan_id?: string
  plan_nombre?: string
  modalidad?: string
  intervalo?: string
  precio_centavos?: number
  estado?: string
  asientos?: number
  asientos_ocupados?: number
  periodo_inicio?: string
  periodo_fin?: string
  prueba_fin?: string
  cancelar_al_terminar?: boolean
  /** Lo calcula el backend: incluye el periodo de gracia de un cobro fallido. */
  acceso_vigente?: boolean
}

/**
 * Estado de la membresía del usuario.
 *
 * `accesoVigente` es la única bandera que la UI debe consultar para decidir si
 * un curso de pago se muestra como "incluido en tu plan". Es un espejo del
 * backend, nunca la autoridad: el servidor vuelve a validarlo al inscribir.
 */
export const useSuscripcionStore = defineStore('suscripcion', () => {
  const planes = ref<Plan[]>([])
  const mia = ref<MiSuscripcion | null>(null)
  const cargandoPlanes = ref(false)
  const cargandoMia = ref(false)
  /** Evita repetir la consulta en cada vista que monta el store. */
  const miaConsultada = ref(false)

  const accesoVigente = computed(() => !!mia.value?.acceso_vigente)
  const tieneSuscripcion = computed(() => !!mia.value?.id)
  const esCorporativa = computed(() => mia.value?.modalidad === 'asientos')

  /** Solo un cobro fallido en dunning: el acceso sigue vivo pero hay que avisar. */
  const enPeriodoDeGracia = computed(() => mia.value?.estado === 'vencida')

  const planesIndividuales = computed(() => planes.value.filter((p) => p.modalidad === 'individual'))
  const planesEquipo = computed(() => planes.value.filter((p) => p.modalidad === 'asientos'))

  async function cargarPlanes(forzar = false) {
    if (planes.value.length && !forzar) return
    cargandoPlanes.value = true
    try {
      const res = await api.get('/planes')
      planes.value = res.data || []
    } catch {
      planes.value = []
    } finally {
      cargandoPlanes.value = false
    }
  }

  async function cargarMia(forzar = false) {
    if (miaConsultada.value && !forzar) return
    cargandoMia.value = true
    try {
      const res = await api.get('/mi-suscripcion')
      // Un id vacío significa "no tiene", no es un error.
      mia.value = res.data?.id ? res.data : null
    } catch {
      // 401 en una vista pública es lo esperado; no se molesta al usuario.
      mia.value = null
    } finally {
      miaConsultada.value = true
      cargandoMia.value = false
    }
  }

  function limpiar() {
    mia.value = null
    miaConsultada.value = false
  }

  async function checkout(planCodigo: string, asientos = 1) {
    const res = await api.post('/suscripcion/checkout', {
      plan_codigo: planCodigo,
      asientos,
    })
    if (!res.data?.url) throw new Error('No se recibió la URL de pago')
    window.location.href = res.data.url
  }

  async function abrirPortal() {
    const res = await api.post('/suscripcion/portal')
    if (!res.data?.url) throw new Error('No se recibió la URL del portal')
    window.location.href = res.data.url
  }

  return {
    planes,
    mia,
    cargandoPlanes,
    cargandoMia,
    accesoVigente,
    tieneSuscripcion,
    esCorporativa,
    enPeriodoDeGracia,
    planesIndividuales,
    planesEquipo,
    cargarPlanes,
    cargarMia,
    limpiar,
    checkout,
    abrirPortal,
  }
})

/** Formatea centavos como moneda, sin decimales cuando el importe es redondo. */
export function precioDesdeCentavos(centavos: number, moneda = 'MXN') {
  return new Intl.NumberFormat('es-MX', {
    style: 'currency',
    currency: moneda || 'MXN',
    maximumFractionDigits: centavos % 100 === 0 ? 0 : 2,
  }).format((centavos || 0) / 100)
}
