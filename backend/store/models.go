package store

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID
	Username      string
	PasswordHash  string
	IsAdmin       bool
	MFAEnabled    bool
	MFASecretEnc  []byte
	MFAVerifiedAt *time.Time
	CreatedAt     time.Time
}

type Session struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	CSRFToken  string
	CreatedAt  time.Time
	ExpiresAt  time.Time
	LastAuthAt time.Time
	LastMFAAt  *time.Time
}

type ExternalIdentity struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Provider  string
	Subject   string
	Email     string
	Groups    []string
	CreatedAt time.Time
}

type Group struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
}

type Role struct {
	ID          uuid.UUID
	Name        string
	Permissions []string
	CreatedAt   time.Time
}

type RoleBinding struct {
	ID        uuid.UUID
	RoleID    uuid.UUID
	UserID    *uuid.UUID
	GroupID   *uuid.UUID
	Resource  string
	CreatedAt time.Time
}

type Policy struct {
	ID        uuid.UUID
	Name      string
	Doc       []byte
	CreatedAt time.Time
}

type Connection struct {
	ID        uuid.UUID
	Name      string
	Type      string
	Host      string
	Port      int
	Database  string
	Username  string
	SecretRef uuid.UUID
	TLS       map[string]any
	Tags      map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
}

type QueryHistory struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	ConnectionID  uuid.UUID
	StatementHash string
	Status        string
	RowCount      int
	DurationMs    int64
	StartedAt     time.Time
	EndedAt       *time.Time
	Action        string
	Resource      string
	ApprovalID    *uuid.UUID
}

type QueryApproval struct {
	ID           uuid.UUID
	ConnectionID uuid.UUID
	UserID       uuid.UUID
	Statement    string
	Status       string
	Environment  string
	CreatedAt    time.Time
	ApprovedBy   *uuid.UUID
	ApprovedAt   *time.Time
	DeniedBy     *uuid.UUID
	DeniedAt     *time.Time
}

type PIIRule struct {
	ID           uuid.UUID
	ConnectionID uuid.UUID
	Resource     string
	Field        string
	MaskType     string
	CreatedAt    time.Time
}

type AuditEntry struct {
	ID        uuid.UUID
	EventType string
	ActorID   *uuid.UUID
	Details   []byte
	CreatedAt time.Time
	PrevHash  string
	Hash      string
	ErrorID   string
}
