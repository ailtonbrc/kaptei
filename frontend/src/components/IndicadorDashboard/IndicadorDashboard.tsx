import type { LucideIcon } from 'lucide-react';
import Card from '@tremor/react/dist/components/layout-elements/Card/Card.js';
import Metric from '@tremor/react/dist/components/text-elements/Metric/Metric.js';
import Text from '@tremor/react/dist/components/text-elements/Text/Text.js';

interface IndicadorDashboardProps {
  rotulo: string;
  valor: number;
  icone: LucideIcon;
  cor: 'blue' | 'emerald' | 'amber' | 'violet';
}

const estilos = { blue: 'text-blue-500', emerald: 'text-emerald-500', amber: 'text-amber-500', violet: 'text-violet-500' };

export function IndicadorDashboard({ rotulo, valor, icone: Icone, cor }: IndicadorDashboardProps) {
  return (
    <Card decoration="top" decorationColor={cor} className="border-slate-200 shadow-sm">
      <Text className="flex items-center gap-2 text-slate-600">
        <Icone className={`h-5 w-5 ${estilos[cor]}`} />
        {rotulo}
      </Text>
      <Metric className="mt-2 text-slate-900">{valor}</Metric>
    </Card>
  );
}
