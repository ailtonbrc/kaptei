import { api } from './api';
import type {
  BloqueioRetencao,
  NovaSolicitacaoTitular,
  NovoBloqueioRetencao,
  PaginaSolicitacoesTitular,
  PoliticaRetencao,
  RelatorioRetencao,
  ResultadoRetencao,
  SolicitacaoTitular,
} from '../types/privacidade';

interface FiltrosSolicitacao { pagina?: number; limite?: number; busca?: string; status?: string; tipo?: string }

export const privacidadeService = {
  async criarPublica(slug: string, dados: NovaSolicitacaoTitular): Promise<{ protocolo: string; mensagem: string }> {
    const resposta = await api.post(`/public/sites/${encodeURIComponent(slug)}/privacidade/solicitacoes`, dados);
    return resposta.data;
  },
  async listar(filtros: FiltrosSolicitacao = {}): Promise<PaginaSolicitacoesTitular> {
    const resposta = await api.get<PaginaSolicitacoesTitular>('/v1/privacidade/solicitacoes', { params: filtros });
    return resposta.data;
  },
  async obter(id: string): Promise<SolicitacaoTitular> {
    const resposta = await api.get<SolicitacaoTitular>(`/v1/privacidade/solicitacoes/${id}`);
    return resposta.data;
  },
  async verificar(id: string, metodo: string, evidencia: string): Promise<void> {
    await api.post(`/v1/privacidade/solicitacoes/${id}/verificacao`, { metodo, evidencia });
  },
  async decidir(id: string, aprovada: boolean, fundamentoLegal: string, observacao: string): Promise<void> {
    await api.post(`/v1/privacidade/solicitacoes/${id}/decisao`, {
      aprovada, fundamento_legal: fundamentoLegal, observacao,
    });
  },
  async exportar(id: string, protocolo: string): Promise<void> {
    const resposta = await api.post(`/v1/privacidade/solicitacoes/${id}/exportacao`, undefined, { responseType: 'blob' });
    const url = URL.createObjectURL(resposta.data);
    const ancora = document.createElement('a');
    ancora.href = url;
    ancora.download = `kaptei-dados-${protocolo}.json`;
    ancora.click();
    URL.revokeObjectURL(url);
  },
  async executar(id: string): Promise<void> {
    await api.post(`/v1/privacidade/solicitacoes/${id}/execucao`);
  },
  async obterPoliticaRetencao(): Promise<PoliticaRetencao> {
    const resposta = await api.get<PoliticaRetencao>('/v1/privacidade/retencao/politica');
    return resposta.data;
  },
  async salvarPoliticaRetencao(politica: PoliticaRetencao): Promise<void> {
    await api.put('/v1/privacidade/retencao/politica', politica);
  },
  async obterRelatorioRetencao(): Promise<RelatorioRetencao> {
    const resposta = await api.get<RelatorioRetencao>('/v1/privacidade/retencao/relatorio');
    return resposta.data;
  },
  async executarRetencao(confirmacao: string): Promise<ResultadoRetencao> {
    const resposta = await api.post<ResultadoRetencao>('/v1/privacidade/retencao/execucao', { confirmacao });
    return resposta.data;
  },
  async listarBloqueiosRetencao(): Promise<BloqueioRetencao[]> {
    const resposta = await api.get<BloqueioRetencao[]>('/v1/privacidade/retencao/bloqueios');
    return resposta.data;
  },
  async salvarBloqueioRetencao(bloqueio: NovoBloqueioRetencao): Promise<BloqueioRetencao> {
    const resposta = await api.post<BloqueioRetencao>('/v1/privacidade/retencao/bloqueios', bloqueio);
    return resposta.data;
  },
  async removerBloqueioRetencao(id: string): Promise<void> {
    await api.delete(`/v1/privacidade/retencao/bloqueios/${id}`);
  },
};

