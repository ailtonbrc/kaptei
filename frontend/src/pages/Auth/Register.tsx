import React, { useState, useEffect } from 'react';
import { GoogleOAuthProvider, GoogleLogin } from '@react-oauth/google';
import type { CredentialResponse } from '@react-oauth/google';
import { User, Users, RefreshCw } from 'lucide-react';
import { useNavigate, Link } from 'react-router-dom';
import { api } from '../../services/api';
import { useAuthStore } from '../../store/useAuthStore';
import { PlanoCard, PasswordInput } from '../../components';
import type { Plano } from '../../constants/planos';
import { obterMensagemErro } from '../../lib/http/erro-api';
import type { UsuarioSessao } from '../../store/useAuthStore';

export const Register: React.FC = () => {
  const [googleClientId, setGoogleClientId] = useState('');
  const [currentStep, setCurrentStep] = useState(0);
  const [perfil, setPerfil] = useState<'CORRETOR_SOLO' | 'IMOBILIARIA' | null>(null);
  const [planoSelecionado, setPlanoSelecionado] = useState<string | null>(null);
  const [errorMsg, setErrorMsg] = useState('');
  
  const [planos, setPlanos] = useState<Plano[]>([]);
  const [loadingPlanos, setLoadingPlanos] = useState(false);

  useEffect(() => {
    // Fetch google client ID from public config
    api.get('/public/configuracoes/GOOGLE_CLIENT_ID')
      .then(response => {
        if (response.data && response.data.valor) {
          setGoogleClientId(response.data.valor);
        }
      })
      .catch(err => {
        console.error("Não foi possível carregar o Google Client ID:", err);
      });
  }, []);

  useEffect(() => {
    if (currentStep === 1 && planos.length === 0) {
      const fetchPlanos = async () => {
        setLoadingPlanos(true);
        setErrorMsg(''); // Limpa mensagens anteriores
        try {
          const response = await api.get('/v1/planos');
          setPlanos(response.data);
        } catch (err) {
          console.error("Erro ao carregar planos", err);
          setErrorMsg('Não foi possível carregar os planos do servidor no momento. Verifique sua conexão ou tente novamente mais tarde.');
        } finally {
          setLoadingPlanos(false);
        }
      };
      fetchPlanos();
    }
  }, [currentStep, planos.length]);
  
  const [nome, setNome] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const setLogin = useAuthStore((state) => state.login);

  const processLoginSuccess = (usuario: UsuarioSessao) => {
    setLogin(usuario);
    navigate('/app');
  };

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMsg('');
    
    if (password !== confirmPassword) {
      setErrorMsg('As senhas não coincidem. Por favor, verifique e tente novamente.');
      return;
    }
    
    if (!perfil || !planoSelecionado) {
      setErrorMsg('Perfil ou Plano não selecionados!');
      return;
    }
    
    setLoading(true);
    try {
      const response = await api.post('/auth/register', {
        nome,
        email,
        senha: password,
        tipo_conta: perfil,
        plano: planoSelecionado,
      });
      processLoginSuccess(response.data);
    } catch (error: unknown) {
      setErrorMsg(obterMensagemErro(error, 'Erro ao criar conta. Verifique os dados.'));
    } finally {
      setLoading(false);
    }
  };

  const handleGoogleSuccess = async (credentialResponse: CredentialResponse) => {
    if (!perfil || !planoSelecionado) {
      setErrorMsg('Perfil ou Plano não selecionados!');
      return;
    }
    setLoading(true);
    setErrorMsg('');
    try {
      const response = await api.post('/auth/google', { 
        token: credentialResponse.credential,
        tipo_conta: perfil,
        plano: planoSelecionado
      });
      processLoginSuccess(response.data);
    } catch (error: unknown) {
      setErrorMsg(obterMensagemErro(error, 'Erro na autenticação com o Google.'));
    } finally {
      setLoading(false);
    }
  };

  const renderStep1 = () => (
    <div className="flex flex-col md:flex-row gap-6 mt-8">
      <div 
        className={`flex-1 flex flex-col items-center justify-center p-8 rounded-2xl cursor-pointer transition-all border-2 ${perfil === 'CORRETOR_SOLO' ? 'border-blue-600 bg-blue-50' : 'border-transparent bg-white shadow-sm hover:shadow-md'}`}
        onClick={() => setPerfil('CORRETOR_SOLO')}
      >
        <User className="text-blue-600 w-12 h-12 mb-4" />
        <h4 className="text-lg font-semibold text-slate-900 m-0">Corretor Autônomo</h4>
        <p className="text-slate-500 text-center text-sm mt-2">Trabalho sozinho e quero gerenciar meus imóveis e clientes.</p>
      </div>
      
      <div 
        className={`flex-1 flex flex-col items-center justify-center p-8 rounded-2xl cursor-pointer transition-all border-2 ${perfil === 'IMOBILIARIA' ? 'border-blue-600 bg-blue-50' : 'border-transparent bg-white shadow-sm hover:shadow-md'}`}
        onClick={() => setPerfil('IMOBILIARIA')}
      >
        <Users className="text-blue-600 w-12 h-12 mb-4" />
        <h4 className="text-lg font-semibold text-slate-900 m-0">Imobiliária / Gestor</h4>
        <p className="text-slate-500 text-center text-sm mt-2">Tenho uma equipe de corretores e preciso gerenciar a empresa.</p>
      </div>
    </div>
  );

  const renderStep2 = () => {
    const tipoFiltro = perfil === 'CORRETOR_SOLO' ? 'CORRETOR' : 'IMOBILIARIA';
    const planosFiltrados = planos.filter(p => p.tipo === tipoFiltro);
    
    if (loadingPlanos) {
      return (
        <div className="flex justify-center mt-12 mb-8">
          <RefreshCw className="w-8 h-8 animate-spin text-blue-600" />
        </div>
      );
    }

    return (
      <div className="flex flex-col items-center mt-8">
        {errorMsg && (
          <div className="mb-6 p-4 w-full max-w-2xl bg-red-50 text-red-600 text-sm rounded-lg border border-red-100 text-center">
            {errorMsg}
          </div>
        )}
        <div className="flex flex-col lg:flex-row gap-6 justify-center items-stretch w-full">
          {planosFiltrados.map((p) => (
            <PlanoCard
              key={p.id}
              plano={p}
              selecionado={planoSelecionado === p.codigo}
              onClick={() => {
                setPlanoSelecionado(p.codigo);
                setCurrentStep(2);
              }}
            />
          ))}
        </div>
      </div>
    );
  };

  const renderStep3 = () => (
    <div className="mt-8 sm:mx-auto sm:w-full sm:max-w-md">
      <div className="bg-white p-8 shadow-sm rounded-2xl border border-slate-100">
        
        {errorMsg && (
          <div className="mb-4 p-3 bg-red-50 text-red-600 text-sm rounded-lg border border-red-100">
            {errorMsg}
          </div>
        )}

        <form onSubmit={handleRegister} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">Nome Completo</label>
            <input
              type="text"
              required
              value={nome}
              onChange={(e) => setNome(e.target.value)}
              className="block w-full px-3 py-2 border border-slate-200 rounded-lg bg-slate-50 text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-600/50 focus:border-blue-600"
              placeholder="Ex: João da Silva"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">E-mail</label>
            <input
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="block w-full px-3 py-2 border border-slate-200 rounded-lg bg-slate-50 text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-600/50 focus:border-blue-600"
              placeholder="seu@email.com"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">Senha</label>
            <PasswordInput
              required
              minLength={6}
              maxLength={72}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="block w-full px-3 py-2 border border-slate-200 rounded-lg bg-slate-50 text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-600/50 focus:border-blue-600"
              placeholder="••••••••"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-1">Confirmar Senha</label>
            <PasswordInput
              required
              minLength={6}
              maxLength={72}
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              className="block w-full px-3 py-2 border border-slate-200 rounded-lg bg-slate-50 text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-600/50 focus:border-blue-600"
              placeholder="Repita a sua senha"
            />
          </div>

          <div className="py-2">
            <div className="w-full border-t border-slate-200"></div>
          </div>

          <button 
            type="submit" 
            disabled={loading}
            className="w-full h-12 text-sm font-semibold rounded-xl bg-blue-600 hover:bg-blue-700 text-white border-0 flex items-center justify-center gap-2 transition-colors disabled:opacity-70 disabled:cursor-not-allowed"
          >
            {loading ? 'Criando...' : (planoSelecionado?.includes('TRIAL') ? 'Criar Conta Gratuita' : 'Criar Conta e Iniciar Trial')}
          </button>
        </form>

        {googleClientId && googleClientId !== "" && !googleClientId.includes("COLOQUE_SEU_CLIENT") && (
          <>
            <div className="relative my-6">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-slate-200"></div>
              </div>
              <div className="relative flex justify-center text-sm">
                <span className="px-2 bg-white text-slate-400 uppercase tracking-wide text-xs font-medium">ou cadastre-se com</span>
              </div>
            </div>

            <div className="flex justify-center mt-4">
              <GoogleOAuthProvider clientId={googleClientId}>
                <GoogleLogin
                  onSuccess={handleGoogleSuccess}
                  onError={() => setErrorMsg('O cadastro com o Google falhou.')}
                  shape="rectangular"
                  theme="outline"
                  size="large"
                  text="continue_with"
                  width="100%"
                />
              </GoogleOAuthProvider>
            </div>
          </>
        )}
      </div>
    </div>
  );

  const steps = ['Perfil', 'Plano', 'Seus Dados'];

  return (
    <div className="min-h-screen bg-slate-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-5xl mx-auto">
        <div className="text-center mb-8 flex flex-col items-center">
          <div className="flex items-center justify-center gap-3 mb-4 cursor-default">
            <div className="bg-gradient-to-br from-blue-500 to-blue-700 p-2.5 rounded-xl shadow-lg shadow-blue-900/50 border border-blue-400/20">
              <svg className="w-8 h-8 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
              </svg>
            </div>
            <h1 className="text-4xl font-extrabold text-transparent bg-clip-text bg-gradient-to-r from-blue-900 to-blue-600 tracking-tight m-0">KAPTEI</h1>
          </div>
          <h2 className="text-3xl font-bold tracking-tight text-slate-900 m-0">Crie sua Conta Kaptei</h2>

        </div>

        {/* Custom Steps Indicator */}
        <div className="max-w-2xl mx-auto mb-8">
          <div className="flex justify-between items-center relative">
            <div className="absolute top-1/2 left-0 right-0 h-0.5 bg-slate-200 -z-10 -translate-y-1/2"></div>
            {steps.map((title, index) => {
              const active = index === currentStep;
              const completed = index < currentStep;
              return (
                <div key={title} className="flex flex-col items-center bg-slate-50 px-2">
                  <div className={`w-8 h-8 flex items-center justify-center rounded-full text-sm font-semibold mb-2 transition-colors ${active ? 'bg-blue-600 text-white border-2 border-blue-600' : completed ? 'bg-blue-100 text-blue-600 border-2 border-blue-600' : 'bg-white text-slate-400 border-2 border-slate-200'}`}>
                    {index + 1}
                  </div>
                  <span className={`text-xs font-medium ${active ? 'text-blue-600' : 'text-slate-500'}`}>{title}</span>
                </div>
              );
            })}
          </div>
        </div>

        {currentStep === 0 && renderStep1()}
        {currentStep === 1 && renderStep2()}
        {currentStep === 2 && renderStep3()}

        <div className="mt-8 flex justify-center gap-4">
          {currentStep > 0 && (
            <button 
              onClick={() => setCurrentStep(prev => prev - 1)}
              className="px-6 py-2.5 rounded-lg border border-slate-300 text-slate-700 font-medium hover:bg-slate-100 transition-colors"
            >
              Voltar
            </button>
          )}
          {currentStep === 0 && (
            <button 
              disabled={!perfil}
              onClick={() => setCurrentStep(1)}
              className="px-6 py-2.5 rounded-lg bg-blue-600 text-white font-medium hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Avançar para Planos
            </button>
          )}
        </div>

        {currentStep === 0 && (
          <div className="text-center mt-12">
            <p className="text-sm text-slate-500">
              Já tem uma conta? <Link to="/login" className="text-blue-600 font-semibold hover:text-blue-700">Faça Login</Link>
            </p>
          </div>
        )}
      </div>
    </div>
  );
};
