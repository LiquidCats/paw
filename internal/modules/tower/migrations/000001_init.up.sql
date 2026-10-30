CREATE TABLE IF NOT EXISTS towers (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tower_versions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    tower_id UUID NOT NULL REFERENCES towers(id),
    version INT NOT NULL,
    signature VARCHAR(255) NOT NULL,
    signature_algorithm VARCHAR(255) NOT NULL,
    digest VARCHAR(255) NOT NULL,
    digest_algorithm VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tower_versions_tower_id ON tower_versions(tower_id);

CREATE TYPE user_role AS ENUM ('ADMIN', 'MEMBER');

CREATE TABLE IF NOT EXISTS tower_users (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    tower_id UUID NOT NULL REFERENCES towers(id),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    role user_role NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tower_uses_tower_id ON tower_users(tower_id);

CREATE TABLE IF NOT EXISTS tower_user_credentials (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    tower_id UUID NOT NULL REFERENCES towers(id),
    user_id UUID NOT NULL REFERENCES tower_users(id),
    type VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TYPE effect_type AS ENUM ('ALLOW', 'DENY');

CREATE TABLE IF NOT EXISTS tower_policies (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    tower_id UUID NOT NULL REFERENCES towers(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    consensus TEXT DEFAULT '',
    constraints TEXT DEFAULT '',
    effect effect_type NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tower_policies_tower_id ON tower_policies(tower_id);

CREATE TABLE IF NOT EXISTS tower_seeds (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    tower_id UUID NOT NULL REFERENCES towers(id),
    envelope TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tower_seeds_tower_id ON tower_seeds(tower_id);

CREATE TABLE IF NOT EXISTS tower_walls (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    tower_id UUID NOT NULL REFERENCES towers(id),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
