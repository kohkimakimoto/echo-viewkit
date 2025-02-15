import { defineConfig } from 'vite';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
  publicDir: false,
  plugins: [
    tailwindcss(),
  ],
  build: {
    manifest: "manifest.json",
    outDir: "public/build",
    rollupOptions: {
      input: ['resources/js/app.js', 'resources/css/app.css'],
    },
  },
});
