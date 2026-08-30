import { Loader2, Settings as SettingsIcon } from 'lucide-react';
import { useAuthStore } from '../../store/useAuthStore';
import { ConfiguracaoCobranca } from './components/ConfiguracaoCobranca';
import { ConfiguracaoGoogle } from './components/ConfiguracaoGoogle';
import { ConfiguracaoMotorLeads } from './components/ConfiguracaoMotorLeads';
import { ConfiguracaoObservabilidade } from './components/ConfiguracaoObservabilidade';
import { ConfiguracaoMetaLeads } from './components/ConfiguracaoMetaLeads';
import { ConfiguracaoWhatsApp } from './components/ConfiguracaoWhatsApp';
import { ConfiguracaoPortais } from './components/ConfiguracaoPortais';
import { ConfiguracaoSite } from './components/ConfiguracaoSite';
import { ConfiguracaoSMTP } from './components/ConfiguracaoSMTP';
import { useConfiguracoesSettings } from './hooks/useConfiguracoesSettings';

export function Settings() {
  const usuario = useAuthStore((estado) => estado.user);
  const superAdmin = usuario?.papel === 'SUPER_ADMIN';
  const podeAdministrarConta = superAdmin || usuario?.papel === 'GESTOR' || usuario?.papel === 'CORRETOR_SOLO';
  const { carregando, smtp, googleClientID, conta, site, metaLeads, whatsApp, observabilidade } = useConfiguracoesSettings(superAdmin, podeAdministrarConta);

  if (carregando) return <div className="grid h-full place-items-center"><Loader2 className="h-8 w-8 animate-spin text-blue-600" /></div>;

  return <main className="mx-auto max-w-4xl space-y-6 p-4 sm:p-6 lg:p-8">
    <header><h1 className="flex items-center gap-2 text-2xl font-bold text-slate-900"><SettingsIcon className="h-7 w-7 text-blue-600" />Configurações</h1><p className="mt-1 text-slate-500">Administre integrações, captação, site e cobrança conforme seu perfil de acesso.</p></header>
    {podeAdministrarConta && site && <section className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm"><ConfiguracaoSite inicial={site} /></section>}
    {superAdmin && <><ConfiguracaoCobranca /><ConfiguracaoSMTP inicial={smtp} /><ConfiguracaoGoogle inicial={googleClientID} /><ConfiguracaoObservabilidade inicial={observabilidade} /></>}
    {podeAdministrarConta && <ConfiguracaoMotorLeads inicial={conta} />}
    {podeAdministrarConta && <ConfiguracaoMetaLeads inicial={metaLeads} />}
    {podeAdministrarConta && <ConfiguracaoWhatsApp inicial={whatsApp} />}
    {podeAdministrarConta && <ConfiguracaoPortais />}
  </main>;
}
