package models

import "time"

// Session represents a Shopify OAuth session stored in the backend.
// This is used by Remix frontend for session management.
type Session struct {
	ID            string     `json:"id" db:"id"`
	Shop          string     `json:"shop" db:"shop"`
	State         string     `json:"state" db:"state"`
	IsOnline      bool       `json:"isOnline" db:"is_online"`
	Scope         string     `json:"scope" db:"scope"`
	Expires       *time.Time `json:"expires,omitempty" db:"expires"`
	AccessToken   string     `json:"accessToken" db:"access_token"`
	UserID        *int64     `json:"userId,omitempty" db:"user_id"`
	FirstName     *string    `json:"firstName,omitempty" db:"first_name"`
	LastName      *string    `json:"lastName,omitempty" db:"last_name"`
	Email         *string    `json:"email,omitempty" db:"email"`
	AccountOwner  bool       `json:"accountOwner" db:"account_owner"`
	Locale        *string    `json:"locale,omitempty" db:"locale"`
	Collaborator  *bool      `json:"collaborator,omitempty" db:"collaborator"`
	EmailVerified *bool      `json:"emailVerified,omitempty" db:"email_verified"`
	CreatedAt     time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time  `json:"updatedAt" db:"updated_at"`
}
