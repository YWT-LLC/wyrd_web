/* wyrd_web
 * Copyright (c) 2025 Empathetech LLC. All rights reserved.
 * See LICENSE for distribution and usage details.
 */

package models

import (
	"time"
	
	"github.com/google/uuid"
)

type Group struct {
	ID        	uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name      	string    `gorm:"not null"`
	Description string    `gorm:"not null"`
	Visibility 	string    `gorm:"not null"`
	CreatedAt 	time.Time
	UpdatedAt 	time.Time
}

type GroupMember struct {
	ID        	uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()"`
	GroupID   	uuid.UUID `gorm:"type:uuid;not null"`
	UserID    	uuid.UUID `gorm:"type:uuid;not null"`
	Role      	string    `gorm:"not null"`
	CreatedAt 	time.Time
	UpdatedAt 	time.Time
}