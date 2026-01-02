/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        brand: {
          bg: '#2b3339',      // Everforest Background (Hard)
          card: '#323c41',    // Everforest Soft Background
          border: '#3a454a',  // Everforest Border
          primary: '#a7c080', // Everforest Green
          secondary: '#dbbc7f', // Everforest Amber/Yellow
          accent: '#e67e80',  // Everforest Red
          blue: '#7fbbb3',    // Everforest Blue
          purple: '#d699b6',  // Everforest Purple
          orange: '#e69875',  // Everforest Orange
        }
      }
    },
  },
  plugins: [],
}
