package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

func (s *Store) DB() *pgxpool.Pool {
	return s.db
}

func (s *Store) CountUsers(ctx context.Context) (int, error) {
	var count int
	if err := s.db.QueryRow(ctx, `SELECT COUNT(1) FROM users`).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Store) CreateUser(ctx context.Context, user User) (User, error) {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	err := s.db.QueryRow(ctx, `
		INSERT INTO users (id, username, password_hash, is_admin, mfa_enabled, mfa_secret_enc, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,now())
		RETURNING created_at
	`, user.ID, user.Username, user.PasswordHash, user.IsAdmin, user.MFAEnabled, user.MFASecretEnc).Scan(&user.CreatedAt)
	return user, err
}

func (s *Store) GetUserByUsername(ctx context.Context, username string) (User, error) {
	var user User
	err := s.db.QueryRow(ctx, `
		SELECT id, username, password_hash, is_admin, mfa_enabled, mfa_secret_enc, mfa_verified_at, created_at
		FROM users WHERE username=$1
	`, username).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.IsAdmin, &user.MFAEnabled, &user.MFASecretEnc, &user.MFAVerifiedAt, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	return user, err
}

func (s *Store) GetUserByID(ctx context.Context, id uuid.UUID) (User, error) {
	var user User
	err := s.db.QueryRow(ctx, `
		SELECT id, username, password_hash, is_admin, mfa_enabled, mfa_secret_enc, mfa_verified_at, created_at
		FROM users WHERE id=$1
	`, id).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.IsAdmin, &user.MFAEnabled, &user.MFASecretEnc, &user.MFAVerifiedAt, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	return user, err
}

func (s *Store) UpdateUserMFA(ctx context.Context, userID uuid.UUID, enabled bool, secretEnc []byte, verifiedAt *time.Time) error {
	_, err := s.db.Exec(ctx, `
		UPDATE users SET mfa_enabled=$1, mfa_secret_enc=$2, mfa_verified_at=$3 WHERE id=$4
	`, enabled, secretEnc, verifiedAt, userID)
	return err
}

func (s *Store) UpdateUserAdmin(ctx context.Context, userID uuid.UUID, isAdmin bool) error {
	_, err := s.db.Exec(ctx, `UPDATE users SET is_admin=$1 WHERE id=$2`, isAdmin, userID)
	return err
}

func (s *Store) CreateSession(ctx context.Context, sess Session) (Session, error) {
	if sess.ID == uuid.Nil {
		sess.ID = uuid.New()
	}
	err := s.db.QueryRow(ctx, `
		INSERT INTO sessions (id, user_id, csrf_token, created_at, expires_at, last_auth_at, last_mfa_at)
		VALUES ($1,$2,$3,now(),$4,$5,$6)
		RETURNING created_at
	`, sess.ID, sess.UserID, sess.CSRFToken, sess.ExpiresAt, sess.LastAuthAt, sess.LastMFAAt).Scan(&sess.CreatedAt)
	return sess, err
}

func (s *Store) GetSession(ctx context.Context, id uuid.UUID) (Session, error) {
	var sess Session
	err := s.db.QueryRow(ctx, `
		SELECT id, user_id, csrf_token, created_at, expires_at, last_auth_at, last_mfa_at
		FROM sessions WHERE id=$1
	`, id).Scan(&sess.ID, &sess.UserID, &sess.CSRFToken, &sess.CreatedAt, &sess.ExpiresAt, &sess.LastAuthAt, &sess.LastMFAAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrNotFound
	}
	return sess, err
}

func (s *Store) DeleteSession(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.Exec(ctx, `DELETE FROM sessions WHERE id=$1`, id)
	return err
}

func (s *Store) UpdateSessionMFA(ctx context.Context, id uuid.UUID, lastMFAAt time.Time) error {
	_, err := s.db.Exec(ctx, `UPDATE sessions SET last_mfa_at=$1 WHERE id=$2`, lastMFAAt, id)
	return err
}

func (s *Store) UpdateSessionAuth(ctx context.Context, id uuid.UUID, lastAuthAt time.Time) error {
	_, err := s.db.Exec(ctx, `UPDATE sessions SET last_auth_at=$1 WHERE id=$2`, lastAuthAt, id)
	return err
}

func (s *Store) UpsertExternalIdentity(ctx context.Context, identity ExternalIdentity) (ExternalIdentity, error) {
	if identity.ID == uuid.Nil {
		identity.ID = uuid.New()
	}
	groups, _ := json.Marshal(identity.Groups)
	err := s.db.QueryRow(ctx, `
		INSERT INTO external_identities (id, user_id, provider, subject, email, groups, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,now())
		ON CONFLICT (provider, subject) DO UPDATE
		SET user_id=excluded.user_id, email=excluded.email, groups=excluded.groups
		RETURNING created_at
	`, identity.ID, identity.UserID, identity.Provider, identity.Subject, identity.Email, groups).Scan(&identity.CreatedAt)
	return identity, err
}

