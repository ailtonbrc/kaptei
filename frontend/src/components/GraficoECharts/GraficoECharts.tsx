import { useId, type CSSProperties } from 'react';
import EChartsReact from 'echarts-for-react';
import * as echarts from 'echarts/core';
import { FunnelChart, LineChart, PieChart } from 'echarts/charts';
import { GridComponent, LegendComponent, TooltipComponent } from 'echarts/components';
import { CanvasRenderer } from 'echarts/renderers';
import type { EChartsOption } from 'echarts';

// Registra apenas os tipos de gráfico e componentes utilizados no sistema,
// evitando importar o bundle completo do ECharts (~1MB).
echarts.use([
  LineChart,
  FunnelChart,
  PieChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  CanvasRenderer,
]);

// Compatibilidade CJS/ESM: o echarts-for-react v3 pode expor o componente
// em .default quando consumido via Vite (ESM). Esta linha garante que
// sempre obtemos a função do componente, independente do bundler.
const ReactEChartsCore = (
  (EChartsReact as unknown as { default?: typeof EChartsReact }).default ?? EChartsReact
);

interface GraficoEChartsProps {
  opcao: EChartsOption;
  rotuloAcessivel: string;
  resumoAcessivel: string;
  style?: CSSProperties;
}

export function GraficoECharts({ opcao, rotuloAcessivel, resumoAcessivel, style }: GraficoEChartsProps) {
  const descricaoId = useId();

  return (
    <figure className="w-full" aria-labelledby={descricaoId}>
      <div role="img" aria-label={rotuloAcessivel}>
        <ReactEChartsCore
          echarts={echarts}
          option={opcao}
          style={style}
          notMerge
          lazyUpdate
        />
      </div>
      <figcaption id={descricaoId} className="sr-only">
        {resumoAcessivel}
      </figcaption>
    </figure>
  );
}
