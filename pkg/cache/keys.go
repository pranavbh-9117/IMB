package cache

import "github.com/google/uuid"

// AdminDashboardKey formats the Redis cache key for an institution's admin dashboard response.
func AdminDashboardKey(institutionID uuid.UUID) string {
	return "admin_dashboard:" + institutionID.String()
}
