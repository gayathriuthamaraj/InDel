module.exports = {
  content: [
    "./app/**/*.{js,ts,jsx,tsx}",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          light: '#E6F3F7', // BlueSoft
          DEFAULT: '#00739D', // BrandBlue
          dark: '#005A7A', // BlueDeep
        },
        surface: '#FFFFFF', // SurfaceWhite
        background: '#F5F7F9', // BackgroundWarmWhite
        text: {
          primary: '#666766',
          secondary: '#919291',
        },
        success: '#1E9E5A',
        warning: '#F59E0B',
        error: '#D92D20',
      },
    },
  },
  plugins: [],
};
