import type { ResumoConversaoSite } from '../../types/dashboard';
import type { EChartsOption } from 'echarts';
import { GraficoDashboard } from './GraficoDashboard';

interface PainelConversaoSiteProps {
  resumo: ResumoConversaoSite;
}

export function PainelConversaoSite({ resumo }: PainelConversaoSiteProps) {
  const etapas = [
    { rotulo: 'Visitas ao site', valor: resumo.visitas_site },
    { rotulo: 'Imóveis visualizados', valor: resumo.imoveis_visualizados },
    { rotulo: 'Formulários iniciados', valor: resumo.formularios_iniciados },
    { rotulo: 'Contatos enviados', valor: resumo.contatos_enviados },
  ];
  const base = Math.max(1, resumo.visitas_site);
  const fontes = Object.entries(resumo.fontes || {}).sort((a, b) => b[1] - a[1]);

  const categories = etapas.map(e => e.rotulo).reverse();
  const values = etapas.map(e => e.valor).reverse();

  const chartOption: EChartsOption = {
    tooltip: { 
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      backgroundColor: 'rgba(255, 255, 255, 0.95)',
      borderColor: '#e2e8f0',
      textStyle: { color: '#1e293b' },
      extraCssText: 'box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1); border-radius: 8px;'
    },
    grid: { left: '2%', right: '6%', bottom: '0%', top: '2%', containLabel: true },
    xAxis: {
      type: 'value',
      axisLabel: { color: '#64748b' },
      splitLine: { lineStyle: { type: 'dashed', color: '#f1f5f9' } },
      axisLine: { show: false },
      axisTick: { show: false }
    },
    yAxis: {
      type: 'category',
      data: categories,
      axisLabel: { color: '#64748b', fontWeight: '500' },
      axisLine: { lineStyle: { color: '#e2e8f0' } },
      axisTick: { show: false }
    },
    series: [
      {
        name: 'Conversões',
        type: 'bar',
        barWidth: '45%',
        data: values,
        itemStyle: {
          color: '#0ea5e9',
          borderRadius: [0, 6, 6, 0]
        },
        label: {
          show: true,
          position: 'right',
          color: '#64748b',
          fontWeight: 'bold'
        }
      }
    ]
  };

  return (
    <section className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm" aria-labelledby="titulo-conversao-site">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h2 id="titulo-conversao-site" className="text-lg font-bold text-slate-900">Conversão do site nos últimos 30 dias</h2>
          <p className="mt-1 text-sm text-slate-500">Sessões first-party, sem IP, user-agent ou dados de contato.</p>
        </div>
        <div className="rounded-xl bg-blue-50 px-4 py-2 text-right">
          <span className="block text-xs font-bold uppercase tracking-wide text-blue-700">Taxa de conversão</span>
          <strong className="text-2xl text-blue-950">{resumo.taxa_conversao.toLocaleString('pt-BR', { maximumFractionDigits: 2 })}%</strong>
        </div>
      </div>

      <div className="mt-6 grid gap-6 lg:grid-cols-[2fr_1fr]">
        <div className="flex-1 w-full relative">
          <GraficoDashboard
            opcao={chartOption}
            rotuloAcessivel="Gráfico de conversão do site"
            resumoAcessivel={etapas.map((item) => `${item.rotulo}: ${item.valor}`).join('; ')}
            style={{ height: '240px', width: '100%' }}
          />
        </div>
        <div className="rounded-xl bg-slate-50 p-4">
          <h3 className="text-sm font-bold text-slate-900">Fontes das visitas</h3>
          {fontes.length ? <ul className="mt-3 space-y-2 text-sm">{fontes.slice(0, 6).map(([fonte, total]) => <li key={fonte} className="flex justify-between gap-3"><span className="truncate text-slate-600">{fonte}</span><strong>{total}</strong></li>)}</ul> : <p className="mt-3 text-sm text-slate-500">Ainda não há visitas registradas.</p>}
          <div className="mt-4 border-t border-slate-200 pt-3 text-xs text-slate-500">
            WhatsApp: <strong>{resumo.cliques_whatsapp}</strong> · Telefone: <strong>{resumo.cliques_telefone}</strong>
          </div>
        </div>
      </div>
    </section>
  );
}
