/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './src/**/*.{js,ts,jsx,tsx,mdx}',
    './app/**/*.{js,ts,jsx,tsx,mdx}',
    './pages/**/*.{js,ts,jsx,tsx,mdx}',
    './components/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      // Paper Portfolio Design System
      colors: {
        // Primary palette
        paper: {
          cream: '#FBF8F3',      // Main background - paper color
          gray: '#E8E4DD',        // Borders, dividers
          'warm-gray': '#6B6456', // Secondary text
        },
        ink: {
          DEFAULT: '#1A1A1A',     // Primary text - ink black
          black: '#1A1A1A',
        },
        accent: {
          blue: '#2C5F77',        // Links, highlights
        },
        // Semantic colors using paper palette
        background: {
          DEFAULT: '#FBF8F3',
          paper: '#FBF8F3',
          white: '#FFFFFF',
        },
        foreground: {
          DEFAULT: '#1A1A1A',
          primary: '#1A1A1A',
          secondary: '#6B6456',
        },
        border: {
          DEFAULT: '#E8E4DD',
        },
      },

      fontFamily: {
        // Serif for headings, numbers, and key elements
        serif: ['Lora', 'Georgia', 'Cambria', 'Times New Roman', 'Times', 'serif'],
        // Sans for body text (system fonts for performance)
        sans: [
          '-apple-system',
          'BlinkMacSystemFont',
          'Segoe UI',
          'Roboto',
          'Helvetica Neue',
          'Arial',
          'sans-serif',
        ],
      },

      fontSize: {
        // Paper Portfolio scale: 14/16/18/24/32/40/48px
        'xs': ['14px', { lineHeight: '1.6' }],
        'sm': ['16px', { lineHeight: '1.6' }],
        'base': ['18px', { lineHeight: '1.6' }],
        'lg': ['24px', { lineHeight: '1.5' }],
        'xl': ['32px', { lineHeight: '1.4' }],
        '2xl': ['40px', { lineHeight: '1.3' }],
        '3xl': ['48px', { lineHeight: '1.2' }],
      },

      fontWeight: {
        normal: '400',
        semibold: '600',
      },

      spacing: {
        // Base unit: 8px
        // Common spacing values for Paper Portfolio
        'statement': '32px',  // Statement-like generous spacing
        'section': '24px',    // Between sections
      },

      borderRadius: {
        // Subtle, professional radius
        'none': '0',
        'sm': '4px',
        DEFAULT: '6px',
        'md': '8px',
        'lg': '12px',
      },

      boxShadow: {
        // Tactile, paper-like shadows
        'paper': '2px 4px 12px rgba(0, 0, 0, 0.08)',
        'paper-lg': '4px 8px 24px rgba(0, 0, 0, 0.12)',
        'paper-hover': '4px 8px 16px rgba(0, 0, 0, 0.12)',
        // Override default shadows with paper-appropriate ones
        'sm': '2px 4px 8px rgba(0, 0, 0, 0.06)',
        DEFAULT: '2px 4px 12px rgba(0, 0, 0, 0.08)',
        'md': '4px 8px 16px rgba(0, 0, 0, 0.10)',
        'lg': '4px 8px 24px rgba(0, 0, 0, 0.12)',
        'xl': '8px 16px 32px rgba(0, 0, 0, 0.15)',
      },

      backgroundImage: {
        // Paper texture - subtle noise overlay
        'paper-texture': `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='100' height='100'%3E%3Cfilter id='noise'%3E%3CfeTurbulence baseFrequency='0.65' /%3E%3C/filter%3E%3Crect width='100' height='100' filter='url(%23noise)' opacity='0.03'/%3E%3C/svg%3E")`,
      },

      transitionDuration: {
        // Subtle, refined transitions
        DEFAULT: '200ms',
        'fast': '150ms',
        'normal': '200ms',
        'slow': '300ms',
      },

      transitionTimingFunction: {
        DEFAULT: 'ease-out',
      },
    },
  },
  plugins: [
    // Custom utilities for Paper Portfolio design system
    function({ addUtilities }) {
      addUtilities({
        '.text-balance': {
          'text-wrap': 'balance',
        },
        '.paper-bg': {
          'background-color': '#FBF8F3',
          'background-image': `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='100' height='100'%3E%3Cfilter id='noise'%3E%3CfeTurbulence baseFrequency='0.65' /%3E%3C/filter%3E%3Crect width='100' height='100' filter='url(%23noise)' opacity='0.03'/%3E%3C/svg%3E")`,
        },
        '.statement-border': {
          'border': '1px solid #E8E4DD',
        },
        '.statement-border-strong': {
          'border': '2px solid #1A1A1A',
        },
      });
    },
  ],
}
