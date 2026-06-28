// Package service provides reminder logic for leave requests.
package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/pranavbh-9117/IMB/internal/domain"
	leaveRepo "github.com/pranavbh-9117/IMB/internal/leave/repository"
	userRepo "github.com/pranavbh-9117/IMB/internal/user/repository"
	"github.com/pranavbh-9117/IMB/pkg/email"
)

// ReminderService defines operations for dispatching grouped leave reminders.
type ReminderService interface {
	DispatchLeaveReminders(ctx context.Context) error
}

type reminderService struct {
	leaveRepo leaveRepo.LeaveRepository
	userRepo  userRepo.UserRepository
	emailSvc  email.EmailService
}

// NewReminderService creates a new ReminderService instance.
func NewReminderService(leaveRepo leaveRepo.LeaveRepository, userRepo userRepo.UserRepository, emailSvc email.EmailService) ReminderService {
	return &reminderService{
		leaveRepo: leaveRepo,
		userRepo:  userRepo,
		emailSvc:  emailSvc,
	}
}

// DispatchLeaveReminders fetches pending leave requests, consolidates them by institution and applicant role, and dispatches reminder emails.
func (s *reminderService) DispatchLeaveReminders(ctx context.Context) error {
	pendingLeaves, err := s.leaveRepo.GetPendingLeavesWithUser(ctx)
	if err != nil {
		return fmt.Errorf("reminder service: fetch pending leaves: %w", err)
	}

	if len(pendingLeaves) == 0 {
		return nil
	}

	studentLeavesByInst := make(map[uuid.UUID][]domain.LeaveRequest)
	facultyLeavesByInst := make(map[uuid.UUID][]domain.LeaveRequest)

	for _, req := range pendingLeaves {
		if req.User.Role == domain.RoleStudent {
			studentLeavesByInst[req.InstitutionID] = append(studentLeavesByInst[req.InstitutionID], req)
		} else if req.User.Role == domain.RoleFaculty {
			facultyLeavesByInst[req.InstitutionID] = append(facultyLeavesByInst[req.InstitutionID], req)
		}
	}

	// Remind faculty about student leaves
	for instID, leaves := range studentLeavesByInst {
		faculties, err := s.userRepo.GetByRoleAndInstitution(ctx, domain.RoleFaculty, instID)
		if err != nil {
			return fmt.Errorf("reminder service: get faculties for inst %s: %w", instID, err)
		}
		if len(faculties) == 0 {
			continue
		}

		subject := fmt.Sprintf("[IMB Reminder] %d Pending Student Leave Request(s)", len(leaves))
		body := buildReminderEmailBody("Faculty", leaves)

		for _, faculty := range faculties {
			if s.emailSvc != nil {
				s.emailSvc.SendAsync(ctx, email.Message{
					To:      faculty.Email,
					Subject: subject,
					Body:    body,
				})
			}
		}
	}

	// Remind institute admins about faculty leaves
	for instID, leaves := range facultyLeavesByInst {
		admins, err := s.userRepo.GetByRoleAndInstitution(ctx, domain.RoleInstituteAdmin, instID)
		if err != nil {
			return fmt.Errorf("reminder service: get admins for inst %s: %w", instID, err)
		}
		if len(admins) == 0 {
			continue
		}

		subject := fmt.Sprintf("[IMB Reminder] %d Pending Faculty Leave Request(s)", len(leaves))
		body := buildReminderEmailBody("Institute Administrator", leaves)

		for _, admin := range admins {
			if s.emailSvc != nil {
				s.emailSvc.SendAsync(ctx, email.Message{
					To:      admin.Email,
					Subject: subject,
					Body:    body,
				})
			}
		}
	}

	return nil
}

func buildReminderEmailBody(recipientRole string, leaves []domain.LeaveRequest) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Hello %s,\n\n", recipientRole))
	sb.WriteString("You have the following pending leave requests waiting for your review:\n\n")

	for i, req := range leaves {
		sb.WriteString(fmt.Sprintf("%d. Applicant: %s (%s)\n", i+1, req.User.Name, req.User.Email))
		sb.WriteString(fmt.Sprintf("   Dates: %s to %s\n", req.StartDate.Format("2006-01-02"), req.EndDate.Format("2006-01-02")))
		if req.Reason != "" {
			sb.WriteString(fmt.Sprintf("   Reason: %s\n", req.Reason))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("Please log in to the IMB dashboard to review these requests.\n\nBest regards,\nIMB System")
	return sb.String()
}
