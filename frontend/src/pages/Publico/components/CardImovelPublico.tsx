import { Bath, BedDouble, Car, MapPin, Maximize2 } from 'lucide-react';
import { Link } from 'react-router-dom';
import { formatarMoedaBR } from '../../../lib/formatadores/moeda';
import { listarValoresImovel } from '../../../lib/imoveis/valores';
import type { ImovelPublico } from '../../../types/sitePublico';
import { rotaPublica } from '../../../lib/sitePublico/rotas';

interface CardImovelPublicoProps {
  imovel: ImovelPublico;
  basePublica: string;
}

export const CardImovelPublico = ({ imovel, basePublica }: CardImovelPublicoProps) => {
  const valores = listarValoresImovel(imovel);
  const capa = imovel.fotos.find((foto) => foto.is_capa) ?? imovel.fotos[0];

  return (
    <Link to={rotaPublica(basePublica, `imoveis/${imovel.slug}`)} className="group overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm transition hover:-translate-y-1 hover:shadow-xl">
      <div className="relative aspect-[4/3] overflow-hidden bg-slate-100">
        {capa ? (
          <img src={capa.url_thumbnail ?? capa.url} alt={imovel.titulo} loading="lazy" className="h-full w-full object-cover transition duration-500 group-hover:scale-105" />
        ) : (
          <div className="grid h-full place-items-center text-slate-400">Imagem em preparação</div>
        )}
        <div className="absolute left-3 top-3 flex gap-2">
          <span className="rounded-full bg-white/95 px-3 py-1 text-xs font-bold text-slate-800 shadow">{imovel.finalidade}</span>
          {imovel.destaque && <span className="rounded-full bg-blue-600 px-3 py-1 text-xs font-bold text-white shadow">Destaque</span>}
        </div>
      </div>
      <div className="p-5">
        <div className="mb-1 min-h-7 text-slate-950">{valores.length ? valores.map((item) => <p key={item.rotulo} className="text-lg font-extrabold"><span className="text-xs font-bold uppercase text-slate-400">{item.rotulo}</span> R$ {formatarMoedaBR(item.valor)}</p>) : <p className="text-xl font-extrabold">Consulte o valor</p>}</div>
        <h3 className="line-clamp-2 min-h-12 font-bold text-slate-800">{imovel.titulo}</h3>
        <p className="mt-2 flex items-center gap-1.5 text-sm text-slate-500"><MapPin className="h-4 w-4" /> {[imovel.bairro, imovel.cidade, imovel.estado].filter(Boolean).join(', ')}</p>
        <div className="mt-4 grid grid-cols-4 gap-2 border-t border-slate-100 pt-4 text-xs font-medium text-slate-600">
          <span className="flex items-center gap-1"><BedDouble className="h-4 w-4" /> {imovel.quartos}</span>
          <span className="flex items-center gap-1"><Bath className="h-4 w-4" /> {imovel.banheiros}</span>
          <span className="flex items-center gap-1"><Car className="h-4 w-4" /> {imovel.vagas}</span>
          <span className="flex items-center gap-1"><Maximize2 className="h-4 w-4" /> {imovel.area_util ?? imovel.area_total ?? '--'}m²</span>
        </div>
      </div>
    </Link>
  );
};
