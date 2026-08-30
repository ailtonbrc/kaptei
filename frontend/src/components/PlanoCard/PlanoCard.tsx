import React from 'react';
import { Check } from 'lucide-react';
import type { Plano } from '../../constants/planos';

interface PlanoCardProps {
  plano: Plano;
  selecionado: boolean;
  onClick: () => void;
  disabled?: boolean;
}

export const PlanoCard: React.FC<PlanoCardProps> = ({
  plano,
  selecionado,
  onClick,
  disabled = false
}) => {
  const { nome, preco, recomendado, subtitle, cor, features, missing } = plano;

  return (
    <div 
      className={`flex-1 rounded-2xl overflow-hidden transition-all duration-300 border-2 relative flex flex-col ${
        disabled 
          ? 'opacity-60 cursor-not-allowed' 
          : 'cursor-pointer transform hover:-translate-y-1 hover:shadow-2xl'
      } ${
        selecionado 
          ? 'shadow-xl' 
          : 'border-transparent'
      }`}
      style={{ 
        backgroundColor: '#0f172a', 
        borderColor: selecionado ? cor : '#1e293b',
        boxShadow: selecionado ? `0 10px 25px -5px ${cor}40` : undefined
      }}
      onClick={() => {
        if (!disabled) onClick();
      }}
    >
      <div className="flex flex-col w-full flex-1">
        <div 
          className="w-full text-center py-1 text-xs font-bold tracking-widest text-white/90 uppercase" 
          style={{ backgroundColor: recomendado ? '#1e293b' : '#0f172a' }}
        >
          {recomendado ? 'Recomendado' : '\u00A0'}
        </div>

        <div className="p-6 text-center" style={{ backgroundColor: cor }}>
          <h3 className="text-white font-bold text-lg uppercase tracking-wide m-0">{nome}</h3>
          <div className="text-white text-4xl font-extrabold mt-2">
            {preco.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL', minimumFractionDigits: 0 })}
            <span className="text-sm font-normal opacity-80">/mês</span>
          </div>
          <div className="text-white/90 text-sm mt-1 min-h-[20px]">
            {subtitle || '\u00A0'}
          </div>
        </div>

        <div className="p-6 bg-[#0f172a] flex-1">
          <ul className="space-y-3 p-0 m-0 list-none">
            {features.map((f, i) => (
              <li key={i} className="flex items-start text-slate-300 text-sm">
                <Check className="text-blue-400 mt-0.5 w-4 h-4 mr-2 flex-shrink-0" />
                <span>{f}</span>
              </li>
            ))}
            {missing.map((f, i) => (
              <li key={`m-${i}`} className="flex items-start text-slate-600 text-sm opacity-50">
                <span className="mr-2 mt-0.5">—</span>
                <span className="line-through">{f}</span>
              </li>
            ))}
          </ul>
        </div>
      </div>

      <div className="p-6 pt-0 bg-[#0f172a] text-center">
        <div 
          className={`w-full py-2.5 rounded-lg text-sm font-semibold transition-colors duration-200 ${
            selecionado 
              ? 'text-white border border-transparent' 
              : 'text-slate-300 border border-slate-700 hover:border-slate-500 hover:text-white'
          }`}
          style={{ 
            backgroundColor: selecionado ? cor : 'transparent'
          }}
        >
          {selecionado ? 'Plano Selecionado' : 'Escolher este Plano'}
        </div>
      </div>
    </div>
  );
};
