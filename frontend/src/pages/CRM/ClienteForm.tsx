import React, { useState, useEffect, useRef } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { User, CheckCircle, FileText, Tag, Search, X } from 'lucide-react';
import { PhoneInput, CrudModal, CurrencyInput } from '../../components';
import { clientesService } from '../../services/clientesService';
import { TimelineLead } from './TimelineLead';
import type { Cliente } from '../../types/cliente';
import { STATUS_FUNIL, ORIGEM_LEAD } from '../../types/cliente';
import { obterMensagemErro } from '../../lib/http/erro-api';
import { SectionHeader as BaseSectionHeader } from '../../components/Formulario';
import { BotaoAgendar } from './components/BotaoAgendar';

// Classe padrão de inputs para este formulário
const inputClass =
  'w-full px-3 py-2 border border-slate-200 rounded-lg focus:ring-2 focus:ring-emerald-600/50 focus:border-emerald-500 outline-none transition-colors text-sm';

const selectClass =
  'w-full px-3 py-2 border border-slate-200 rounded-lg focus:ring-2 focus:ring-emerald-600/50 focus:border-emerald-500 outline-none transition-colors text-sm bg-white';

// Cabeçalho de seção reutilizável internamente
const SectionHeader = ({ icon, title }: { icon: React.ReactNode; title: string }) => <BaseSectionHeader icon={icon} title={title} color="emerald" />;

