// Package templates provides email content formatting for leave notifications.
package templates

import (
	"fmt"
	"strings"

	"github.com/pranavbh-9117/IMB/internal/domain"
)

// BuildLeaveNotification formats the subject and body for a leave status decision email.
func BuildLeaveNotification(req *domain.LeaveRequest, status domain.LeaveStatus, note string) (subject, body string) {
	statusUpper := strings.ToUpper(string(status))
	statusTitle := strings.ToUpper(string(status[:1])) + string(status[1:])

	subject = fmt.Sprintf("Your leave request has been %s", statusTitle)

	instName := "Institution"
	if req.Institution.Name != "" {
		instName = req.Institution.Name
	}

	name := "Applicant"
	if req.User.Name != "" {
		name = req.User.Name
	}

	startStr := req.StartDate.Format("2006-01-02")
	endStr := req.EndDate.Format("2006-01-02")

	body = fmt.Sprintf(`Dear %s,

Your leave request from %s to %s has been %s.

Reviewer Note: %s

Please log in to the portal for more details.

Regards,
%s Administration`,
		name,
		startStr,
		endStr,
		statusUpper,
		note,
		instName,
	)

	return subject, body
}
