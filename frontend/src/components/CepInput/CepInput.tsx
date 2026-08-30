import React, { useRef, useState } from 'react';

interface EnderecoViaCEP {
  logradouro?: string;
  bairro?: string;
  localidade?: string;
  uf?: string;
  erro?: boolean;
}

interface AddressData {
  logradouro: string;
  bairro: string;
  cidade: string;
  estado: string;
}

interface CepInputProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'onChange'> {
  value: string;
  onChange: (evento: React.ChangeEvent<HTMLInputElement>) => void;
  onAddressFetch?: (endereco: AddressData) => void;
  name?: string;
  label?: string;
}

const formatarCEP = (valor: string): string => {
  const digitos = valor.replace(/\D/g, '').slice(0, 8);
  return digitos.length > 5 ? `${digitos.slice(0, 5)}-${digitos.slice(5)}` : digitos;
};

export const CepInput: React.FC<CepInputProps> = ({
  value,
  onChange,
  onAddressFetch,
  name = 'cep',
  label = 'CEP',
  className = '',
  ...props
}) => {
  const [carregando, setCarregando] = useState(false);
  const [erro, setErro] = useState<string | null>(null);
  const ultimaConsulta = useRef('');

  const buscarEndereco = async (cep: string) => {
    if (ultimaConsulta.current === cep) return;
    ultimaConsulta.current = cep;

    try {
      setCarregando(true);
      setErro(null);
      const resposta = await fetch(`https://viacep.com.br/ws/${cep}/json/`);
      if (!resposta.ok) throw new Error(`ViaCEP respondeu com status ${resposta.status}`);
      const dados = await resposta.json() as EnderecoViaCEP;

      if (dados.erro) {
        setErro('CEP não encontrado');
        return;
      }

      onAddressFetch?.({
        logradouro: dados.logradouro ?? '',
        bairro: dados.bairro ?? '',
        cidade: dados.localidade ?? '',
        estado: dados.uf ?? '',
      });
    } catch {
      ultimaConsulta.current = '';
      setErro('Erro ao buscar o CEP');
    } finally {
      setCarregando(false);
    }
  };

  const handleChange = (evento: React.ChangeEvent<HTMLInputElement>) => {
    const valorFormatado = formatarCEP(evento.target.value);
    const digitos = valorFormatado.replace(/\D/g, '');
    if (digitos.length !== 8) {
      ultimaConsulta.current = '';
      setErro(null);
    }

    const eventoFormatado = {
      ...evento,
      target: { ...evento.target, name, value: valorFormatado },
    };
    onChange(eventoFormatado as React.ChangeEvent<HTMLInputElement>);
    if (digitos.length === 8) void buscarEndereco(digitos);
  };

  return (
    <div className="space-y-1.5 w-full">
      {label && (
        <label className="block text-sm font-medium text-slate-700">
          {label}
          {props.required && <span className="text-red-500 ml-1">*</span>}
        </label>
      )}
      <div className="relative">
        <input
          type="text"
          name={name}
          value={value}
          onChange={handleChange}
          placeholder="00000-000"
          maxLength={9}
          className={`w-full px-3 py-2 border rounded-lg focus:ring-2 outline-none transition-colors ${
            erro
              ? 'border-red-300 focus:border-red-400 focus:ring-red-500/20 bg-red-50/30'
              : 'border-slate-200 focus:border-blue-600 focus:ring-blue-600/50'
          } ${className}`}
          {...props}
        />
        {carregando && (
          <div className="absolute right-3 top-1/2 -translate-y-1/2 text-xs text-blue-600 font-medium animate-pulse">
            Buscando...
          </div>
        )}
      </div>
      {erro && <span className="text-xs font-medium text-red-500 mt-1 block">{erro}</span>}
    </div>
  );
};
