import {defineConfig} from 'vite'
import react from '@vitejs/plugin-react'
import {resolve} from 'path';

export default defineConfig({
    plugins: [react()],
    resolve: {
        alias: {
            '@': resolve(__dirname, './src'),
        },
    },
    server: {
        proxy: {
            '/api': {
                target: 'http://localhost:9001',
                changeOrigin: true,
            },
            '/get': {
                target: 'http://localhost:9001',
                changeOrigin: true,
            },
            '/post': {
                target: 'http://localhost:9001',
                changeOrigin: true,
            },
            '/data': {
                target: 'http://localhost:9001',
                changeOrigin: true,
            }
        }
    }
})
