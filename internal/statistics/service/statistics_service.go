// Package service provides business logic for daily statistics aggregation.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	attemptRepo "github.com/pranavbh-9117/IMB/internal/attempt/repository"
	"github.com/pranavbh-9117/IMB/internal/domain"
	leaveRepo "github.com/pranavbh-9117/IMB/internal/leave/repository"
	statsRepo "github.com/pranavbh-9117/IMB/internal/statistics/repository"
	"github.com/pranavbh-9117/IMB/pkg/database"
)

// StatisticsService defines operations for generating daily institution statistics.
type StatisticsService interface {
	GenerateInstitutionStatistics(ctx context.Context, institutionID uuid.UUID, reportDate time.Time) error
}

type statisticsService struct {
	uow         database.UnitOfWork
	attemptRepo attemptRepo.AttemptRepository
	leaveRepo   leaveRepo.LeaveRepository
	statsRepo   statsRepo.StatsRepository
}

// NewStatisticsService creates a new StatisticsService instance.
func NewStatisticsService(
	uow database.UnitOfWork,
	attemptRepo attemptRepo.AttemptRepository,
	leaveRepo leaveRepo.LeaveRepository,
	statsRepo statsRepo.StatsRepository,
) StatisticsService {
	return &statisticsService{
		uow:         uow,
		attemptRepo: attemptRepo,
		leaveRepo:   leaveRepo,
		statsRepo:   statsRepo,
	}
}

// GenerateInstitutionStatistics aggregates quiz and leave metrics for the specified reporting date and persists them.
func (s *statisticsService) GenerateInstitutionStatistics(ctx context.Context, institutionID uuid.UUID, reportDate time.Time) error {
	startTime := time.Date(reportDate.Year(), reportDate.Month(), reportDate.Day(), 0, 0, 0, 0, time.UTC)
	endTime := startTime.Add(24 * time.Hour)

	var totalAttempts int
	var uniqueStudents int
	var topStudents []domain.TopStudentEntry
	var approvedLeaves, rejectedLeaves, pendingLeaves int
	var facultyLeaveStats []domain.FacultyLeaveEntry

	err := s.uow.WithinReadOnlyTransaction(ctx, func(txCtx context.Context) error {
		var err error
		totalAttempts, uniqueStudents, err = s.attemptRepo.GetInstitutionQuizStatsByWindow(txCtx, institutionID, startTime, endTime)
		if err != nil {
			return fmt.Errorf("statistics service: get quiz stats: %w", err)
		}

		topStudents, err = s.attemptRepo.GetTopStudentsByWindow(txCtx, institutionID, startTime, endTime, 5)
		if err != nil {
			return fmt.Errorf("statistics service: get top students: %w", err)
		}

		approvedLeaves, rejectedLeaves, pendingLeaves, err = s.leaveRepo.GetInstitutionLeaveStatsByWindow(txCtx, institutionID, startTime, endTime)
		if err != nil {
			return fmt.Errorf("statistics service: get leave stats: %w", err)
		}

		facultyLeaveStats, err = s.leaveRepo.GetFacultyLeaveStatsByWindow(txCtx, institutionID, startTime, endTime)
		if err != nil {
			return fmt.Errorf("statistics service: get faculty leave stats: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	topStudentsJSON, err := json.Marshal(topStudents)
	if err != nil {
		return fmt.Errorf("statistics service: marshal top students: %w", err)
	}

	facultyStatsJSON, err := json.Marshal(facultyLeaveStats)
	if err != nil {
		return fmt.Errorf("statistics service: marshal faculty stats: %w", err)
	}

	statRecord := &domain.DailyInstituteStatistic{
		InstitutionID:        institutionID,
		ReportDate:           startTime,
		TotalQuizAttempts:    totalAttempts,
		UniqueStudentsTested: uniqueStudents,
		LeavesApproved:       approvedLeaves,
		LeavesRejected:       rejectedLeaves,
		LeavesPending:        pendingLeaves,
		TopStudents:          datatypes.JSON(topStudentsJSON),
		FacultyLeaveStats:    datatypes.JSON(facultyStatsJSON),
	}

	if err := s.statsRepo.UpsertStatistic(ctx, statRecord); err != nil {
		return fmt.Errorf("statistics service: upsert statistic: %w", err)
	}

	return nil
}
