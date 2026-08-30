interface ImovelComValores {
  finalidade: string;
  valor_venda?: number | null;
  valor_locacao?: number | null;
}

export interface ValorImovel {
  rotulo: 'Venda' | 'Locação';
  valor: number;
}

export const listarValoresImovel = (imovel: ImovelComValores): ValorImovel[] => {
  const valores: ValorImovel[] = [];
  if (imovel.finalidade.includes('Venda') && imovel.valor_venda) valores.push({ rotulo: 'Venda', valor: imovel.valor_venda });
  if (imovel.finalidade.includes('Locação') && imovel.valor_locacao) valores.push({ rotulo: 'Locação', valor: imovel.valor_locacao });
  return valores;
};
