import { useState } from 'react';
import { CheckCircle2, Download, FileCheck2, ShieldCheck, Trash2 } from 'lucide-react';
import { Button } from '../../../components/ui/button';
import { privacidadeService } from '../../../services/privacidadeService';
import type { SolicitacaoTitular } from '../../../types/privacidade';

const exportaveis = ['CONFIRMACAO', 'ACESSO', 'PORTABILIDADE', 'INFORMACAO_COMPARTILHAMENTO'];
const executaveis = ['CORRECAO', 'ANONIMIZACAO', 'BLOQUEIO', 'EXCLUSAO', 'REVOGACAO'];

interface Propriedades {
  solicitacao: SolicitacaoTitular;
  aoAtualizar: () => Promise<void>;
}

export const PainelSolicitacao = ({ solicitacao, aoAtualizar }: Propriedades) => {
  const [metodo, setMetodo] = useState('Confirmação por contato cadastrado');
  const [evidencia, setEvidencia] = useState('');
  const [fundamento, setFundamento] = useState('Atendimento aos direitos do titular previstos na LGPD.');
  const [observacao, setObservacao] = useState('');
  const [processando, setProcessando] = useState(false);
  const [erro, setErro] = useState('');

  const executar = async (acao: () => Promise<void>) => {
    setErro(''); setProcessando(true);
    try { await acao(); await aoAtualizar(); } catch { setErro('Não foi possível concluir a operação. Revise o estado da solicitação.'); }
    finally { setProcessando(false); }
  };

  const executarDireito = () => {
    const destrutiva = ['ANONIMIZACAO', 'EXCLUSAO'].includes(solicitacao.tipo);
    const mensagem = destrutiva
      ? 'Esta operação altera ou remove dados pessoais de forma irreversível. Confirma que a decisão e as exceções legais foram revisadas?'
      : 'Confirma que o tratamento aprovado foi concluído?';
    if (window.confirm(mensagem)) void executar(() => privacidadeService.executar(solicitacao.id));
  };

  const campo = 'mt-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-100';
  return <aside className="space-y-6 rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
    <div className="flex flex-wrap items-start justify-between gap-3">
      <div><p className="font-mono text-sm font-bold text-blue-700">{solicitacao.protocolo}</p><h2 className="mt-1 text-xl font-black text-slate-950">{solicitacao.nome}</h2><p className="mt-1 text-sm text-slate-500">{solicitacao.email || solicitacao.telefone}</p></div>
      <span className="rounded-full bg-slate-100 px-3 py-1 text-xs font-bold text-slate-700">{solicitacao.status.replaceAll('_', ' ')}</span>
    </div>
    <dl className="grid gap-3 text-sm sm:grid-cols-2"><div><dt className="text-slate-500">Direito</dt><dd className="font-semibold text-slate-900">{solicitacao.tipo}</dd></div><div><dt className="text-slate-500">Prazo</dt><dd className="font-semibold text-slate-900">{new Date(solicitacao.prazo_resposta_em).toLocaleDateString('pt-BR')}</dd></div></dl>
    {solicitacao.detalhes && <div className="rounded-xl bg-slate-50 p-4 text-sm leading-6 text-slate-700"><strong className="block text-slate-900">Detalhes do titular</strong>{solicitacao.detalhes}</div>}

    {!solicitacao.identidade_verificada_em && <section className="border-t border-slate-200 pt-5"><h3 className="flex items-center gap-2 font-bold text-slate-900"><ShieldCheck className="h-5 w-5 text-blue-600" /> Verificar identidade</h3><p className="mt-1 text-xs text-slate-500">Não registre documentos ou segredos em texto aberto; a evidência é criptografada.</p><label className="mt-4 block text-sm font-semibold">Método<input className={campo} maxLength={80} value={metodo} onChange={(e) => setMetodo(e.target.value)} /></label><label className="mt-3 block text-sm font-semibold">Evidência<textarea className={campo} rows={3} maxLength={2000} value={evidencia} onChange={(e) => setEvidencia(e.target.value)} /></label><Button className="mt-4" disabled={processando || evidencia.trim().length < 3} onClick={() => void executar(() => privacidadeService.verificar(solicitacao.id, metodo, evidencia))}><CheckCircle2 className="h-4 w-4" /> Confirmar verificação</Button></section>}

    {solicitacao.status === 'EM_ANALISE' && <section className="border-t border-slate-200 pt-5"><h3 className="flex items-center gap-2 font-bold text-slate-900"><FileCheck2 className="h-5 w-5 text-blue-600" /> Decisão do controlador</h3><label className="mt-4 block text-sm font-semibold">Fundamento legal<textarea className={campo} rows={3} maxLength={2000} value={fundamento} onChange={(e) => setFundamento(e.target.value)} /></label><label className="mt-3 block text-sm font-semibold">Observação<textarea className={campo} rows={2} maxLength={2000} value={observacao} onChange={(e) => setObservacao(e.target.value)} /></label><div className="mt-4 flex gap-3"><Button disabled={processando} onClick={() => void executar(() => privacidadeService.decidir(solicitacao.id, true, fundamento, observacao))}>Aprovar</Button><Button variant="outline" disabled={processando} onClick={() => void executar(() => privacidadeService.decidir(solicitacao.id, false, fundamento, observacao))}>Rejeitar</Button></div></section>}

    {solicitacao.status === 'APROVADA' && <section className="border-t border-slate-200 pt-5"><h3 className="font-bold text-slate-900">Cumprimento da decisão</h3><p className="mt-1 text-sm text-slate-600">Revise a base legal e as hipóteses de retenção antes de prosseguir.</p><div className="mt-4 flex flex-wrap gap-3">{exportaveis.includes(solicitacao.tipo) && <Button disabled={processando} onClick={() => void executar(() => privacidadeService.exportar(solicitacao.id, solicitacao.protocolo))}><Download className="h-4 w-4" /> Gerar exportação</Button>}{executaveis.includes(solicitacao.tipo) && <Button variant={['ANONIMIZACAO', 'EXCLUSAO'].includes(solicitacao.tipo) ? 'destructive' : 'default'} disabled={processando} onClick={executarDireito}>{solicitacao.tipo === 'EXCLUSAO' && <Trash2 className="h-4 w-4" />} Confirmar execução</Button>}</div></section>}

    {erro && <p role="alert" className="text-sm font-medium text-red-600">{erro}</p>}
    {!!solicitacao.eventos?.length && <section className="border-t border-slate-200 pt-5"><h3 className="font-bold text-slate-900">Histórico</h3><ol className="mt-3 space-y-3">{solicitacao.eventos.map((evento) => <li key={evento.id} className="border-l-2 border-blue-200 pl-3 text-sm"><p className="font-semibold text-slate-800">{evento.descricao}</p><time className="text-xs text-slate-500">{new Date(evento.criado_em).toLocaleString('pt-BR')}</time></li>)}</ol></section>}
  </aside>;
};

