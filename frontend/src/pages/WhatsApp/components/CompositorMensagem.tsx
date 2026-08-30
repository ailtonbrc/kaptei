import { useState } from 'react';
import { FileText, Send } from 'lucide-react';
import { Button } from '../../../components/ui/button';
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger,
} from '../../../components/ui/dialog';

interface Props {
  janelaAberta: boolean;
  consentimento: boolean;
  enviando: boolean;
  aoEnviarTexto: (texto: string) => Promise<boolean>;
  aoEnviarTemplate: (nome: string, idioma: string, parametros: string[]) => Promise<boolean>;
}

export function CompositorMensagem({ janelaAberta, consentimento, enviando, aoEnviarTexto, aoEnviarTemplate }: Props) {
  const [texto, setTexto] = useState('');
  const [templateAberto, setTemplateAberto] = useState(false);
  const [nome, setNome] = useState('');
  const [idioma, setIdioma] = useState('pt_BR');
  const [parametros, setParametros] = useState('');

  async function enviarTexto() {
    if (await aoEnviarTexto(texto)) setTexto('');
  }

  async function enviarTemplate() {
    const valores = parametros.split('\n').map((valor) => valor.trim()).filter(Boolean);
    if (await aoEnviarTemplate(nome, idioma, valores)) {
      setNome('');
      setParametros('');
      setTemplateAberto(false);
    }
  }

  return (
    <div className="border-t border-slate-200 bg-white p-3 sm:p-4">
      {!janelaAberta && (
        <p className="mb-2 text-xs font-medium text-amber-700">A janela de 24 horas terminou. Use um template aprovado.</p>
      )}
      <div className="flex items-end gap-2">
        <Dialog open={templateAberto} onOpenChange={setTemplateAberto}>
          <DialogTrigger asChild>
            <Button type="button" variant="outline" size="icon" title="Enviar template" disabled={!consentimento || enviando}>
              <FileText className="h-4 w-4" />
              <span className="sr-only">Enviar template aprovado</span>
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Enviar template aprovado</DialogTitle>
              <DialogDescription>Informe exatamente o nome, idioma e parâmetros cadastrados no WhatsApp Manager.</DialogDescription>
            </DialogHeader>
            <div className="space-y-3">
              <label className="block text-sm font-medium text-slate-700">Nome do template
                <input value={nome} onChange={(e) => setNome(e.target.value)} placeholder="retorno_interesse_imovel" className="mt-1 h-10 w-full rounded-lg border border-slate-300 px-3" />
              </label>
              <label className="block text-sm font-medium text-slate-700">Idioma
                <input value={idioma} onChange={(e) => setIdioma(e.target.value)} placeholder="pt_BR" className="mt-1 h-10 w-full rounded-lg border border-slate-300 px-3" />
              </label>
              <label className="block text-sm font-medium text-slate-700">Parâmetros do corpo
                <textarea value={parametros} onChange={(e) => setParametros(e.target.value)} placeholder="Um parâmetro por linha" rows={4} className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2" />
              </label>
            </div>
            <DialogFooter><Button type="button" onClick={() => void enviarTemplate()} disabled={enviando || !nome || !idioma}>Enfileirar template</Button></DialogFooter>
          </DialogContent>
        </Dialog>
        <textarea
          value={texto}
          onChange={(evento) => setTexto(evento.target.value)}
          onKeyDown={(evento) => {
            if (evento.key === 'Enter' && !evento.shiftKey) {
              evento.preventDefault();
              if (janelaAberta && texto.trim()) void enviarTexto();
            }
          }}
          disabled={!janelaAberta || enviando}
          maxLength={4096}
          rows={2}
          placeholder={janelaAberta ? 'Digite uma mensagem...' : 'Envio livre indisponível fora da janela'}
          className="min-h-10 flex-1 resize-none rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-100 disabled:bg-slate-100"
        />
        <Button type="button" size="icon" onClick={() => void enviarTexto()} disabled={!janelaAberta || enviando || !texto.trim()}>
          <Send className="h-4 w-4" />
          <span className="sr-only">Enviar mensagem</span>
        </Button>
      </div>
      {!consentimento && <p className="mt-2 text-xs text-slate-500">Registre o consentimento para habilitar mensagens iniciadas por template.</p>}
    </div>
  );
}
