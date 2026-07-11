import { defineConfig } from '@vben/vite-config';

import ElementPlus from 'unplugin-element-plus/vite';

export default defineConfig(async () => {
  return {
    application: {},
    vite: {
      plugins: [
        ElementPlus({
          format: 'esm',
        }),
      ],
      server: {
        port: 11526,
        host: true,
        proxy: {
          '/admin': {
            changeOrigin: true,
            target: 'http://127.0.0.1:11505',
          },
          '/api': {
            changeOrigin: true,
            target: 'http://127.0.0.1:11505',
          },
        },
      },
    },
  };
});
