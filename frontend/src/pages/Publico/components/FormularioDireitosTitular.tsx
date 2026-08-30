import { useState } from 'react';
import { CheckCircle2, Send } from 'lucide-react';
import { privacidadeService } from '../../../services/privacidadeService';
import type { NovaSolicitacaoTitular, TipoSolicitacaoTitular } from '../../../types/privacidade';

const tipos: Array<{ valor: TipoSolicitacaoTitular; rotulo: string }> = [
  { valor: 'CONFIRMACAO', rotulo: 'Confirmar tratamento de dados' },
  { valor: 'ACESSO', rotulo: 'Acessar meus dados' },
  { valor: 'CORRECAO', rotulo: 'Corrigir meus dados' },
  { valor: 'ANONIMIZACAO', rotulo: 'Anonimizar dados' },
  { valor: 'BLOQUEIO', rotulo: 'Bloquear tratamento' },
  { valor: 'EXCLUSAO', rotulo: 'Excluir dados' },
  { valor: 'PORTABILIDADE', rotulo: 'Solicitar portabilidade' },
  { valor: 'REVOGACAO', rotulo: 'Revogar consentimento' },
  { valor: 'INFORMACAO_COMPARTILHAMENTO', rotulo: 'Saber com quem os dados foram compartilhados' },
];

const inicial: NovaSolicitacaoTitular = { tipo: 'ACESSO', nome: '', email: '', telefone: '', detalhes: '' };

export const FormularioDireitosTitular = ({ slug }: { slug: string }) => {
  const [dados, setDados] = useState(inicial);
  const [enviando, setEnviando] = useState(false);
  const [erro, setErro] = useState('');
  const [protocolo, setProtocolo] = useState('');

  const alterar = (campo: keyof NovaSolicitacaoTitular, valor: string) => setDados((atual) => ({ ...atual, [campo]: valor }));
  const enviar = async (evento: React.FormEvent) => {
    evento.preventDefault(); setErro(''); setEnviando(true);
    try {
      const resposta = await privacidadeService.criarPublica(slug, dados);
      setProtocolo(resposta.protocolo); setDados(inicial);
    } catch {
      setErro('Não foi possível registrar a solicitação. Revise os dados e tente novamente.');
    } finally { setEnviando(false); }
  };

  if (protocolo) return <div className="rounded-2xl border border-emerald-200 bg-emerald-50 p-6" role="status">
    <CheckCircle2 className="h-8 w-8 text-emerald-600" />
    <h2 className="mt-3 text-xl font-bold text-emerald-950">Solicitação registrada</h2>
    <p className="mt-2 text-sm text-emerald-800">Guarde este protocolo para contato com o controlador:</p>
    <p className="mt-3 break-all rounded-lg bg-white px-4 py-3 font-mono font-bold text-emerald-950">{protocolo}</p>
    <button type="button" onClick={() => setProtocolo('')} className="mt-5 text-sm font-bold text-emerald-800 underline">Registrar outra solicitação</button>
  </div>;

  const classeCampo = 'mt-1 w-full rounded-xl border border-slate-300 bg-white px-4 py-3 text-sm outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-100';
  return <form onSubmit={enviar} className="space-y-5 rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
    <div><h2 className="text-xl font-bold text-slate-950">Exercer meus direitos</h2><p className="mt-1 text-sm text-slate-600">O pedido será analisado pelo controlador. Nenhuma exclusão ocorre automaticamente.</p></div>
    <label className="block text-sm font-semibold text-slate-800">Direito solicitado<select className={classeCampo} value={dados.tipo} onChange={(e) => alterar('tipo', e.target.value)}>{tipos.map((tipo) => <option key={tipo.valor} value={tipo.valor}>{tipo.rotulo}</option>)}</select></label>
    <label className="block text-sm font-semibold text-slate-800">Nome completo<input required minLength={2} maxLength={160} autoComplete="name" className={classeCampo} value={dados.nome} onChange={(e) => alterar('nome', e.target.value)} /></label>
    <div className="grid gap-4 sm:grid-cols-2">
      <label className="block text-sm font-semibold text-slate-800">E-mail<input type="email" maxLength={254} autoComplete="email" className={classeCampo} value={dados.email} onChange={(e) => alterar('email', e.target.value)} /></label>
      <label className="block text-sm font-semibold text-slate-800">Telefone<input maxLength={24} autoComplete="tel" className={classeCampo} value={dados.telefone} onChange={(e) => alterar('telefone', e.target.value)} /></label>
    </div>
    <p className="-mt-3 text-xs text-slate-500">Informe ao menos um contato. Ele será usado também para a verificação de identidade.</p>
    <label className="block text-sm font-semibold text-slate-800">Detalhes<textarea maxLength={4000} rows={4} className={classeCampo} value={dados.detalhes} onChange={(e) => alterar('detalhes', e.target.value)} /></label>
    {erro && <p role="alert" className="text-sm font-medium text-red-600">{erro}</p>}
    <button disabled={enviando || (!dados.email && !dados.telefone)} className="inline-flex w-full items-center justify-center gap-2 rounded-xl bg-blue-600 px-5 py-3 font-bold text-white hover:bg-blue-700 disabled:opacity-50"><Send className="h-4 w-4" />{enviando ? 'Registrando...' : 'Registrar solicitação'}</button>
  </form>;
};
