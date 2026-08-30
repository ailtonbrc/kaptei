import React from 'react';
import { Save } from 'lucide-react';

interface SaveButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  isSaving?: boolean;
  label?: string;
  savingLabel?: string;
}

export const SaveButton: React.FC<SaveButtonProps> = ({ 
  isSaving = false, 
  label = 'Salvar Alterações', 
  savingLabel = 'Salvando...', 
  disabled, 
  className = '', 
  ...props 
}) => {
  return (
    <button
      type="submit"
      disabled={isSaving || disabled}
      className={`px-6 py-2.5 bg-blue-600 text-white font-medium rounded-xl hover:bg-blue-700 transition-colors flex items-center justify-center gap-2 disabled:opacity-70 ${className}`}
      {...props}
    >
      <Save className="w-4 h-4" />
      {isSaving ? savingLabel : label}
    </button>
  );
};
