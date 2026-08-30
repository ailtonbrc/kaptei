import { Suspense } from 'react';
import { RouterProvider } from 'react-router-dom';
import { router } from './routes';
import { Toaster } from '@/components/ui/sonner';
import { Loader2 } from 'lucide-react';

const PageLoader = () => (
  <div className="min-h-screen w-full flex flex-col items-center justify-center bg-slate-950">
    <div className="flex items-center gap-3 mb-6 transition-transform hover:scale-105 duration-300">
      <div className="bg-gradient-to-br from-blue-500 to-blue-700 p-2.5 rounded-xl shadow-lg shadow-blue-900/50 border border-blue-400/20">
        <svg className="w-8 h-8 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
        </svg>
      </div>
      <h1 className="text-4xl font-extrabold text-transparent bg-clip-text bg-gradient-to-r from-white to-blue-100 tracking-tight m-0">KAPTEI</h1>
    </div>
    <Loader2 className="h-8 w-8 animate-spin text-blue-500" />
    <p className="mt-4 text-slate-400 text-sm font-medium">Carregando sistema...</p>
  </div>
);

function App() {
  return (
    <Suspense fallback={<PageLoader />}>
      <RouterProvider router={router} />
      <Toaster />
    </Suspense>
  );
}

export default App;
