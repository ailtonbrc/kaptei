export interface Plano {
  id: string;
  codigo: string;
  tipo: string;
  nome: string;
  preco: number;
  recomendado?: boolean;
  subtitle?: string;
  cor: string;
  features: string[];
  missing: string[];
}
