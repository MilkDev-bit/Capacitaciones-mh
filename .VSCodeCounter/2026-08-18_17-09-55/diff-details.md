# Diff Details

Date : 2026-08-18 17:09:55

Directory /home/MictDev/Projects/Capacitaciones-mh

Total : 103 files,  7316 codes, 1819 comments, 1040 blanks, all 10175 lines

[Summary](results.md) / [Details](details.md) / [Diff Summary](diff.md) / Diff Details

## Files
| filename | language | code | comment | blank | total |
| :--- | :--- | ---: | ---: | ---: | ---: |
| [Makefile](/Makefile) | Makefile | 9 | 2 | 1 | 12 |
| [docker-compose.yml](/docker-compose.yml) | YAML | 102 | 64 | 7 | 173 |
| [frontend/RESPONSIVE.md](/frontend/RESPONSIVE.md) | Markdown | 151 | 0 | 27 | 178 |
| [frontend/src/App.vue](/frontend/src/App.vue) | vue | 2 | 2 | 0 | 4 |
| [frontend/src/assets/main.css](/frontend/src/assets/main.css) | PostCSS | 105 | 136 | 21 | 262 |
| [frontend/src/components/AppSidebar.vue](/frontend/src/components/AppSidebar.vue) | vue | 6 | 0 | 0 | 6 |
| [frontend/src/components/CartDrawer.vue](/frontend/src/components/CartDrawer.vue) | vue | 1 | 0 | 1 | 2 |
| [frontend/src/components/CourseEditorDrawer.vue](/frontend/src/components/CourseEditorDrawer.vue) | vue | 126 | 17 | 11 | 154 |
| [frontend/src/components/CourseWizardModal.vue](/frontend/src/components/CourseWizardModal.vue) | vue | 3 | 0 | 0 | 3 |
| [frontend/src/components/CreateGroupModal.vue](/frontend/src/components/CreateGroupModal.vue) | vue | 5 | 0 | 1 | 6 |
| [frontend/src/components/DC3EmpresaInstructor.vue](/frontend/src/components/DC3EmpresaInstructor.vue) | vue | 236 | 11 | 33 | 280 |
| [frontend/src/components/DC3Modal.vue](/frontend/src/components/DC3Modal.vue) | vue | 121 | 9 | 16 | 146 |
| [frontend/src/components/DC3Panel.test.ts](/frontend/src/components/DC3Panel.test.ts) | TypeScript | 73 | 24 | 20 | 117 |
| [frontend/src/components/DC3Panel.vue](/frontend/src/components/DC3Panel.vue) | vue | 402 | 35 | 41 | 478 |
| [frontend/src/components/LlamadaTimbrando.vue](/frontend/src/components/LlamadaTimbrando.vue) | vue | 210 | 0 | 25 | 235 |
| [frontend/src/components/NotificationBell.vue](/frontend/src/components/NotificationBell.vue) | vue | 57 | 5 | 13 | 75 |
| [frontend/src/components/SearchUserModal.vue](/frontend/src/components/SearchUserModal.vue) | vue | 5 | 0 | 1 | 6 |
| [frontend/src/components/VideoCallModal.vue](/frontend/src/components/VideoCallModal.vue) | vue | 59 | 0 | 0 | 59 |
| [frontend/src/components/VideoPlayer.vue](/frontend/src/components/VideoPlayer.vue) | vue | 38 | 0 | 1 | 39 |
| [frontend/src/composables/notificaciones.test.ts](/frontend/src/composables/notificaciones.test.ts) | TypeScript | 89 | 12 | 17 | 118 |
| [frontend/src/composables/notificaciones.ts](/frontend/src/composables/notificaciones.ts) | TypeScript | 39 | 42 | 12 | 93 |
| [frontend/src/composables/useLlamadas.test.ts](/frontend/src/composables/useLlamadas.test.ts) | TypeScript | 170 | 4 | 28 | 202 |
| [frontend/src/composables/useLlamadas.ts](/frontend/src/composables/useLlamadas.ts) | TypeScript | 156 | 38 | 29 | 223 |
| [frontend/src/composables/useScrollLock.test.ts](/frontend/src/composables/useScrollLock.test.ts) | TypeScript | 75 | 30 | 20 | 125 |
| [frontend/src/composables/useScrollLock.ts](/frontend/src/composables/useScrollLock.ts) | TypeScript | 16 | 13 | 6 | 35 |
| [frontend/src/composables/useTheme.ts](/frontend/src/composables/useTheme.ts) | TypeScript | 13 | 13 | 2 | 28 |
| [frontend/src/router/index.ts](/frontend/src/router/index.ts) | TypeScript | 3 | 3 | 0 | 6 |
| [frontend/src/stores/dc3.ts](/frontend/src/stores/dc3.ts) | TypeScript | 17 | 18 | 5 | 40 |
| [frontend/src/utils/dc3.test.ts](/frontend/src/utils/dc3.test.ts) | TypeScript | 71 | 20 | 12 | 103 |
| [frontend/src/utils/dc3.ts](/frontend/src/utils/dc3.ts) | TypeScript | 74 | 55 | 9 | 138 |
| [frontend/src/views/CheckoutSuccessView.vue](/frontend/src/views/CheckoutSuccessView.vue) | vue | 26 | 1 | 1 | 28 |
| [frontend/src/views/CursoPublicView.vue](/frontend/src/views/CursoPublicView.vue) | vue | 31 | 0 | 5 | 36 |
| [frontend/src/views/LoginView.vue](/frontend/src/views/LoginView.vue) | vue | 15 | 0 | 0 | 15 |
| [frontend/src/views/admin/AdminLayout.vue](/frontend/src/views/admin/AdminLayout.vue) | vue | 3 | 0 | 0 | 3 |
| [frontend/src/views/instructor/EntregasView.vue](/frontend/src/views/instructor/EntregasView.vue) | vue | -3 | 0 | -2 | -5 |
| [frontend/src/views/instructor/EstudiantesView.vue](/frontend/src/views/instructor/EstudiantesView.vue) | vue | 13 | 0 | 0 | 13 |
| [frontend/src/views/instructor/ExamenesInstructor.vue](/frontend/src/views/instructor/ExamenesInstructor.vue) | vue | 4 | 0 | 0 | 4 |
| [frontend/src/views/instructor/InstructorDashboard.vue](/frontend/src/views/instructor/InstructorDashboard.vue) | vue | 1 | 0 | 0 | 1 |
| [frontend/src/views/instructor/InstructorLayout.vue](/frontend/src/views/instructor/InstructorLayout.vue) | vue | 3 | 0 | 0 | 3 |
| [frontend/src/views/instructor/InstructorPerfilView.vue](/frontend/src/views/instructor/InstructorPerfilView.vue) | vue | 2 | 5 | 1 | 8 |
| [frontend/src/views/shared/MensajesView.vue](/frontend/src/views/shared/MensajesView.vue) | vue | 26 | 7 | 1 | 34 |
| [frontend/src/views/shared/StoreView.vue](/frontend/src/views/shared/StoreView.vue) | vue | 33 | 0 | -2 | 31 |
| [frontend/src/views/shared/VerificarConstancia.vue](/frontend/src/views/shared/VerificarConstancia.vue) | vue | 239 | 7 | 34 | 280 |
| [frontend/src/views/user/MisCapacitaciones.vue](/frontend/src/views/user/MisCapacitaciones.vue) | vue | -5 | 0 | -3 | -8 |
| [frontend/src/views/user/MisConstancias.vue](/frontend/src/views/user/MisConstancias.vue) | vue | 173 | 1 | 24 | 198 |
| [frontend/src/views/user/MisExamenes.vue](/frontend/src/views/user/MisExamenes.vue) | vue | 10 | 0 | 0 | 10 |
| [frontend/src/views/user/MisLicencias.vue](/frontend/src/views/user/MisLicencias.vue) | vue | 224 | 6 | 52 | 282 |
| [frontend/src/views/user/PerfilView.vue](/frontend/src/views/user/PerfilView.vue) | vue | -64 | -1 | -2 | -67 |
| [frontend/src/views/user/ResponderExamen.vue](/frontend/src/views/user/ResponderExamen.vue) | vue | 9 | 2 | 0 | 11 |
| [frontend/src/views/user/UserLayout.vue](/frontend/src/views/user/UserLayout.vue) | vue | 30 | 6 | 4 | 40 |
| [frontend/src/views/user/VerCapacitacion.vue](/frontend/src/views/user/VerCapacitacion.vue) | vue | 25 | 10 | 2 | 37 |
| [frontend/tailwind.config.js](/frontend/tailwind.config.js) | JavaScript | 8 | 22 | 0 | 30 |
| [gateway/Dockerfile](/gateway/Dockerfile) | Docker | 2 | 5 | 0 | 7 |
| [gateway/cmd/server/main.go](/gateway/cmd/server/main.go) | Go | 6 | 5 | 2 | 13 |
| [gateway/go.mod](/gateway/go.mod) | Go Module File | 4 | 0 | 1 | 5 |
| [gateway/go.sum](/gateway/go.sum) | Go Checksum File | 5 | 0 | 0 | 5 |
| [gateway/internal/config/config.go](/gateway/internal/config/config.go) | Go | 19 | 13 | 3 | 35 |
| [gateway/internal/handler/compras.go](/gateway/internal/handler/compras.go) | Go | 48 | 20 | 9 | 77 |
| [gateway/internal/handler/cursos.go](/gateway/internal/handler/cursos.go) | Go | 136 | 78 | 10 | 224 |
| [gateway/internal/handler/dc3.go](/gateway/internal/handler/dc3.go) | Go | 330 | 138 | 45 | 513 |
| [gateway/internal/handler/diagnostic.go](/gateway/internal/handler/diagnostic.go) | Go | 47 | 16 | 6 | 69 |
| [gateway/internal/handler/email.go](/gateway/internal/handler/email.go) | Go | -38 | -11 | -12 | -61 |
| [gateway/internal/handler/foros.go](/gateway/internal/handler/foros.go) | Go | 34 | 9 | 6 | 49 |
| [gateway/internal/handler/lecciones.go](/gateway/internal/handler/lecciones.go) | Go | -14 | 8 | -2 | -8 |
| [gateway/internal/handler/llamadas.go](/gateway/internal/handler/llamadas.go) | Go | 176 | 56 | 27 | 259 |
| [gateway/internal/handler/mensajes.go](/gateway/internal/handler/mensajes.go) | Go | 23 | 12 | 1 | 36 |
| [gateway/internal/handler/notificar.go](/gateway/internal/handler/notificar.go) | Go | 73 | 38 | 12 | 123 |
| [gateway/internal/handler/presign.go](/gateway/internal/handler/presign.go) | Go | 1 | 4 | 0 | 5 |
| [gateway/internal/handler/suscripciones.go](/gateway/internal/handler/suscripciones.go) | Go | 18 | 14 | 2 | 34 |
| [gateway/internal/handler/usuarios.go](/gateway/internal/handler/usuarios.go) | Go | -10 | -1 | -1 | -12 |
| [gateway/internal/handler/ws.go](/gateway/internal/handler/ws.go) | Go | 29 | 16 | 5 | 50 |
| [gateway/internal/hub/hub.go](/gateway/internal/hub/hub.go) | Go | 6 | 5 | 1 | 12 |
| [gateway/internal/hub/llamadas.go](/gateway/internal/hub/llamadas.go) | Go | 376 | 102 | 47 | 525 |
| [gateway/internal/hub/llamadas\_test.go](/gateway/internal/hub/llamadas_test.go) | Go | 368 | 19 | 66 | 453 |
| [gateway/internal/pdf/gotenberg.go](/gateway/internal/pdf/gotenberg.go) | Go | 68 | 47 | 13 | 128 |
| [gateway/internal/router/router.go](/gateway/internal/router/router.go) | Go | 9 | 21 | 4 | 34 |
| [gateway/internal/storage/r2.go](/gateway/internal/storage/r2.go) | Go | 6 | 6 | 1 | 13 |
| [gen/cursos/cursos.pb.go](/gen/cursos/cursos.pb.go) | Go | 815 | 77 | 121 | 1,013 |
| [gen/cursos/cursos\_grpc.pb.go](/gen/cursos/cursos_grpc.pb.go) | Go | 252 | 24 | 14 | 290 |
| [gen/foros/foros.pb.go](/gen/foros/foros.pb.go) | Go | 26 | 8 | 3 | 37 |
| [gen/usuarios/usuarios.pb.go](/gen/usuarios/usuarios.pb.go) | Go | 126 | 10 | 20 | 156 |
| [gen/usuarios/usuarios\_grpc.pb.go](/gen/usuarios/usuarios_grpc.pb.go) | Go | 0 | 4 | 0 | 4 |
| [go.work](/go.work) | Go Work File | 1 | 0 | 0 | 1 |
| [gotenberg/Dockerfile](/gotenberg/Dockerfile) | Docker | 11 | 25 | 6 | 42 |
| [gotenberg/fonts-dc3.conf](/gotenberg/fonts-dc3.conf) | Properties | 30 | 0 | 6 | 36 |
| [internal/handlers/perfil.go](/internal/handlers/perfil.go) | Go | -40 | 2 | -7 | -45 |
| [main.go](/main.go) | Go | -1 | 2 | 0 | 1 |
| [pkg/dc3/dc3.go](/pkg/dc3/dc3.go) | Go | 203 | 97 | 36 | 336 |
| [pkg/dc3/dc3\_test.go](/pkg/dc3/dc3_test.go) | Go | 197 | 38 | 25 | 260 |
| [pkg/dc3/go.mod](/pkg/dc3/go.mod) | Go Module File | 4 | 0 | 4 | 8 |
| [pkg/dc3/go.sum](/pkg/dc3/go.sum) | Go Checksum File | 11 | 0 | 1 | 12 |
| [services/cursos/cmd/server/main.go](/services/cursos/cmd/server/main.go) | Go | 45 | 48 | 5 | 98 |
| [services/cursos/go.sum](/services/cursos/go.sum) | Go Checksum File | -6 | 0 | 0 | -6 |
| [services/cursos/internal/handler/dc3\_handler.go](/services/cursos/internal/handler/dc3_handler.go) | Go | 69 | 14 | 12 | 95 |
| [services/cursos/internal/repository/cursos\_repository.go](/services/cursos/internal/repository/cursos_repository.go) | Go | 22 | 18 | 2 | 42 |
| [services/cursos/internal/repository/dc3\_repository.go](/services/cursos/internal/repository/dc3_repository.go) | Go | 241 | 69 | 24 | 334 |
| [services/cursos/internal/service/cursos\_service.go](/services/cursos/internal/service/cursos_service.go) | Go | 9 | 2 | 0 | 11 |
| [services/cursos/internal/service/dc3\_service.go](/services/cursos/internal/service/dc3_service.go) | Go | 212 | 83 | 28 | 323 |
| [services/cursos/internal/service/metodos\_pago.go](/services/cursos/internal/service/metodos_pago.go) | Go | 56 | 29 | 13 | 98 |
| [services/foros/internal/repository/foros\_repository.go](/services/foros/internal/repository/foros_repository.go) | Go | 13 | 5 | 1 | 19 |
| [services/usuarios/internal/handler/usuarios\_handler.go](/services/usuarios/internal/handler/usuarios_handler.go) | Go | 6 | 3 | 0 | 9 |
| [services/usuarios/internal/repository/usuario\_repository.go](/services/usuarios/internal/repository/usuario_repository.go) | Go | 30 | 9 | 3 | 42 |
| [services/usuarios/internal/service/usuarios\_service.go](/services/usuarios/internal/service/usuarios_service.go) | Go | 25 | 13 | 3 | 41 |

[Summary](results.md) / [Details](details.md) / [Diff Summary](diff.md) / Diff Details