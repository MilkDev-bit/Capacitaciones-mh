/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  theme: {
    /**
     * ESCALA ÚNICA DE BREAKPOINTS DEL PROYECTO.
     *
     * Se declara explícita (aunque sm–2xl sean los valores por defecto de
     * Tailwind) porque es la referencia que debe usar TAMBIÉN el CSS scoped de
     * los componentes. La auditoría encontró 68 breakpoints distintos repartidos
     * por los .vue —600, 640, 680, 700, 720, 860, 900, 1200, 1240…—, cada uno
     * inventado por su componente. Con eso, un mismo ancho de pantalla rompe
     * unas piezas y otras no, y ningún arreglo local compone con el de al lado.
     *
     * Al migrar un componente, sus @media pasan a estos valores. No se añaden
     * intermedios: si algo "solo se ve mal en 730px", casi siempre es un ancho
     * fijo o una rejilla rígida, no un breakpoint que falte.
     *
     *   sm   640  móvil grande / móvil apaisado
     *   md   768  tablet vertical
     *   lg  1024  tablet apaisada / portátil pequeño
     *   xl  1280  portátil / monitor
     *   2xl 1536  monitor grande
     *   tv  1920  televisión y ultrawide — a partir de aquí NO se sigue
     *             estirando: el contenido se centra (ver .app-container).
     */
    screens: {
      sm: '640px',
      md: '768px',
      lg: '1024px',
      xl: '1280px',
      '2xl': '1536px',
      tv: '1920px',
    },
    extend: {
      colors: {
        brand: {
          DEFAULT: '#f97316',
          dark: '#ea580c',
          darker: '#c2410c',
          light: '#fff7ed',
          border: '#fed7aa',
        },
      },
    },
  },
  plugins: [],
}
