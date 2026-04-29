import type { Config } from "tailwindcss"

const config: Config = {
  content: [
    "./src/pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/components/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        // Official RIT Brand Colors
        rit: {
          orange: "#F76902", // Primary orange
          white: "#FFFFFF", // Primary white
          black: "#000000", // Secondary black
          charcoal: "#2C2C2C", // Charcoal for text
          gray: {
            light: "#D0D3D4", // PMS427C
            medium: "#A2AAAD", // PMS429C
            dark: "#7C878E", // PMS430C
            warmLight: "#D7D2CB", // Warm Gray 1
            warmMedium: "#ACA39A", // Warm Gray 5
          },
        },
        // Standard orange palette for hover states
        orange: {
          50: "#FFF7ED",
          100: "#FFEDD5",
          200: "#FED7AA",
          300: "#FDBA74",
          400: "#FB923C",
          500: "#F76902", // RIT Orange
          600: "#EA580C",
          700: "#C2410C",
          800: "#9A360B",
          900: "#7C2D12",
        },
        // Standard gray palette
        gray: {
          50: "#F9FAFB",
          100: "#F3F4F6",
          200: "#E5E7EB",
          300: "#D1D5DB",
          400: "#9CA3AF",
          500: "#6B7280",
          600: "#4B5563",
          700: "#374151",
          800: "#1F2937",
          900: "#111827",
        },
        // Accent colors from RIT palette (use sparingly)
        accent: {
          green: "#84BD00", // PMS376C
          yellowGreen: "#C4D600", // PMS382C
          cyan: "#009CBD", // PMS7703C
          purple: "#7D55C7", // PMS2665C
          red: "#DA291C", // PMS485C
          gold: "#F6BE00", // PMS7408C
        },
      },
      fontFamily: {
        sans: ["var(--font-geist-sans)", "system-ui", "sans-serif"],
        mono: ["var(--font-geist-mono)", "monospace"],
      },
      borderRadius: {
        xl: "1.5rem",
        "2xl": "2rem",
      },
      boxShadow: {
        bubble: "0 8px 32px 0 rgba(247, 105, 2, 0.15)",
        "bubble-lg": "0 12px 48px 0 rgba(247, 105, 2, 0.25)",
      },
      animation: {
        "bounce-gentle": "bounce-gentle 3s infinite",
        float: "float 3s ease-in-out infinite",
      },
      keyframes: {
        "bounce-gentle": {
          "0%, 100%": { transform: "translateY(0)" },
          "50%": { transform: "translateY(-10px)" },
        },
        float: {
          "0%, 100%": { transform: "translateY(0px)" },
          "50%": { transform: "translateY(-15px)" },
        },
      },
    },
  },
  plugins: [],
}

export default config
