import React, { useState } from 'react';
import type { InputHTMLAttributes } from 'react';
import { Eye, EyeOff } from 'lucide-react';

type PasswordInputProps = InputHTMLAttributes<HTMLInputElement>;

export const PasswordInput: React.FC<PasswordInputProps> = ({ className = '', ...props }) => {
  const [showPassword, setShowPassword] = useState(false);

  // Garantimos que a classe pr-10 exista para dar espaço ao ícone
  const combinedClassName = `${className} pr-10`.trim();

  return (
    <div className="relative w-full">
      <input
        {...props}
        type={showPassword ? 'text' : 'password'}
        className={combinedClassName}
      />
      <button
        type="button"
        onClick={() => setShowPassword(!showPassword)}
        className="absolute inset-y-0 right-0 pr-3 flex items-center text-slate-400 hover:text-slate-600 focus:outline-none"
      >
        {showPassword ? <EyeOff className="h-5 w-5" /> : <Eye className="h-5 w-5" />}
      </button>
    </div>
  );
};
