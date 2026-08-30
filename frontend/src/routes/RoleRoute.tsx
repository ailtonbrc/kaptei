import { Navigate, Outlet } from 'react-router-dom';
import type { UsuarioSessao } from '../store/useAuthStore';
import { useAuthStore } from '../store/useAuthStore';

interface RoleRouteProps {
  papeis: UsuarioSessao['papel'][];
}

export const RoleRoute = ({ papeis }: RoleRouteProps) => {
  const usuario = useAuthStore((estado) => estado.user);
  if (!usuario || !papeis.includes(usuario.papel)) {
    return <Navigate to="/app" replace />;
  }
  return <Outlet />;
};
