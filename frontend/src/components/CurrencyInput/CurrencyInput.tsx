import React, { useState, useEffect, useRef } from 'react';

// ---------------------------------------------------------------
// Utilitários de formatação de moeda (BR: R$ 1.234.567,89)
// ---------------------------------------------------------------

/**
 * Formata um número para o padrão monetário brasileiro.
 * Ex: 1234567.89 → "1.234.567,89"
 */
const formatarMoedaBR = (valor: number): string => {
  if (!valor && valor !== 0) return '';
  return valor.toLocaleString('pt-BR', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
};

/**
 * Extrai o valor numérico de uma string formatada em BR.
 * Trata os dois últimos dígitos como centavos.
 * Ex: "1.234,56" → 1234.56 | "12356" → 123.56
 */
// ---------------------------------------------------------------
// Interface do componente
// ---------------------------------------------------------------

interface CurrencyInputProps {
  /** Nome do campo (usado para identificação no handler do formulário) */
  name: string;
  /** Valor numérico atual (ex: 1234.56) */
  value: number;
  /** Callback chamado com o nome e o valor numérico atualizado */
  onValueChange: (name: string, value: number) => void;
  /** Rótulo exibido acima do campo */
  label?: string;
  /** Marca o campo como obrigatório */
  required?: boolean;
  /** Classes adicionais para o input */
  className?: string;
  /** Placeholder exibido quando vazio */
  placeholder?: string;
  /** Desabilita o campo */
  disabled?: boolean;
}

// ---------------------------------------------------------------
// Componente
// ---------------------------------------------------------------

export const CurrencyInput: React.FC<CurrencyInputProps> = ({
  name,
  value,
  onValueChange,
  label,
  required,
  className = '',
  placeholder = '0,00',
  disabled = false,
}) => {
  const [displayValue, setDisplayValue] = useState<string>(
    value > 0 ? formatarMoedaBR(value) : ''
  );

  // Flag para evitar loop entre prop e state
  const mudancaInterna = useRef(false);

  // Sincroniza quando o valor externo muda (ex: carregamento de dados da API)
  useEffect(() => {
    if (!mudancaInterna.current) {
      setDisplayValue(value > 0 ? formatarMoedaBR(value) : '');
    }
    mudancaInterna.current = false;
  }, [value]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const rawInput = e.target.value;
    const apenasDigitos = rawInput.replace(/\D/g, '');

    const valorNumerico = apenasDigitos
      ? parseInt(apenasDigitos, 10) / 100
      : 0;

    const valorFormatado = apenasDigitos
      ? formatarMoedaBR(valorNumerico)
      : '';

    mudancaInterna.current = true;
    setDisplayValue(valorFormatado);
    onValueChange(name, valorNumerico);
  };

  // Ao perder o foco, se vazio, garante que o display está limpo
  const handleBlur = () => {
    if (!displayValue && value === 0) {
      setDisplayValue('');
    }
  };

  return (
    <div className="space-y-1.5 w-full">
      {label && (
        <label className="block text-sm font-medium text-slate-700">
          {label}
          {required && <span className="text-red-500 ml-1">*</span>}
        </label>
      )}
      <div className="relative">
        <span className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 text-xs font-semibold select-none pointer-events-none">
          R$
        </span>
        <input
          type="text"
          inputMode="numeric"
          name={name}
          value={displayValue}
          onChange={handleChange}
          onBlur={handleBlur}
          placeholder={placeholder}
          disabled={disabled}
          className={`
            w-full pl-9 pr-3 py-2 text-right text-sm
            border border-slate-200 rounded-lg outline-none transition-colors
            focus:ring-2 focus:ring-blue-600/50 focus:border-blue-500
            disabled:bg-slate-50 disabled:text-slate-400 disabled:cursor-not-allowed
            ${className}
          `}
        />
      </div>
    </div>
  );
};
