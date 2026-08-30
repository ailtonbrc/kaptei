import { useRef, useState } from 'react';
import { ImagePlus, Link, Loader2, Star, Trash2, Upload } from 'lucide-react';
import { imovelService } from '../../../services/imovelService';
import type { ImovelFoto } from '../../../types/imovel';
import { obterMensagemErro } from '../../../lib/http/erro-api';

interface FotosImovelProps {
  imovelId: string;
  fotos: ImovelFoto[];
  onChange: (fotos: ImovelFoto[]) => void;
}

export const FotosImovel = ({ imovelId, fotos, onChange }: FotosImovelProps) => {
  const [url, setURL] = useState('');
  const [arquivo, setArquivo] = useState<File | null>(null);
  const [capa, setCapa] = useState(fotos.length === 0);
  const [processando, setProcessando] = useState(false);
  const [confirmando, setConfirmando] = useState<string | null>(null);
  const [erro, setErro] = useState('');
  const inputArquivo = useRef<HTMLInputElement>(null);

  const recarregar = async () => {
    const imovel = await imovelService.buscarPorId(imovelId);
    onChange(imovel.fotos ?? []);
  };

  const adicionar = async () => {
    if (!url.trim()) return;
    setProcessando(true); setErro('');
    try {
      await imovelService.adicionarFoto(imovelId, url.trim(), capa);
      await recarregar();
      setURL(''); setCapa(false);
    } catch (falha: unknown) {
      setErro(obterMensagemErro(falha, 'Não foi possível adicionar a foto.'));
    } finally {
      setProcessando(false);
    }
  };

  const enviar = async () => {
    if (!arquivo) return;
    setProcessando(true);
    setErro('');
    try {
      await imovelService.enviarFoto(imovelId, arquivo, capa);
      await recarregar();
      setArquivo(null);
      setCapa(false);
      if (inputArquivo.current) inputArquivo.current.value = '';
    } catch (falha: unknown) {
      setErro(obterMensagemErro(falha, 'Não foi possível enviar a imagem.'));
    } finally {
      setProcessando(false);
    }
  };

  const excluir = async (fotoId: string) => {
    if (confirmando !== fotoId) {
      setConfirmando(fotoId);
      return;
    }
    setProcessando(true); setErro('');
    try {
      await imovelService.excluirFoto(imovelId, fotoId);
      await recarregar();
      setConfirmando(null);
    } catch (falha: unknown) {
      setErro(obterMensagemErro(falha, 'Não foi possível excluir a foto.'));
    } finally {
      setProcessando(false);
    }
  };

  return (
    <section>
      <div className="mb-5 flex items-center gap-3 border-b border-slate-100 pb-3">
        <div className="rounded-lg bg-blue-50 p-2 text-blue-600"><ImagePlus className="h-4 w-4" /></div>
        <div><h3 className="font-semibold text-slate-800">Fotos do imóvel</h3><p className="text-xs text-slate-500">Envie JPEG, PNG ou WebP. O Kaptei otimiza a imagem e gera a miniatura automaticamente.</p></div>
      </div>

      <div className="rounded-xl border border-dashed border-blue-200 bg-blue-50/50 p-4">
        <div className="grid gap-3 sm:grid-cols-[1fr_auto_auto] sm:items-center">
          <label className="flex cursor-pointer items-center gap-3 rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-700">
            <Upload className="h-4 w-4 text-blue-600" />
            <span className="min-w-0 truncate">{arquivo?.name ?? 'Selecionar imagem'}</span>
            <input ref={inputArquivo} type="file" accept="image/jpeg,image/png,image/webp" className="sr-only" onChange={(e) => setArquivo(e.target.files?.[0] ?? null)} />
          </label>
          <label className="flex items-center gap-2 text-sm font-medium text-slate-700"><input type="checkbox" checked={capa} onChange={(e) => setCapa(e.target.checked)} /> Usar como capa</label>
          <button type="button" onClick={enviar} disabled={processando || !arquivo} className="inline-flex items-center justify-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-bold text-white disabled:opacity-60">{processando ? <Loader2 className="h-4 w-4 animate-spin" /> : <Upload className="h-4 w-4" />}Enviar</button>
        </div>
      </div>

      <details className="mt-3 rounded-lg border border-slate-200 bg-white p-3">
        <summary className="flex cursor-pointer items-center gap-2 text-sm font-semibold text-slate-600"><Link className="h-4 w-4" />Usar uma URL externa legada</summary>
        <div className="mt-3 grid gap-3 sm:grid-cols-[1fr_auto] sm:items-center">
          <input type="url" value={url} onChange={(e) => setURL(e.target.value)} placeholder="https://cdn.seudominio.com.br/imovel/foto.jpg" className="w-full rounded-lg border border-slate-200 px-3 py-2 text-sm outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-600/30" />
          <button type="button" onClick={adicionar} disabled={processando || !url.trim()} className="inline-flex items-center justify-center gap-2 rounded-lg border border-slate-300 px-4 py-2 text-sm font-bold text-slate-700 disabled:opacity-60">Adicionar URL</button>
        </div>
      </details>
      {erro && <p role="alert" className="mt-3 text-sm font-medium text-red-600">{erro}</p>}

      {fotos.length > 0 && <div className="mt-5 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">{fotos.map((foto) => (
        <article key={foto.id} className="overflow-hidden rounded-xl border border-slate-200 bg-white">
          <div className="aspect-video bg-slate-100"><img src={foto.url_thumbnail ?? foto.url} alt="Foto do imóvel" className="h-full w-full object-cover" loading="lazy" /></div>
          <div className="flex items-center justify-between gap-3 p-3">
            <span className="inline-flex items-center gap-1 text-xs font-semibold text-slate-600">{foto.is_capa && <><Star className="h-3.5 w-3.5 fill-amber-400 text-amber-500" /> Capa</>}</span>
            <button type="button" onClick={() => excluir(foto.id)} disabled={processando} className={`inline-flex items-center gap-1 rounded-lg px-3 py-1.5 text-xs font-bold ${confirmando === foto.id ? 'bg-red-600 text-white' : 'bg-red-50 text-red-700'}`}><Trash2 className="h-3.5 w-3.5" />{confirmando === foto.id ? 'Confirmar' : 'Excluir'}</button>
          </div>
        </article>
      ))}</div>}
    </section>
  );
};
