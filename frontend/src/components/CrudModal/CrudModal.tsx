import React from 'react';
import { X, ChevronRight, Save, Loader2 } from 'lucide-react';

// ---------------------------------------------------------------
// Interface do Componente
// ---------------------------------------------------------------

interface CrudModalProps {
  /** Rótulo da tela pai no breadcrumb (ex: "Imóveis", "CRM") */
  breadcrumbPai?: string;
  /** Título da operação exibida no cabeçalho (ex: "Novo Imóvel") */
  titulo: string;
  /** Chamado ao clicar em Cancelar ou no botão X */
  onCancel: () => void;
  /** ID do formulário a ser submetido pelo botão Salvar */
  formId: string;
  /** Indica que o salvamento está em progresso */
  isSaving?: boolean;
  /** Rótulo personalizado do botão Salvar */
  labelSalvar?: string;
  /** Rótulo personalizado do botão Cancelar */
  labelCancelar?: string;
  /** Altura máxima do modal (Tailwind: ex. "max-h-[65vh]") */
  maxHeight?: string;
  /** Mensagem de sucesso exibida no topo */
  successMsg?: string;
  /** Mensagem de erro exibida no topo */
  errorMsg?: string;
  /** Conteúdo interno do modal (formulários, seções, etc.) */
  children: React.ReactNode;
}

// ---------------------------------------------------------------
// Componente — usa forwardRef para expor o scroll interno
// ao componente pai, permitindo scrollToTop/scrollToBottom
// ---------------------------------------------------------------

export const CrudModal = React.forwardRef<HTMLDivElement, CrudModalProps>(
  (
    {
      breadcrumbPai,
      titulo,
      onCancel,
      formId,
      isSaving = false,
      labelSalvar = 'Salvar',
      labelCancelar = 'Cancelar',
      maxHeight = 'max-h-[65vh]',
      successMsg,
      errorMsg,
      children,
    },
    ref
  ) => {
    return (
      <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-sm flex items-center justify-center p-4 animate-in fade-in duration-200">

        {/* ── Notificações flutuantes centralizadas ── */}
        {(successMsg || errorMsg) && (
          <div className="fixed top-6 left-1/2 -translate-x-1/2 z-[60] flex flex-col gap-3 animate-in fade-in slide-in-from-top-4 duration-300 pointer-events-none">
            {successMsg && (
              <div className="flex items-center gap-3 px-5 py-3.5 bg-emerald-50 border border-emerald-200 shadow-lg rounded-xl text-emerald-800 min-w-[320px]">
                <span className="text-emerald-600">✓</span>
                <p className="font-medium text-sm">{successMsg}</p>
              </div>
            )}
            {errorMsg && (
              <div className="flex items-center gap-3 px-5 py-3.5 bg-red-50 border border-red-200 shadow-lg rounded-xl text-red-700 min-w-[320px]">
                <span className="text-red-500">✕</span>
                <p className="font-medium text-sm">{errorMsg}</p>
              </div>
            )}
          </div>
        )}

        {/* ── Caixa do Modal ── */}
        <div className={`bg-white rounded-2xl shadow-2xl w-full max-w-4xl flex flex-col ${maxHeight} overflow-hidden`}>

          {/* Cabeçalho com breadcrumb e botão fechar */}
          <div className="flex items-center justify-between px-6 py-4 border-b border-slate-200 bg-slate-50 rounded-t-2xl shrink-0">
            <div className="flex items-center gap-2 text-slate-500 text-sm">
              {breadcrumbPai && (
                <>
                  <span
                    onClick={onCancel}
                    className="hover:text-blue-600 cursor-pointer transition-colors"
                  >
                    {breadcrumbPai}
                  </span>
                  <ChevronRight className="w-4 h-4 text-slate-400" />
                </>
              )}
              <span className="font-semibold text-slate-800">{titulo}</span>
            </div>

            <button
              type="button"
              onClick={onCancel}
              className="p-2 hover:bg-slate-200 rounded-lg text-slate-500 hover:text-slate-800 transition-colors"
              title="Fechar"
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          {/* Corpo com scroll gerenciado pelo componente pai via ref */}
          <div ref={ref} className="overflow-y-auto flex-1">
            {children}
          </div>

          {/* Rodapé com botões outlined alinhados à direita */}
          <div className="border-t border-slate-200 bg-white px-6 py-4 flex items-center justify-end gap-3 shrink-0 rounded-b-2xl">
            {/* Botão Cancelar — outlined neutro */}
            <button
              type="button"
              onClick={onCancel}
              disabled={isSaving}
              className="
                px-5 py-2.5 text-sm font-medium rounded-xl
                border border-slate-300 text-slate-600
                hover:bg-slate-50 hover:border-slate-400
                transition-colors disabled:opacity-50 disabled:cursor-not-allowed
                flex items-center gap-2
              "
            >
              <X className="w-4 h-4" />
              {labelCancelar}
            </button>

            {/* Botão Salvar — outlined primário */}
            <button
              type="submit"
              form={formId}
              disabled={isSaving}
              className="
                px-5 py-2.5 text-sm font-medium rounded-xl
                border border-blue-500 text-blue-600
                hover:bg-blue-50 hover:border-blue-600
                transition-colors disabled:opacity-50 disabled:cursor-not-allowed
                flex items-center gap-2
              "
            >
              {isSaving ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  Salvando...
                </>
              ) : (
                <>
                  <Save className="w-4 h-4" />
                  {labelSalvar}
                </>
              )}
            </button>
          </div>
        </div>
      </div>
    );
  }
);

CrudModal.displayName = 'CrudModal';
