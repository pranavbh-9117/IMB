# Quiz Module Requirements & Business Rules

Version: 2.0

Purpose:

This document defines the complete Quiz Module requirements, workflows, RBAC rules, multi-tenant constraints, scoring rules, ownership rules, and implementation constraints.

All Quiz Module implementation must follow this document.

If implementation details conflict with this document, this document takes precedence.

---

# 1. Module Overview

The Quiz Module allows:

* Faculty to create quizzes.
* Faculty to manage quizzes they created.
* Faculty to publish quizzes.
* Students to view published quizzes belonging to their institution.
* Students to attempt quizzes.
* Faculty to view performance of students who attempted their quizzes.
* Students to view their own results.

The module follows:

* Clean Architecture
* Repository Pattern
* Dependency Injection
* RBAC
* Multi-Tenant Architecture

---

# 2. RBAC Rules

## SUPER_ADMIN

No quiz permissions.

Cannot:

* Create Quiz
* Update Quiz
* Delete Quiz
* Attempt Quiz
* View Results

---

## INSTITUTE_ADMIN

No quiz permissions.

Cannot:

* Create Quiz
* Update Quiz
* Delete Quiz
* Attempt Quiz
* View Results

---

## FACULTY

Can:

* Create Quiz
* Update Own Quiz
* Delete Own Quiz
* Publish Own Quiz
* View Own Quizzes
* View Results For Own Quizzes

Cannot:

* Manage another faculty's quizzes
* View results of another faculty's quizzes
* Attempt quizzes

---

## STUDENT

Can:

* View Published Quizzes
* Attempt Quiz
* View Own Results

Cannot:

* Create Quiz
* Update Quiz
* Delete Quiz
* View Other Students' Results

---

# 3. Multi-Tenant Rules

Every quiz belongs to exactly one institution.

Rules:

Faculty:

* Can create quizzes only within their institution.
* Can manage quizzes only within their institution.

Students:

* Can view quizzes only from their institution.
* Can attempt quizzes only from their institution.

Quiz attempts:

* Must belong to the same institution as the quiz.

Cross-institution access is prohibited.

---

# 4. Quiz Ownership Rules

Each Quiz contains:

```go
InstitutionID uuid.UUID

CreatedBy uuid.UUID
```

CreatedBy refers to the faculty user who created the quiz.

Rules:

Faculty can:

* View own quizzes
* Update own quizzes
* Delete own quizzes
* Publish own quizzes
* View results of own quizzes

Faculty cannot:

* Manage another faculty's quizzes
* View another faculty's quiz results

---

# 5. Quiz Lifecycle

Draft Quiz

```text
Faculty
    ↓
Create Quiz
    ↓
Draft
```

Published Quiz

```text
Faculty
    ↓
Publish Quiz
    ↓
Visible To Students
```

---

# 6. Quiz Publication Rules

Field:

```go
IsPublished bool
```

Rules:

IsPublished = false

* Visible only to creator faculty.

IsPublished = true

* Visible to students of same institution.

Published quizzes become read-only.

After publication:

Faculty cannot:

* Modify quiz metadata
* Modify questions
* Modify options

Reason:

Quiz integrity must be preserved.

---

# 7. Quiz Deletion Rules

Faculty may delete a quiz only when:

```text
No attempts exist.
```

If attempts exist:

Return:

```text
409 Conflict
```

Reason:

Attempt history must remain valid.

---

# 8. Quiz Structure

Quiz

Contains:

* Quiz metadata
* Questions

Question

Contains:

* Question text
* Marks
* Options

Option

Contains:

* Option text
* Correct answer flag

Only MCQ questions are supported.

---

# 9. Quiz Entity

Quiz contains:

```go
InstitutionID
CreatedBy

Title
Description

DurationMinutes

TotalMarks

IsPublished
```

Rules:

* TotalMarks is automatically calculated.
* Faculty cannot manually set TotalMarks.
* Quiz belongs to exactly one institution.
* Quiz belongs to exactly one faculty creator.

---

# 10. Question Entity

Question contains:

```go
QuizID

Text

Marks

OrderIndex
```

Rules:

* Marks must be greater than zero.
* Questions belong to one quiz.
* OrderIndex controls display order.

---

# 11. Option Entity

Option contains:

```go
QuestionID

Text

IsCorrect

OrderIndex
```

Rules:

* Each question must have exactly one correct option.
* Multiple correct answers are not supported.
* MCQ only.

Validation must reject:

```text
No correct option

More than one correct option
```

---

# 12. Total Marks Calculation

Quiz contains:

```go
TotalMarks
```

