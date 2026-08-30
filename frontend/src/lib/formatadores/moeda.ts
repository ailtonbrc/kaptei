export const formatarMoedaBR = (valor: number): string => {
  if (!Number.isFinite(valor)) return '';
  return valor.toLocaleString('pt-BR', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
};

export const converterMoedaBRParaNumero = (texto: string): number => {
  const apenasDigitos = texto.replace(/\D/g, '');
  return apenasDigitos ? Number.parseInt(apenasDigitos, 10) / 100 : 0;
};
