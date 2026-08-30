import axios from 'axios';

interface CorpoErroAPI {
  erro?: string;
  mensagem?: string;
  message?: string;
}

export const obterMensagemErro = (erro: unknown, mensagemPadrao: string): string => {
  if (!axios.isAxiosError(erro)) return mensagemPadrao;
  const dados = erro.response?.data;
  if (typeof dados === 'string' && dados.trim()) return dados;
  if (dados && typeof dados === 'object') {
    const corpo = dados as CorpoErroAPI;
    return corpo.mensagem ?? corpo.erro ?? corpo.message ?? mensagemPadrao;
  }
  return erro.message || mensagemPadrao;
};
