package models

import "time"

// Session represents a Shopify OAuth session stored in the backend.
// This is used by Remix frontend for session management.
type Session struct {
	ID            string     `json:"id" bson:"_id"`
	Shop          string     `json:"shop" bson:"shop"`
	State         string     `json:"state" bson:"state"`
	IsOnline      bool       `json:"isOnline" bson:"is_online"`
	Scope         string     `json:"scope" bson:"scope"`
	Expires       *time.Time `json:"expires,omitempty" bson:"expires,omitempty"`
	AccessToken   string     `json:"accessToken" bson:"access_token"`
	UserID        *int64     `json:"userId,omitempty" bson:"user_id,omitempty"`
	FirstName     *string    `json:"firstName,omitempty" bson:"first_name,omitempty"`
	LastName      *string    `json:"lastName,omitempty" bson:"last_name,omitempty"`
	Email         *string    `json:"email,omitempty" bson:"email,omitempty"`
	AccountOwner  bool       `json:"accountOwner" bson:"account_owner"`
	Locale        *string    `json:"locale,omitempty" bson:"locale,omitempty"`
	Collaborator  *bool      `json:"collaborator,omitempty" bson:"collaborator,omitempty"`
	EmailVerified *bool      `json:"emailVerified,omitempty" bson:"email_verified,omitempty"`
	CreatedAt     time.Time  `json:"createdAt" bson:"created_at"`
	UpdatedAt     time.Time  `json:"updatedAt" bson:"updated_at"`
}
