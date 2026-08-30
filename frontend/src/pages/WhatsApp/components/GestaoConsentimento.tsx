import { useState } from 'react';
import { ShieldCheck } from 'lucide-react';
import { Button } from '../../../components/ui/button';
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger,
} from '../../../components/ui/dialog';

interface Props {
  consentimento: boolean;
  salvando: boolean;
  aoSalvar: (consentiu: boolean, origem: string, evidencia: string) => Promise<boolean>;
}

export function GestaoConsentimento({ consentimento, salvando, aoSalvar }: Props) {
  const [aberto, setAberto] = useState(false);
  const [origem, setOrigem] = useState('FORMULARIO_SITE');
  const [evidencia, setEvidencia] = useState('');

  async function salvar(consentiu: boolean) {
    if (await aoSalvar(consentiu, consentiu ? origem : '', consentiu ? evidencia : '')) {
      setAberto(false);
      setEvidencia('');
    }
  }

  return (
    <Dialog open={aberto} onOpenChange={setAberto}>
      <DialogTrigger asChild>
        <Button variant="outline" size="sm"><ShieldCheck className="h-4 w-4" />{consentimento ? 'Consentimento ativo' : 'Registrar consentimento'}</Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Consentimento para contato</DialogTitle>
          <DialogDescription>Registre somente quando houver prova verificável da autorização do contato.</DialogDescription>
        </DialogHeader>
        <div className="space-y-3">
          <label className="block text-sm font-medium text-slate-700">Origem
            <input value={origem} onChange={(e) => setOrigem(e.target.value)} maxLength={80} className="mt-1 h-10 w-full rounded-lg border border-slate-300 px-3" />
          </label>
          <label className="block text-sm font-medium text-slate-700">Evidência
            <textarea value={evidencia} onChange={(e) => setEvidencia(e.target.value)} maxLength={500} rows={3} placeholder="Ex.: aceite no formulário em 06/08/2026, protocolo..." className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2" />
          </label>
        </div>
        <DialogFooter>
          {consentimento && <Button type="button" variant="destructive" disabled={salvando} onClick={() => void salvar(false)}>Revogar</Button>}
          <Button type="button" disabled={salvando || !origem.trim() || !evidencia.trim()} onClick={() => void salvar(true)}>Registrar autorização</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
