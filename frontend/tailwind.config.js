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
        main: 'rgb(var(--bg-main-rgb) / <alpha-value>)',
        surface: 'rgb(var(--bg-surface-rgb) / <alpha-value>)',
        card: 'rgb(var(--bg-card-rgb) / <alpha-value>)',
        subtle: 'rgb(var(--border-color-rgb) / <alpha-value>)',
        hover: 'rgb(var(--bg-hover-rgb) / <alpha-value>)',
        brand: {
          periwinkle: 'rgb(var(--accent-blue-rgb) / <alpha-value>)',
          'periwinkle-hover': 'rgb(var(--accent-blue-hover-rgb) / <alpha-value>)',
          'periwinkle-muted': 'rgb(var(--accent-blue-rgb) / 0.15)',
        },
        status: {
          up: 'rgb(var(--status-up-rgb) / <alpha-value>)',
          'up-bg': 'rgb(var(--status-up-rgb) / 0.12)',
          down: 'rgb(var(--status-down-rgb) / <alpha-value>)',
          'down-bg': 'rgb(var(--status-down-rgb) / 0.12)',
          'down-border': 'var(--status-down-border)',
          warning: 'rgb(var(--status-warning-rgb) / <alpha-value>)',
          'warning-bg': 'rgb(var(--status-warning-rgb) / 0.12)'
        },
        text: {
          main: 'rgb(var(--text-primary-rgb) / <alpha-value>)',
          secondary: 'rgb(var(--text-secondary-rgb) / <alpha-value>)',
          muted: 'rgb(var(--text-muted-rgb) / <alpha-value>)'
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
