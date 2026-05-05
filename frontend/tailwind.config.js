/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  theme: {
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
