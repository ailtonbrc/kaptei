import React, { useState, useEffect, useRef } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Home, MapPin, Tag, FileText, Bed, Bath, Car, Maximize2, Globe2 } from 'lucide-react';
import { CepInput, CurrencyInput, CrudModal } from '../../components';
import { imovelService } from '../../services/imovelService';
import type { Imovel } from '../../types/imovel';
import { obterMensagemErro } from '../../lib/http/erro-api';
import { FotosImovel } from './components/FotosImovel';
import { FormField, SectionHeader } from '../../components/Formulario';
import { BotaoAgendar } from '../CRM/components/BotaoAgendar';

// ---------------------------------------------------------------
// Campo de formulário reutilizável para manter consistência visual
// ---------------------------------------------------------------
// Input padrão reutilizável
const inputClass =
  'w-full px-3 py-2 border border-slate-200 rounded-lg focus:ring-2 focus:ring-blue-600/50 focus:border-blue-500 outline-none transition-colors text-sm';

const selectClass =
  'w-full px-3 py-2 border border-slate-200 rounded-lg focus:ring-2 focus:ring-blue-600/50 focus:border-blue-500 outline-none transition-colors text-sm bg-white';

// ---------------------------------------------------------------
// Cabeçalho de seção reutilizável
// ---------------------------------------------------------------
// ---------------------------------------------------------------
// Componente principal
// ---------------------------------------------------------------
export const ImovelForm: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const isEditing = Boolean(id);

  const [loading, setLoading] = useState(isEditing);
  const [saving, setSaving] = useState(false);
  const [successMsg, setSuccessMsg] = useState('');
  const [errorMsg, setErrorMsg] = useState('');

  const scrollContainerRef = useRef<HTMLDivElement>(null);

  const [formData, setFormData] = useState<Partial<Imovel>>({
    titulo:           '',
    tipo:             'Casa',
    finalidade:       'Venda',
    status:           'Ativo',
    valor_venda:      0,
    valor_locacao:    0,
    valor_condominio: 0,
    valor_iptu:       0,
    area_total:       0,
    area_util:        0,
    quartos:          0,
    suites:           0,
    banheiros:        0,
    vagas:            0,
    cep:              '',
    logradouro:       '',
    numero:           '',
    complemento:      '',
    bairro:           '',
    cidade:           '',
    estado:           'MS',
    descricao:        '',
		slug_publico:     '',
		publicado:        false,
		destaque:         false,
		titulo_seo:       '',
		descricao_seo:    ''
  });

  async function carregarImovel(imovelId: string) {
    try {
      const dados = await imovelService.buscarPorId(imovelId);
      setFormData(dados);
    } catch {
      setErrorMsg('Erro ao carregar dados do imóvel.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- sincroniza o formulário com o identificador da rota.
    if (isEditing && id) void carregarImovel(id);
  }, [id, isEditing]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
    const { name, value, type } = e.target;
    if (type === 'number') {
      setFormData(prev => ({ ...prev, [name]: value ? Number(value) : 0 }));
    } else {
      setFormData(prev => ({ ...prev, [name]: value }));
    }
  };

  // Handler específico para CurrencyInput — recebe nome e valor numérico direto
  const handleCurrencyChange = (name: string, valor: number) => {
    setFormData(prev => ({ ...prev, [name]: valor }));
  };

  // Callback chamado pelo CepInput após busca no ViaCEP
  const handleAddressFetch = (data: { logradouro: string; bairro: string; cidade: string; estado: string }) => {
    setFormData(prev => ({
      ...prev,
      logradouro: data.logradouro || prev.logradouro,
      bairro:     data.bairro     || prev.bairro,
      cidade:     data.cidade     || prev.cidade,
      estado:     data.estado     || prev.estado,
    }));
  };

  // Rola o scroll do modal para o topo (seções: básica, valores, descrição)
  const scrollToTop = () => {
    scrollContainerRef.current?.scrollTo({ top: 0, behavior: 'smooth' });
  };

  // Rola o scroll do modal para o fim (seção: endereço)
  const scrollToBottom = () => {
    const el = scrollContainerRef.current;
    if (el) el.scrollTo({ top: el.scrollHeight, behavior: 'smooth' });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setErrorMsg('');
    setSuccessMsg('');

    try {
      if (isEditing && id) {
        await imovelService.atualizar(id, formData);
        setSuccessMsg('Imóvel atualizado com sucesso!');
      } else {
        await imovelService.criar(formData);
        setSuccessMsg('Imóvel cadastrado com sucesso!');
      }
      setTimeout(() => navigate('/app/imoveis'), 2000);
    } catch (error: unknown) {
      setErrorMsg(obterMensagemErro(error, 'Erro ao salvar imóvel. Verifique os campos.'));
      setTimeout(() => setErrorMsg(''), 5000);
    } finally {
      setSaving(false);
    }
  };

  // Fecha o modal voltando para a lista
  const fechar = () => navigate('/app/imoveis');

  // ---------------------------------------------------------------
  // Loading state
  // ---------------------------------------------------------------
  if (loading) {
    return (
      <div className="fixed inset-0 z-50 bg-black/40 flex items-center justify-center">
        <div className="bg-white rounded-2xl p-10 flex flex-col items-center gap-4 shadow-2xl">
          <div className="animate-spin rounded-full h-10 w-10 border-4 border-blue-600 border-t-transparent" />
          <p className="text-slate-600 font-medium">Carregando dados do imóvel...</p>
        </div>
      </div>
    );
  }

  // ---------------------------------------------------------------
  // Renderização principal — usa CrudModal como estrutura base
  // ---------------------------------------------------------------
  return (
    <CrudModal
      ref={scrollContainerRef}
      breadcrumbPai="Imóveis"
      titulo={isEditing ? 'Editar Imóvel' : 'Novo Imóvel'}
      onCancel={fechar}
      formId="imovel-form"
      isSaving={saving}
      labelSalvar={isEditing ? 'Salvar Alterações' : 'Cadastrar Imóvel'}
      successMsg={successMsg}
      errorMsg={errorMsg}
    >
		  <form id="imovel-form" onSubmit={handleSubmit} className="p-6 sm:p-8 space-y-10">
			{isEditing && id && <div className="flex justify-end"><BotaoAgendar imovelID={id} contexto={formData.titulo || 'Imóvel'} /></div>}

            {/* ── INFORMAÇÕES BÁSICAS ── */}
            <section>
              <SectionHeader
                icon={<Home className="w-4 h-4" />}
                title="Informações Básicas"
                subtitle="Tipo, finalidade e status do imóvel"
              />
              <div className="grid grid-cols-1 md:grid-cols-12 gap-5">
                {/* Título — ocupa linha inteira */}
                <FormField label="Título do Anúncio" required className="md:col-span-12">
                  <input
                    type="text"
                    name="titulo"
                    value={formData.titulo}
                    onChange={handleChange}
                    required
                    placeholder="Ex: Lindo Apartamento de 3 Quartos no Centro"
                    className={inputClass}
                  />
                </FormField>

                <FormField label="Tipo" required className="md:col-span-4">
                  <select name="tipo" value={formData.tipo} onChange={handleChange} required className={selectClass}>
                    <option>Casa</option>
                    <option>Apartamento</option>
                    <option>Terreno</option>
                    <option>Comercial</option>
                    <option>Galpão</option>
                    <option>Rural</option>
                  </select>
                </FormField>

                <FormField label="Finalidade" required className="md:col-span-4">
                  <select name="finalidade" value={formData.finalidade} onChange={handleChange} required className={selectClass}>
                    <option value="Venda">Venda</option>
                    <option value="Locação">Locação</option>
                    <option value="Venda e Locação">Venda e Locação</option>
                  </select>
                </FormField>

                <FormField label="Status" required className="md:col-span-4">
                  <select name="status" value={formData.status} onChange={handleChange} required className={selectClass}>
                    <option value="Ativo">Disponível (Ativo)</option>
                    <option value="Inativo">Inativo</option>
                    <option value="Vendido">Vendido</option>
                    <option value="Alugado">Alugado</option>
                  </select>
                </FormField>
              </div>
            </section>

            {/* ── VALORES E CARACTERÍSTICAS ── */}
            <section onFocusCapture={scrollToTop}>
              <SectionHeader
                icon={<Tag className="w-4 h-4" />}
                title="Valores e Detalhes"
                subtitle="Preços, áreas e atributos do imóvel"
              />

              {/* Valores monetários — CurrencyInput formata em R$ ##.##0,00 durante a digitação */}
              <div className="grid grid-cols-2 md:grid-cols-4 gap-5 mb-6">
                {([
                  { label: 'Valor de Venda',       name: 'valor_venda'      },
                  { label: 'Valor de Locação',      name: 'valor_locacao'    },
                  { label: 'Valor do Condomínio',   name: 'valor_condominio' },
                  { label: 'IPTU Anual',            name: 'valor_iptu'       },
                ] as const).map(({ label, name }) => (
                  <CurrencyInput
                    key={name}
                    name={name}
                    label={label}
                    value={formData[name] ?? 0}
                    onValueChange={handleCurrencyChange}
                  />
                ))}
              </div>

              {/* Áreas e cômodos */}
              <div className="grid grid-cols-3 md:grid-cols-6 gap-4">
                {([
                  { label: 'Área Útil (m²)',  name: 'area_util',  icon: <Maximize2 className="w-3.5 h-3.5" /> },
                  { label: 'Área Total (m²)', name: 'area_total', icon: <Maximize2 className="w-3.5 h-3.5" /> },
                  { label: 'Quartos',         name: 'quartos',    icon: <Bed className="w-3.5 h-3.5" /> },
                  { label: 'Suítes',          name: 'suites',     icon: <Bed className="w-3.5 h-3.5" /> },
                  { label: 'Banheiros',       name: 'banheiros',  icon: <Bath className="w-3.5 h-3.5" /> },
                  { label: 'Vagas',           name: 'vagas',      icon: <Car className="w-3.5 h-3.5" /> },
                ] as const).map(({ label, name, icon }) => (
                  <FormField key={name} label={label}>
                    <div className="relative">
                      <span className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400">{icon}</span>
                      <input
                        type="number"
                        name={name}
                        value={formData[name] || ''}
                        onChange={handleChange}
                        min={0}
                        className={`${inputClass} pl-8`}
                      />
                    </div>
                  </FormField>
                ))}
              </div>
            </section>

            {/* ── ENDEREÇO ── */}
            <section onFocusCapture={scrollToBottom}>
              <SectionHeader
                icon={<MapPin className="w-4 h-4" />}
                title="Localização"
                subtitle="Endereço completo do imóvel — CEP preenche automaticamente"
              />

              {/* Linha 1: CEP + Logradouro + Número */}
              <div className="grid grid-cols-1 md:grid-cols-12 gap-5 mb-5">
                <div className="md:col-span-3">
                  <CepInput
                    name="cep"
                    value={formData.cep || ''}
                    onChange={handleChange}
                    onAddressFetch={handleAddressFetch}
                  />
                </div>
                <FormField label="Logradouro (Rua, Av, etc.)" className="md:col-span-7">
                  <input
                    type="text"
                    name="logradouro"
                    value={formData.logradouro || ''}
                    onChange={handleChange}
                    placeholder="Preenchido automaticamente pelo CEP"
                    className={inputClass}
                  />
                </FormField>
                <FormField label="Número" className="md:col-span-2">
                  <input
                    type="text"
                    name="numero"
                    value={formData.numero || ''}
                    onChange={handleChange}
                    placeholder="Nº"
                    className={inputClass}
                  />
                </FormField>
              </div>

              {/* Linha 2: Complemento + Bairro + Cidade + Estado */}
              <div className="grid grid-cols-1 md:grid-cols-12 gap-5">
                <FormField label="Complemento" className="md:col-span-4">
                  <input
                    type="text"
                    name="complemento"
                    value={formData.complemento || ''}
                    onChange={handleChange}
                    placeholder="Apto, Bloco, Casa..."
                    className={inputClass}
                  />
                </FormField>
                <FormField label="Bairro" className="md:col-span-3">
                  <input
                    type="text"
                    name="bairro"
                    value={formData.bairro || ''}
                    onChange={handleChange}
                    placeholder="Preenchido pelo CEP"
                    className={inputClass}
                  />
                </FormField>
                <FormField label="Cidade" className="md:col-span-3">
                  <input
                    type="text"
                    name="cidade"
                    value={formData.cidade || ''}
                    onChange={handleChange}
                    placeholder="Preenchido pelo CEP"
                    className={inputClass}
                  />
                </FormField>
                <FormField label="Estado (UF)" className="md:col-span-2">
                  <input
                    type="text"
                    name="estado"
                    value={formData.estado || ''}
                    onChange={handleChange}
                    maxLength={2}
                    placeholder="MS"
                    className={`${inputClass} uppercase`}
                  />
                </FormField>
              </div>
            </section>

            {/* ── DESCRIÇÃO ── */}
            <section onFocusCapture={scrollToBottom}>
              <SectionHeader
                icon={<FileText className="w-4 h-4" />}
                title="Descrição"
                subtitle="Destaque os diferenciais e informações adicionais do imóvel"
              />
              <textarea
                name="descricao"
                value={formData.descricao || ''}
                onChange={handleChange}
                rows={5}
                placeholder="Descreva os diferenciais do imóvel, infraestrutura do condomínio, etc."
                className={`${inputClass} resize-y`}
              />
            </section>

			{isEditing && id && <FotosImovel imovelId={id} fotos={formData.fotos ?? []} onChange={(fotos) => setFormData((atual) => ({ ...atual, fotos }))} />}

			<section onFocusCapture={scrollToBottom}>
				<SectionHeader icon={<Globe2 className="w-4 h-4" />} title="Publicação no site" subtitle="Controle a URL, destaque e metadados do catálogo público" />
				<div className="grid gap-5 md:grid-cols-2">
					<FormField label="Endereço público do imóvel">
						<input name="slug_publico" value={formData.slug_publico || ''} onChange={handleChange} pattern="[a-z0-9]+(?:-[a-z0-9]+)*" maxLength={180} placeholder="casa-3-quartos-centro" className={inputClass} />
					</FormField>
					<div className="flex items-center gap-6 rounded-xl border border-slate-200 bg-slate-50 px-4">
						<label className="flex items-center gap-2 text-sm font-semibold text-slate-700"><input type="checkbox" checked={formData.publicado ?? false} onChange={(e) => setFormData((atual) => ({ ...atual, publicado: e.target.checked }))} /> Publicado</label>
						<label className="flex items-center gap-2 text-sm font-semibold text-slate-700"><input type="checkbox" checked={formData.destaque ?? false} onChange={(e) => setFormData((atual) => ({ ...atual, destaque: e.target.checked }))} /> Destaque</label>
					</div>
					<FormField label="Título SEO" className="md:col-span-2"><input name="titulo_seo" value={formData.titulo_seo || ''} onChange={handleChange} maxLength={180} className={inputClass} /></FormField>
					<FormField label="Descrição SEO" className="md:col-span-2"><textarea name="descricao_seo" value={formData.descricao_seo || ''} onChange={handleChange} maxLength={320} rows={3} className={`${inputClass} resize-y`} /></FormField>
				</div>
			</section>

          </form>
    </CrudModal>
  );
};
