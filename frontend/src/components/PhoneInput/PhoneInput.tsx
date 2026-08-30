import React from 'react';

interface PhoneInputProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'onChange'> {
  value: string;
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  name?: string;
  label?: string;
  allowUsername?: boolean; // Se true, permite iniciar com '@' (ex: @ailtonbrc) e ignora a máscara de telefone
}

// Formatação para telefone fixo (10 dígitos) ou celular (11 dígitos)
const formatPhone = (value: string) => {
  const digits = value.replace(/\D/g, '').slice(0, 11);
  
  if (digits.length === 0) return '';
  if (digits.length <= 2) return `(${digits}`;
  if (digits.length <= 6) return `(${digits.slice(0, 2)}) ${digits.slice(2)}`;
  if (digits.length <= 10) return `(${digits.slice(0, 2)}) ${digits.slice(2, 6)}-${digits.slice(6)}`;
  
  // Celular (11 dígitos)
  return `(${digits.slice(0, 2)}) ${digits.slice(2, 7)}-${digits.slice(7)}`;
};

export const PhoneInput: React.FC<PhoneInputProps> = ({ 
  value, 
  onChange, 
  name,
  label,
  allowUsername = false,
  className = '',
  ...props 
}) => {
  
  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const rawValue = e.target.value;
    let formatted: string;

    if (allowUsername && rawValue.startsWith('@')) {
      // Se permitir username e começar com @, não aplica máscara de telefone e remove espaços
      formatted = rawValue.replace(/\s/g, '');
    } else {
      // Caso contrário, formata como número de telefone
      formatted = formatPhone(rawValue);
    }

    const event = {
      ...e,
      target: { ...e.target, name, value: formatted }
    };
    
    onChange(event as unknown as React.ChangeEvent<HTMLInputElement>);
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
        placeholder={allowUsername ? "(00) 00000-0000 ou @usuario" : "(00) 00000-0000"}
        maxLength={allowUsername && value.startsWith('@') ? 50 : 15} // Limita o tel a 15 chars e o username a 50
        className={`w-full px-3 py-2 border border-slate-200 rounded-lg focus:ring-2 focus:border-blue-600 focus:ring-blue-600/50 outline-none transition-colors ${className}`}
        {...props}
      />
    </div>
  );
};