func (s *Store) GetExternalIdentity(ctx context.Context, provider string, subject string) (ExternalIdentity, error) {
	var identity ExternalIdentity
	var groups []byte
	err := s.db.QueryRow(ctx, `
		SELECT id, user_id, provider, subject, email, groups, created_at
		FROM external_identities WHERE provider=$1 AND subject=$2
	`, provider, subject).Scan(&identity.ID, &identity.UserID, &identity.Provider, &identity.Subject, &identity.Email, &groups, &identity.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ExternalIdentity{}, ErrNotFound
	}
	_ = json.Unmarshal(groups, &identity.Groups)
	return identity, err
}

func (s *Store) UpsertGroup(ctx context.Context, name string) (Group, error) {
	var group Group
	err := s.db.QueryRow(ctx, `
		INSERT INTO groups (id, name, created_at)
		VALUES (gen_random_uuid(), $1, now())
		ON CONFLICT (name) DO UPDATE SET name=excluded.name
		RETURNING id, name, created_at
	`, name).Scan(&group.ID, &group.Name, &group.CreatedAt)
	return group, err
}

func (s *Store) SetUserGroups(ctx context.Context, userID uuid.UUID, groupIDs []uuid.UUID) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM group_members WHERE user_id=$1`, userID); err != nil {
		return err
	}
	for _, gid := range groupIDs {
		if _, err := tx.Exec(ctx, `INSERT INTO group_members (user_id, group_id) VALUES ($1,$2)`, userID, gid); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *Store) ListUserRoles(ctx context.Context, userID uuid.UUID) ([]Role, error) {
	rows, err := s.db.Query(ctx, `
		SELECT r.id, r.name, r.permissions, r.created_at
		FROM roles r
		JOIN role_bindings rb ON rb.role_id = r.id
		LEFT JOIN group_members gm ON gm.group_id = rb.group_id
		WHERE rb.user_id=$1 OR gm.user_id=$1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var roles []Role
	for rows.Next() {
		var role Role
		var perms []byte
		if err := rows.Scan(&role.ID, &role.Name, &perms, &role.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(perms, &role.Permissions)
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (s *Store) CreateRoleIfMissing(ctx context.Context, name string, permissions []string) (Role, error) {
	var role Role
	perms, _ := json.Marshal(permissions)
	err := s.db.QueryRow(ctx, `
		INSERT INTO roles (id, name, permissions, created_at)
		VALUES (gen_random_uuid(), $1, $2, now())
		ON CONFLICT (name) DO UPDATE SET permissions=excluded.permissions
		RETURNING id, name, permissions, created_at
	`, name, perms).Scan(&role.ID, &role.Name, &perms, &role.CreatedAt)
	if err != nil {
		return Role{}, err
	}
	_ = json.Unmarshal(perms, &role.Permissions)
	return role, nil
}

func (s *Store) BindRoleToUser(ctx context.Context, roleID uuid.UUID, userID uuid.UUID, resource string) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO role_bindings (id, role_id, user_id, resource, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, now())
	`, roleID, userID, resource)
	return err
}

func (s *Store) ListPolicies(ctx context.Context) ([]Policy, error) {
	rows, err := s.db.Query(ctx, `SELECT id, name, doc, created_at FROM policies`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var policies []Policy
	for rows.Next() {
		var p Policy
		if err := rows.Scan(&p.ID, &p.Name, &p.Doc, &p.CreatedAt); err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}
	return policies, rows.Err()
}

func (s *Store) UpsertPolicy(ctx context.Context, name string, doc []byte) (Policy, error) {
	var p Policy
	err := s.db.QueryRow(ctx, `
		INSERT INTO policies (id, name, doc, created_at)
		VALUES (gen_random_uuid(), $1, $2, now())
		ON CONFLICT (name) DO UPDATE SET doc=excluded.doc
		RETURNING id, name, doc, created_at
	`, name, doc).Scan(&p.ID, &p.Name, &p.Doc, &p.CreatedAt)
	return p, err
}

func (s *Store) CreateConnectionSecret(ctx context.Context, encrypted []byte) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.db.QueryRow(ctx, `
		INSERT INTO connection_secrets (id, encrypted, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, now(), now())
		RETURNING id
	`, encrypted).Scan(&id)
	return id, err
}

func (s *Store) UpdateConnectionSecret(ctx context.Context, id uuid.UUID, encrypted []byte) error {
	_, err := s.db.Exec(ctx, `
		UPDATE connection_secrets SET encrypted=$1, updated_at=now() WHERE id=$2
	`, encrypted, id)
	return err
}

func (s *Store) GetConnectionSecret(ctx context.Context, id uuid.UUID) ([]byte, error) {
	var encrypted []byte
	err := s.db.QueryRow(ctx, `SELECT encrypted FROM connection_secrets WHERE id=$1`, id).Scan(&encrypted)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return encrypted, err
}

func (s *Store) CreateConnection(ctx context.Context, conn Connection) (Connection, error) {
	if conn.ID == uuid.Nil {
		conn.ID = uuid.New()
	}
	tlsData, _ := json.Marshal(conn.TLS)
	tagsData, _ := json.Marshal(conn.Tags)
	err := s.db.QueryRow(ctx, `
		INSERT INTO connections (id, name, type, host, port, database, username, secret_ref, tls, tags, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,now(),now())
		RETURNING created_at, updated_at
	`, conn.ID, conn.Name, conn.Type, conn.Host, conn.Port, conn.Database, conn.Username, conn.SecretRef, tlsData, tagsData).Scan(&conn.CreatedAt, &conn.UpdatedAt)
	return conn, err
}

func (s *Store) UpdateConnection(ctx context.Context, conn Connection) error {
	tlsData, _ := json.Marshal(conn.TLS)
	tagsData, _ := json.Marshal(conn.Tags)
	_, err := s.db.Exec(ctx, `
		UPDATE connections
		SET name=$1, type=$2, host=$3, port=$4, database=$5, username=$6, tls=$7, tags=$8, updated_at=now()
		WHERE id=$9
	`, conn.Name, conn.Type, conn.Host, conn.Port, conn.Database, conn.Username, tlsData, tagsData, conn.ID)
	return err
}

func (s *Store) DeleteConnection(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.Exec(ctx, `DELETE FROM connections WHERE id=$1`, id)
	return err
}

func (s *Store) GetConnection(ctx context.Context, id uuid.UUID) (Connection, error) {
	var conn Connection
	var tlsData []byte
	var tagsData []byte
	err := s.db.QueryRow(ctx, `
		SELECT id, name, type, host, port, database, username, secret_ref, tls, tags, created_at, updated_at
		FROM connections WHERE id=$1
	`, id).Scan(&conn.ID, &conn.Name, &conn.Type, &conn.Host, &conn.Port, &conn.Database, &conn.Username, &conn.SecretRef, &tlsData, &tagsData, &conn.CreatedAt, &conn.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Connection{}, ErrNotFound
	}
	_ = json.Unmarshal(tlsData, &conn.TLS)
	_ = json.Unmarshal(tagsData, &conn.Tags)
	return conn, err
}

func (s *Store) ListConnections(ctx context.Context) ([]Connection, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, name, type, host, port, database, username, secret_ref, tls, tags, created_at, updated_at
		FROM connections ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var conns []Connection
	for rows.Next() {
		var conn Connection
		var tlsData []byte
		var tagsData []byte
		if err := rows.Scan(&conn.ID, &conn.Name, &conn.Type, &conn.Host, &conn.Port, &conn.Database, &conn.Username, &conn.SecretRef, &tlsData, &tagsData, &conn.CreatedAt, &conn.UpdatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(tlsData, &conn.TLS)
		_ = json.Unmarshal(tagsData, &conn.Tags)
		conns = append(conns, conn)
	}
	return conns, rows.Err()
}

func (s *Store) CreateQueryHistory(ctx context.Context, history QueryHistory) (QueryHistory, error) {
	if history.ID == uuid.Nil {
		history.ID = uuid.New()
	}
	err := s.db.QueryRow(ctx, `
		INSERT INTO query_history (id, user_id, connection_id, statement_hash, status, row_count, duration_ms, started_at, ended_at, action, resource, approval_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING started_at
	`, history.ID, history.UserID, history.ConnectionID, history.StatementHash, history.Status, history.RowCount, history.DurationMs, history.StartedAt, history.EndedAt, history.Action, history.Resource, history.ApprovalID).Scan(&history.StartedAt)
	return history, err
}

func (s *Store) UpdateQueryHistory(ctx context.Context, history QueryHistory) error {
	_, err := s.db.Exec(ctx, `
		UPDATE query_history
		SET status=$1, row_count=$2, duration_ms=$3, ended_at=$4
		WHERE id=$5
	`, history.Status, history.RowCount, history.DurationMs, history.EndedAt, history.ID)
	return err
}

func (s *Store) ListHistory(ctx context.Context, limit int, offset int) ([]QueryHistory, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, user_id, connection_id, statement_hash, status, row_count, duration_ms, started_at, ended_at, action, resource, approval_id
		FROM query_history ORDER BY started_at DESC LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []QueryHistory
	for rows.Next() {
		var q QueryHistory
		if err := rows.Scan(&q.ID, &q.UserID, &q.ConnectionID, &q.StatementHash, &q.Status, &q.RowCount, &q.DurationMs, &q.StartedAt, &q.EndedAt, &q.Action, &q.Resource, &q.ApprovalID); err != nil {
			return nil, err
		}
		list = append(list, q)
	}
	return list, rows.Err()
}

func (s *Store) ListUsers(ctx context.Context, limit int, offset int) ([]User, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, username, password_hash, is_admin, mfa_enabled, mfa_secret_enc, mfa_verified_at, created_at
		FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.IsAdmin, &u.MFAEnabled, &u.MFASecretEnc, &u.MFAVerifiedAt, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (s *Store) ListGroups(ctx context.Context, limit int, offset int) ([]Group, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, name, created_at FROM groups ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var groups []Group
	for rows.Next() {
		var g Group
		if err := rows.Scan(&g.ID, &g.Name, &g.CreatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

func (s *Store) CreateQueryApproval(ctx context.Context, approval QueryApproval) (QueryApproval, error) {
	if approval.ID == uuid.Nil {
		approval.ID = uuid.New()
	}
	err := s.db.QueryRow(ctx, `
		INSERT INTO query_approvals (id, connection_id, user_id, statement, status, environment, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,now())
		RETURNING created_at
	`, approval.ID, approval.ConnectionID, approval.UserID, approval.Statement, approval.Status, approval.Environment).Scan(&approval.CreatedAt)
	return approval, err
}

func (s *Store) GetQueryApproval(ctx context.Context, id uuid.UUID) (QueryApproval, error) {
	var q QueryApproval
	err := s.db.QueryRow(ctx, `
		SELECT id, connection_id, user_id, statement, status, environment, created_at, approved_by, approved_at, denied_by, denied_at
		FROM query_approvals WHERE id=$1
	`, id).Scan(&q.ID, &q.ConnectionID, &q.UserID, &q.Statement, &q.Status, &q.Environment, &q.CreatedAt, &q.ApprovedBy, &q.ApprovedAt, &q.DeniedBy, &q.DeniedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return QueryApproval{}, ErrNotFound
	}
	return q, err
}

func (s *Store) ListPendingApprovals(ctx context.Context) ([]QueryApproval, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, connection_id, user_id, statement, status, environment, created_at, approved_by, approved_at, denied_by, denied_at
		FROM query_approvals WHERE status='pending' ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []QueryApproval
	for rows.Next() {
		var q QueryApproval
		if err := rows.Scan(&q.ID, &q.ConnectionID, &q.UserID, &q.Statement, &q.Status, &q.Environment, &q.CreatedAt, &q.ApprovedBy, &q.ApprovedAt, &q.DeniedBy, &q.DeniedAt); err != nil {
			return nil, err
		}
		list = append(list, q)
	}
	return list, rows.Err()
}

func (s *Store) ApproveQuery(ctx context.Context, id uuid.UUID, approver uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE query_approvals SET status='approved', approved_by=$1, approved_at=now() WHERE id=$2
	`, approver, id)
	return err
}

func (s *Store) DenyQuery(ctx context.Context, id uuid.UUID, approver uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE query_approvals SET status='denied', denied_by=$1, denied_at=now() WHERE id=$2
	`, approver, id)
	return err
}

func (s *Store) ListPIIRules(ctx context.Context, connectionID uuid.UUID) ([]PIIRule, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, connection_id, resource, field, mask_type, created_at
		FROM pii_rules WHERE connection_id=$1
	`, connectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rules []PIIRule
	for rows.Next() {
		var r PIIRule
		if err := rows.Scan(&r.ID, &r.ConnectionID, &r.Resource, &r.Field, &r.MaskType, &r.CreatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}
	return rules, rows.Err()
}

func (s *Store) ListAudit(ctx context.Context, limit int, offset int) ([]AuditEntry, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, event_type, actor_user_id, details, created_at, prev_hash, hash, error_id
		FROM audit_log ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []AuditEntry
	for rows.Next() {
		var entry AuditEntry
		if err := rows.Scan(&entry.ID, &entry.EventType, &entry.ActorID, &entry.Details, &entry.CreatedAt, &entry.PrevHash, &entry.Hash, &entry.ErrorID); err != nil {
			return nil, err
		}
		list = append(list, entry)
	}
	return list, rows.Err()
}

var ErrNotFound = errors.New("not found")
