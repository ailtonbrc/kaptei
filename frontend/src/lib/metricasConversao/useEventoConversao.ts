import { useEffect } from 'react';
import { registrarEventoConversao } from './metricasConversao';
import type { TipoEventoConversao } from '../../types/sitePublico';

export function useEventoConversao(slugSite: string, tipo: TipoEventoConversao, slugImovel?: string) {
  useEffect(() => {
    void registrarEventoConversao(slugSite, tipo, slugImovel);
  }, [slugImovel, slugSite, tipo]);
}
