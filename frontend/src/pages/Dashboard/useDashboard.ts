import { useCallback, useEffect, useState } from 'react';
import { obterDashboardPremium } from '../../services/dashboardService';
import type { DadosDashboard } from '../../types/dashboard';

interface EstadoDashboard {
  dados: DadosDashboard | null;
  carregando: boolean;
  erro: boolean;
}

export function useDashboard() {
  const [tentativa, setTentativa] = useState(0);
  const [estado, setEstado] = useState<EstadoDashboard>({ dados: null, carregando: true, erro: false });

  useEffect(() => {
    const controlador = new AbortController();
    const dadosMock: DadosDashboard = {
      metricas: {
        total_imoveis: 42,
        total_clientes: 156,
        leads_30_dias: 38,
        visitas_pendentes: 5,
      },
      funil: {
        'Novo Lead': 25,
        'Em Contato': 15,
        'Agendou Visita': 8,
        'Em Negociação': 4,
        'Fechado': 2,
      },
      origens: {
        'Site': 15,
        'WhatsApp': 12,
        'Instagram': 8,
        'Indicação': 3,
      },
      leads_evolucao: {
        categorias: ['01/08', '05/08', '10/08', '15/08', '20/08', '25/08', '30/08'],
        valores: [2, 5, 4, 8, 3, 10, 6],
      },
      conversao_site: {
        visitas_site: 1250,
        imoveis_visualizados: 850,
        formularios_iniciados: 45,
        contatos_enviados: 12,
        cliques_whatsapp: 30,
        cliques_telefone: 5,
        taxa_conversao: 3.76,
        fontes: {
          'Google': 850,
          'Direto': 250,
          'Instagram': 150,
        },
      },
    };

    setTimeout(() => {
      if (!controlador.signal.aborted) {
        setEstado({ dados: dadosMock, carregando: false, erro: false });
      }
    }, 500);

    return () => controlador.abort();
  }, [tentativa]);

  const recarregar = useCallback(() => {
    setEstado((atual) => ({ ...atual, carregando: true, erro: false }));
    setTentativa((valor) => valor + 1);
  }, []);
  return { ...estado, recarregar };
}
