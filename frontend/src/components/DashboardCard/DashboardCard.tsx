import React from 'react';

interface DashboardCardProps {
  title: string;
  value: string | number;
  icon: React.ReactNode;
  borderColor: string; // Ex: 'bg-blue-800' ou 'bg-orange-500'
  iconColor: string; // Ex: 'text-blue-800'
}

export const DashboardCard: React.FC<DashboardCardProps> = ({
  title,
  value,
  icon,
  borderColor,
  iconColor,
}) => {
  return (
    <div 
      className="bg-white rounded-xl overflow-hidden shadow-sm hover:shadow-md transition-all duration-300 relative cursor-default p-5 border border-slate-100"
    >
      {/* Borda lateral esquerda colorida */}
      <div className={`absolute top-0 left-0 w-1.5 h-full ${borderColor} rounded-l-xl`} />
      
      <div className="flex flex-col ml-2">
        {/* Título */}
        <p className="text-slate-500 text-xs font-semibold uppercase tracking-wider mb-3">
          {title}
        </p>
        
        {/* Ícone e Valor na mesma linha */}
        <div className="flex items-center space-x-3">
          <div className={`text-2xl flex items-center justify-center ${iconColor}`}>
            {icon}
          </div>
          <h3 className="text-3xl font-bold text-slate-800 m-0 leading-none">
            {value}
          </h3>
        </div>
      </div>
    </div>
  );
};
