import type { ReactNode } from 'react';

interface FormFieldProps {
  label: string;
  required?: boolean;
  children: ReactNode;
  className?: string;
}

export function FormField({ label, required, children, className = '' }: FormFieldProps) {
  return <div className={`space-y-1.5 ${className}`}>
    <label className="block text-sm font-medium text-slate-700">{label}{required && <span className="ml-1 text-red-500">*</span>}</label>
    {children}
  </div>;
}
