package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RunMigrations applies idempotent SQL migrations required for the service.
// For production, prefer using a dedicated migration tool, but this ensures
// local development works out-of-the-box.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	statements := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`,
		`CREATE TABLE IF NOT EXISTS stores (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			shop_domain TEXT UNIQUE NOT NULL,
			shop_name TEXT,
			encrypted_access_token BYTEA NOT NULL,
			api_key_public TEXT UNIQUE NOT NULL,
			api_key_private TEXT NOT NULL,
			product_index_uid TEXT NOT NULL,
			meilisearch_index_uid TEXT NOT NULL,
			meilisearch_document_type TEXT NOT NULL DEFAULT 'product',
			meilisearch_url TEXT,
			meilisearch_api_key BYTEA,
			plan_level TEXT NOT NULL DEFAULT 'free',
			status TEXT NOT NULL DEFAULT 'active',
			webhook_secret TEXT NOT NULL,
			installed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			uninstalled_at TIMESTAMPTZ,
			sync_state JSONB DEFAULT '{}'::jsonb,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
		`CREATE INDEX IF NOT EXISTS idx_stores_shop_domain ON stores (shop_domain);`,
		`CREATE INDEX IF NOT EXISTS idx_stores_api_key_public ON stores (api_key_public);`,
		`ALTER TABLE stores ADD COLUMN IF NOT EXISTS meilisearch_index_uid TEXT;`,
		`UPDATE stores SET meilisearch_index_uid = product_index_uid WHERE meilisearch_index_uid IS NULL AND product_index_uid IS NOT NULL;`,
		`ALTER TABLE stores ALTER COLUMN meilisearch_index_uid SET NOT NULL;`,
		`ALTER TABLE stores ADD COLUMN IF NOT EXISTS meilisearch_document_type TEXT;`,
		`UPDATE stores SET meilisearch_document_type = 'product' WHERE meilisearch_document_type IS NULL;`,
		`ALTER TABLE stores ALTER COLUMN meilisearch_document_type SET NOT NULL;`,
		`ALTER TABLE stores ADD COLUMN IF NOT EXISTS meilisearch_url TEXT;`,
		`ALTER TABLE stores ADD COLUMN IF NOT EXISTS meilisearch_api_key BYTEA;`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id VARCHAR(255) PRIMARY KEY,
			shop VARCHAR(255) NOT NULL,
			state VARCHAR(255) NOT NULL,
			is_online BOOLEAN DEFAULT FALSE,
			scope TEXT,
			expires TIMESTAMPTZ,
			access_token TEXT NOT NULL,
			user_id BIGINT,
			first_name VARCHAR(255),
			last_name VARCHAR(255),
			email VARCHAR(255),
			account_owner BOOLEAN DEFAULT FALSE,
			locale VARCHAR(10),
			collaborator BOOLEAN,
			email_verified BOOLEAN,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_shop ON sessions(shop);`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires);`,
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin migration tx: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, stmt := range statements {
		if _, err := tx.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("migration failed for statement %q: %w", stmt, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit migrations: %w", err)
	}

	return nil
}
