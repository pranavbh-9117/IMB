package domain

import (
	"time"

	"github.com/google/uuid"
)

// Quiz is a set of questions created by a faculty member within an
// institution. It remains hidden from students until IsPublished is true.
// DurationMinutes of 0 means no time limit is enforced.
type Quiz struct {
	Base

	InstitutionID    uuid.UUID  `gorm:"type:uuid;not null"`
	Institution      Institution `gorm:"foreignKey:InstitutionID"`
	CreatedBy        uuid.UUID  `gorm:"type:uuid;not null"`
	Creator          User       `gorm:"foreignKey:CreatedBy"`
	Title            string     `gorm:"type:varchar(255);not null"`
	Description      string     `gorm:"type:text"`
	DurationMinutes  int        `gorm:"default:0"`
	PassMarkPercent  float64
	IsPublished      bool       `gorm:"default:false"`
	Questions        []Question `gorm:"foreignKey:QuizID"`
}

// Question is a single item within a Quiz. OrderIndex controls the display
// sequence presented to the student.
type Question struct {
	Base

	QuizID     uuid.UUID `gorm:"type:uuid;not null"`
	Quiz       Quiz      `gorm:"foreignKey:QuizID"`
	Text       string    `gorm:"type:text;not null"`
	OrderIndex int       `gorm:"default:0"`
	Options    []Option  `gorm:"foreignKey:QuestionID"`
}

// Option is one of the possible answers for a Question. Exactly one Option
// per Question should have IsCorrect set to true for a standard
// single-answer multiple-choice quiz.
type Option struct {
	Base

	QuestionID uuid.UUID `gorm:"type:uuid;not null"`
	Question   Question  `gorm:"foreignKey:QuestionID"`
	Text       string    `gorm:"type:text;not null"`
	IsCorrect  bool      `gorm:"default:false"`
	OrderIndex int       `gorm:"default:0"`
}

// QuizAttempt records a single sitting of a student taking a Quiz.
// SubmittedAt is nil while the attempt is still in progress. Score holds
// the raw numeric result; IsPassed is derived from PassMarkPercent at
// submission time by the service layer.
type QuizAttempt struct {
	Base

	QuizID      uuid.UUID    `gorm:"type:uuid;not null"`
	Quiz        Quiz         `gorm:"foreignKey:QuizID"`
	UserID      uuid.UUID    `gorm:"type:uuid;not null"`
	User        User         `gorm:"foreignKey:UserID"`
	StartedAt   time.Time    `gorm:"not null"`
	SubmittedAt *time.Time
	Score       float64      `gorm:"default:0"`
	IsPassed    bool         `gorm:"default:false"`
	Answers     []QuizAnswer `gorm:"foreignKey:AttemptID"`
}

// QuizAnswer records the option a student selected for a specific question
// within a QuizAttempt. SelectedOptionID is nil if the student skipped the
// question.
type QuizAnswer struct {
	Base

	AttemptID        uuid.UUID  `gorm:"type:uuid;not null"`
	Attempt          QuizAttempt `gorm:"foreignKey:AttemptID"`
	QuestionID       uuid.UUID  `gorm:"type:uuid;not null"`
	Question         Question   `gorm:"foreignKey:QuestionID"`
	SelectedOptionID *uuid.UUID `gorm:"type:uuid"`
	SelectedOption   *Option    `gorm:"foreignKey:SelectedOptionID"`
}
