import { useEffect } from 'react';

interface MetadadosPagina {
  titulo: string;
  descricao?: string;
  imagem?: string;
  dadosEstruturados?: Record<string, unknown>;
}

const definirMeta = (nome: string, conteudo: string, propriedade = false) => {
  const atributo = propriedade ? 'property' : 'name';
  let elemento = document.head.querySelector<HTMLMetaElement>(`meta[${atributo}="${nome}"]`);
  if (!elemento) {
    elemento = document.createElement('meta');
    elemento.setAttribute(atributo, nome);
    document.head.appendChild(elemento);
  }
  elemento.content = conteudo;
};

export const useMetadadosPagina = ({ titulo, descricao, imagem, dadosEstruturados }: MetadadosPagina) => {
  useEffect(() => {
    const tituloAnterior = document.title;
    document.title = titulo;
    if (descricao) {
      definirMeta('description', descricao);
      definirMeta('og:description', descricao, true);
    }
    definirMeta('og:title', titulo, true);
    if (imagem) definirMeta('og:image', imagem, true);

    let script: HTMLScriptElement | null = null;
    if (dadosEstruturados) {
      script = document.createElement('script');
      script.type = 'application/ld+json';
      script.dataset.kapteiSeo = 'true';
      script.text = JSON.stringify(dadosEstruturados);
      document.head.appendChild(script);
    }

    return () => {
      document.title = tituloAnterior;
      script?.remove();
    };
  }, [dadosEstruturados, descricao, imagem, titulo]);
};
