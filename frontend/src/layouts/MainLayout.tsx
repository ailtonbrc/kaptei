import React, { useEffect, useRef, useState } from 'react';
import { Outlet, useNavigate, useLocation, Link } from 'react-router-dom';
import { Menu, LayoutDashboard, Home, Users, Settings, LogOut, ChevronDown, User, Lock, Inbox, Calendar, UserRoundCog, MessageCircleMore, Scale, X } from 'lucide-react';
import { useAuthStore } from '../store/useAuthStore';
import { api } from '../services/api';

const rotaEstaAtiva = (caminhoAtual: string, caminhoItem: string) => caminhoAtual === caminhoItem || (caminhoItem !== '/app' && caminhoAtual.startsWith(`${caminhoItem}/`));

export const MainLayout: React.FC = () => {
  const [drawerVisible, setDrawerVisible] = useState(false);
  const drawerRef = useRef<HTMLDialogElement>(null);

  useEffect(() => {
    const drawer = drawerRef.current;
    if (!drawer) return;
    if (drawerVisible && !drawer.open) {
      drawer.showModal();
    } else if (!drawerVisible && drawer.open) {
      drawer.close();
    }
  }, [drawerVisible]);

  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout } = useAuthStore();

  const handleLogout = async () => {
    try {
      await api.post('/auth/logout');
    } finally {
      logout();
      navigate('/login');
    }
  };

  const getPapelName = (papel?: string) => {
    switch (papel) {
      case 'SUPER_ADMIN': return 'Administrador do Sistema';
      case 'GESTOR': return 'Imobiliária (Gestor)';
      case 'CORRETOR_EQUIPE': return 'Corretor Associado';
      case 'CORRETOR_SOLO': return 'Corretor Autônomo';
      default: return 'Conta Gratuita';
    }
  };

  const menuItems = [
    { key: '/app', icon: <LayoutDashboard size={20} />, label: 'Dashboard' },
    { key: '/app/imoveis', icon: <Home size={20} />, label: 'Imóveis' },
    { key: '/app/crm', icon: <Users size={20} />, label: 'CRM (Clientes)' },
    { key: '/app/leads', icon: <Inbox size={20} />, label: 'Leads (Inbox)' },
    { key: '/app/agendamentos', icon: <Calendar size={20} />, label: 'Agendamentos' },
	{ key: '/app/whatsapp', icon: <MessageCircleMore size={20} />, label: 'WhatsApp' },
	...(user?.papel === 'GESTOR' ? [{ key: '/app/equipe', icon: <UserRoundCog size={20} />, label: 'Equipe' }] : []),
	...(user?.papel !== 'CORRETOR_EQUIPE' ? [{ key: '/app/privacidade', icon: <Scale size={20} />, label: 'Privacidade' }] : []),
	...(user?.papel !== 'CORRETOR_EQUIPE' ? [{ key: '/app/configuracoes', icon: <Settings size={20} />, label: 'Configurações' }] : []),
  ];

  return (
    <div className="min-h-screen bg-slate-50 flex">
      <a href="#conteudo-principal" className="sr-only z-[100] rounded-md bg-white px-4 py-2 text-slate-900 shadow focus:not-sr-only focus:fixed focus:left-4 focus:top-4">
        Pular para o conteúdo principal
      </a>
      {/* Sider (Desktop) */}
      <aside className="hidden lg:flex flex-col w-64 bg-slate-950 text-white fixed h-full z-20">
        <div className="h-16 flex items-center justify-center border-b border-slate-800 mt-4 mb-4 gap-2">
          <div className="bg-gradient-to-br from-blue-500 to-blue-700 p-1.5 rounded-lg">
            <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
            </svg>
          </div>
          <span className="text-xl font-extrabold text-transparent bg-clip-text bg-gradient-to-r from-white to-blue-200 tracking-tight">KAPTEI</span>
        </div>
        <nav className="flex-1 py-4">
          <ul className="space-y-1 px-3">
            {menuItems.map(item => {
              const active = rotaEstaAtiva(location.pathname, item.key);
              return (
                <li key={item.key}>
                  <Link 
                    to={item.key} 
                    className={`flex items-center space-x-3 px-4 py-3 rounded-xl transition-colors ${active ? 'bg-blue-600 text-white' : 'text-slate-400 hover:text-white hover:bg-slate-900'}`}
                    aria-current={active ? 'page' : undefined}
                  >
                    {item.icon}
                    <span className="font-medium">{item.label}</span>
                  </Link>
                </li>
              );
            })}
          </ul>
        </nav>
      </aside>

      <dialog
        ref={drawerRef}
        aria-label="Navegação principal"
        onCancel={() => setDrawerVisible(false)}
        onClose={() => setDrawerVisible(false)}
        className="fixed inset-y-0 left-0 z-50 m-0 h-full max-h-none w-64 max-w-none overflow-y-auto border-0 bg-slate-950 p-0 text-white backdrop:bg-slate-950/70 lg:hidden"
      >
        <button type="button" onClick={() => setDrawerVisible(false)} className="absolute right-3 top-3 rounded-lg p-2 text-slate-300 hover:bg-slate-800 hover:text-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-400" aria-label="Fechar navegação principal">
          <X className="h-5 w-5" aria-hidden="true" />
        </button>
        <div className="h-16 flex items-center justify-center border-b border-slate-800 mt-4 mb-4">
          <img src="/logo.png" alt="Kaptei" className="h-12 w-auto object-contain" />
        </div>
        <nav className="flex-1 py-4">
          <ul className="space-y-1 px-3">
            {menuItems.map(item => {
              const active = rotaEstaAtiva(location.pathname, item.key);
              return (
                <li key={item.key}>
                  <Link 
                    to={item.key}
                    onClick={() => setDrawerVisible(false)}
                    className={`flex items-center space-x-3 px-4 py-3 rounded-xl transition-colors ${active ? 'bg-blue-600 text-white' : 'text-slate-400 hover:text-white hover:bg-slate-900'}`}
                    aria-current={active ? 'page' : undefined}
                  >
                    {item.icon}
                    <span className="font-medium">{item.label}</span>
                  </Link>
                </li>
              );
            })}
          </ul>
        </nav>
      </dialog>

      {/* Content wrapper */}
      <div className="flex-1 lg:ml-64 flex flex-col min-h-screen">
        {/* Header */}
        <header className="h-16 bg-white border-b border-slate-200 flex items-center justify-between px-4 sm:px-6 z-10">
            <button
              className="rounded-lg p-2 text-slate-600 hover:bg-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-600 lg:hidden"
              aria-label="Abrir navegação principal"
              onClick={() => setDrawerVisible(true)}
            >
              <Menu size={24} aria-hidden="true" />
            </button>
          
          <details className="group relative ml-auto">
            <summary className="flex cursor-pointer list-none items-center space-x-3 rounded-lg p-2 text-slate-700 transition-colors hover:bg-slate-50 hover:text-slate-900 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-600 [&::-webkit-details-marker]:hidden" aria-label="Abrir menu da conta">
              <div className="flex flex-col text-right">
                <span className="font-semibold text-sm leading-tight">Olá, {user?.nome || 'Corretor'}</span>
                <span className="text-xs text-slate-500">{getPapelName(user?.papel)}</span>
              </div>
              <ChevronDown size={16} className="text-slate-400 transition-transform group-open:rotate-180" aria-hidden="true" />
            </summary>
            <div className="absolute right-0 z-50 mt-2 w-48 rounded-xl border border-slate-200 bg-white p-1 shadow-lg">
              <Link to="/app/perfil" onClick={(evento) => evento.currentTarget.closest('details')?.removeAttribute('open')} className="flex items-center gap-2 rounded-lg px-3 py-2 text-sm text-slate-700 hover:bg-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-600">
                <User size={16} aria-hidden="true" />
                <span>Meu perfil</span>
              </Link>
              <Link to="/app/trocar-senha" onClick={(evento) => evento.currentTarget.closest('details')?.removeAttribute('open')} className="flex items-center gap-2 rounded-lg px-3 py-2 text-sm text-slate-700 hover:bg-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-600">
                <Lock size={16} aria-hidden="true" />
                <span>Trocar senha</span>
              </Link>
              <div className="my-1 h-px bg-slate-100" />
              <button type="button" className="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-sm text-red-600 hover:bg-red-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-red-600" onClick={() => void handleLogout()}>
                <LogOut size={16} aria-hidden="true" />
                <span>Sair do sistema</span>
              </button>
            </div>
          </details>
        </header>

        {/* Main Content — cada página gerencia seu próprio padding e largura */}
        <main id="conteudo-principal" tabIndex={-1} className="flex-1 overflow-hidden flex flex-col">
          <Outlet />
        </main>
      </div>
    </div>
  );
};
