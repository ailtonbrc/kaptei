# Qualidade do frontend

## Estratégia de interface

O Kaptei mantém o layout atual e seu design system baseado em Shadcn/ui, Tailwind, Tremor para indicadores e ECharts modular para visualizações analíticas. Ant Design não faz parte da arquitetura do projeto. A evolução deve preservar essa identidade e extrair primitives reutilizáveis, sem introduzir uma segunda biblioteca visual concorrente.

## Orçamento de JavaScript

O build de produção aplica limites por chunk e falha quando houver regressão relevante:

- entrada principal: até 350 kB sem compressão;
- rota do Dashboard sem o motor de gráficos: até 50 kB;
- calendário: até 225 kB;
- motor ECharts isolado: até 575 kB;
- demais chunks: até 250 kB.

Os limites são guardrails, não metas de consumo. Qualquer aumento precisa ser explicado por uma capacidade de produto e medido também em rede e aparelho reais. Bibliotecas pesadas devem permanecer fora da entrada principal e ser carregadas apenas na rota ou interação que as utiliza.

Na medição de 7 de agosto de 2026, o Dashboard passou de aproximadamente 589 kB para 31 kB sem compressão na entrada da rota. O ECharts ficou em chunk próprio, carregado depois dos indicadores e acompanhado por um resumo textual acessível.

## Requisitos de acessibilidade

- uma região `main` identificável e link para pular a navegação;
- navegação ativa com `aria-current="page"`;
- menu móvel modal, com foco contido, fechamento por Escape e sem controles focáveis quando fechado;
- controles apenas com ícone devem possuir nome acessível;
- estados de carregamento e erro devem ser anunciados;
- gráficos precisam de título e resumo textual equivalente;
- formulários devem manter rótulo associado, mensagem de erro e foco visível;
- contraste, zoom em 200%, teclado e leitores de tela devem ser homologados em interface renderizada.

## Matriz mínima de validação

1. Desktop: Chrome/Edge atual, Firefox atual e Safari atual.
2. Mobile: Android/Chrome e iOS/Safari em aparelho real.
3. Teclado: ordem de foco, menus, modais, formulários e ausência de armadilhas.
4. Leitor de tela: NVDA + Edge/Chrome no Windows e VoiceOver + Safari no iOS/macOS.
5. Responsividade: 320, 375, 768, 1024 e 1440 px, além de zoom de 200%.
6. Desempenho: build dentro do orçamento, LCP/INP/CLS em homologação e rede móvel simulada.

Nesta revisão, tipagem, lint e build foram executados. A validação visual automatizada não ocorreu porque o processo do navegador local não conseguiu ler o workspace devido à ACL do Windows; ela permanece uma porta obrigatória da homologação final.
