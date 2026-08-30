export const validarCPF = (cpf: string): boolean => {
  const digitos = cpf.replace(/\D/g, '');
  if (digitos.length !== 11 || /^(\d)\1+$/.test(digitos)) return false;

  const calcularDigito = (tamanho: number): number => {
    let soma = 0;
    for (let indice = 0; indice < tamanho; indice += 1) {
      soma += Number(digitos[indice]) * (tamanho + 1 - indice);
    }
    const resto = (soma * 10) % 11;
    return resto === 10 ? 0 : resto;
  };

  return calcularDigito(9) === Number(digitos[9])
    && calcularDigito(10) === Number(digitos[10]);
};

export const formatarCPF = (valor: string): string => {
  const digitos = valor.replace(/\D/g, '').slice(0, 11);
  if (digitos.length > 9) return `${digitos.slice(0, 3)}.${digitos.slice(3, 6)}.${digitos.slice(6, 9)}-${digitos.slice(9)}`;
  if (digitos.length > 6) return `${digitos.slice(0, 3)}.${digitos.slice(3, 6)}.${digitos.slice(6)}`;
  if (digitos.length > 3) return `${digitos.slice(0, 3)}.${digitos.slice(3)}`;
  return digitos;
};
