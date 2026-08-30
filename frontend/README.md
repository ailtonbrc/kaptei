# Frontend do Kaptei

Aplicação web do painel SaaS e dos sites públicos de captação imobiliária.

## Tecnologias e responsabilidades

- React 18, TypeScript e Vite;
- Shadcn/Radix para os primitivos acessíveis e Tailwind para composição visual;
- Tremor para indicadores executivos;
- ECharts modular para gráficos analíticos;
- Zustand para sessão no cliente e React Router para rotas e carregamento sob demanda;
- serviços HTTP isolados em `src/services`, componentes reutilizáveis em `src/components` e regras auxiliares em `src/lib` e `src/hooks`.

Ant Design não faz parte desta aplicação. Essa decisão preserva o layout atual, evita dois sistemas visuais concorrentes e atende à diretriz arquitetural do repositório.

## Configuração local

1. Copie `.env.example` para `.env`.
2. Informe apenas valores do ambiente, sem versionar segredos.
3. Instale as dependências com `npm install`.
4. Inicie com `npm run dev`.

## Validação

```powershell
npm.cmd run typecheck
npm.cmd run lint
npm.cmd run build
```

As rotas são carregadas sob demanda. Componentes do Tremor devem usar importações granulares para não incorporar toda a biblioteca no chunk do Dashboard.
