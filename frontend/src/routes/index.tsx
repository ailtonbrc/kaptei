import { createBrowserRouter } from 'react-router-dom';
import { MainLayout } from '../layouts/MainLayout';
import { ProtectedRoute } from './ProtectedRoute';
import { RoleRoute } from './RoleRoute';
import { ErroPublico } from '../pages/Publico/ErroPublico';
import { carregarImovelPublico, carregarPoliticaPrivacidade, carregarSitePublico } from '../pages/Publico/carregadores';
import { carregarImovelDoDominio, carregarPrivacidadeDoDominio, carregarSiteDoDominio } from '../pages/Publico/carregadoresDominio';

const paginaLogin = async () => ({ Component: (await import('../pages/Auth/Login')).Login });
const paginaCadastro = async () => ({ Component: (await import('../pages/Auth/Register')).Register });
const paginaEsqueciSenha = async () => ({ Component: (await import('../pages/Auth/EsqueciSenha')).EsqueciSenha });
const paginaRedefinirSenha = async () => ({ Component: (await import('../pages/Auth/RedefinirSenha')).RedefinirSenha });
const paginaAceitarConvite = async () => ({ Component: (await import('../pages/Auth/AceitarConvite')).AceitarConvite });
const paginaSitePublico = async () => ({ Component: (await import('../pages/Publico/SitePublicoPage')).SitePublicoPage });
const paginaImovelPublico = async () => ({ Component: (await import('../pages/Publico/ImovelPublicoPage')).ImovelPublicoPage });
const paginaPrivacidade = async () => ({ Component: (await import('../pages/Publico/PoliticaPrivacidadePage')).PoliticaPrivacidadePage });
const paginaCheckout = async () => ({ Component: (await import('../pages/Billing/Checkout')).Checkout });
const paginaAssinatura = async () => ({ Component: (await import('../pages/Billing/Billing')).Billing });
const paginaDashboard = async () => ({ Component: (await import('../pages/Dashboard/Dashboard')).Dashboard });
const paginaImoveis = async () => ({ Component: (await import('../pages/Imoveis/ImoveisList')).ImoveisList });
const paginaImovelFormulario = async () => ({ Component: (await import('../pages/Imoveis/ImovelForm')).ImovelForm });
const paginaClientes = async () => ({ Component: (await import('../pages/CRM/ClientesList')).ClientesList });
const paginaClienteFormulario = async () => ({ Component: (await import('../pages/CRM/ClienteForm')).ClienteForm });
const paginaLeads = async () => ({ Component: (await import('../pages/CRM/LeadsInbox')).LeadsInbox });
const paginaAgendamentos = async () => ({ Component: (await import('../pages/CRM/Agendamentos')).Agendamentos });
const paginaPerfil = async () => ({ Component: (await import('../pages/Profile/Profile')).Profile });
const paginaTrocarSenha = async () => ({ Component: (await import('../pages/Profile/ChangePassword')).ChangePassword });
const paginaConfiguracoes = async () => ({ Component: (await import('../pages/Settings/Settings')).Settings });
const paginaEquipe = async () => ({ Component: (await import('../pages/Equipe/EquipePage')).EquipePage });
const paginaWhatsApp = async () => ({ Component: (await import('../pages/WhatsApp/ConversasWhatsApp')).ConversasWhatsApp });
const paginaPrivacidadeAdmin = async () => ({ Component: (await import('../pages/Privacidade/CentralPrivacidade')).CentralPrivacidade });

export const router = createBrowserRouter([
  {
    path: '/',
    lazy: paginaSitePublico,
    loader: carregarSiteDoDominio,
    errorElement: <ErroPublico />,
  },
  {
    path: '/s/:slug',
	lazy: paginaSitePublico,
    loader: carregarSitePublico,
    errorElement: <ErroPublico />,
  },
  {
    path: '/s/:slug/imoveis/:slugImovel',
	lazy: paginaImovelPublico,
    loader: carregarImovelPublico,
    errorElement: <ErroPublico />,
  },
  {
    path: '/s/:slug/privacidade',
	lazy: paginaPrivacidade,
    loader: carregarPoliticaPrivacidade,
    errorElement: <ErroPublico />,
  },
  {
    path: '/imoveis/:slugImovel',
    lazy: paginaImovelPublico,
    loader: carregarImovelDoDominio,
    errorElement: <ErroPublico />,
  },
  {
    path: '/privacidade',
    lazy: paginaPrivacidade,
    loader: carregarPrivacidadeDoDominio,
    errorElement: <ErroPublico />,
  },
  {
    path: '/login',
	lazy: paginaLogin,
  },
  {
    path: '/register',
	lazy: paginaCadastro,
  },
  {
    path: '/esqueci-senha',
	lazy: paginaEsqueciSenha,
  },
  {
    path: '/redefinir-senha',
	lazy: paginaRedefinirSenha,
  },
  {
    path: '/aceitar-convite',
	lazy: paginaAceitarConvite,
  },
  {
    path: '/checkout',
    element: <ProtectedRoute />,
    children: [
      {
        path: '',
		lazy: paginaCheckout,
      }
    ]
  },
	{
		path: '/checkout/sucesso',
		element: <ProtectedRoute />,
		children: [{ path: '', lazy: paginaCheckout }],
	},
  {
    path: '/app',
    element: <ProtectedRoute />,
    children: [
      {
        path: 'assinatura',
		lazy: paginaAssinatura,
      },
      {
        path: '',
        element: <MainLayout />,
        children: [
          {
            path: '',
			lazy: paginaDashboard,
          },
          {
            path: 'imoveis',
			lazy: paginaImoveis,
          },
          {
            path: 'imoveis/novo',
			lazy: paginaImovelFormulario,
          },
          {
            path: 'imoveis/:id/editar',
			lazy: paginaImovelFormulario,
          },
          {
            path: 'crm',
			lazy: paginaClientes,
          },
          {
            path: 'leads',
			lazy: paginaLeads,
          },
          {
            path: 'agendamentos',
			lazy: paginaAgendamentos,
          },
          {
            path: 'whatsapp',
			lazy: paginaWhatsApp,
          },
          {
            path: 'crm/novo',
			lazy: paginaClienteFormulario,
          },
          {
            path: 'crm/:id',
			lazy: paginaClienteFormulario,
          },
          {
            path: 'perfil',
			lazy: paginaPerfil,
          },
          {
            path: 'trocar-senha',
			lazy: paginaTrocarSenha,
          },
          {
            element: <RoleRoute papeis={['SUPER_ADMIN', 'GESTOR', 'CORRETOR_SOLO']} />,
			children: [
              { path: 'configuracoes', lazy: paginaConfiguracoes },
              { path: 'privacidade', lazy: paginaPrivacidadeAdmin },
            ],
          },
          {
            element: <RoleRoute papeis={['GESTOR']} />,
			children: [{ path: 'equipe', lazy: paginaEquipe }],
          },
        ],
      },
    ],
  },
]);
