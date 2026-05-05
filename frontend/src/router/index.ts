import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', redirect: '/login' },
    { path: '/login', component: () => import('../views/LoginView.vue') },
    { path: '/unirse/:codigo', component: () => import('../views/UnirseView.vue') },
    {
      path: '/admin',
      component: () => import('../views/admin/AdminLayout.vue'),
      meta: { requiresAuth: true, role: 'admin' },
      children: [
        { path: '', redirect: '/admin/capacitaciones' },
        { path: 'capacitaciones', component: () => import('../views/admin/CapacitacionesView.vue') },
        { path: 'examenes', component: () => import('../views/admin/ExamenesView.vue') },
        { path: 'usuarios', component: () => import('../views/admin/UsuariosView.vue') },
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
      ],
    },
    {
      path: '/usuario',
      component: () => import('../views/user/UserLayout.vue'),
      meta: { requiresAuth: true, role: 'user' },
      children: [
        { path: '', redirect: '/usuario/capacitaciones' },
        { path: 'capacitaciones', component: () => import('../views/user/MisCapacitaciones.vue') },
        { path: 'examenes', component: () => import('../views/user/MisExamenes.vue') },
        { path: 'examenes/:id', component: () => import('../views/user/ResponderExamen.vue') },
        { path: 'capacitaciones/:id', component: () => import('../views/user/VerCapacitacion.vue') },
        { path: 'perfil', component: () => import('../views/user/PerfilView.vue') },
      ],
    },
  ],
})

router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('token')
  const user = JSON.parse(localStorage.getItem('user') || 'null')

  if (to.meta.requiresAuth && !token) {
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

export default router
