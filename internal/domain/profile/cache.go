package profile

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const defaultTTL = 6 * time.Hour

func profileIDCacheKey(tenantID, profileID uuid.UUID) string {
	return fmt.Sprintf("%s:profile:id:%s", tenantID, profileID)
}

func (s *Service) getProfileIDCache(ctx context.Context, tenantID, profileID uuid.UUID) (*Profile, error) {
	cacheKey := profileIDCacheKey(tenantID, profileID)

	var p *Profile
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != "" {
		if err := json.Unmarshal([]byte(cached), p); err == nil {
			return p, nil
		}
	}
	return nil, ErrProfileNotFound
}

func (s *Service) setProfileIDCache(ctx context.Context, tenantID uuid.UUID, p *Profile) {
	cacheKey := profileIDCacheKey(tenantID, p.ID)

	if data, err := json.Marshal(p); err == nil {
		_ = s.cache.Set(ctx, cacheKey, string(data), defaultTTL)
	}
}

func profileListByRoleCacheKey(tenantID uuid.UUID, role string) string {
	return fmt.Sprintf("%s:profile:role:%s", tenantID, role)
}

func (s *Service) getProfileListByRoleCache(ctx context.Context, tenantID uuid.UUID, role string) ([]Profile, error) {
	cacheKey := profileListByRoleCacheKey(tenantID, role)

	var profiles []Profile
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != "" {
		if err := json.Unmarshal([]byte(cached), &profiles); err == nil {
			return profiles, nil
		}
	}
	return nil, ErrProfileNotFound
}

func (s *Service) setProfileListByRoleCache(ctx context.Context, tenantID uuid.UUID, role string, profiles []Profile) {
	cacheKey := profileListByRoleCacheKey(tenantID, role)

	if data, err := json.Marshal(profiles); err == nil {
		_ = s.cache.Set(ctx, cacheKey, string(data), defaultTTL)
	}
}

func (s *Service) invalidateProfileCache(ctx context.Context, tenantID uuid.UUID) {
	_, _ = s.cache.DelWithIndex(ctx, fmt.Sprintf("%s:profile:*", tenantID))
	_, _ = s.cache.DelWithIndex(ctx, fmt.Sprintf("%s:*:profile:*", tenantID))
}
