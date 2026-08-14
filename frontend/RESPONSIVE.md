# Migración responsiva

Trabajo por fases. Este documento existe porque abarca más de una sesión: es la
referencia para retomarlo sin volver a decidir lo ya decidido.

## El problema, medido

Auditoría inicial sobre 68 archivos `.vue`:

| Métrica | Valor |
|---|---|
| Breakpoints **distintos** en CSS scoped | **68** |
| Archivos sin ninguna media query | 27 |
| `min-width` propios (pantallas grandes) | 0 |
| Tipografía en `rem` | 878 usos (1 en px) |

No faltaba código responsivo: 41 archivos ya tenían media queries. Faltaba un
**sistema**. Cada componente inventó sus cortes —600, 640, 680, 700, 720, 860,
900, 1200, 1240…— así que un mismo ancho de pantalla rompía unas piezas y otras
no, y ningún arreglo local componía con el de al lado.

Además, nadie había tratado pantallas grandes: sin tope de ancho, en un
ultrawide o una TV el contenido se estiraba hasta hacer las líneas ilegibles.

## Decisiones

**La escala es la de Tailwind, que el proyecto ya arrastraba.** No se inventa
una nueva; se hace explícita en `tailwind.config.js` y se usa también en el CSS
scoped. Un breakpoint nuevo debe justificarse: si algo "solo se ve mal en 730px",
casi siempre es un ancho fijo o una rejilla rígida, no un corte que falte.

```
sm   640   móvil grande / apaisado
md   768   tablet vertical
lg  1024   tablet apaisada / portátil pequeño
xl  1280   portátil / monitor
2xl 1536   monitor grande
tv  1920   TV y ultrawide — a partir de aquí el contenido se centra
```

**Pantallas grandes: contenedor centrado**, no más columnas. `--content-max`
(1440px) con `margin-inline: auto`. La única excepción deliberada es el visor de
curso (`.ver-curso-shell`), que es una interfaz de dos paneles y sí aprovecha
todo el ancho.

**Preferir layout intrínseco antes que media queries.** La mayoría de los 68
breakpoints resuelven rejillas rígidas, y eso se arregla sin un solo `@media`
con `auto-grid`. Menos código y funciona en anchos que nadie previó.

## Herramientas disponibles (fase 1, ya hechas)

Definidas en `src/assets/main.css`:

| Clase / variable | Para qué |
|---|---|
| `.app-container` | Contenedor centrado con tope y gutter |
| `.auto-grid` | Rejilla que se adapta sola, sin breakpoints |
| `.auto-grid-sm` / `.auto-grid-lg` | Igual con mínimo de 180px / 320px |
| `.table-scroll` | Scroll propio para tablas anchas |
| `.flex-wrap-row` | Fila que se apila sin media query |
| `.touch-target` | Área de toque de 44px en móvil |
| `.safe-bottom` / `.safe-top` | Notch y barra de gestos |
| `--content-max`, `--gutter`, `--touch-min` | Tokens compartidos |

## Fases

- [x] **1 · Sistema base.** Escala explícita, contenedor máximo, utilidades,
      tokens. `.page-content` con tope centrado, que cubre los tres layouts de
      una sola vez.
- [x] **2 · Layouts.** Todo el CSS de layout es compartido en `main.css`, así
      que un solo cambio cubre los tres paneles.
      - Colapso del sidebar movido de **768 → 1024**. Con 768, una tablet de
        800-1000px conservaba el sidebar fijo de 260px y dejaba al contenido
        540px: menos que un móvil apaisado.
      - `100vh` → `100dvh` con respaldo. En móvil, `100vh` no cuenta la barra
        de direcciones y recortaba el pie de pantalla.
      - `padding-bottom: env(safe-area-inset-bottom)` en el sidebar: el botón
        de cerrar sesión quedaba bajo la barra de gestos.
      - Breadcrumbs con elipsis en <640: eran lo primero que desbordaba.
      - Cajón limitado a `82vw`: en un móvil de 320px, 260px no dejaban ver
        que había página detrás.
      - `useScrollLock` (composable con 5 pruebas): la página ya no se
        desplaza por detrás del cajón abierto.
- [x] **3 · Pantallas de alto tráfico.** `StoreView`, `LoginView`,
      `VerCapacitacion`, `MisCapacitaciones`. De **13 breakpoints a 6**, todos
      de la escala.
      - Rejillas rígidas → `auto-fit`: cifras del hero, bloque B2B, rol del
        login, panel de estadísticas. Cinco breakpoints desaparecieron porque
        dejaron de hacer falta.
      - **Bug real corregido**: el catálogo usaba `auto-fill`, que crea las
        pistas vacías igualmente. Al filtrar y quedar dos resultados, las
        tarjetas se quedaban a un cuarto de ancho con la fila medio en blanco.
        Ahora `auto-fit`.
      - Login: el panel ilustrado desaparecía en 860, pero entre 860 y 1024 se
        veía en una franja tan estrecha que salía recortado. Subido a lg.
      - `VerCapacitacion`: índice de lecciones alineado al colapso del sidebar
        (1024). Por debajo no cabían tres columnas de navegación.
