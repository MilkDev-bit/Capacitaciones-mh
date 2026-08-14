import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    // La raíz es la tienda: un visitante nuevo aterriza en el catálogo, no en
    // un muro de login. La sesión se pide solo al pagar o entrar al panel.
    { path: '/', component: () => import('../views/shared/StoreView.vue') },
    { path: '/login', component: () => import('../views/LoginView.vue') },
    { path: '/reset-password', component: () => import('../views/ResetPasswordView.vue') },
    { path: '/verificar-correo', component: () => import('../views/VerifyEmailView.vue') },
    // Pantalla de retorno de Stripe. Pública a propósito: si la cookie tarda en
    // propagarse, el usuario ve la confirmación en vez de un rebote al login.
    { path: '/checkout/exito', component: () => import('../views/CheckoutSuccessView.vue') },
    { path: '/unirse/:codigo', component: () => import('../views/UnirseView.vue') },
    { path: '/tienda', component: () => import('../views/shared/StoreView.vue') },
    // Pública: el precio es parte del argumento de venta, no se esconde tras login.
    { path: '/planes', component: () => import('../views/shared/PlanesView.vue') },
    // Escena guiada por scroll. Fuera de la tienda a propósito: son ~6 pantallas
    // de recorrido y meterlas en medio del catálogo alargaría el camino a la
    // compra para todo el mundo, incluido quien ya sabe lo que quiere.
    { path: '/como-funciona', component: () => import('../views/shared/ComoFuncionaView.vue') },
    { path: '/curso/:id', component: () => import('../views/CursoPublicView.vue') },
    { path: '/examen/:id', component: () => import('../views/user/ExamenFormView.vue'), meta: { requiresAuth: true } },
    {
      path: '/admin',
      component: () => import('../views/admin/AdminLayout.vue'),
      meta: { requiresAuth: true, role: 'admin' },
      children: [
        { path: '', redirect: '/admin/dashboard' },
        { path: 'dashboard', component: () => import('../views/admin/AdminDashboard.vue') },
        { path: 'capacitaciones', component: () => import('../views/admin/CapacitacionesView.vue') },
        { path: 'examenes', component: () => import('../views/admin/ExamenesView.vue') },
        { path: 'usuarios', component: () => import('../views/admin/UsuariosView.vue') },
        { path: 'perfil/:id', component: () => import('../views/shared/PublicProfileView.vue') },
      ],
    },
    {
      path: '/instructor',
      component: () => import('../views/instructor/InstructorLayout.vue'),
      meta: { requiresAuth: true, role: 'instructor' },
      children: [
        { path: '', redirect: '/instructor/dashboard' },
        { path: 'dashboard', component: () => import('../views/instructor/InstructorDashboard.vue') },
        { path: 'capacitaciones', component: () => import('../views/instructor/CapacitacionesInstructor.vue') },
        { path: 'examenes', component: () => import('../views/instructor/ExamenesInstructor.vue') },
        { path: 'estudiantes', component: () => import('../views/instructor/EstudiantesView.vue') },
        { path: 'entregas', component: () => import('../views/instructor/EntregasView.vue') },
        { path: 'perfil', component: () => import('../views/instructor/InstructorPerfilView.vue') },
        { path: 'perfil/:id', component: () => import('../views/shared/PublicProfileView.vue') },
        { path: 'mensajes', component: () => import('../views/shared/MensajesView.vue') },
        { path: 'mensajes/:peer_id', component: () => import('../views/shared/MensajesView.vue') },
      ],
    },
    {
      path: '/usuario',
      component: () => import('../views/user/UserLayout.vue'),
      meta: { requiresAuth: true, role: 'user' },
      children: [
        { path: '', redirect: '/usuario/dashboard' },
        { path: 'dashboard', component: () => import('../views/user/DashboardView.vue') },
        { path: 'capacitaciones', component: () => import('../views/user/MisCapacitaciones.vue') },
        { path: 'examenes', component: () => import('../views/user/MisExamenes.vue') },
        { path: 'examenes/:id', component: () => import('../views/user/ResponderExamen.vue') },
        { path: 'capacitaciones/:id', component: () => import('../views/user/VerCapacitacion.vue') },
        { path: 'licencias', component: () => import('../views/user/MisLicencias.vue') },
        { path: 'constancias', component: () => import('../views/user/MisConstancias.vue') },
        { path: 'perfil', component: () => import('../views/user/PerfilView.vue') },
        { path: 'perfil/:id', component: () => import('../views/shared/PublicProfileView.vue') },
        { path: 'mensajes', component: () => import('../views/shared/MensajesView.vue') },
        { path: 'mensajes/:peer_id', component: () => import('../views/shared/MensajesView.vue') },
      ],
    },
  ],

  /**
   * Al cambiar de ruta se vuelve arriba; con el botón atrás se restaura la
   * posición guardada.
   *
   * Sin esto, entrar a /como-funciona desde media tienda arrancaría la escena
   * a mitad del recorrido —el cubo ya girado, el cielo ya de noche— porque toda
   * ella se deriva de scrollY. El anclaje por hash se respeta igual.
   */
  scrollBehavior(to, _from, savedPosition) {
    if (savedPosition) return savedPosition
    if (to.hash) return { el: to.hash, behavior: 'smooth' }
    return { top: 0 }
  },
})

router.beforeEach((to, _from, next) => {
  const auth = useAuthStore()
  const user = auth.user

  if (to.meta.requiresAuth && !auth.isLoggedIn) {
    auth.handleSessionExpired()
    next('/login')
    return
  }
  if (to.meta.role === 'admin' && user?.role !== 'admin') {
    next(user?.role === 'instructor' ? '/instructor' : '/usuario')
    return
  }
  if (to.meta.role === 'instructor' && user?.role !== 'instructor' && user?.role !== 'admin') {
    next('/usuario')
    return
  }
  next()
})

// Manejo automático de ChunkLoadErrors tras un despliegue
router.onError((error, to) => {
  if (error.message.includes('Failed to fetch dynamically imported module') || error.name === 'ChunkLoadError') {
    // Evitar bucle infinito recargando
    const reloadKey = `chunk_reload_${to.fullPath}`
    if (!sessionStorage.getItem(reloadKey)) {
      sessionStorage.setItem(reloadKey, 'true')
      window.location.href = to.fullPath
    }
  }
})

export default router
