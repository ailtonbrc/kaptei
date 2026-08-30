import { useCallback, useEffect, useState, type FormEvent } from 'react';
import { CheckCircle2, Copy, ExternalLink, Loader2, RefreshCw, RotateCcw, Save, TriangleAlert } from 'lucide-react';
import { toast } from 'sonner';
import { obterMensagemErro } from '../../../lib/http/erro-api';
import { portalImobiliarioService } from '../../../services/portalImobiliarioService';
import type { ConfiguracaoPortal, DiagnosticoFeedPortal, PublicacaoPortal } from '../../../types/portalImobiliario';

export const ConfiguracaoPortais = () => {
  const [configuracao, setConfiguracao] = useState<ConfiguracaoPortal>();
  const [publicacoes, setPublicacoes] = useState<PublicacaoPortal[]>([]);
  const [diagnostico, setDiagnostico] = useState<DiagnosticoFeedPortal>();
  const [urlFeed, setURLFeed] = useState('');
  const [carregando, setCarregando] = useState(true);
  const [urlWebhook, setURLWebhook] = useState('');
  const [processando, setProcessando] = useState(false);
  const [imovelProcessando, setImovelProcessando] = useState('');

  const carregar = useCallback(async () => {
    setCarregando(true);
    try {
      const [configuracaoAtual, publicacoesAtuais, diagnosticoAtual] = await Promise.all([
        portalImobiliarioService.obterConfiguracao(),
        portalImobiliarioService.listarPublicacoes(),
        portalImobiliarioService.diagnosticar(),
      ]);
      setConfiguracao(configuracaoAtual);
      setPublicacoes(publicacoesAtuais);
      setDiagnostico(diagnosticoAtual);
    } catch (erro: unknown) {
      toast.error(obterMensagemErro(erro, 'Não foi possível carregar a integração com portais.'));
    } finally { setCarregando(false); }
  }, []);

  useEffect(() => {
    let ativo = true;
    Promise.all([
      portalImobiliarioService.obterConfiguracao(),
      portalImobiliarioService.listarPublicacoes(),
      portalImobiliarioService.diagnosticar(),
    ]).then(([configuracaoAtual, publicacoesAtuais, diagnosticoAtual]) => {
      if (!ativo) return;
      setConfiguracao(configuracaoAtual); setPublicacoes(publicacoesAtuais); setDiagnostico(diagnosticoAtual);
    }).catch((erro: unknown) => { if (ativo) toast.error(obterMensagemErro(erro, 'Não foi possível carregar a integração com portais.')); })
      .finally(() => { if (ativo) setCarregando(false); });
    return () => { ativo = false; };
  }, []);

  const salvarConfiguracao = async (evento: FormEvent) => {
    evento.preventDefault();
    if (!configuracao) return;
    setProcessando(true);
    try {
      setConfiguracao(await portalImobiliarioService.salvarConfiguracao(configuracao));
      toast.success('Configuração do Grupo OLX salva.');
      await carregar();
    } catch (erro: unknown) { toast.error(obterMensagemErro(erro, 'Não foi possível salvar a configuração.')); }
    finally { setProcessando(false); }
  };

  const gerarURL = async () => {
    if (configuracao?.token_feed_prefixo && !window.confirm('A URL atual deixará de funcionar imediatamente. Deseja rotacionar?')) return;
    setProcessando(true);
    try {
      const credencial = await portalImobiliarioService.rotacionarToken();
      setURLFeed(credencial.url_feed);
      setURLWebhook(credencial.url_webhook);
      toast.success('Nova URL gerada. Copie-a agora; o token não será exibido novamente.');
      await carregar();
    } catch (erro: unknown) { toast.error(obterMensagemErro(erro, 'Não foi possível gerar a URL do feed.')); }
    finally { setProcessando(false); }
  };

  const atualizarPublicacao = async (publicacao: PublicacaoPortal, alteracao: Partial<Pick<PublicacaoPortal, 'ativa' | 'tipo_publicacao'>>) => {
    setImovelProcessando(publicacao.imovel_id);
    try {
      await portalImobiliarioService.salvarPublicacao(publicacao.imovel_id, alteracao.ativa ?? publicacao.ativa, alteracao.tipo_publicacao ?? publicacao.tipo_publicacao);
      await carregar();
    } catch (erro: unknown) { toast.error(obterMensagemErro(erro, 'Não foi possível atualizar a publicação.')); }
    finally { setImovelProcessando(''); }
  };

  const copiar = async (valor: string, descricao: string) => {
    await navigator.clipboard.writeText(valor);
    toast.success(`${descricao} copiada.`);
  };

  if (carregando && !configuracao) return <section className="rounded-2xl border border-slate-200 bg-white p-6"><p className="flex items-center gap-2 text-sm text-slate-500"><Loader2 className="h-4 w-4 animate-spin" /> Carregando portais...</p></section>;
  if (!configuracao) return null;

  return <section className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
    <header className="border-b border-slate-100 px-6 py-5"><div className="flex flex-wrap items-start justify-between gap-3"><div><h2 className="flex items-center gap-2 font-semibold text-slate-900"><ExternalLink className="h-5 w-5 text-blue-600" /> Grupo OLX — ZAP Imóveis e Viva Real</h2><p className="mt-1 text-sm text-slate-500">Feed completo no formato VRSync oficial, protegido por URL rotacionável.</p></div>{configuracao.ativa ? <span className="inline-flex items-center gap-1 rounded-full bg-emerald-100 px-3 py-1 text-xs font-bold text-emerald-800"><CheckCircle2 className="h-4 w-4" /> Integração ativa</span> : <span className="rounded-full bg-slate-100 px-3 py-1 text-xs font-bold text-slate-600">Desativada</span>}</div></header>

    <form onSubmit={salvarConfiguracao} className="space-y-5 p-6">

      <div className="grid gap-4 md:grid-cols-2">
        <label className="text-sm font-medium text-slate-700">Nome de contato<input required maxLength={120} className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2" value={configuracao.nome_contato} onChange={(e) => setConfiguracao({ ...configuracao, nome_contato: e.target.value })} /></label>
        <label className="text-sm font-medium text-slate-700">E-mail de contato<input required type="email" maxLength={254} className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2" value={configuracao.email_contato} onChange={(e) => setConfiguracao({ ...configuracao, email_contato: e.target.value })} /></label>
        <label className="text-sm font-medium text-slate-700">Telefone<input maxLength={32} className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2" value={configuracao.telefone_contato} onChange={(e) => setConfiguracao({ ...configuracao, telefone_contato: e.target.value })} /></label>
        <label className="text-sm font-medium text-slate-700">Endereço exibido<select className="mt-1 w-full rounded-lg border border-slate-300 bg-white px-3 py-2" value={configuracao.exibicao_endereco} onChange={(e) => setConfiguracao({ ...configuracao, exibicao_endereco: e.target.value as ConfiguracaoPortal['exibicao_endereco'] })}><option value="BAIRRO">Somente bairro</option><option value="LOGRADOURO">Bairro e logradouro</option><option value="COMPLETO">Endereço completo</option></select></label>
      </div>
      <label className="flex items-start gap-3 rounded-xl border border-slate-200 bg-slate-50 p-4"><input type="checkbox" className="mt-1" checked={configuracao.ativa} onChange={(e) => setConfiguracao({ ...configuracao, ativa: e.target.checked })} /><span><strong className="text-sm text-slate-900">Ativar feed</strong><span className="mt-1 block text-xs text-slate-500">A ativação só será aceita com token, site publicado e todos os imóveis selecionados válidos.</span></span></label>
      <div className="flex flex-wrap items-center justify-between gap-3 border-t border-slate-100 pt-5"><div><p className="text-sm font-medium text-slate-700">URL protegida {configuracao.token_feed_prefixo && <code className="ml-1 text-xs">{configuracao.token_feed_prefixo}…</code>}</p><p className="text-xs text-slate-500">Cadastre esta URL uma única vez no Canal Pro como Desenvolvedor Próprio.</p></div><div className="flex gap-2"><button type="button" onClick={() => void gerarURL()} disabled={processando} className="inline-flex items-center gap-2 rounded-lg border border-slate-300 px-4 py-2 text-sm font-medium"><RotateCcw className="h-4 w-4" /> {configuracao.token_feed_prefixo ? 'Rotacionar URL' : 'Gerar URL'}</button><button disabled={processando} className="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-5 py-2 text-sm font-bold text-white"><Save className="h-4 w-4" /> Salvar</button></div></div>
      {urlFeed && <div className="space-y-2 rounded-xl border border-amber-200 bg-amber-50 p-3"><URLProtegida rotulo="Feed VRSync" valor={urlFeed} aoCopiar={copiar} /><URLProtegida rotulo="Webhook de leads" valor={urlWebhook} aoCopiar={copiar} /><p className="text-xs text-amber-800">A equipe de homologação do Grupo OLX também precisa configurar a chave Basic global fornecida ao Kaptei.</p></div>}
    </form>

    <div className="border-t border-slate-100 p-6">
      <div className="flex flex-wrap items-center justify-between gap-3"><div><h3 className="font-bold text-slate-900">Inventário enviado</h3><p className="text-sm text-slate-500">O feed nunca entrega uma carga parcial: qualquer erro o torna temporariamente indisponível.</p></div><button type="button" onClick={() => void carregar()} disabled={carregando} className="inline-flex items-center gap-2 rounded-lg border border-slate-300 px-3 py-2 text-sm font-medium"><RefreshCw className={`h-4 w-4 ${carregando ? 'animate-spin' : ''}`} /> Validar</button></div>
      {diagnostico && <div className={`mt-4 rounded-xl border p-4 ${diagnostico.valido ? 'border-emerald-200 bg-emerald-50' : 'border-amber-200 bg-amber-50'}`}><p className="flex items-center gap-2 text-sm font-bold">{diagnostico.valido ? <CheckCircle2 className="h-4 w-4 text-emerald-700" /> : <TriangleAlert className="h-4 w-4 text-amber-700" />}{diagnostico.total_valido} de {diagnostico.total_selecionado} selecionado(s) válidos</p>{diagnostico.erros_gerais.map((erro) => <p key={erro} className="mt-1 text-xs text-amber-800">{erro}</p>)}</div>}
      <div className="mt-4 overflow-x-auto"><table className="w-full min-w-[760px] text-left text-sm"><thead className="bg-slate-50 text-xs uppercase text-slate-500"><tr><th className="px-3 py-2">Enviar</th><th className="px-3 py-2">Imóvel</th><th className="px-3 py-2">Situação</th><th className="px-3 py-2">Publicação</th><th className="px-3 py-2">Validação</th></tr></thead><tbody>{publicacoes.map((publicacao) => <tr key={publicacao.imovel_id} className="border-t border-slate-100"><td className="px-3 py-3"><input aria-label={`Publicar ${publicacao.titulo}`} type="checkbox" checked={publicacao.ativa} disabled={imovelProcessando === publicacao.imovel_id} onChange={(e) => void atualizarPublicacao(publicacao, { ativa: e.target.checked })} /></td><td className="px-3 py-3"><p className="font-semibold text-slate-900">{publicacao.titulo}</p><p className="text-xs text-slate-500">{publicacao.tipo} · {publicacao.finalidade}</p></td><td className="px-3 py-3">{publicacao.status}</td><td className="px-3 py-3"><select aria-label={`Tipo de publicação de ${publicacao.titulo}`} className="rounded-lg border border-slate-300 bg-white px-2 py-1.5" value={publicacao.tipo_publicacao} disabled={imovelProcessando === publicacao.imovel_id} onChange={(e) => void atualizarPublicacao(publicacao, { tipo_publicacao: e.target.value as PublicacaoPortal['tipo_publicacao'] })}><option value="STANDARD">Padrão</option><option value="PREMIUM">Destaque</option><option value="SUPER_PREMIUM">Super destaque</option></select></td><td className="px-3 py-3">{imovelProcessando === publicacao.imovel_id ? <Loader2 className="h-4 w-4 animate-spin" /> : publicacao.erros.length ? <ul className="max-w-xs list-disc pl-4 text-xs text-amber-700">{publicacao.erros.map((erro) => <li key={erro}>{erro}</li>)}</ul> : publicacao.ativa ? <span className="text-xs font-bold text-emerald-700">Pronto</span> : <span className="text-xs text-slate-400">Não selecionado</span>}</td></tr>)}</tbody></table>{!publicacoes.length && <p className="p-6 text-center text-sm text-slate-500">Cadastre imóveis para configurar o inventário.</p>}</div>
    </div>
  </section>;
};

const URLProtegida = ({ rotulo, valor, aoCopiar }: { rotulo: string; valor: string; aoCopiar: (valor: string, descricao: string) => Promise<void> }) => <div className="flex items-center gap-2"><span className="w-32 shrink-0 text-xs font-bold text-amber-900">{rotulo}</span><code className="min-w-0 flex-1 break-all text-xs text-amber-950">{valor}</code><button type="button" aria-label={`Copiar ${rotulo}`} onClick={() => void aoCopiar(valor, rotulo)} className="rounded-lg p-2 text-amber-800 hover:bg-amber-100"><Copy className="h-4 w-4" /></button></div>;
