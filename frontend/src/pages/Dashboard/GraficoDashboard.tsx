import { lazy, Suspense, type CSSProperties } from 'react';
import type { EChartsOption } from 'echarts';

const GraficoECharts = lazy(() =>
  import('../../components/GraficoECharts/GraficoECharts').then((modulo) => ({ default: modulo.GraficoECharts })),
);

interface GraficoDashboardProps {
  opcao: EChartsOption;
  rotuloAcessivel: string;
  resumoAcessivel: string;
  style: CSSProperties;
}

export function GraficoDashboard(props: GraficoDashboardProps) {
  return (
    <Suspense
      fallback={
        <div
          className="flex animate-pulse items-center justify-center rounded-lg bg-slate-100 text-sm text-slate-600"
          style={props.style}
          role="status"
        >
          Carregando visualização…
        </div>
      }
    >
      <GraficoECharts {...props} />
    </Suspense>
  );
}
