import React, { useEffect, useState } from 'react';
import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { api } from '../services/api';
import type { UsuarioSessao } from '../store/useAuthStore';
import { useAuthStore } from '../store/useAuthStore';

export const ProtectedRoute: React.FC = () => {
  const { isAuthenticated, login, logout, user } = useAuthStore();
  const [validandoSessao, setValidandoSessao] = useState(true);
  const location = useLocation();

  useEffect(() => {
    let ativo = true;

    api
      .get<UsuarioSessao>('/v1/sessao')
      .then(({ data }) => {
        if (ativo) login(data);
      })
      .catch(() => {
        if (ativo) logout();
      })
      .finally(() => {
        if (ativo) setValidandoSessao(false);
      });

    return () => {
      ativo = false;
    };
  }, [login, logout]);

  if (validandoSessao) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-slate-50 text-sm text-slate-600" role="status">
        Validando sua sessão...
      </div>
    );
  }

  if (!isAuthenticated || !user) {
    return <Navigate to="/login" replace />;
  }

  // Mantém as jornadas de cobrança acessíveis mesmo quando o restante da conta está bloqueado.
  if (user.status_plano === 'AGUARDANDO_PAGAMENTO' && !['/app/assinatura', '/checkout/sucesso'].includes(location.pathname)) {
    return <Navigate to="/app/assinatura" replace />;
  }

  // Bloqueia se o Trial expirou
  if (user.status_plano === 'TRIAL' && user.trial_vence_em) {
    const dataVencimento = new Date(user.trial_vence_em);
    const agora = new Date();
    
    if (agora > dataVencimento && location.pathname !== '/app/assinatura') {
      return <Navigate to="/app/assinatura" replace />;
    }
  }

  return <Outlet />;
};
