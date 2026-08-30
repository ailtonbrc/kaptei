import { create } from 'zustand';
import { persist } from 'zustand/middleware';

// Remove a sessão legada que armazenava JWT no navegador.
localStorage.removeItem('kaptei-auth-storage');

export interface UsuarioSessao {
  id: string;
  nome: string;
  email: string;
  papel: 'SUPER_ADMIN' | 'GESTOR' | 'CORRETOR_EQUIPE' | 'CORRETOR_SOLO';
  conta_id: string;
  avatar: string | null;
  status_plano: string;
  plano: string;
  trial_vence_em: string | null;
}

interface AuthState {
  user: UsuarioSessao | null;
  isAuthenticated: boolean;
  login: (user: UsuarioSessao) => void;
  atualizarConta: (dados: Pick<UsuarioSessao, 'status_plano' | 'plano' | 'trial_vence_em'>) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      isAuthenticated: false,
      login: (user) => set({ user, isAuthenticated: true }),
      atualizarConta: (dados) => set((state) => ({ user: state.user ? { ...state.user, ...dados } : null })),
      logout: () => set({ user: null, isAuthenticated: false }),
    }),
    {
      name: 'kaptei-sessao',
      partialize: (state) => ({ user: state.user, isAuthenticated: state.isAuthenticated }),
    },
  ),
);
