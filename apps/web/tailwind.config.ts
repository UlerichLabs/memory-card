import type { Config } from "tailwindcss";

export default {
  darkMode: ["class"],
  content: [
    "./index.html",
    "./src/**/*.{ts,tsx}"
  ],
  theme: {
    extend: {
      borderRadius: {
        lg: "var(--radius)",
        md: "calc(var(--radius) - 2px)",
        sm: "calc(var(--radius) - 4px)"
      },
      colors: {
        background: "var(--background)",
        surface: "var(--surface)",
        "surface-raised": "var(--surface-raised)",
        border: "var(--border)",
        "border-strong": "var(--border-strong)",
        foreground: "var(--foreground)",
        muted: "var(--muted)",
        accent: "var(--accent)",
        "accent-dim": "var(--accent-dim)",
        gold: "var(--gold)",
        "gold-dim": "var(--gold-dim)",
        green: "var(--green)",
        red: "var(--red)",
        orange: "var(--orange)",
        purple: "var(--purple)",
        input: "var(--border)",
        ring: "var(--accent)",
        subtle: "var(--muted)",
        "muted-foreground": "var(--muted)",
        "accent-foreground": "var(--background)",
        "primary-foreground": "var(--background)",
        "secondary-foreground": "var(--foreground)",
        "destructive-foreground": "var(--foreground)",
        primary: {
          DEFAULT: "var(--accent)",
          foreground: "var(--background)"
        },
        secondary: {
          DEFAULT: "var(--surface-raised)",
          foreground: "var(--foreground)"
        },
        destructive: {
          DEFAULT: "var(--red)",
          foreground: "var(--foreground)"
        },
        card: {
          DEFAULT: "var(--surface)",
          foreground: "var(--foreground)"
        },
        popover: {
          DEFAULT: "var(--surface)",
          foreground: "var(--foreground)"
        }
      }
    }
  },
  plugins: [require("tailwindcss-animate")]
} satisfies Config;