Formula:

```text
Sum(All Question Marks)
```

Example:

```text
Question 1 = 2

Question 2 = 3

Question 3 = 5

TotalMarks = 10
```

The service layer is responsible for maintaining TotalMarks.

Faculty must not manually provide TotalMarks.

---

# 13. Quiz Attempt Rules

Students may attempt:

Published quizzes only.

Rules:

Student must belong to quiz institution.

Student may attempt a quiz only once.

Validation:

```text
Already Attempted
    ↓
Reject
```

Return:

```text
409 Conflict
```

---

# 14. Quiz Attempt Entity

QuizAttempt contains:

```go
InstitutionID

QuizID

StudentID

StartedAt

SubmittedAt

Score

TotalMarks
```

Rules:

* One attempt per student per quiz.
* Score stores earned marks.
* TotalMarks stores quiz total marks at submission time.

---

# 15. Quiz Answer Entity

QuizAnswer contains:

```go
AttemptID

QuestionID

SelectedOptionID
```

Rules:

SelectedOptionID may be null if skipped.

---

# 16. Quiz Attempt Workflow

```text
Student
    ↓
View Quiz
    ↓
Submit Answers
    ↓
Automatic Evaluation
    ↓
Store Result
    ↓
Result Available
```

---

# 17. Scoring Rules

Evaluation Type:

MCQ Only

Rules:

Correct Answer

```text
Award Question.Marks
```

Incorrect Answer

```text
Award 0 Marks
```

Skipped Question

```text
Award 0 Marks
```

Negative Marking

```text
Not Supported
```

---

# 18. Score Calculation

Formula:

```text
Score =
Sum(Marks For Correctly Answered Questions)
```

Example:

```text
Question 1 = 2 Marks

Question 2 = 3 Marks

Question 3 = 5 Marks

Student Correct:
Q1
Q3

Score = 7
```

---

# 19. Student Result Rules

Students can view:

Only their own results.

Endpoint:

```text
GET /results
```

Response:

```text
Quiz Title

Score

TotalMarks

SubmittedAt
```

Students cannot view:

* Other students' results
* Other students' attempts

---

# 20. Faculty Performance Rules

Faculty can view:

Only results of quizzes they created.

Endpoint:

```text
GET /quizzes/{id}/results
```

Faculty cannot view:

Results of another faculty's quiz.

---

# 21. Faculty Performance Response

Response contains:

```text
Student Name

Student Email

Score

TotalMarks

SubmittedAt
```

No percentage calculation.

No grading system.

No pass/fail.

---

# 22. API Authorization Matrix

## Quiz Management

POST /quizzes

```text
FACULTY
```

GET /quizzes

```text
FACULTY
STUDENT
```

GET /quizzes/{id}

```text
FACULTY
STUDENT
```

PUT /quizzes/{id}

```text
FACULTY
```

DELETE /quizzes/{id}

```text
FACULTY
```

---

## Quiz Attempts

POST /quizzes/{id}/attempt

```text
STUDENT
```

GET /results

```text
STUDENT
```

GET /quizzes/{id}/results

```text
FACULTY
```

---

# 23. Expected Endpoint Behaviour

## GET /quizzes

FACULTY

```text
Return quizzes created by faculty.
```

STUDENT

```text
Return published quizzes belonging to student's institution.
```

---

## GET /quizzes/{id}

FACULTY

```text
Can view own quiz.
```

STUDENT

```text
Can view published quiz from same institution.
```

---

## GET /quizzes/{id}/results

FACULTY

```text
Can view results only if quiz.CreatedBy == facultyID
```

Otherwise:

```text
404 Not Found
```

to prevent information leakage.

---

# 24. Architectural Constraints

Repository Layer

Allowed:

* Database access

Not Allowed:

* Business logic
* Authorization
* Scoring

---

Service Layer

Allowed:

* Validation
* Ownership checks
* Institution checks
* Scoring
* Business rules

Not Allowed:

* HTTP concerns
* Gin usage

---

Handler Layer

Allowed:

* Request binding
* Response formatting
* Service calls

Not Allowed:

* Business logic
* Database access

---

# 25. Implementation Order

Step Q0

* Quiz Domain Models

Step 6.1

* Quiz Repository

Step 6.2

* Quiz Service

Step 6.3

* Quiz Handlers

Step 6.4

* Quiz Wiring & Testing

Step 7.1

* Attempt Repository

Step 7.2

* Scoring Logic

Step 7.3

* Attempt Service

Step 7.4

* Attempt Handlers

Step 7.5

* Attempt Wiring & Testing

Implementation must follow this sequence.
