/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        dark: {
          bg: '#0A0A0B',
          surface: '#151517',
          card: '#18181B',
          border: '#26262A',
          hover: '#222226'
        },
        brand: {
          periwinkle: '#7B96F5',
          'periwinkle-hover': '#95ABF7',
          'periwinkle-muted': 'rgba(123, 150, 245, 0.15)',
        },
        status: {
          up: '#3ECF8E',
          'up-bg': 'rgba(62, 207, 142, 0.12)',
          down: '#F16565',
          'down-bg': 'rgba(241, 101, 101, 0.12)',
          'down-border': 'rgba(241, 101, 101, 0.35)',
          warning: '#F5A65B',
          'warning-bg': 'rgba(245, 166, 91, 0.12)'
        }
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'sans-serif'],
        mono: ['JetBrains Mono', 'Fira Code', 'Consolas', 'monospace']
      }
    },
  },
  plugins: [],
}
