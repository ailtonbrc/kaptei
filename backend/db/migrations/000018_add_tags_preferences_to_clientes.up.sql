ALTER TABLE clientes ADD COLUMN tags JSONB DEFAULT '[]'::jsonb;
ALTER TABLE clientes ADD COLUMN preferencias JSONB DEFAULT '{}'::jsonb;
CREATE INDEX idx_clientes_tags ON clientes USING GIN (tags);