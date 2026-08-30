import React from 'react';
import { formatarCPF, validarCPF } from '../../lib/validacoes/cpf';

interface CpfInputProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'onChange'> {
  value: string;
  onChange: (evento: React.ChangeEvent<HTMLInputElement>) => void;
  name?: string;
  label?: string;
}

export const CpfInput: React.FC<CpfInputProps> = ({
  value,
  onChange,
  name = 'cpf',
  label = 'CPF',
  className = '',
  ...props
}) => {
  const digitos = value.replace(/\D/g, '');
  const erro = digitos.length === 11 && !validarCPF(digitos) ? 'CPF inválido' : null;

  const handleChange = (evento: React.ChangeEvent<HTMLInputElement>) => {
    const eventoFormatado = {
      ...evento,
      target: { ...evento.target, name, value: formatarCPF(evento.target.value) },
    };
    onChange(eventoFormatado as React.ChangeEvent<HTMLInputElement>);
  };

  return (
    <div className="space-y-1.5 w-full">
      {label && (
        <label className="block text-sm font-medium text-slate-700">
          {label}
          {props.required && <span className="text-red-500 ml-1">*</span>}
        </label>
      )}
      <input
        type="text"
        name={name}
        value={value}
        onChange={handleChange}
        placeholder="___.___.___-__"
        maxLength={14}
        className={`w-full px-3 py-2 border rounded-lg focus:ring-2 outline-none transition-colors ${
          erro
            ? 'border-red-300 focus:border-red-400 focus:ring-red-500/20 bg-red-50/30'
            : 'border-slate-200 focus:border-blue-600 focus:ring-blue-600/50'
        } ${className}`}
        {...props}
      />
      {erro && <span className="text-xs font-medium text-red-500 mt-1 block">{erro}</span>}
    </div>
  );
};
