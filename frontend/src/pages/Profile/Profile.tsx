import React, { useState, useEffect, useRef } from 'react';
import { api } from '../../services/api';
import { User, MapPin, Briefcase, RefreshCw, CheckCircle, Lock } from 'lucide-react';
import { Link } from 'react-router-dom';
import { SaveButton, CpfInput, CepInput, PhoneInput } from '../../components';
import { validarCPF } from '../../lib/validacoes/cpf';

export const Profile: React.FC = () => {
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [successMsg, setSuccessMsg] = useState('');
  const [errorMsg, setErrorMsg] = useState('');
  const scrollContainerRef = useRef<HTMLDivElement>(null);

  const [formData, setFormData] = useState({
    nome_completo: '',
    cpf: '',
    rg: '',
    rg_estado: 'MS',
    rg_orgao_expedidor: 'SSP',
    nacionalidade: 'Brasileira',
    estado_civil: 'Solteiro(a)',
    telefone: '',
    numero_whatsapp: '',
    creci: '',
    creci_estado: 'MS',
    cep: '',
    logradouro: '',
    numero: '',
    complemento: '',
    bairro: '',
    cidade: '',
    estado: 'MS',
  });

  useEffect(() => {
    const fetchProfile = async () => {
      try {
        const response = await api.get('/v1/me');
        const user = response.data;
        setFormData({
          nome_completo: user.nome_completo || '',
          cpf: user.cpf || '',
          rg: user.rg || '',
          rg_estado: user.rg_estado || 'MS',
          rg_orgao_expedidor: user.rg_orgao_expedidor || 'SSP',
          nacionalidade: user.nacionalidade || 'Brasileira',
          estado_civil: user.estado_civil || 'Solteiro(a)',
          telefone: user.telefone || '',
          numero_whatsapp: user.numero_whatsapp || '',
          creci: user.creci || '',
          creci_estado: user.creci_estado || 'MS',
          cep: user.cep || '',
          logradouro: user.logradouro || '',
          numero: user.numero || '',
          complemento: user.complemento || '',
          bairro: user.bairro || '',
          cidade: user.cidade || '',
          estado: user.estado || 'MS',
        });
      } catch {
        setErrorMsg('Erro ao carregar os dados do perfil.');
      } finally {
        setLoading(false);
      }
    };
    fetchProfile();
  }, []);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  const handleAddressFetch = (data: { logradouro: string; bairro: string; cidade: string; estado: string; }) => {
    setFormData(prev => ({
      ...prev,
      logradouro: data.logradouro || prev.logradouro,
      bairro: data.bairro || prev.bairro,
      cidade: data.cidade || prev.cidade,
      estado: data.estado || prev.estado,
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    // Validação de CPF antes de salvar
    if (formData.cpf) {
      const rawCpf = formData.cpf.replace(/\D/g, '');
      if (rawCpf.length > 0 && rawCpf.length !== 11) {
        setErrorMsg('Por favor, preencha o CPF completamente com 11 dígitos.');
        setTimeout(() => setErrorMsg(''), 5000);
        return;
      }
      if (rawCpf.length === 11 && !validarCPF(rawCpf)) {
        setErrorMsg('O CPF informado é inválido. Corrija antes de salvar.');
        setTimeout(() => setErrorMsg(''), 5000);
        return;
      }
    }

    setSaving(true);
    setSuccessMsg('');
    setErrorMsg('');

    try {
      await api.put('/v1/usuarios/perfil', formData);
      setSuccessMsg('Perfil atualizado com sucesso!');
      setTimeout(() => setSuccessMsg(''), 4000); // Some após 4 segundos
    } catch {
      setErrorMsg('Erro ao atualizar perfil. Verifique os dados e tente novamente.');
      setTimeout(() => setErrorMsg(''), 5000); // Some após 5 segundos
    } finally {
      setSaving(false);
    }
  };

  const handleAddressFocus = () => {
    if (scrollContainerRef.current) {
      scrollContainerRef.current.scrollTo({
        top: scrollContainerRef.current.scrollHeight,
        behavior: 'smooth'
      });
    }
  };

  const handlePersonalFocus = () => {
    if (scrollContainerRef.current) {
      scrollContainerRef.current.scrollTo({
        top: 0,
        behavior: 'smooth'
      });
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <RefreshCw className="w-8 h-8 text-blue-600 animate-spin" />
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto p-4 sm:p-6 lg:p-8">
      <div className="mb-8 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">Meu Perfil</h1>
          <p className="text-slate-500 mt-1">
            Complete suas informações. Estes dados são necessários para gerar contratos e documentos legalmente válidos.
          </p>
        </div>
        <Link 
          to="/app/trocar-senha"
          className="shrink-0 flex items-center gap-2 px-4 py-2 text-sm font-medium text-slate-600 bg-white border border-slate-200 rounded-lg hover:bg-slate-50 hover:text-slate-900 transition-colors"
        >
          <Lock size={16} />
          Trocar Senha
        </Link>
      </div>

      {/* Notificações Flutuantes (Toasts) */}
      {(successMsg || errorMsg) && (
        <div className="fixed top-20 left-4 md:left-72 z-50 flex flex-col gap-3 animate-in fade-in slide-in-from-top-4 duration-300">
          {successMsg && (
            <div className="p-4 bg-emerald-50 border border-emerald-200 shadow-lg rounded-xl flex items-center gap-3 text-emerald-800 min-w-[300px]">
              <CheckCircle className="w-5 h-5" />
              <p className="font-medium">{successMsg}</p>
            </div>
          )}
          {errorMsg && (
            <div className="p-4 bg-red-50 border border-red-200 shadow-lg rounded-xl text-red-700 min-w-[300px]">
              <p className="font-medium">{errorMsg}</p>
            </div>
          )}
        </div>
      )}

      <div className="bg-white rounded-2xl shadow-sm border border-slate-200 overflow-hidden flex flex-col max-h-[65vh]">
        
        {/* Scrollable Form Content */}
        <div ref={scrollContainerRef} className="overflow-y-auto p-6 sm:p-8 flex-1">
          <form id="profile-form" onSubmit={handleSubmit} className="space-y-10">
            
            {/* SEÇÃO 1: DADOS PESSOAIS */}
            <section onFocusCapture={handlePersonalFocus}>
              <div className="flex items-center gap-2 mb-4 border-b border-slate-100 pb-2">
                <User className="w-5 h-5 text-blue-600" />
                <h3 className="text-lg font-semibold text-slate-800">Dados Pessoais</h3>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
                <div className="space-y-1.5 md:col-span-2">
                  <label className="block text-sm font-medium text-slate-700">Nome Completo<span className="text-red-500 ml-1">*</span></label>
                  <input type="text" name="nome_completo" value={formData.nome_completo} onChange={handleChange} required className="w-full px-3 py-2 border border-slate-200 rounded-lg focus:ring-2 focus:ring-blue-600/50" />
                </div>
                <div className="md:col-span-2 grid grid-cols-1 md:grid-cols-12 gap-5">
                  {/* CPF - 3 colunas */}
                  <div className="md:col-span-3">
                    <CpfInput 
                      name="cpf" 
                      value={formData.cpf} 
                      onChange={handleChange} 
                      required
                    />
                  </div>
                  {/* RG - 4 colunas */}
                  <div className="space-y-1.5 md:col-span-4">
                    <label className="block text-sm font-medium text-slate-700">RG<span className="text-red-500 ml-1">*</span></label>
                    <input type="text" name="rg" value={formData.rg} onChange={handleChange} required className="w-full px-3 py-2 border border-slate-200 rounded-lg focus:ring-2 focus:ring-blue-600/50" />
                  </div>
                  {/* Órgão Expedidor - 3 colunas */}
                  <div className="space-y-1.5 md:col-span-3">
                    <label className="block text-sm font-medium text-slate-700">Órgão Expedidor<span className="text-red-500 ml-1">*</span></label>
                    <input type="text" name="rg_orgao_expedidor" value={formData.rg_orgao_expedidor} onChange={handleChange} required placeholder="Ex: SSP" className="w-full px-3 py-2 border border-slate-200 rounded-lg focus:ring-2 focus:ring-blue-600/50" />
                  </div>
                  {/* Estado (UF) - 2 colunas */}
                  <div className="space-y-1.5 md:col-span-2">
                    <label className="block text-sm font-medium text-slate-700">Estado (UF)<span className="text-red-500 ml-1">*</span></label>
                    <select name="rg_estado" value={formData.rg_estado} onChange={handleChange} required className="w-full px-3 py-2 border border-slate-200 rounded-lg focus:ring-2 focus:ring-blue-600/50 bg-white">
                      <option value="AC">AC</option><option value="AL">AL</option><option value="AP">AP</option>
                      <option value="AM">AM</option><option value="BA">BA</option><option value="CE">CE</option>
                      <option value="DF">DF</option><option value="ES">ES</option><option value="GO">GO</option>
                      <option value="MA">MA</option><option value="MT">MT</option><option value="MS">MS</option>
                      <option value="MG">MG</option><option value="PA">PA</option><option value="PB">PB</option>
                      <option value="PR">PR</option><option value="PE">PE</option><option value="PI">PI</option>
                      <option value="RJ">RJ</option><option value="RN">RN</option><option value="RS">RS</option>
                      <option value="RO">RO</option><option value="RR">RR</option><option value="SC">SC</option>
                      <option value="SP">SP</option><option value="SE">SE</option><option value="TO">TO</option>
                    </select>
                  </div>
                </div>
                <div className="space-y-1.5">
                  <label className="block text-sm font-medium text-slate-700">Estado Civil<span className="text-red-500 ml-1">*</span></label>
                  <select name="estado_civil" value={formData.estado_civil} onChange={handleChange} required className="w-full px-3 py-2 border border-slate-200 rounded-lg focus:ring-2 focus:ring-blue-600/50 bg-white">
                    <option>Solteiro(a)</option>
                    <option>Casado(a)</option>
                    <option>Divorciado(a)</option>
                    <option>Viúvo(a)</option>
                    <option>União Estável</option>
                  </select>
                </div>
                <div className="space-y-1.5">
                  <label className="block text-sm font-medium text-slate-700">Nacionalidade<span className="text-red-500 ml-1">*</span></label>
                  <input type="text" name="nacionalidade" value={formData.nacionalidade} onChange={handleChange} required className="w-full px-3 py-2 border border-slate-200 rounded-lg focus:ring-2 focus:ring-blue-600/50" />
                </div>
                <PhoneInput 
                  name="telefone" 
                  label="Telefone"
                  value={formData.telefone} 
                  onChange={handleChange} 
                />
                <PhoneInput 
                  name="numero_whatsapp" 
                  label="WhatsApp ou @Usuário"
                  value={formData.numero_whatsapp} 
                  onChange={handleChange} 
                  allowUsername={true}
                />
              </div>
            </section>

            {/* SEÇÃO 2: DADOS PROFISSIONAIS */}
            <section>
              <div className="flex items-center gap-2 mb-4 border-b border-slate-100 pb-2">
                <Briefcase className="w-5 h-5 text-blue-600" />
                <h3 className="text-lg font-semibold text-slate-800">Dados Profissionais</h3>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
                <div className="space-y-1.5">
                  <label className="block text-sm font-medium text-slate-700">Número do CRECI<span className="text-red-500 ml-1">*</span></label>
                  <input type="text" name="creci" value={formData.creci} onChange={handleChange} required className="w-full px-3 py-2 border border-slate-200 rounded-lg focus:ring-2 focus:ring-blue-600/50" />
                </div>
                <div className="space-y-1.5">
                  <label className="block text-sm font-medium text-slate-700">Estado (CRECI)<span className="text-red-500 ml-1">*</span></label>
                  <select name="creci_estado" value={formData.creci_estado} onChange={handleChange} required className="w-full px-3 py-2 border border-slate-200 rounded-lg focus:ring-2 focus:ring-blue-600/50 bg-white">
                    <option value="AC">Acre</option>
                    <option value="AL">Alagoas</option>
                    <option value="AP">Amapá</option>
                    <option value="AM">Amazonas</option>
                    <option value="BA">Bahia</option>
                    <option value="CE">Ceará</option>
                    <option value="DF">Distrito Federal</option>
                    <option value="ES">Espírito Santo</option>
                    <option value="GO">Goiás</option>
                    <option value="MA">Maranhão</option>
                    <option value="MT">Mato Grosso</option>
                    <option value="MS">Mato Grosso do Sul</option>
                    <option value="MG">Minas Gerais</option>
                    <option value="PA">Pará</option>
                    <option value="PB">Paraíba</option>
                    <option value="PR">Paraná</option>
                    <option value="PE">Pernambuco</option>
                    <option value="PI">Piauí</option>
                    <option value="RJ">Rio de Janeiro</option>
                    <option value="RN">Rio Grande do Norte</option>
                    <option value="RS">Rio Grande do Sul</option>
                    <option value="RO">Rondônia</option>
                    <option value="RR">Roraima</option>
                    <option value="SC">Santa Catarina</option>
                    <option value="SP">São Paulo</option>
                    <option value="SE">Sergipe</option>
                    <option value="TO">Tocantins</option>
                  </select>
                </div>
              </div>
            </section>

            {/* SEÇÃO 3: ENDEREÇO */}
            <section onFocusCapture={handleAddressFocus}>
              <div className="flex items-center gap-2 mb-4 border-b border-slate-100 pb-2">
                <MapPin className="w-5 h-5 text-blue-600" />
                <h3 className="text-lg font-semibold text-slate-800">Endereço</h3>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-4 gap-5">
                <div className="md:col-span-1">
                  <CepInput 
                    name="cep" 
                    value={formData.cep} 
                    onChange={handleChange} 
                    onAddressFetch={handleAddressFetch}
                    required
                  />
                </div>
                <div className="space-y-1.5 md:col-span-3">
                  <label className="block text-sm font-medium text-slate-700">Logradouro (Rua, Av, etc)<span className="text-red-500 ml-1">*</span></label>
                  <input type="text" name="logradouro" value={formData.logradouro} onChange={handleChange} required className="w-full px-3 py-2 border border-slate-200 rounded-lg focus:ring-2 focus:ring-blue-600/50" />
                </div>
                <div className="space-y-1.5 md:col-span-1">
                  <label className="block text-sm font-medium text-slate-700">Número<span className="text-red-500 ml-1">*</span></label>
                  <input type="text" name="numero" value={formData.numero} onChange={handleChange} required className="w-full px-3 py-2 border border-slate-200 rounded-lg focus:ring-2 focus:ring-blue-600/50" />
                </div>
                <div className="space-y-1.5 md:col-span-3">
                  <label className="block text-sm font-medium text-slate-700">Complemento</label>
                  <input type="text" name="complemento" value={formData.complemento} onChange={handleChange} className="w-full px-3 py-2 border border-slate-200 rounded-lg focus:ring-2 focus:ring-blue-600/50" />
                </div>
                <div className="space-y-1.5 md:col-span-2">
                  <label className="block text-sm font-medium text-slate-700">Bairro<span className="text-red-500 ml-1">*</span></label>
                  <input type="text" name="bairro" value={formData.bairro} onChange={handleChange} required className="w-full px-3 py-2 border border-slate-200 rounded-lg focus:ring-2 focus:ring-blue-600/50" />
                </div>
                <div className="space-y-1.5 md:col-span-1">
                  <label className="block text-sm font-medium text-slate-700">Cidade<span className="text-red-500 ml-1">*</span></label>
                  <input type="text" name="cidade" value={formData.cidade} onChange={handleChange} required className="w-full px-3 py-2 border border-slate-200 rounded-lg focus:ring-2 focus:ring-blue-600/50" />
                </div>
                <div className="space-y-1.5 md:col-span-1">
                  <label className="block text-sm font-medium text-slate-700">Estado<span className="text-red-500 ml-1">*</span></label>
                  <input type="text" name="estado" value={formData.estado} onChange={handleChange} required className="w-full px-3 py-2 border border-slate-200 rounded-lg focus:ring-2 focus:ring-blue-600/50" />
                </div>
              </div>
            </section>
          </form>
        </div>

        {/* Footer actions fixed at bottom */}
        <div className="bg-slate-50 border-t border-slate-200 p-4 sm:px-8 flex justify-end shrink-0">
          <SaveButton isSaving={saving} form="profile-form" />
        </div>
      </div>
    </div>
  );
};
