export interface ImovelFoto {
  id: string;
  imovel_id: string;
  url: string;
  url_thumbnail?: string;
  tipo_conteudo?: string;
  tamanho_bytes?: number;
  largura?: number;
  altura?: number;
  hash_sha256?: string;
  ordem: number;
  is_capa: boolean;
  criado_em: string;
}

export interface Imovel {
  id: string;
  conta_id: string;
  usuario_id: string;
  
  titulo: string;
  tipo: string;
  finalidade: string;
  status: string;
  
  valor_venda?: number;
  valor_locacao?: number;
  valor_condominio?: number;
  valor_iptu?: number;
  
  area_total?: number;
  area_util?: number;
  quartos: number;
  suites: number;
  banheiros: number;
  vagas: number;
  
  cep?: string;
  logradouro?: string;
  numero?: string;
  complemento?: string;
  bairro?: string;
  cidade?: string;
  estado?: string;
  
  descricao?: string;
	slug_publico?: string;
	publicado: boolean;
	destaque: boolean;
	titulo_seo?: string;
	descricao_seo?: string;
  
  criado_em: string;
  atualizado_em: string;

  fotos?: ImovelFoto[];
}
