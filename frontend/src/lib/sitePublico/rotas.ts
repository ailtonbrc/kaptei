export const rotaPublica = (basePublica: string, caminho = '') => {
  const base = basePublica === '/' ? '' : basePublica.replace(/\/$/, '');
  const sufixo = caminho.replace(/^\//, '');
  return sufixo ? `${base}/${sufixo}` : (base || '/');
};

