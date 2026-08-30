import React from 'react';
import { CalendarCheck, FileText, UserPlus, Users, RefreshCw } from 'lucide-react';
import { IndicadorDashboard } from '../../components/IndicadorDashboard/IndicadorDashboard';
import { Button } from '../../components/ui/button';
import { GraficoDashboard } from './GraficoDashboard';
import { useDashboard } from './useDashboard';
import { PainelConversaoSite } from './PainelConversaoSite';
import type { EChartsOption } from 'echarts';

export const Dashboard: React.FC = () => {
  const { dados, carregando, erro, recarregar } = useDashboard();

  if (carregando) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center" role="status" aria-live="polite">
        <RefreshCw className="h-8 w-8 animate-spin text-blue-600" aria-hidden="true" />
        <span className="sr-only">Carregando dashboard</span>
      </div>
    );
  }

  if (erro || !dados) {
    return (
      <div className="m-4 rounded-xl border border-red-200 bg-red-50 p-6 text-red-900" role="alert">
        <h1 className="text-lg font-semibold">Não foi possível carregar o dashboard</h1>
        <p className="mt-1 text-sm">Verifique sua conexão e tente novamente.</p>
        <Button className="mt-4" variant="outline" onClick={recarregar}>
          <RefreshCw className="h-4 w-4" aria-hidden="true" />
          Tentar novamente
        </Button>
      </div>
    );
  }

  const metricas = dados.metricas || {};
  const funil = dados.funil || {};
  const origens = dados.origens || {};
  const leadsEvolucao = dados.leads_evolucao || {};

  const origensData = Object.keys(origens).map(key => ({ name: key, value: origens[key] }));
  const totalOrigens = origensData.reduce((acc, curr) => acc + curr.value, 0);

  const chartOptionEvolucao: EChartsOption = {
    tooltip: { 
      trigger: 'axis',
      backgroundColor: 'rgba(255, 255, 255, 0.95)',
      borderColor: '#e2e8f0',
      textStyle: { color: '#1e293b' },
      extraCssText: 'box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1); border-radius: 8px;'
    },
    grid: { left: '3%', right: '4%', bottom: '8%', top: '10%', containLabel: true },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: leadsEvolucao.categorias || [],
      axisLine: { lineStyle: { color: '#e2e8f0' } },
      axisLabel: { color: '#64748b', margin: 16 }
    },
    yAxis: {
      type: 'value',
      axisLabel: { color: '#64748b' },
      splitLine: { lineStyle: { type: 'dashed', color: '#f1f5f9' } }
    },
    series: [
      {
        name: 'Leads',
        type: 'line',
        smooth: true,
        symbol: 'none',
        lineStyle: { color: '#10b981', width: 3 }, // Verde vibrante da imagem
        itemStyle: { color: '#10b981' },
        areaStyle: {
          color: {
            type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(16, 185, 129, 0.3)' },
              { offset: 1, color: 'rgba(16, 185, 129, 0.02)' }
            ]
          }
        },
        data: leadsEvolucao.valores || []
      }
    ]
  };

  const chartOptionOrigens: EChartsOption = {
    tooltip: { 
      trigger: 'item',
      backgroundColor: 'rgba(255, 255, 255, 0.95)',
      borderColor: '#e2e8f0',
      textStyle: { color: '#1e293b' },
      extraCssText: 'box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1); border-radius: 8px;'
    },
    legend: { 
      bottom: '0%', 
      icon: 'circle', 
      itemWidth: 10, 
      itemHeight: 10, 
      textStyle: { color: '#64748b', fontSize: 12 },
      itemGap: 15
    },
    series: [
      {
        name: 'Origem',
        type: 'pie',
        radius: ['60%', '85%'],
        center: ['50%', '42%'],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 6,
          borderColor: '#fff',
          borderWidth: 3
        },
        label: {
          show: true,
          position: 'center',
          formatter: () => `{total|${totalOrigens}}\n{text|Total Leads}`,
          rich: {
            total: { fontSize: 36, fontWeight: 'bold', color: '#0f172a' },
            text: { fontSize: 14, color: '#64748b', padding: [4, 0, 0, 0] }
          }
        },
        emphasis: { label: { show: true } },
        labelLine: { show: false },
        data: origensData,
        color: ['#0ea5e9', '#10b981', '#f59e0b', '#8b5cf6', '#64748b', '#f43f5e']
      }
    ]
  };

  const funilCategories = Object.keys(funil).reverse();
  const funilValues = funilCategories.map(key => funil[key]);

  const chartOptionFunil: EChartsOption = {
    tooltip: { 
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      backgroundColor: 'rgba(255, 255, 255, 0.95)',
      borderColor: '#e2e8f0',
      textStyle: { color: '#1e293b' },
      extraCssText: 'box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1); border-radius: 8px;'
    },
    grid: { left: '3%', right: '4%', bottom: '3%', top: '5%', containLabel: true },
    xAxis: {
      type: 'value',
      axisLabel: { color: '#64748b' },
      splitLine: { lineStyle: { type: 'dashed', color: '#f1f5f9' } },
      axisLine: { show: false },
      axisTick: { show: false }
    },
    yAxis: {
      type: 'category',
      data: funilCategories,
      axisLabel: { color: '#64748b', fontWeight: '500' },
      axisLine: { lineStyle: { color: '#e2e8f0' } },
      axisTick: { show: false }
    },
    series: [
      {
        name: 'Leads',
        type: 'bar',
        barWidth: '45%',
        data: funilValues,
        itemStyle: {
          color: '#0ea5e9', // Azul claro clássico de barra horizontal
          borderRadius: [0, 6, 6, 0] // Bordas arredondadas na direita
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
    <div className="space-y-6 bg-slate-50 min-h-full p-4 sm:p-6 lg:p-8">
      <div>
        <h1 className="text-2xl font-bold tracking-tight text-slate-900">Dashboard Analítico</h1>
        <p className="text-slate-500">Visão macro de movimentações e captações no sistema.</p>
      </div>
      
	  <section className="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
		<IndicadorDashboard rotulo="Total de imóveis ativos" valor={metricas.total_imoveis || 0} icone={FileText} cor="blue" />
		<IndicadorDashboard rotulo="Clientes no CRM" valor={metricas.total_clientes || 0} icone={Users} cor="emerald" />
		<IndicadorDashboard rotulo="Leads em 30 dias" valor={metricas.leads_30_dias || 0} icone={UserPlus} cor="amber" />
		<IndicadorDashboard rotulo="Visitas futuras" valor={metricas.visitas_pendentes || 0} icone={CalendarCheck} cor="violet" />
	  </section>
      {dados.conversao_site && <PainelConversaoSite resumo={dados.conversao_site} />}


      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mt-6">
		<section className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
          <h3 className="text-lg font-bold text-slate-900 mb-4">Funil de Conversão (Leads)</h3>
		  <GraficoDashboard
            opcao={chartOptionFunil}
            rotuloAcessivel="Gráfico do funil de conversão de leads"
            resumoAcessivel={funilCategories.map((name, i) => `${name}: ${funilValues[i]}`).join('; ') || 'Sem dados no período.'}
            style={{ height: '350px', width: '100%' }}
          />
		</section>

		<section className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
          <h3 className="text-lg font-bold text-slate-900 mb-4">Origem dos Leads</h3>
		  <GraficoDashboard
            opcao={chartOptionOrigens}
            rotuloAcessivel="Gráfico das origens dos leads"
            resumoAcessivel={origensData.map((item) => `${item.name}: ${item.value}`).join('; ') || 'Sem dados no período.'}
            style={{ height: '350px', width: '100%' }}
          />
		</section>
      </div>

	  <section className="mt-6 rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
        <h3 className="text-lg font-bold text-slate-900 mb-4">Evolução da captação de leads</h3>
		<GraficoDashboard
          opcao={chartOptionEvolucao}
          rotuloAcessivel="Gráfico da evolução da captação de leads"
          resumoAcessivel={(leadsEvolucao.categorias || []).map((categoria, indice) => `${categoria}: ${(leadsEvolucao.valores || [])[indice] ?? 0}`).join('; ') || 'Sem dados no período.'}
          style={{ height: '380px', width: '100%' }}
        />
	  </section>
    </div>
  );
};
