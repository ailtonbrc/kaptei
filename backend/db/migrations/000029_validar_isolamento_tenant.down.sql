-- O PostgreSQL não permite tornar uma FK validada novamente NOT VALID sem recriá-la.
-- A reversão não enfraquece deliberadamente o isolamento multiempresa.
SELECT 1;
