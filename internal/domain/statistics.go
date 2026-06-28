// Package domain provides domain functionality for the IMB platform.
package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// TopStudentEntry represents a top performing student in JSONB format.
type TopStudentEntry struct {
	StudentID string  `json:"student_id"`
	Name      string  `json:"name"`
	AvgScore  float64 `json:"avg_score"`
}

// FacultyLeaveEntry represents faculty leave request statistics in JSONB format.
type FacultyLeaveEntry struct {
	FacultyID string `json:"faculty_id"`
	Name      string `json:"name"`
	Pending   int    `json:"pending"`
	Approved  int    `json:"approved"`
	Rejected  int    `json:"rejected"`
}

// DailyInstituteStatistic represents the aggregated daily statistics for an institution.
type DailyInstituteStatistic struct {
	Base
	InstitutionID        uuid.UUID      `gorm:"type:uuid;not null;index"`
	ReportDate           time.Time      `gorm:"type:date;not null"`
	TotalQuizAttempts    int            `gorm:"not null;default:0"`
	UniqueStudentsTested int            `gorm:"not null;default:0"`
	LeavesApproved       int            `gorm:"not null;default:0"`
	LeavesRejected       int            `gorm:"not null;default:0"`
	LeavesPending        int            `gorm:"not null;default:0"`
	TopStudents          datatypes.JSON `gorm:"type:jsonb"`
	FacultyLeaveStats    datatypes.JSON `gorm:"type:jsonb"`
}
