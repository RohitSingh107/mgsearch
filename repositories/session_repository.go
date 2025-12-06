package repositories

import (
	"context"
	"errors"
	"fmt"

	"mgsearch/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

func (r *SessionRepository) scanSession(row pgx.Row) (*models.Session, error) {
	session := &models.Session{}
	err := row.Scan(
		&session.ID,
		&session.Shop,
		&session.State,
		&session.IsOnline,
		&session.Scope,
		&session.Expires,
		&session.AccessToken,
		&session.UserID,
		&session.FirstName,
		&session.LastName,
		&session.Email,
		&session.AccountOwner,
		&session.Locale,
		&session.Collaborator,
		&session.EmailVerified,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (r *SessionRepository) CreateOrUpdate(ctx context.Context, session *models.Session) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO sessions (
			id, shop, state, is_online, scope, expires, access_token,
			user_id, first_name, last_name, email, account_owner,
			locale, collaborator, email_verified, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW(), NOW())
		ON CONFLICT (id) DO UPDATE SET
			shop = EXCLUDED.shop,
			state = EXCLUDED.state,
			is_online = EXCLUDED.is_online,
			scope = EXCLUDED.scope,
			expires = EXCLUDED.expires,
			access_token = EXCLUDED.access_token,
			user_id = EXCLUDED.user_id,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			email = EXCLUDED.email,
			account_owner = EXCLUDED.account_owner,
			locale = EXCLUDED.locale,
			collaborator = EXCLUDED.collaborator,
			email_verified = EXCLUDED.email_verified,
			updated_at = NOW()
	`,
		session.ID,
		session.Shop,
		session.State,
		session.IsOnline,
		session.Scope,
		session.Expires,
		session.AccessToken,
		session.UserID,
		session.FirstName,
		session.LastName,
		session.Email,
		session.AccountOwner,
		session.Locale,
		session.Collaborator,
		session.EmailVerified,
	)
	return err
}

func (r *SessionRepository) GetByID(ctx context.Context, id string) (*models.Session, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, shop, state, is_online, scope, expires, access_token,
		       user_id, first_name, last_name, email, account_owner,
		       locale, collaborator, email_verified, created_at, updated_at
		FROM sessions WHERE id = $1
	`, id)

	session, err := r.scanSession(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("session not found")
		}
		return nil, err
	}
	return session, nil
}

func (r *SessionRepository) DeleteByID(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, id)
	return err
}

func (r *SessionRepository) DeleteByIDs(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE id = ANY($1)`, ids)
	return err
}

func (r *SessionRepository) DeleteExpired(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE expires < NOW()`)
	return err
}

func (r *SessionRepository) GetByShop(ctx context.Context, shop string) ([]*models.Session, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, shop, state, is_online, scope, expires, access_token,
		       user_id, first_name, last_name, email, account_owner,
		       locale, collaborator, email_verified, created_at, updated_at
		FROM sessions WHERE shop = $1 ORDER BY created_at DESC
	`, shop)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*models.Session
	for rows.Next() {
		session, err := r.scanSession(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}
