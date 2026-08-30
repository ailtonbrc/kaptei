import React, { useState } from 'react';
import { Lock, Save, ArrowLeft } from 'lucide-react';
import { useNavigate, Link } from 'react-router-dom';
import { api } from '../../services/api';
import { useAuthStore } from '../../store/useAuthStore';
import { PasswordInput } from '../../components';
import { obterMensagemErro } from '../../lib/http/erro-api';

export const ChangePassword: React.FC = () => {
  const [senhaAtual, setSenhaAtual] = useState('');
  const [novaSenha, setNovaSenha] = useState('');
  const [confirmarNovaSenha, setConfirmarNovaSenha] = useState('');
  const [loading, setLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState('');
  const [successMsg, setSuccessMsg] = useState('');
  
  const navigate = useNavigate();
  const { logout } = useAuthStore();

  const handleUpdatePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMsg('');
    setSuccessMsg('');

    if (novaSenha.length < 6) {
      setErrorMsg('A nova senha deve ter no mínimo 6 caracteres.');
      return;
    }

    if (novaSenha !== confirmarNovaSenha) {
      setErrorMsg('A confirmação da nova senha não coincide. Verifique e tente novamente.');
      return;
    }

    setLoading(true);
    try {
      await api.put('/v1/usuarios/senha', {
        senha_atual: senhaAtual,
        nova_senha: novaSenha
      });

      setSuccessMsg('Senha alterada com sucesso! Você será redirecionado para o login...');
      
      // Aguarda 2 segundos para o usuário ler a mensagem, depois faz o logout
      setTimeout(async () => {
        await api.post('/auth/logout');
        logout();
        navigate('/login');
      }, 2000);

    } catch (error: unknown) {
      setErrorMsg(obterMensagemErro(error, 'Erro ao alterar a senha. Verifique a senha atual e tente novamente.'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-xl mx-auto">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900 flex items-center gap-2">
            <Lock className="text-blue-600" /> Trocar Senha
          </h1>
          <p className="text-slate-500 mt-1">Atualize a sua senha de acesso ao sistema.</p>
        </div>
        <Link 
          to="/app/perfil"
          className="flex items-center gap-2 px-4 py-2 text-sm font-medium text-slate-600 bg-white border border-slate-200 rounded-lg hover:bg-slate-50 hover:text-slate-900 transition-colors"
        >
          <ArrowLeft size={16} />
          Voltar ao Perfil
        </Link>
      </div>

      <div className="bg-white rounded-2xl shadow-sm border border-slate-200 overflow-hidden">
        <form onSubmit={handleUpdatePassword} className="p-6 sm:p-8">
          
          {errorMsg && (
            <div className="mb-6 p-4 bg-red-50 text-red-700 text-sm rounded-xl border border-red-100 flex items-start gap-3">
              <svg className="w-5 h-5 mt-0.5 text-red-500 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span>{errorMsg}</span>
            </div>
          )}

          {successMsg && (
            <div className="mb-6 p-4 bg-green-50 text-green-700 text-sm rounded-xl border border-green-100 flex items-start gap-3">
              <svg className="w-5 h-5 mt-0.5 text-green-500 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span>{successMsg}</span>
            </div>
          )}

          <div className="max-w-md space-y-6">
            <div>
              <label className="block text-sm font-semibold text-slate-900 mb-2">
                Senha Atual
              </label>
              <PasswordInput
                required
                value={senhaAtual}
                onChange={(e) => setSenhaAtual(e.target.value)}
                className="w-full px-4 py-2.5 border border-slate-300 rounded-xl focus:ring-2 focus:ring-blue-600/50 focus:border-blue-600 bg-slate-50 text-slate-900 transition-colors"
                placeholder="Sua senha atual"
                disabled={loading || !!successMsg}
              />
            </div>

            <div className="pt-4 border-t border-slate-100">
              <label className="block text-sm font-semibold text-slate-900 mb-2">
                Nova Senha
              </label>
              <PasswordInput
                required
                minLength={6}
                maxLength={72}
                value={novaSenha}
                onChange={(e) => setNovaSenha(e.target.value)}
                className="w-full px-4 py-2.5 border border-slate-300 rounded-xl focus:ring-2 focus:ring-blue-600/50 focus:border-blue-600 bg-slate-50 text-slate-900 transition-colors"
                placeholder="Digite a nova senha"
                disabled={loading || !!successMsg}
              />
              <p className="mt-1.5 text-xs text-slate-500">A senha deve ter no mínimo 6 caracteres.</p>
            </div>

            <div>
              <label className="block text-sm font-semibold text-slate-900 mb-2">
                Confirmar Nova Senha
              </label>
              <PasswordInput
                required
                minLength={6}
                maxLength={72}
                value={confirmarNovaSenha}
                onChange={(e) => setConfirmarNovaSenha(e.target.value)}
                className="w-full px-4 py-2.5 border border-slate-300 rounded-xl focus:ring-2 focus:ring-blue-600/50 focus:border-blue-600 bg-slate-50 text-slate-900 transition-colors"
                placeholder="Repita a nova senha"
                disabled={loading || !!successMsg}
              />
            </div>
          </div>

          <div className="mt-8 pt-6 border-t border-slate-100">
            <button
              type="submit"
              disabled={loading || !!successMsg}
              className="px-6 py-2.5 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-xl flex items-center gap-2 transition-colors disabled:opacity-50 disabled:cursor-not-allowed shadow-sm shadow-blue-600/20"
            >
              <Save size={18} />
              {loading ? 'Salvando...' : 'Alterar Senha'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
