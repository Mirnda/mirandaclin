package repository

import (
	"context"
	"time"

	profileblock "github.com/Mirnda/mirandaclin/internal/domain/profile_block"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type profileBlockRepository struct{}

func NewProfileBlockRepository() profileblock.Repository {
	return &profileBlockRepository{}
}

func (r *profileBlockRepository) Create(ctx context.Context, db *gorm.DB, b *profileblock.ProfileBlock) error {
	return db.WithContext(ctx).Create(b).Error
}

func (r *profileBlockRepository) Delete(ctx context.Context, db *gorm.DB, tenantID, id uuid.UUID) error {
	return db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Delete(&profileblock.ProfileBlock{}).Error
}

// FindBlocksForSlot retorna bloqueios que cobrem a data e intervalo solicitados.
// Considera bloqueios sem clinic_id (valem para todas as clínicas) e com clinic_id específico.
func (r *profileBlockRepository) FindBlocksForSlot(
	ctx context.Context, db *gorm.DB,
	tenantID, profileID uuid.UUID, clinicID *uuid.UUID,
	date time.Time, start, end string,
) ([]profileblock.ProfileBlock, error) {
	var blocks []profileblock.ProfileBlock
	q := db.WithContext(ctx).
		Where("tenant_id = ? AND profile_id = ? AND blocked_date = ?", tenantID, profileID, date.Format("2006-01-02")).
		Where("(clinic_id IS NULL OR clinic_id = ?)", clinicID).
		Where("(start_time IS NULL OR (start_time <= ? AND end_time >= ?))", end, start)

	err := q.Find(&blocks).Error
	return blocks, err
}
