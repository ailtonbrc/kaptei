import type { ReactNode } from 'react';

interface SectionHeaderProps {
  icon: ReactNode;
  title: string;
  subtitle?: string;
  color?: 'blue' | 'emerald';
}

export function SectionHeader({ icon, title, subtitle, color = 'blue' }: SectionHeaderProps) {
  const emphasis = color === 'emerald' ? 'bg-emerald-50 text-emerald-600' : 'bg-blue-50 text-blue-600';
  return <header className="mb-5 flex items-center gap-3 border-b border-slate-100 pb-3">
    <span className={`shrink-0 rounded-lg p-2 ${emphasis}`}>{icon}</span>
    <span><h3 className="text-base font-semibold text-slate-800">{title}</h3>{subtitle && <p className="mt-0.5 text-xs text-slate-500">{subtitle}</p>}</span>
  </header>;
}
