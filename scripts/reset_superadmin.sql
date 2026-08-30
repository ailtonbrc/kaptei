-- ============================================================
-- Script: reset_superadmin.sql
-- Objetivo: Remover o superadmin fictício para permitir
--           o re-bootstrap via ferramenta segura.
--
-- ATENÇÃO: Execute apenas em ambiente local/desenvolvimento.
-- Em produção, utilize o procedimento de recuperação de senha.
-- ============================================================

BEGIN;

-- Remove o usuário superadmin e sua conta administrativa
DELETE FROM usuarios
WHERE papel = 'SUPER_ADMIN';

-- Remove a conta administrativa vinculada (sem usuários)
DELETE FROM contas_saas
WHERE nome_conta = 'Administração Kaptei'
  AND NOT EXISTS (
    SELECT 1 FROM usuarios u WHERE u.conta_id = contas_saas.id
  );

COMMIT;

-- Confirmar que não existe mais superadmin
SELECT COUNT(*) AS superadmins_restantes
FROM usuarios
WHERE papel = 'SUPER_ADMIN';
