import { useCallback, useEffect, useState } from 'react';
import { MessageCircleMore, RefreshCw } from 'lucide-react';
import { Button } from '../../components/ui/button';
import { whatsappService } from '../../services/whatsappService';
import type { ConversaWhatsApp, MensagemWhatsApp } from '../../types/whatsapp';
import { CompositorMensagem } from './components/CompositorMensagem';
import { GestaoConsentimento } from './components/GestaoConsentimento';
import { HistoricoMensagens } from './components/HistoricoMensagens';
import { ListaConversas } from './components/ListaConversas';

function mensagemErro(erro: unknown): string {
  const resposta = erro as { response?: { data?: { erro?: string } } };
  return resposta.response?.data?.erro || 'Não foi possível concluir a operação.';
}

export function ConversasWhatsApp() {
  const [conversas, setConversas] = useState<ConversaWhatsApp[]>([]);
  const [selecionada, setSelecionada] = useState<ConversaWhatsApp>();
  const [mensagens, setMensagens] = useState<MensagemWhatsApp[]>([]);
  const [busca, setBusca] = useState('');
  const [carregandoConversas, setCarregandoConversas] = useState(true);
  const [carregandoMensagens, setCarregandoMensagens] = useState(false);
  const [processando, setProcessando] = useState(false);
  const [erro, setErro] = useState('');
  const [janelaAberta, setJanelaAberta] = useState(false);

  const carregarMensagens = useCallback(async (conversaID: string) => {
    setCarregandoMensagens(true);
    try {
      const resultado = await whatsappService.listarMensagens(conversaID);
      setMensagens(resultado.dados);
      setErro('');
    } catch (falha) {
      setMensagens([]);
      setErro(mensagemErro(falha));
    } finally {
      setCarregandoMensagens(false);
    }
  }, []);

  const selecionarConversa = useCallback((conversa: ConversaWhatsApp) => {
    setSelecionada(conversa);
    setJanelaAberta(Boolean(conversa.janela_atendimento_ate && new Date(conversa.janela_atendimento_ate).getTime() > Date.now()));
    void carregarMensagens(conversa.id);
  }, [carregarMensagens]);

  const carregarConversas = useCallback(async (termo = '') => {
    setCarregandoConversas(true);
    try {
      const resultado = await whatsappService.listarConversas(1, termo);
      setConversas(resultado.dados);
      const proxima = resultado.dados.find((item) => item.id === selecionada?.id) || resultado.dados[0];
      if (proxima) selecionarConversa(proxima);
      else {
        setSelecionada(undefined);
        setMensagens([]);
        setJanelaAberta(false);
      }
      setErro('');
    } catch (falha) {
      setErro(mensagemErro(falha));
    } finally {
      setCarregandoConversas(false);
    }
  }, [selecionada?.id, selecionarConversa]);

  useEffect(() => {
    const temporizador = window.setTimeout(() => void carregarConversas(busca), 350);
    return () => window.clearTimeout(temporizador);
  }, [busca, carregarConversas]);

  async function executar(acao: () => Promise<void>): Promise<boolean> {
    setProcessando(true);
    try {
      await acao();
      if (selecionada) await carregarMensagens(selecionada.id);
      await carregarConversas(busca);
      setErro('');
      return true;
    } catch (falha) {
      setErro(mensagemErro(falha));
      return false;
    } finally {
      setProcessando(false);
    }
  }

  return (
    <div className="flex h-full min-h-0 flex-col bg-slate-50">
      <header className="flex flex-wrap items-center justify-between gap-3 border-b border-slate-200 bg-white px-4 py-3 sm:px-6">
        <div>
          <h1 className="flex items-center gap-2 text-xl font-bold text-slate-900"><MessageCircleMore className="h-6 w-6 text-emerald-600" />WhatsApp</h1>
          <p className="text-xs text-slate-500">Conversas recebidas e envios pela API oficial.</p>
        </div>
        <Button variant="outline" size="sm" onClick={() => void carregarConversas(busca)}><RefreshCw className="h-4 w-4" />Atualizar</Button>
      </header>
      {erro && <div role="alert" className="border-b border-rose-200 bg-rose-50 px-4 py-2 text-sm text-rose-700">{erro}</div>}
      <div className="grid min-h-0 flex-1 grid-cols-1 lg:grid-cols-[20rem_minmax(0,1fr)]">
        <ListaConversas conversas={conversas} conversaAtiva={selecionada?.id} busca={busca} carregando={carregandoConversas} aoBuscar={setBusca} aoSelecionar={selecionarConversa} />
        {!selecionada ? (
          <div className="m-auto p-8 text-center text-slate-500"><MessageCircleMore className="mx-auto mb-3 h-12 w-12 text-slate-300" /><p>Selecione uma conversa para iniciar o atendimento.</p></div>
        ) : (
          <section className="flex min-h-0 flex-col">
            <div className="flex flex-wrap items-center justify-between gap-3 border-b border-slate-200 bg-white px-4 py-3">
              <div><h2 className="font-semibold text-slate-900">{selecionada.nome_contato || `+${selecionada.numero_contato}`}</h2><p className="text-xs text-slate-500">+{selecionada.numero_contato}</p></div>
              <GestaoConsentimento consentimento={selecionada.consentimento_marketing} salvando={processando} aoSalvar={(consentiu, origem, evidencia) => executar(() => whatsappService.registrarConsentimento(selecionada.id, consentiu, origem, evidencia))} />
            </div>
            <div className="min-h-0 flex-1 overflow-y-auto bg-slate-100/70"><HistoricoMensagens mensagens={mensagens} carregando={carregandoMensagens} /></div>
            <CompositorMensagem janelaAberta={janelaAberta} consentimento={selecionada.consentimento_marketing} enviando={processando} aoEnviarTexto={(texto) => executar(() => whatsappService.enviarTexto(selecionada.id, texto))} aoEnviarTemplate={(nome, idioma, parametros) => executar(() => whatsappService.enviarTemplate(selecionada.id, { nome, idioma, parametros }))} />
          </section>
        )}
      </div>
    </div>
  );
}