- [x] **4 · Paneles de gestión.** Instructor y admin. De **10 breakpoints
      distintos a 3** (639 · 767 · 1023), todos de la escala.
      - Las 12 rejillas `repeat(N, 1fr)` pasaron a `auto-fit`. Con ellas se
        fueron 6 media queries que solo cambiaban el número de columnas.
      - **Bug real en las 3 tablas**: tenían `overflow-x: auto` (o ni eso) pero
        la tabla con `width: 100%` **se comprime en vez de desbordar**, así que
        el scroll nunca se activaba y las celdas quedaban ilegibles — una
        columna de correos en 90px. Con `min-width` la tabla desborda y el
        contenedor sí ofrece scroll.
      - `EstudiantesView` tenía `overflow: hidden`: las columnas de la derecha
        se recortaban sin forma de alcanzarlas.
- [x] **5 · Modales y drawers.** 8 de los 11 no tenían **ninguna** media query.
      - Utilidad compartida `.sheet-panel`: por debajo de sm el panel pasa a
        hoja a pantalla completa. Se añade junto a la clase que ya tuviera el
        componente, así que no cambia nada en escritorio. Lleva `!important`
        porque los estilos *scoped* de Vue añaden `[data-v-…]` y ganan en
        especificidad a una utilidad global.
      - **`vh` → `dvh` en los tres modales que lo usaban.** Con `vh` la altura
        se calcula sin la barra de direcciones del navegador móvil: el modal
        salía más alto que el área visible y **sus botones quedaban fuera de
        alcance**, sin forma de confirmar ni cancelar.
      - `SearchUserModal` tenía `max-height` pero no `overflow-y`: con más
        resultados que altura, la lista se cortaba sin poder desplazarla.
      - `useScrollLock` en los cuatro con estado propio, y `overscroll-behavior:
        contain` para que el scroll del modal no arrastre la página al llegar
        al final.
- [x] **6 · Barrido final.**
      - **17 archivos** con breakpoints fuera de escala migrados. Mapeo usado:
        400/420/480/560/600/620 → **639** · 680/720 → **767** ·
        860/900/1100 → **1023**.
      - **Cero rejillas `repeat(N)` en todo el proyecto.** Las 7 restantes
        pasaron a `auto-fit`, y con ellas 4 media queries más.
      - `MisExamenes` tenía el mismo bug de `auto-fill` que la tienda.
      - `vh` → `dvh` con respaldo en 17 archivos. Los `max-height` de paneles
        flotantes pasaron a `dvh` directo.
      - Cortafuegos contra desborde por palabras largas: códigos de licencia,
        correos y URLs son una sola "palabra" para el navegador y desbordan el
        contenedor. `overflow-wrap: anywhere` y no `word-break: break-all`,
        que parte cualquier palabra aunque quepa y destroza la lectura.

## Estado final

| Métrica | Antes | Ahora |
|---|---|---|
| Valores de breakpoint en `@media` | **68** | **7** |
| Breakpoints lógicos | sin escala | 4 (sm · md · lg · xl) |
| Rejillas `repeat(N)` rígidas | 12+ | **0** |
| `min-width` para pantallas grandes | 0 | contenedor a 1440px |
| Archivos sin ninguna media query | 27 | irrelevante: ya no hacen falta |

Los 7 valores son 4 breakpoints expresados como pares complementarios
(`max-width: 639` / `min-width: 640`), no 7 cortes distintos.

Que un archivo no tenga media queries **ya no es un síntoma**: con `auto-fit`,
`clamp()` y las utilidades compartidas, la mayoría de componentes se adaptan sin
un solo `@media`. Es el objetivo, no una carencia.

## Cómo migrar un componente

1. Sustituir sus `@media` por los valores de la escala.
2. Cambiar rejillas de N columnas fijas por `auto-grid`.
3. Envolver tablas en `.table-scroll`.
4. Buscar anchos fijos (`width: 420px`) y pasarlos a `max-width` o `%`.
5. Comprobar a 320px de ancho, que es el móvil más estrecho en uso real.

## Verificación

```
npx vite build && npx oxlint src/ && npx vitest run
```

El desborde horizontal es el fallo más difícil de rastrear: un solo hijo
demasiado ancho saca scroll lateral en toda la página y el culpable puede estar
a diez niveles. Para localizarlo, en la consola del navegador:

```js
document.querySelectorAll('*').forEach(el => {
  if (el.scrollWidth > document.documentElement.clientWidth) console.log(el)
})
```
