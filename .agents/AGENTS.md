<RULE[project]>
# Diretrizes do Projeto KPTEI

1. **Stack Visual do Frontend (Substituição)**: Para este projeto, o **Ant Design está estritamente PROIBIDO**. Utilize exclusivamente a tríade baseada em Tailwind CSS:
   - **Shadcn/ui**: Para toda a estrutura da aplicação (botões, painéis, modais, etc.). 
   - **Tremor**: Para blocos visuais de resumo (cards de totais, mini gráficos de tendência).
   - **Apache ECharts**: Exclusivamente para gráficos complexos e grandes volumes de dados.
   
2. **Qualidade e Atualização**: Sempre trabalhar com os componentes e bibliotecas mais atualizados do mercado. 

3. **Arquitetura e Boas Práticas**:
   - Manter forte **Separação de Responsabilidades**.
   - **NUNCA usar Hardcode**. 
   
4. **Variáveis de Ambiente (.env)**:
   - O `.env` deve conter **APENAS** as configurações necessárias para INICIAR o projeto (DB DSN, PORTA, Chave Secreta mestre).
   - O restante das parametrizações **deve estar no Banco de Dados**.
</RULE[project]>
