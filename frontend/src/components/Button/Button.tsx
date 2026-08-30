import React from 'react';

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'danger' | 'dark' | 'outlined';
  icon?: React.ReactNode;
  isLoading?: boolean;
  loadingText?: string;
}

export const Button: React.FC<ButtonProps> = ({
  children,
  variant = 'primary',
  icon,
  isLoading,
  loadingText,
  className = '',
  disabled,
  ...props
}) => {
  const baseStyles = 'px-5 py-2.5 rounded-xl font-medium transition-all shadow-sm flex items-center justify-center gap-2 disabled:opacity-70';
  const variants = {
    primary: 'bg-blue-600 text-white hover:bg-blue-700 hover:shadow-md border border-transparent',
    secondary: 'bg-white text-slate-700 border-2 border-slate-200 hover:bg-slate-50 hover:border-slate-300',
    danger: 'bg-red-600 text-white hover:bg-red-700 hover:shadow-md border border-transparent',
    dark: 'bg-slate-900 text-white hover:bg-slate-800 hover:shadow-md border border-transparent',
    outlined: 'bg-transparent text-slate-700 border-2 border-slate-300 hover:border-slate-800 hover:text-slate-900',
  };

  return (
    <button disabled={disabled || isLoading} className={`${baseStyles} ${variants[variant]} ${className}`} {...props}>
      {isLoading ? (
        <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-current" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
      ) : icon ? icon : null}
      {isLoading && loadingText ? loadingText : children}
    </button>
  );
};