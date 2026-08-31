package entity

import "time"

type Audit struct {
	CreatedAt time.Time  `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time  `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP"`
	DeletedAt *time.Time `gorm:"column:deleted_at;type:timestamp;default:NULL"`
	IsDeleted bool       `gorm:"column:is_deleted;type:boolean;default:false"`
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
