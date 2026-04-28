package repository

import (
	"context"
	"time"

	dentistblock "github.com/Mirnda/mirandaclin/internal/domain/dentist_block"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type dentistBlockRepository struct{}

func NewDentistBlockRepository() dentistblock.Repository {
	return &dentistBlockRepository{}
}

func (r *dentistBlockRepository) Create(ctx context.Context, db *gorm.DB, b *dentistblock.DentistBlock) error {
	return db.WithContext(ctx).Create(b).Error
}

func (r *dentistBlockRepository) Delete(ctx context.Context, db *gorm.DB, tenantID, id uuid.UUID) error {
	return db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Delete(&dentistblock.DentistBlock{}).Error
}

// FindBlocksForSlot retorna bloqueios que cobrem a data e intervalo solicitados.
// Considera bloqueios sem clinic_id (valem para todas as clínicas) e com clinic_id específico.
func (r *dentistBlockRepository) FindBlocksForSlot(
	ctx context.Context, db *gorm.DB,
	tenantID, dentistID uuid.UUID, clinicID *uuid.UUID,
	date time.Time, start, end string,
) ([]dentistblock.DentistBlock, error) {
	var blocks []dentistblock.DentistBlock
	q := db.WithContext(ctx).
		Where("tenant_id = ? AND dentist_id = ? AND blocked_date = ?", tenantID, dentistID, date.Format("2006-01-02")).
		Where("(clinic_id IS NULL OR clinic_id = ?)", clinicID).
		Where("(start_time IS NULL OR (start_time <= ? AND end_time >= ?))", end, start)

	err := q.Find(&blocks).Error
	return blocks, err
}
