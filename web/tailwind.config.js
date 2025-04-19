// tailwind.config.js
export default {
  darkMode: "class", // ✅ Enables `dark:` variants using class on <html>
  content: ["./index.html", "./src/**/*.{vue,js,ts,jsx,tsx}"],
  theme: {
    extend: {},
  },
  plugins: ["@tailwindcss/postcss", "@tailwindcss/vite"],
};
