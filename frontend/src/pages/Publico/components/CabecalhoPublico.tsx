import { Building2, MessageCircle, Phone } from 'lucide-react';
import { Link } from 'react-router-dom';
import type { SitePublico } from '../../../types/sitePublico';
import { rotaPublica } from '../../../lib/sitePublico/rotas';
import { registrarEventoConversao } from '../../../lib/metricasConversao/metricasConversao';

interface CabecalhoPublicoProps {
  site: SitePublico;
  basePublica: string;
}

export const CabecalhoPublico = ({ site, basePublica }: CabecalhoPublicoProps) => {
  const { configuracao } = site;
  const whatsapp = configuracao.whatsapp?.replace(/\D/g, '');
  return (
    <header className="sticky top-0 z-40 border-b border-slate-200/80 bg-white/95 backdrop-blur">
      <div className="mx-auto flex h-18 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
        <Link to={rotaPublica(basePublica)} className="flex items-center gap-3 text-slate-950">
          {configuracao.logo_url ? (
            <img src={configuracao.logo_url} alt={`Logo ${site.nome}`} className="h-10 max-w-44 object-contain" />
          ) : (
            <span className="grid h-10 w-10 place-items-center rounded-xl bg-blue-600 text-white">
              <Building2 className="h-5 w-5" />
            </span>
          )}
          <span className="hidden text-lg font-extrabold tracking-tight sm:block">{site.nome}</span>
        </Link>
        <nav className="flex items-center gap-3">
          <a href="#imoveis" className="hidden text-sm font-semibold text-slate-600 hover:text-blue-700 sm:block">Imóveis</a>
          {configuracao.telefone && (
            <a href={`tel:${configuracao.telefone.replace(/\D/g, '')}`} onClick={() => void registrarEventoConversao(site.slug, 'TELEFONE_CLICADO')} className="inline-flex items-center gap-2 rounded-xl bg-slate-950 px-4 py-2.5 text-sm font-semibold text-white hover:bg-blue-700">
              <Phone className="h-4 w-4" /> Falar com especialista
            </a>
          )}
        </nav>
      </div>
      {whatsapp && <a href={`https://wa.me/${whatsapp}`} onClick={() => void registrarEventoConversao(site.slug, 'WHATSAPP_CLICADO')} target="_blank" rel="noreferrer" aria-label={`Conversar com ${site.nome} pelo WhatsApp`} className="fixed bottom-5 right-5 z-50 grid h-14 w-14 place-items-center rounded-full bg-emerald-500 text-white shadow-xl transition hover:scale-105 hover:bg-emerald-600"><MessageCircle className="h-7 w-7" /></a>}
    </header>
  );
};
