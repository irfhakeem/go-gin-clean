package entities

import "time"

type Audit struct {
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"column:deleted_at;type:timestamp;default:NULL"`
	IsDeleted bool       `json:"is_deleted" gorm:"column:is_deleted;type:boolean;default:false"`
}

func (a *Audit) MarkAsDeleted() {
	now := time.Now()
	a.DeletedAt = &now
	a.IsDeleted = true
}

func (a *Audit) RestoreFromDeletion() {
	a.DeletedAt = nil
	a.IsDeleted = false
}

func (a *Audit) UpdateTimestamp() {
	a.UpdatedAt = time.Now()
}
