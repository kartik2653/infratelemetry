/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./app/**/*.{ts,tsx}", "./components/**/*.{ts,tsx}", "./lib/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        paper: "#f2f4f1",
        line: "#e2e6e1",
        ink: { DEFAULT: "#17211c", soft: "#5d6b64", faint: "#8b9891" },
        pine: { DEFAULT: "#1c6b45", dark: "#11462d", soft: "#e7f3ec" },
        sea: "#1e4f86",
        clay: "#8d5a16",
        rose: "#9d3140",
        side: "#101614",
      },
      boxShadow: {
        card: "0 1px 2px rgba(16, 22, 18, 0.05)",
      },
      fontFamily: {
        sans: ["Segoe UI", "ui-sans-serif", "system-ui", "sans-serif"],
      },
    },
  },
  plugins: [],
};
