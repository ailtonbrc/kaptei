export interface ResumoConversaoSite {
  visitas_site: number;
  imoveis_visualizados: number;
  formularios_iniciados: number;
  contatos_enviados: number;
  cliques_whatsapp: number;
  cliques_telefone: number;
  taxa_conversao: number;
  fontes: Record<string, number>;
}

export interface DadosDashboard {
  metricas?: {
    total_imoveis?: number;
    total_clientes?: number;
    leads_30_dias?: number;
    visitas_pendentes?: number;
  };
  funil?: Record<string, number>;
  origens?: Record<string, number>;
  leads_evolucao?: {
    categorias?: string[];
    valores?: number[];
  };
  conversao_site?: ResumoConversaoSite;
}
