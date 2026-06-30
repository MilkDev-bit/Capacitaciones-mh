import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', redirect: '/tienda' },
    { path: '/login', component: () => import('../views/LoginView.vue') },
    { path: '/reset-password', component: () => import('../views/ResetPasswordView.vue') },
    { path: '/unirse/:codigo', component: () => import('../views/UnirseView.vue') },
    { path: '/invitacion/:id', component: () => import('../views/InvitacionVideollamada.vue') },
    { path: '/tienda', component: () => import('../views/shared/StoreView.vue') },
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
        { path: 'schedules', component: () => import('../views/admin/SchedulesView.vue') },
        { path: 'perfil/:id', component: () => import('../views/shared/PublicProfileView.vue') },
      ],
    },
    {
      path: '/instructor',
      component: () => import('../views/instructor/InstructorLayout.vue'),
      meta: { requiresAuth: true, role: 'instructor' },
      children: [
        { path: '', redirect: '/instructor/capacitaciones' },
        { path: 'capacitaciones', component: () => import('../views/instructor/CapacitacionesInstructor.vue') },
        { path: 'examenes', component: () => import('../views/instructor/ExamenesInstructor.vue') },
        { path: 'estudiantes', component: () => import('../views/instructor/EstudiantesView.vue') },
        { path: 'perfil', component: () => import('../views/instructor/InstructorPerfilView.vue') },
        { path: 'perfil/:id', component: () => import('../views/shared/PublicProfileView.vue') },
        { path: 'mensajes', component: () => import('../views/shared/MensajesView.vue') },
        { path: 'mensajes/:peer_id', component: () => import('../views/shared/MensajesView.vue') },
        { path: 'videocall/:id', component: () => import('../views/instructor/InstructorVideocallView.vue') },
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
        { path: 'perfil', component: () => import('../views/user/PerfilView.vue') },
        { path: 'perfil/:id', component: () => import('../views/shared/PublicProfileView.vue') },
        { path: 'mensajes', component: () => import('../views/shared/MensajesView.vue') },
        { path: 'mensajes/:peer_id', component: () => import('../views/shared/MensajesView.vue') },
        { path: 'videocall/:id', component: () => import('../views/user/VideocallView.vue') },
      ],
    },
  ],
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
