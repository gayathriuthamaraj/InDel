module.exports = {
  content: [
    "./app/**/*.{js,ts,jsx,tsx}",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'Fira Code', 'monospace'],
      },
      colors: {
        primary: {
          light: '#E6F3F7',
          DEFAULT: '#00739D',
          dark: '#005A7A',
        },
        surface: '#FFFFFF',
        background: '#F5F7F9',
        text: {
          primary: '#1A2533',
          secondary: '#64748B',
          muted: '#94A3B8',
        },
        success: '#1E9E5A',
        warning: '#F59E0B',
        error: '#D92D20',
        dark: {
          900: '#0A1628',
          800: '#0F1F3D',
          700: '#162845',
        },
      },
      backgroundImage: {
        'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))',
        'hero-gradient': 'linear-gradient(135deg, #0A1628 0%, #0F1F3D 50%, #003D54 100%)',
        'card-gradient': 'linear-gradient(135deg, rgba(0,115,157,0.08) 0%, rgba(0,90,122,0.04) 100%)',
      },
      animation: {
        'fade-up': 'fadeUp 0.7s ease-out forwards',
        'fade-in': 'fadeIn 0.5s ease-out forwards',
        'shimmer': 'shimmer 2.5s linear infinite',
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        'counter': 'counter 1.5s ease-out forwards',
        'flow-line': 'flowLine 1.5s ease-out forwards',
        'glow': 'glow 2s ease-in-out infinite alternate',
        'slide-down': 'slideDown 0.3s ease-out forwards',
        'blob': 'blob 7s infinite',
      },
      keyframes: {
        fadeUp: {
          '0%': { opacity: '0', transform: 'translateY(24px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        shimmer: {
          '0%': { backgroundPosition: '-200% 0' },
          '100%': { backgroundPosition: '200% 0' },
        },
        flowLine: {
          '0%': { height: '0%' },
          '100%': { height: '100%' },
        },
        glow: {
          '0%': { boxShadow: '0 0 20px rgba(0,115,157,0.3)' },
          '100%': { boxShadow: '0 0 40px rgba(0,115,157,0.6)' },
        },
        slideDown: {
          '0%': { opacity: '0', transform: 'translateY(-8px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        blob: {
          '0%': { transform: 'translate(0px, 0px) scale(1)' },
          '33%': { transform: 'translate(30px, -50px) scale(1.1)' },
          '66%': { transform: 'translate(-20px, 20px) scale(0.9)' },
          '100%': { transform: 'translate(0px, 0px) scale(1)' },
        },
      },
      backdropBlur: {
        xs: '2px',
      },
      transitionDuration: {
        '400': '400ms',
      },
    },
  },
  plugins: [],
};
