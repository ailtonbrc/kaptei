import path from 'path';
import { defineConfig } from 'vite';
import type { Plugin } from 'vite';
import react from '@vitejs/plugin-react';

function validarOrcamentoDeChunks(): Plugin {
  return {
    name: 'validar-orcamento-de-chunks',
    apply: 'build',
    generateBundle(_, bundle) {
      for (const artefato of Object.values(bundle)) {
        if (artefato.type !== 'chunk') continue;

        const limite = artefato.isEntry
          ? 350_000
          : artefato.name.includes('GraficoECharts')
            ? 575_000
            : artefato.name.includes('Agendamentos')
              ? 225_000
              : artefato.name.includes('Dashboard')
                ? 50_000
                : 250_000;
        const tamanho = Buffer.byteLength(artefato.code, 'utf8');
        if (tamanho > limite) {
          throw new Error(`Chunk ${artefato.name} excedeu o orçamento: ${tamanho} bytes de ${limite} permitidos.`);
        }
      }
    },
  };
}

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), validarOrcamentoDeChunks()],
  // ECharts é carregado somente na rota analítica; o limite reflete esse pacote isolado.
  build: { chunkSizeWarningLimit: 575 },
  resolve: {
    alias: {
      '@': path.resolve(import.meta.dirname, './src'),
    },
  },
  // Força o Vite a pré-processar o echarts-for-react (CommonJS) para ESM,
  // evitando o erro "Element type is invalid: got object" no modo desenvolvimento.
  optimizeDeps: {
    include: ['echarts-for-react', 'echarts/core'],
  },
});

