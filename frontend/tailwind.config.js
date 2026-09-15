/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./index.html', './src/**/*.{html,ts}'],
  theme: {
    extend: {
      fontFamily: {
        sans: ['"Space Grotesk"', 'Inter', 'system-ui', 'sans-serif'],
      },
      colors: {
        brand: {
          primary: '#4c63ed',
          secondary: '#d946ef',
          accent: '#f97316',
          slate: '#0f172a',
          muted: '#94a3b8',
          surface: '#0b1120',
        },
      },
    },
  },
  plugins: [],
};
