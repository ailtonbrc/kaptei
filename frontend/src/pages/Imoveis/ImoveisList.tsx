import React from 'react';
import { Link } from 'react-router-dom';
import { Plus, MapPin, Bed, Bath, Car, Image as ImageIcon } from 'lucide-react';
import { imovelService } from '../../services/imovelService';
import type { Imovel } from '../../types/imovel';
import { formatarMoedaBR } from '../../lib/formatadores/moeda';
import { usePaginacao } from '../../hooks/usePaginacao';

const formatarValorImovel = (valor?: number) => valor ? `R$ ${formatarMoedaBR(valor)}` : 'Consulte';

export const ImoveisList: React.FC = () => {
  const {
    items: imoveis,
    loading,
    pagina,
    total,
    carregarItems: carregarImoveis,
  } = usePaginacao<Imovel>({
    fetchData: async (pag) => await imovelService.listar({ pagina: pag })
  });

  return (
    <div className="max-w-7xl mx-auto p-4 sm:p-6 lg:p-8">
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center mb-8 gap-4">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">Imóveis</h1>
          <p className="text-slate-500 mt-1">Gerencie a sua carteira de imóveis.</p>
        </div>
        <Link 
          to="/app/imoveis/novo" 
          className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg font-medium transition-colors shadow-sm"
        >
          <Plus className="w-5 h-5" />
          Novo Imóvel
        </Link>
      </div>

      {/* Grid de Imóveis */}
      {loading ? (
        <div className="flex justify-center p-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        </div>
      ) : imoveis.length === 0 ? (
        <div className="bg-white rounded-2xl shadow-sm border border-slate-200 p-12 text-center">
          <div className="w-16 h-16 bg-blue-50 text-blue-600 rounded-full flex items-center justify-center mx-auto mb-4">
            <ImageIcon className="w-8 h-8" />
          </div>
          <h3 className="text-lg font-medium text-slate-900 mb-2">Nenhum imóvel cadastrado</h3>
          <p className="text-slate-500 mb-6 max-w-md mx-auto">
            Você ainda não possui imóveis na sua carteira. Clique no botão abaixo para adicionar o seu primeiro imóvel.
          </p>
          <Link 
            to="/app/imoveis/novo" 
            className="inline-flex items-center gap-2 bg-slate-900 hover:bg-slate-800 text-white px-5 py-2.5 rounded-lg font-medium transition-colors shadow-sm"
          >
            <Plus className="w-5 h-5" />
            Cadastrar Primeiro Imóvel
          </Link>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          {imoveis.map((imovel) => (
            <Link key={imovel.id} to={`/app/imoveis/${imovel.id}/editar`} className="group block bg-white rounded-2xl shadow-sm hover:shadow-md border border-slate-200 overflow-hidden transition-all duration-300">
              {/* Imagem de Capa */}
              <div className="aspect-[4/3] bg-slate-100 relative overflow-hidden">
                {imovel.fotos && imovel.fotos.length > 0 ? (
                  <img src={imovel.fotos[0].url_thumbnail ?? imovel.fotos[0].url} alt={imovel.titulo} className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500" />
                ) : (
                  <div className="w-full h-full flex items-center justify-center text-slate-400">
                    <ImageIcon className="w-10 h-10 opacity-50" />
                  </div>
                )}
                <div className="absolute top-3 left-3 bg-white/90 backdrop-blur-sm px-2.5 py-1 rounded-md text-xs font-semibold text-slate-700 shadow-sm">
                  {imovel.finalidade}
                </div>
                {imovel.status === 'Vendido' || imovel.status === 'Alugado' ? (
                  <div className="absolute top-3 right-3 bg-emerald-500 text-white px-2.5 py-1 rounded-md text-xs font-semibold shadow-sm">
                    {imovel.status}
                  </div>
                ) : null}
              </div>
              
              {/* Conteúdo do Card */}
              <div className="p-4">
                <div className="text-xl font-bold text-slate-900 mb-1">
				  {formatarValorImovel(imovel.finalidade === 'Venda' ? imovel.valor_venda : imovel.valor_locacao)}
                </div>
                <h3 className="text-slate-800 font-medium text-sm line-clamp-1 mb-2" title={imovel.titulo}>
                  {imovel.titulo}
                </h3>
                
                {/* Localização */}
                <div className="flex items-center gap-1.5 text-slate-500 text-sm mb-4">
                  <MapPin className="w-4 h-4 shrink-0" />
                  <span className="truncate">{imovel.bairro}{imovel.bairro && imovel.cidade ? ', ' : ''}{imovel.cidade}</span>
                </div>
                
                {/* Atributos */}
                <div className="flex items-center justify-between border-t border-slate-100 pt-3 text-slate-600 text-sm">
                  <div className="flex items-center gap-1.5" title="Quartos">
                    <Bed className="w-4 h-4" />
                    <span>{imovel.quartos}</span>
                  </div>
                  <div className="flex items-center gap-1.5" title="Banheiros">
                    <Bath className="w-4 h-4" />
                    <span>{imovel.banheiros}</span>
                  </div>
                  <div className="flex items-center gap-1.5" title="Vagas">
                    <Car className="w-4 h-4" />
                    <span>{imovel.vagas}</span>
                  </div>
                  <div className="font-medium text-slate-700">
                    {imovel.area_util ? `${imovel.area_util}m²` : '--'}
                  </div>
                </div>
              </div>
            </Link>
          ))}
        </div>
      )}
	  {imoveis.length < total && <div className="mt-8 text-center"><button type="button" onClick={() => void carregarImoveis(pagina + 1, true)} disabled={loading} className="rounded-lg border border-slate-300 bg-white px-4 py-2 text-sm font-semibold text-slate-700 hover:bg-slate-50 disabled:opacity-60">Carregar mais imóveis ({imoveis.length} de {total})</button></div>}
    </div>
  );
};
