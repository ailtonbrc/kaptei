import { api } from './api';
import type { DadosDashboard } from '../types/dashboard';

export async function obterDashboardPremium(signal?: AbortSignal): Promise<DadosDashboard> {
  const resposta = await api.get<DadosDashboard>('/v1/negocios/dashboard/premium', { signal });
  return resposta.data;
}
