import { defineConfig } from 'vite';

export default defineConfig({
  publicDir: false,
  build: {
    manifest: "manifest.json",
    outDir: "public/build",
    rollupOptions: {
      input: ['assets/app.js', "assets/app.css"],
    },
  },
});
