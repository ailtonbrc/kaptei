import { Building2 } from 'lucide-react';
import { Link, useRouteError } from 'react-router-dom';

export const ErroPublico = () => {
  const erro = useRouteError();
  const naoEncontrado = erro instanceof Response && erro.status === 404;
  return (
    <main className="grid min-h-screen place-items-center bg-slate-950 p-6 text-center text-white">
      <div>
        <Building2 className="mx-auto mb-5 h-14 w-14 text-blue-400" />
        <h1 className="text-3xl font-extrabold">{naoEncontrado ? 'Página não encontrada' : 'Não foi possível carregar esta página'}</h1>
        <p className="mx-auto mt-3 max-w-md text-slate-400">Confira o endereço informado ou tente novamente em alguns instantes.</p>
        <Link to="/login" className="mt-7 inline-flex rounded-xl bg-blue-600 px-5 py-3 font-bold hover:bg-blue-700">Acessar o Kaptei</Link>
      </div>
    </main>
  );
};
