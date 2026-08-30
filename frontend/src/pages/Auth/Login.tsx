import React, { useState, useEffect } from 'react';
import { GoogleOAuthProvider, GoogleLogin } from '@react-oauth/google';
import type { CredentialResponse } from '@react-oauth/google';
import { useNavigate, Link } from 'react-router-dom';
import { Mail, Lock, LogIn } from 'lucide-react';
import { api } from '../../services/api';
import { useAuthStore } from '../../store/useAuthStore';
import { PasswordInput } from '../../components';
import { obterMensagemErro } from '../../lib/http/erro-api';
import type { UsuarioSessao } from '../../store/useAuthStore';

export const Login: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [googleClientId, setGoogleClientId] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [errorMsg, setErrorMsg] = useState('');
  const navigate = useNavigate();
  const setLogin = useAuthStore((state) => state.login);

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

  const processLoginSuccess = (usuario: UsuarioSessao) => {
    setLogin(usuario);
    navigate('/app');
  };

  const handleStandardLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setErrorMsg('');

    // Permite login com atalho: "admin" é expandido para "admin@msdev.com.br"
    const emailEnvio = email.includes('@') ? email.trim() : `${email.trim()}@msdev.com.br`;

    try {
      const response = await api.post('/auth/login', { email: emailEnvio, senha: password });
      processLoginSuccess(response.data);
    } catch (error: unknown) {
      setErrorMsg(obterMensagemErro(error, 'Erro ao fazer login. Verifique as credenciais.'));
    } finally {
      setLoading(false);
    }
  };

  const handleGoogleSuccess = async (credentialResponse: CredentialResponse) => {
    setLoading(true);
    setErrorMsg('');
    try {
      const response = await api.post('/auth/google', { token: credentialResponse.credential });
      processLoginSuccess(response.data);
    } catch (error: unknown) {
      setErrorMsg(obterMensagemErro(error, 'Erro na autenticação com o Google.'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-slate-950 flex items-center justify-center p-4 sm:p-6 md:p-8">
      
      <div className="flex flex-col md:flex-row max-w-4xl w-full min-h-[550px] shadow-2xl rounded-3xl overflow-hidden bg-white">
        
        <div className="md:w-1/2 bg-slate-900 p-8 md:p-12 flex flex-col justify-between items-center text-center relative border-r border-slate-800">
          <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_center,_var(--tw-gradient-stops))] from-blue-900/20 to-transparent pointer-events-none" />
          
          <div className="my-auto flex flex-col items-center z-10">
            <div className="flex items-center justify-center gap-3 mb-6 transition-transform hover:scale-105 duration-300 cursor-default">
              <div className="bg-gradient-to-br from-blue-500 to-blue-700 p-2.5 rounded-xl shadow-lg shadow-blue-900/50 border border-blue-400/20">
                <svg className="w-8 h-8 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
                </svg>
              </div>
              <h1 className="text-4xl font-extrabold text-transparent bg-clip-text bg-gradient-to-r from-white to-blue-100 tracking-tight m-0">KAPTEI</h1>
            </div>
            
            <h2 className="text-3xl text-white font-bold tracking-tight m-0">Portal de Vendas</h2>
            
            <p className="text-slate-400 mt-3 max-w-[280px] text-sm m-0">
              CRM inteligente e gestão completa para corretores e imobiliárias de alto nível.
            </p>
            
            <div className="mt-6 inline-flex px-3 py-1 rounded-full bg-blue-600/20 text-blue-400 text-xs font-semibold tracking-wider border border-blue-500/20">
              v1.0.0-beta
            </div>
          </div>
          <div className="text-slate-500 text-xs mt-4 z-10">
            Kaptei &copy; 2026
          </div>
        </div>

        <div className="md:w-1/2 p-8 md:p-12 flex flex-col justify-between bg-white">
          <div className="my-auto">
            <div className="mb-8">
              <h3 className="text-2xl text-slate-900 font-semibold m-0">Bem-vindo de volta!</h3>
              <p className="text-slate-500 text-sm mt-1">Faça login para acessar o sistema.</p>
            </div>

            {errorMsg && (
              <div className="mb-4 p-3 bg-red-50 text-red-600 text-sm rounded-lg border border-red-100">
                {errorMsg}
              </div>
            )}

            <form onSubmit={handleStandardLogin} className="space-y-4">
              <div className="relative">
                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                  <Mail className="h-5 w-5 text-slate-400" />
                </div>
                <input
                  type="text"
                  required
                  autoComplete="username"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="block w-full pl-10 pr-3 py-3 border border-slate-200 rounded-xl bg-slate-50 text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-600/50 focus:border-blue-600 transition-colors"
                  placeholder="E-mail ou usuário (ex: admin)"
                />
              </div>

              <div className="relative mb-6">
                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none z-10">
                  <Lock className="h-5 w-5 text-slate-400" />
                </div>
                <PasswordInput
                  required
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="block w-full pl-10 py-3 border border-slate-200 rounded-xl bg-slate-50 text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-600/50 focus:border-blue-600 transition-colors"
                  placeholder="Senha"
                />
              </div>
              <div className="flex justify-end -mt-4 mb-4">
                <Link to="/esqueci-senha" className="text-sm font-semibold text-blue-600 hover:text-blue-700 transition-colors">
                  Esqueci a senha?
                </Link>
              </div>

              <button 
                type="submit" 
                disabled={loading}
                className="w-full h-12 text-sm font-semibold rounded-xl bg-blue-600 hover:bg-blue-700 text-white border-0 flex items-center justify-center gap-2 transition-colors disabled:opacity-70 disabled:cursor-not-allowed"
              >
                <LogIn className="w-4 h-4" />
                {loading ? 'Entrando...' : 'Entrar no Sistema'}
              </button>
            </form>

            {googleClientId && googleClientId !== "" && !googleClientId.includes("COLOQUE_SEU_CLIENT") && (
              <>
                <div className="relative my-6">
                  <div className="absolute inset-0 flex items-center">
                    <div className="w-full border-t border-slate-200"></div>
                  </div>
                  <div className="relative flex justify-center text-sm">
                    <span className="px-2 bg-white text-slate-400 uppercase tracking-wide text-xs font-medium">ou login social</span>
                  </div>
                </div>

                <div className="flex justify-center mb-6">
                  <GoogleOAuthProvider clientId={googleClientId}>
                    <GoogleLogin
                      onSuccess={handleGoogleSuccess}
                      onError={() => setErrorMsg('O Login com o Google falhou.')}
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

          <div className="text-center mt-4 border-t border-slate-100 pt-4">
            <p className="text-slate-500 text-sm">
              Ainda não tem conta? <Link to="/register" className="text-blue-600 font-semibold hover:text-blue-700">Cadastre-se grátis</Link>
            </p>
          </div>
        </div>

      </div>
    </div>
  );
};
