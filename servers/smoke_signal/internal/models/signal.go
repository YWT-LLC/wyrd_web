/* wyrd_web
 * Copyright (c) 2025 Empathetech LLC. All rights reserved.
 * See LICENSE for distribution and usage details.
 */

package models

import (
	"time"
	"github.com/google/uuid"
)

type User struct {
	ID  				uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	GroupID			uuid.UUID `gorm:"type:uuid;not null"`
	CreatorID		uuid.UUID `gorm:"type:uuid;not null"`
	Name				string    `gorm:"not null"`
	Description	string    `gorm:"not null"`
	CreatedAt		time.Time
	UpdatedAt		time.Time
}
