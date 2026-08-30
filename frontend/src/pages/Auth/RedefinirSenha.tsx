import React, { useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { Lock, ArrowRight, CheckCircle2 } from 'lucide-react';
import { api } from '../../services/api';
import { PasswordInput } from '../../components';
import { obterMensagemErro } from '../../lib/http/erro-api';

export const RedefinirSenha: React.FC = () => {
  const [searchParams] = useSearchParams();
  const token = searchParams.get('token');


  const [loading, setLoading] = useState(false);
  const [senha, setSenha] = useState('');
  const [confirmarSenha, setConfirmarSenha] = useState('');

  const [erro, setErro] = useState(() => token ? '' : 'Token de recuperação não fornecido na URL.');
  const [sucesso, setSucesso] = useState(false);

  const handleRedefinir = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (senha !== confirmarSenha) {
      setErro('As senhas não coincidem.');
      return;
    }

    if (senha.length < 6) {
      setErro('A senha deve ter pelo menos 6 caracteres.');
      return;
    }

    setLoading(true);

    setErro('');

    try {
      await api.post('/auth/redefinir-senha', { token, senha });
      setSucesso(true);
    } catch (error: unknown) {
      setErro(obterMensagemErro(error, 'Erro ao redefinir a senha.'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-slate-950 flex items-center justify-center p-4 sm:p-6 md:p-8">
      <div className="max-w-md w-full bg-white p-8 md:p-10 rounded-3xl shadow-2xl">
        
        <div className="mb-8 text-center">
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-blue-50 text-blue-600 mb-4">
            {sucesso ? <CheckCircle2 className="w-8 h-8 text-green-500" /> : <Lock className="w-8 h-8" />}
          </div>
          <h2 className="text-2xl font-bold text-slate-900">
            {sucesso ? 'Tudo certo!' : 'Redefinir Senha'}
          </h2>
          <p className="text-slate-500 mt-2 text-sm">
            {sucesso 
              ? 'Sua senha foi atualizada. Você já pode acessar sua conta.' 
              : 'Crie uma nova senha segura para sua conta Kaptei.'}
          </p>
        </div>

        {erro && (
          <div className="mb-6 p-4 bg-red-50 border border-red-100 text-red-700 rounded-xl text-sm text-center font-medium">
            {erro}
          </div>
        )}

        {!sucesso ? (
          <form onSubmit={handleRedefinir} className="space-y-4">
            <div className="relative">
              <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none z-10">
                <Lock className="h-5 w-5 text-slate-400" />
              </div>
              <PasswordInput
                required
                value={senha}
                onChange={(e) => setSenha(e.target.value)}
                className="block w-full pl-10 py-3 border border-slate-200 rounded-xl bg-slate-50 text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-600/50 focus:border-blue-600 transition-colors"
                placeholder="Nova Senha"
              />
            </div>

            <div className="relative pb-2">
              <div className="absolute inset-y-0 left-0 pl-3 top-0 bottom-2 flex items-center pointer-events-none z-10">
                <Lock className="h-5 w-5 text-slate-400" />
              </div>
              <PasswordInput
                required
                value={confirmarSenha}
                onChange={(e) => setConfirmarSenha(e.target.value)}
                className="block w-full pl-10 py-3 border border-slate-200 rounded-xl bg-slate-50 text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-600/50 focus:border-blue-600 transition-colors"
                placeholder="Confirme a Nova Senha"
              />
            </div>

            <button
              type="submit"
              disabled={loading || !token || !senha || !confirmarSenha}
              className="w-full h-12 text-sm font-semibold rounded-xl bg-blue-600 hover:bg-blue-700 text-white border-0 flex items-center justify-center gap-2 transition-colors disabled:opacity-70 disabled:cursor-not-allowed"
            >
              {loading ? 'Salvando...' : 'Salvar Nova Senha'}
            </button>
          </form>
        ) : (
          <Link
            to="/login"
            className="w-full h-12 text-sm font-semibold rounded-xl bg-green-600 hover:bg-green-700 text-white flex items-center justify-center gap-2 transition-colors"
          >
            Ir para o Login <ArrowRight className="w-4 h-4" />
          </Link>
        )}

      </div>
    </div>
  );
};
