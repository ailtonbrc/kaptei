import { useState } from 'react';
import { ExternalLink, Globe2, Loader2, Save } from 'lucide-react';
import { toast } from 'sonner';
import { obterMensagemErro } from '../../../lib/http/erro-api';
import { siteAdminService, type AtualizacaoSitePublico } from '../../../services/siteAdminService';
import { ConfiguracaoDominio } from './ConfiguracaoDominio';

interface ConfiguracaoSiteProps { inicial: AtualizacaoSitePublico }

export const ConfiguracaoSite = ({ inicial }: ConfiguracaoSiteProps) => {
  const [dados, setDados] = useState(inicial);
  const [salvando, setSalvando] = useState(false);
  const atualizarConfiguracao = (campo: string, valor: string) => setDados((atual) => ({ ...atual, configuracao: { ...atual.configuracao, [campo]: valor } }));
  const inputClass = 'w-full rounded-xl border border-slate-200 px-4 py-2.5 text-sm outline-none focus:border-blue-600 focus:ring-2 focus:ring-blue-600/15';

  const salvar = async (evento: React.FormEvent) => {
    evento.preventDefault(); setSalvando(true);
    try { await siteAdminService.salvar(dados); toast.success('Site público atualizado.'); }
    catch (erro: unknown) { toast.error(obterMensagemErro(erro, 'Não foi possível atualizar o site.')); }
    finally { setSalvando(false); }
  };

  return (
    <>
      <form onSubmit={salvar} className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-4"><div><h2 className="flex items-center gap-2 text-lg font-bold"><Globe2 className="h-5 w-5 text-blue-600" /> Site e catálogo público</h2><p className="mt-1 text-sm text-slate-500">Configure a presença digital usada para captar clientes.</p></div>{dados.publicado && dados.slug && <a href={`/s/${dados.slug}`} target="_blank" rel="noreferrer" className="inline-flex items-center gap-2 text-sm font-bold text-blue-700">Visualizar site <ExternalLink className="h-4 w-4" /></a>}</div>
      <div className="grid gap-5 md:grid-cols-2">
        <label className="text-sm font-semibold text-slate-700">Endereço público<span className="mt-1 flex overflow-hidden rounded-xl border border-slate-200 bg-slate-50"><span className="px-3 py-2.5 text-slate-400">/s/</span><input required={dados.publicado} pattern="[a-z0-9]+(?:-[a-z0-9]+)*" maxLength={80} className="min-w-0 flex-1 bg-transparent py-2.5 pr-3 outline-none" value={dados.slug} onChange={(e) => setDados((atual) => ({ ...atual, slug: e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, '') }))} /></span></label>
        <label className="text-sm font-semibold text-slate-700">URL do logotipo<input type="url" className={`${inputClass} mt-1`} value={dados.configuracao.logo_url ?? ''} onChange={(e) => atualizarConfiguracao('logo_url', e.target.value)} /></label>
        <label className="text-sm font-semibold text-slate-700 md:col-span-2">Título principal<input maxLength={120} className={`${inputClass} mt-1`} value={dados.configuracao.titulo ?? ''} onChange={(e) => atualizarConfiguracao('titulo', e.target.value)} /></label>
        <label className="text-sm font-semibold text-slate-700 md:col-span-2">Subtítulo<input maxLength={220} className={`${inputClass} mt-1`} value={dados.configuracao.subtitulo ?? ''} onChange={(e) => atualizarConfiguracao('subtitulo', e.target.value)} /></label>
        <label className="text-sm font-semibold text-slate-700 md:col-span-2">Apresentação<textarea rows={4} maxLength={1000} className={`${inputClass} mt-1`} value={dados.configuracao.descricao ?? ''} onChange={(e) => atualizarConfiguracao('descricao', e.target.value)} /></label>
        <label className="text-sm font-semibold text-slate-700">Telefone<input className={`${inputClass} mt-1`} value={dados.configuracao.telefone ?? ''} onChange={(e) => atualizarConfiguracao('telefone', e.target.value)} /></label>
        <label className="text-sm font-semibold text-slate-700">WhatsApp<input className={`${inputClass} mt-1`} value={dados.configuracao.whatsapp ?? ''} onChange={(e) => atualizarConfiguracao('whatsapp', e.target.value)} /></label>
        <label className="text-sm font-semibold text-slate-700">E-mail público e de privacidade<input required={dados.publicado} type="email" maxLength={254} className={`${inputClass} mt-1`} value={dados.configuracao.email ?? ''} onChange={(e) => atualizarConfiguracao('email', e.target.value)} /></label>
        <label className="text-sm font-semibold text-slate-700">CRECI<input className={`${inputClass} mt-1`} value={dados.configuracao.creci ?? ''} onChange={(e) => atualizarConfiguracao('creci', e.target.value)} /></label>
      </div>
      <label className="flex items-start gap-3 rounded-xl border border-slate-200 bg-slate-50 p-4"><input type="checkbox" className="mt-1" checked={dados.publicado} onChange={(e) => setDados((atual) => ({ ...atual, publicado: e.target.checked }))} /><span><strong className="block text-sm text-slate-800">Publicar site</strong><span className="text-xs text-slate-500">Ao ativar, o catálogo ficará acessível na internet.</span></span></label>
      <button disabled={salvando} className="inline-flex items-center gap-2 rounded-xl bg-blue-600 px-5 py-2.5 font-bold text-white hover:bg-blue-700 disabled:opacity-60">{salvando ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />} Salvar site</button>
      </form>
      <ConfiguracaoDominio />
    </>
  );
};