// ---------------------------------------------------------------
// Componente principal
// ---------------------------------------------------------------
export const ClienteForm: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const isEditing = Boolean(id);

  const [loading, setLoading] = useState(isEditing);
  const [saving, setSaving] = useState(false);
  const [successMsg, setSuccessMsg] = useState('');
  const [errorMsg, setErrorMsg]   = useState('');

  const scrollContainerRef = useRef<HTMLDivElement>(null);
  const [currentTag, setCurrentTag] = useState('');
  const [currentBairro, setCurrentBairro] = useState('');

  const [formData, setFormData] = useState<Partial<Cliente>>({
    cpf: '',
    data_nascimento: '',
    estado_civil: '',
    nome:          '',
    email:         '',
    telefone:      '',
    status_funil:  'NOVO',
    origem:        'OUTROS',
    interesse_tipo: '',
    notas:         '',
    tags:          [],
    temperatura:   'MORNO',
    corretor_id:   '',
    preferencias:  {
      tipo_imovel: [],
      finalidade: '',
      orcamento_min: null,
      orcamento_max: null,
      bairros: [],
      quartos_min: null,
    },
    financeiro: {
      renda_mensal: null,
      precisa_financiamento: '',
      possui_fgts: false,
      forma_pagamento: ''
    }
  });

  async function carregarCliente(clienteId: string) {
    try {
      const dados = await clientesService.obterPorId(clienteId);
      if (!dados.tags) dados.tags = [];
      if (!dados.preferencias) {
        dados.preferencias = {
          tipo_imovel: [],
          bairros: []
        };
      }
      setFormData(dados);
    } catch {
      setErrorMsg('Erro ao carregar dados do cliente.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- sincroniza o formulário com o identificador da rota.
    if (isEditing && id) void carregarCliente(id);
  }, [id, isEditing]);

  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>
  ) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  const handlePrefChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>
  ) => {
    const { name, value } = e.target;
    // Converte quartos_min para number se existir
    const finalValue = name === 'quartos_min' ? (value ? Number(value) : null) : value;
    setFormData(prev => ({
      ...prev,
      preferencias: {
        ...prev.preferencias,
        [name]: finalValue
      }
    }));
  };

  const handleCurrencyChange = (name: string, value: number) => {
    setFormData(prev => ({
      ...prev,
      preferencias: {
        ...prev.preferencias,
        [name]: value || null
      }
    }));
  };

  const handleAddTag = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      if (currentTag.trim() && !formData.tags?.includes(currentTag.trim())) {
        setFormData(prev => ({
          ...prev,
          tags: [...(prev.tags || []), currentTag.trim()]
        }));
        setCurrentTag('');
      }
    }
  };

  const removeTag = (tagToRemove: string) => {
    setFormData(prev => ({
      ...prev,
      tags: (prev.tags || []).filter(t => t !== tagToRemove)
    }));
  };

  const handleAddBairro = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      if (currentBairro.trim() && !formData.preferencias?.bairros?.includes(currentBairro.trim())) {
        setFormData(prev => ({
          ...prev,
          preferencias: {
            ...prev.preferencias,
            bairros: [...(prev.preferencias?.bairros || []), currentBairro.trim()]
          }
        }));
        setCurrentBairro('');
      }
    }
  };

  const removeBairro = (bairroToRemove: string) => {
    setFormData(prev => ({
      ...prev,
      preferencias: {
        ...prev.preferencias,
        bairros: (prev.preferencias?.bairros || []).filter(b => b !== bairroToRemove)
      }
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setErrorMsg('');
    setSuccessMsg('');

    try {
      const payload = {
        ...formData,
        tags: formData.tags || [],
        preferencias: {
          ...formData.preferencias,
          orcamento_min: formData.preferencias?.orcamento_min ? Number(formData.preferencias.orcamento_min) : null,
          orcamento_max: formData.preferencias?.orcamento_max ? Number(formData.preferencias.orcamento_max) : null,
          quartos_min: formData.preferencias?.quartos_min ? Number(formData.preferencias.quartos_min) : null,
        },
        financeiro: {
          ...formData.financeiro,
          renda_mensal: formData.financeiro?.renda_mensal ? Number(formData.financeiro.renda_mensal) : null
        }
      };

      if (isEditing && id) {
        await clientesService.atualizar(id, payload);
        setSuccessMsg('Cliente atualizado com sucesso!');
      } else {
        await clientesService.criar(payload as Cliente);
        setSuccessMsg('Cliente criado com sucesso!');
      }
      setTimeout(() => navigate('/app/crm'), 2000);
    } catch (error: unknown) {
      setErrorMsg(obterMensagemErro(error, 'Erro ao salvar cliente. Verifique os campos.'));
      setTimeout(() => setErrorMsg(''), 5000);
    } finally {
      setSaving(false);
    }
  };

  const fechar = () => navigate('/app/crm');

  // ---------------------------------------------------------------
  // Loading state
  // ---------------------------------------------------------------
  if (loading) {
    return (
      <div className="fixed inset-0 z-50 bg-black/40 flex items-center justify-center">
        <div className="bg-white rounded-2xl p-10 flex flex-col items-center gap-4 shadow-2xl">
          <div className="animate-spin rounded-full h-10 w-10 border-4 border-emerald-600 border-t-transparent" />
          <p className="text-slate-600 font-medium">Carregando dados do cliente...</p>
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
      breadcrumbPai="CRM"
      titulo={isEditing ? 'Editar Cliente / Lead' : 'Novo Cliente / Lead'}
      onCancel={fechar}
      formId="cliente-form"
      isSaving={saving}
      labelSalvar={isEditing ? 'Salvar Alterações' : 'Cadastrar Cliente'}
      successMsg={successMsg}
      errorMsg={errorMsg}
    >
	  <form id="cliente-form" onSubmit={handleSubmit} className="p-6 sm:p-8 space-y-10">
		{isEditing && id && <div className="flex justify-end"><BotaoAgendar clienteID={id} contexto={formData.nome || 'Cliente'} /></div>}

        {/* ── DADOS DO CONTATO ── */}
        <section>
          <SectionHeader icon={<User className="w-4 h-4" />} title="Dados do Contato" />
          <div className="grid grid-cols-1 md:grid-cols-2 gap-5">

            <div className="space-y-1.5 md:col-span-2">
              <label className="block text-sm font-medium text-slate-700">
                Nome Completo <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                name="nome"
                value={formData.nome}
                onChange={handleChange}
                required
                placeholder="Ex: Maria Silva"
                className={inputClass}
              />
            </div>

            <div className="space-y-1.5">
              <label className="block text-sm font-medium text-slate-700">CPF</label>
              <input
                type="text"
                name="cpf"
                value={formData.cpf || ''}
                onChange={handleChange}
                placeholder="000.000.000-00"
                className={inputClass}
              />
            </div>

            <div className="space-y-1.5">
              <label className="block text-sm font-medium text-slate-700">Data de Nascimento</label>
              <input
                type="date"
                name="data_nascimento"
                value={formData.data_nascimento || ''}
                onChange={handleChange}
                className={inputClass}
              />
            </div>
            
            <div className="space-y-1.5">
              <label className="block text-sm font-medium text-slate-700">Estado Civil</label>
              <select
                name="estado_civil"
                value={formData.estado_civil || ''}
                onChange={handleChange}
                className={selectClass}
              >
                <option value="">Selecione...</option>
                <option value="Solteiro(a)">Solteiro(a)</option>
                <option value="Casado(a)">Casado(a)</option>
                <option value="Divorciado(a)">Divorciado(a)</option>
                <option value="Viúvo(a)">Viúvo(a)</option>
                <option value="União Estável">União Estável</option>
              </select>
            </div>

            <div className="space-y-1.5">
              <label className="block text-sm font-medium text-slate-700">E-mail</label>
              <input
                type="email"
                name="email"
                value={formData.email || ''}
                onChange={handleChange}
                placeholder="maria@email.com"
                className={inputClass}
              />
            </div>

            <PhoneInput
              name="telefone"
              label="Telefone / WhatsApp"
              value={formData.telefone || ''}
              onChange={handleChange}
            />
          </div>
        </section>

        {/* ── DADOS FINANCEIROS ── */}
        <section>
          <SectionHeader icon={<FileText className="w-4 h-4" />} title="Dados Financeiros" />
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-5">
            <div className="space-y-1.5">
              <label className="block text-sm font-medium text-slate-700">Renda Mensal</label>
              <CurrencyInput
                name="renda_mensal"
                value={formData.financeiro?.renda_mensal || 0}
                onValueChange={(_, value) => {
                  setFormData(prev => ({
                    ...prev,
                    financeiro: { ...prev.financeiro, renda_mensal: value }
                  }));
                }}
                placeholder="0,00"
              />
            </div>

            <div className="space-y-1.5">
              <label className="block text-sm font-medium text-slate-700">Forma de Pagamento</label>
              <select
                name="forma_pagamento"
                value={formData.financeiro?.forma_pagamento || ''}
                onChange={(e) => setFormData(prev => ({ ...prev, financeiro: { ...prev.financeiro, forma_pagamento: e.target.value } }))}
                className={selectClass}
              >
                <option value="">Selecione...</option>
                <option value="A Vista">A Vista</option>
                <option value="Financiamento">Financiamento</option>
                <option value="Consorcio">Consórcio</option>
                <option value="Parcelado">Parcelado Direto</option>
              </select>
            </div>

            <div className="space-y-1.5">
              <label className="block text-sm font-medium text-slate-700">Financiamento?</label>
              <select
                name="precisa_financiamento"
                value={formData.financeiro?.precisa_financiamento || ''}
                onChange={(e) => setFormData(prev => ({ ...prev, financeiro: { ...prev.financeiro, precisa_financiamento: e.target.value } }))}
                className={selectClass}
              >
                <option value="">Selecione...</option>
                <option value="Sim">Sim, precisa aprovar</option>
                <option value="Nao">Não precisa</option>
                <option value="JaAprovado">Já tem crédito aprovado</option>
              </select>
            </div>

            <div className="space-y-1.5 flex items-center mt-6">
              <label className="flex items-center gap-2 cursor-pointer text-sm font-medium text-slate-700">
                <input
                  type="checkbox"
                  name="possui_fgts"
                  checked={formData.financeiro?.possui_fgts || false}
                  onChange={(e) => setFormData(prev => ({ ...prev, financeiro: { ...prev.financeiro, possui_fgts: e.target.checked } }))}
                  className="w-4 h-4 rounded border-slate-300 text-emerald-600 focus:ring-emerald-600"
                />
                Utilizará FGTS?
              </label>
            </div>
          </div>
        </section>

        {/* ── GESTÃO DE VENDAS ── */}
        <section>
          <SectionHeader icon={<CheckCircle className="w-4 h-4" />} title="Gestão de Vendas (CRM)" />
          <div className="grid grid-cols-1 md:grid-cols-3 gap-5">

            <div className="space-y-1.5">
              <label className="block text-sm font-medium text-slate-700">
                Status no Funil <span className="text-red-500">*</span>
              </label>
              <select
                name="status_funil"
                value={formData.status_funil}
                onChange={handleChange}
                required
                className={selectClass}
              >
                {STATUS_FUNIL.map(s => (
                  <option key={s.value} value={s.value}>{s.label}</option>
                ))}
              </select>
            </div>

            <div className="space-y-1.5">
              <label className="block text-sm font-medium text-slate-700">Origem do Lead</label>
              <select
                name="origem"
                value={formData.origem}
                onChange={handleChange}
                className={selectClass}
              >
                {ORIGEM_LEAD.map(o => (
                  <option key={o.value} value={o.value}>{o.label}</option>
                ))}
              </select>
            </div>

            <div className="space-y-1.5">
              <label className="block text-sm font-medium text-slate-700">Tipo de Interesse</label>
              <select
                name="interesse_tipo"
                value={formData.interesse_tipo || ''}
                onChange={handleChange}
                className={selectClass}
              >
                <option value="">Indefinido</option>
                <option value="Compra">Compra</option>
                <option value="Locação">Locação</option>
              </select>
            </div>

            <div className="space-y-1.5">
              <label className="block text-sm font-medium text-slate-700">Temperatura</label>
              <select
                name="temperatura"
                value={formData.temperatura || ''}
                onChange={handleChange}
                className={selectClass}
              >
                <option value="FRIO">Frio (Sem pressa)</option>
                <option value="MORNO">Morno (Avaliando opções)</option>
                <option value="QUENTE">Quente (Pronto para fechar)</option>
              </select>
            </div>

            <div className="space-y-1.5">
              <label className="block text-sm font-medium text-slate-700">Próxima Ação (Data)</label>
              <input
                type="date"
                name="proxima_acao"
                value={formData.proxima_acao?.substring(0, 10) || ''}
                onChange={handleChange}
                className={inputClass}
              />
            </div>
          </div>
          
          {/* Subseção de Origem / UTM */}
          <div className="mt-6 pt-5 border-t border-slate-100 grid grid-cols-1 md:grid-cols-3 gap-5">
            <div className="space-y-1.5">
              <label className="block text-sm font-medium text-slate-700">Canal (UTM Source)</label>
              <input
                type="text"
                name="canal"
                value={formData.origem_utm?.canal || ''}
                onChange={(e) => setFormData(prev => ({ ...prev, origem_utm: { ...prev.origem_utm, canal: e.target.value } }))}
                placeholder="Ex: instagram, google"
                className={inputClass}
              />
            </div>
            <div className="space-y-1.5">
              <label className="block text-sm font-medium text-slate-700">Campanha (UTM Campaign)</label>
              <input
                type="text"
                name="campanha"
                value={formData.origem_utm?.campanha || ''}
                onChange={(e) => setFormData(prev => ({ ...prev, origem_utm: { ...prev.origem_utm, campanha: e.target.value } }))}
                placeholder="Ex: black_friday, imovel_luxo"
                className={inputClass}
              />
            </div>
            <div className="space-y-1.5">
              <label className="block text-sm font-medium text-slate-700">Imóvel de Origem (ID)</label>
              <input
                type="text"
                name="imovel_origem_id"
                value={formData.origem_utm?.imovel_origem_id || ''}
                onChange={(e) => setFormData(prev => ({ ...prev, origem_utm: { ...prev.origem_utm, imovel_origem_id: e.target.value } }))}
                placeholder="Cód do imóvel que gerou o lead"
                className={inputClass}
              />
            </div>
          </div>
        </section>

        {/* ── TAGS DE SEGMENTAÇÃO ── */}
        <section>
          <SectionHeader icon={<Tag className="w-4 h-4" />} title="Tags e Perfil (Keywords)" />
          <div className="space-y-4">
            <div className="space-y-1.5">
              <label className="block text-sm font-medium text-slate-700">
                Adicionar Tags (Aperte Enter para adicionar)
              </label>
              <div className="flex flex-wrap gap-2 mb-2">
                {formData.tags?.map((tag) => (
                  <span key={tag} className="inline-flex items-center gap-1 px-3 py-1 bg-emerald-100 text-emerald-700 rounded-full text-sm font-medium">
                    #{tag}
                    <button type="button" onClick={() => removeTag(tag)} className="text-emerald-500 hover:text-emerald-800">
                      <X className="w-3.5 h-3.5" />
                    </button>
                  </span>
                ))}
              </div>
              <input
                type="text"
                value={currentTag}
                onChange={(e) => setCurrentTag(e.target.value)}
                onKeyDown={handleAddTag}
                placeholder="Ex: investidor, permuta, urgente..."
                className={inputClass}
              />
            </div>
            
            <div className="flex flex-wrap gap-2 pt-2">
              <span className="text-xs text-slate-500 font-medium self-center mr-2">Sugestões rápidas:</span>
              {['investidor', 'primeiro_imovel', 'alto_padrao', 'permuta', 'urgente'].map(sug => (
                <button
                  key={sug}
                  type="button"
                  onClick={() => {
                    if (!formData.tags?.includes(sug)) {
                      setFormData(prev => ({ ...prev, tags: [...(prev.tags || []), sug] }));
                    }
                  }}
                  className="px-2.5 py-1 text-xs border border-slate-200 text-slate-600 rounded-lg hover:bg-slate-50 hover:border-slate-300 transition-colors"
                >
                  +{sug}
                </button>
              ))}
            </div>
          </div>
        </section>

        {/* ── PERFIL DE BUSCA / PREFERÊNCIAS ── */}
        <section>
          <SectionHeader icon={<Search className="w-4 h-4" />} title="Perfil de Busca (Match Inteligente)" />
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-5 mb-5">
            <CurrencyInput
              name="orcamento_min"
              label="Orçamento Mínimo"
              value={formData.preferencias?.orcamento_min || 0}
              onValueChange={handleCurrencyChange}
              placeholder="0,00"
            />
            
            <CurrencyInput
              name="orcamento_max"
              label="Orçamento Máximo"
              value={formData.preferencias?.orcamento_max || 0}
              onValueChange={handleCurrencyChange}
              placeholder="0,00"
            />
          </div>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
            <div className="space-y-1.5">
              <label className="block text-sm font-medium text-slate-700">Bairros de Interesse (Aperte Enter para adicionar)</label>
              <div className="flex flex-wrap gap-2 mb-2">
                {formData.preferencias?.bairros?.map((bairro) => (
                  <span key={bairro} className="inline-flex items-center gap-1 px-3 py-1 bg-slate-100 text-slate-700 border border-slate-200 rounded-lg text-sm font-medium">
                    {bairro}
                    <button type="button" onClick={() => removeBairro(bairro)} className="text-slate-400 hover:text-red-500">
                      <X className="w-3.5 h-3.5" />
                    </button>
                  </span>
                ))}
              </div>
              <input
                type="text"
                value={currentBairro}
                onChange={(e) => setCurrentBairro(e.target.value)}
                onKeyDown={handleAddBairro}
                placeholder="Ex: Centro, Vila Nova..."
                className={inputClass}
              />
            </div>

            <div className="space-y-1.5">
              <label className="block text-sm font-medium text-slate-700">Quartos (Mínimo)</label>
              <input
                type="number"
                name="quartos_min"
                value={formData.preferencias?.quartos_min || ''}
                onChange={handlePrefChange}
                placeholder="Ex: 3"
                className={inputClass}
              />
            </div>
          </div>
        </section>

        {/* ── ANOTAÇÕES ── */}
        <section>
          <SectionHeader icon={<FileText className="w-4 h-4" />} title="Anotações Iniciais" />
          <div className="space-y-1.5">
            <textarea
              name="notas"
              value={formData.notas || ''}
              onChange={handleChange}
              rows={4}
              placeholder="Ex: Cliente tem 2 filhos, prefere apartamento na zona sul, limite de R$ 500.000..."
              className={`${inputClass} resize-y`}
            />
          </div>
        </section>

      </form>

      {/* TIMELINE (Aparece apenas ao editar um Lead existente) */}
      {isEditing && id && (
        <div className="p-6 sm:p-8 pt-0 border-t border-slate-100">
          <TimelineLead clienteId={id} />
        </div>
      )}

    </CrudModal>
  );
};
