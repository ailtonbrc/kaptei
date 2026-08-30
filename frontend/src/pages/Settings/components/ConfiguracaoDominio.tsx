import { useEffect, useState } from 'react';
import { CheckCircle2, Copy, GlobeLock, Loader2, RefreshCw } from 'lucide-react';
import { toast } from 'sonner';
import { obterMensagemErro } from '../../../lib/http/erro-api';
import { siteAdminService } from '../../../services/siteAdminService';
import type { DominioSite } from '../../../types/dominioSite';

export const ConfiguracaoDominio = () => {
  const [dominio, setDominio] = useState<DominioSite>();
  const [hostname, setHostname] = useState('');
  const [carregando, setCarregando] = useState(true);
  const [processando, setProcessando] = useState(false);

  useEffect(() => {
    let ativo = true;
    siteAdminService.obterDominio().then((resultado) => {
      if (!ativo) return;
      setDominio(resultado); setHostname(resultado?.hostname ?? '');
    }).catch(() => undefined).finally(() => { if (ativo) setCarregando(false); });
    return () => { ativo = false; };
  }, []);

  const configurar = async () => {
    setProcessando(true);
    try { const resultado = await siteAdminService.configurarDominio(hostname); setDominio(resultado); toast.success('Domínio salvo. Configure o registro TXT para comprovar a posse.'); }
    catch (erro: unknown) { toast.error(obterMensagemErro(erro, 'Não foi possível configurar o domínio.')); }
    finally { setProcessando(false); }
  };
  const verificar = async () => {
    setProcessando(true);
    try { const resultado = await siteAdminService.verificarDominio(); setDominio(resultado); toast.success('Domínio verificado e ativado.'); }
    catch (erro: unknown) { toast.error(obterMensagemErro(erro, 'O registro TXT ainda não foi localizado.')); }
    finally { setProcessando(false); }
  };
  const copiar = async (valor: string) => { await navigator.clipboard.writeText(valor); toast.success('Valor copiado.'); };

  if (carregando) return <div className="flex items-center gap-2 text-sm text-slate-500"><Loader2 className="h-4 w-4 animate-spin" /> Carregando domínio...</div>;
  return <section className="mt-8 space-y-5 border-t border-slate-200 pt-7">
    <div><h3 className="flex items-center gap-2 text-base font-bold text-slate-900"><GlobeLock className="h-5 w-5 text-blue-600" /> Domínio próprio</h3><p className="mt-1 text-sm text-slate-500">Use um domínio que você controla. A ativação exige comprovação por DNS.</p></div>
    <div className="flex flex-col gap-3 sm:flex-row"><input aria-label="Domínio próprio" placeholder="imoveis.suaempresa.com.br" maxLength={253} className="min-w-0 flex-1 rounded-xl border border-slate-200 px-4 py-2.5 text-sm outline-none focus:border-blue-600 focus:ring-2 focus:ring-blue-600/15" value={hostname} onChange={(e) => setHostname(e.target.value.toLowerCase().trim())} /><button type="button" disabled={processando || !hostname} onClick={() => void configurar()} className="rounded-xl bg-slate-900 px-5 py-2.5 text-sm font-bold text-white hover:bg-slate-800 disabled:opacity-50">Salvar domínio</button></div>
    {dominio && <div className="space-y-4 rounded-xl border border-slate-200 bg-slate-50 p-5">
      <div className="flex flex-wrap items-center justify-between gap-3"><div><p className="font-bold text-slate-900">{dominio.hostname}</p><p className="text-xs text-slate-500">Status: {dominio.status}</p></div>{dominio.status === 'ATIVO' && <span className="inline-flex items-center gap-1 rounded-full bg-emerald-100 px-3 py-1 text-xs font-bold text-emerald-800"><CheckCircle2 className="h-4 w-4" /> Ativo</span>}</div>
      {dominio.status !== 'ATIVO' && <><p className="text-sm text-slate-600">Crie um registro <strong>TXT</strong> no seu provedor DNS:</p><div className="grid gap-3"><CampoDNS rotulo="Nome" valor={dominio.registro_txt_nome} aoCopiar={copiar} /><CampoDNS rotulo="Valor" valor={dominio.registro_txt_valor} aoCopiar={copiar} /></div>{dominio.ultimo_erro && <p className="text-sm text-amber-700">{dominio.ultimo_erro}</p>}<button type="button" disabled={processando} onClick={() => void verificar()} className="inline-flex items-center gap-2 rounded-xl bg-blue-600 px-5 py-2.5 text-sm font-bold text-white hover:bg-blue-700 disabled:opacity-50">{processando ? <Loader2 className="h-4 w-4 animate-spin" /> : <RefreshCw className="h-4 w-4" />} Verificar DNS</button></>}
    </div>}
  </section>;
};

const CampoDNS = ({ rotulo, valor, aoCopiar }: { rotulo: string; valor: string; aoCopiar: (valor: string) => Promise<void> }) => <div><span className="text-xs font-bold uppercase text-slate-500">{rotulo}</span><div className="mt-1 flex items-center gap-2 rounded-lg border border-slate-200 bg-white px-3 py-2"><code className="min-w-0 flex-1 break-all text-xs text-slate-700">{valor}</code><button type="button" aria-label={`Copiar ${rotulo}`} onClick={() => void aoCopiar(valor)} className="rounded p-1 text-slate-500 hover:bg-slate-100"><Copy className="h-4 w-4" /></button></div></div>;
