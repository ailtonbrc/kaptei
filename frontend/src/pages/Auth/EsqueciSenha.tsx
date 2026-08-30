import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { Mail, ArrowLeft, Send } from 'lucide-react';
import { api } from '../../services/api';
import { obterMensagemErro } from '../../lib/http/erro-api';

export const EsqueciSenha: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [email, setEmail] = useState('');
  const [mensagem, setMensagem] = useState('');
  const [erro, setErro] = useState('');

  const handleSolicitar = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setMensagem('');
    setErro('');

    try {
      const response = await api.post('/auth/esqueci-senha', { email });
      setMensagem(response.data.mensagem || 'Link de recuperação enviado com sucesso!');
    } catch (error: unknown) {
      setErro(obterMensagemErro(error, 'Erro ao solicitar recuperação de senha.'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-slate-950 flex items-center justify-center p-4 sm:p-6 md:p-8">
      <div className="max-w-md w-full bg-white p-8 md:p-10 rounded-3xl shadow-2xl">
        
        <div className="mb-8 text-center">
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-blue-50 text-blue-600 mb-4">
            <Mail className="w-8 h-8" />
          </div>
          <h2 className="text-2xl font-bold text-slate-900">Esqueceu a senha?</h2>
          <p className="text-slate-500 mt-2 text-sm">
            Digite seu e-mail abaixo e enviaremos instruções para redefinir sua senha.
          </p>
        </div>

        {mensagem && (
          <div className="mb-6 p-4 bg-green-50 border border-green-100 text-green-700 rounded-xl text-sm text-center font-medium">
            {mensagem}
          </div>
        )}

        {erro && (
          <div className="mb-6 p-4 bg-red-50 border border-red-100 text-red-700 rounded-xl text-sm text-center font-medium">
            {erro}
          </div>
        )}

        <form onSubmit={handleSolicitar} className="space-y-6">
          <div className="relative">
            <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
              <Mail className="h-5 w-5 text-slate-400" />
            </div>
            <input
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="block w-full pl-11 pr-4 py-3.5 border border-slate-200 rounded-xl bg-slate-50 text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-600/50 focus:border-blue-600 transition-colors"
              placeholder="Digite seu e-mail cadastrado"
            />
          </div>

          <button
            type="submit"
            disabled={loading || !email}
            className="w-full h-12 text-sm font-semibold rounded-xl bg-blue-600 hover:bg-blue-700 text-white border-0 flex items-center justify-center gap-2 transition-colors disabled:opacity-70 disabled:cursor-not-allowed"
          >
            {loading ? 'Enviando...' : (
              <>
                Enviar link <Send className="w-4 h-4" />
              </>
            )}
          </button>
        </form>

        <div className="mt-8 text-center">
          <Link to="/login" className="inline-flex items-center justify-center gap-2 text-sm font-semibold text-slate-500 hover:text-slate-700 transition-colors">
            <ArrowLeft className="w-4 h-4" /> Voltar para o Login
          </Link>
        </div>

      </div>
    </div>
  );
};
