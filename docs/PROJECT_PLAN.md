# Institute Management & Quiz Platform

## AI-Assisted Development Plan

Version: 1.0

---

# Stage 1: Architecture Approval

No code should be generated before the following are approved:

## System Scope

Modules:

1. Authentication
2. Institution Management
3. User Management
4. Leave Management

Excluded:

* Frontend
* Dashboard
* Quiz Management
* Quiz Attempt Management
* Quiz File Upload
* Notifications
* Caching
* Microservices

---

## Architecture Style

Clean Architecture

Flow:

Handler
↓
Service
↓
Repository Interface
↓
Repository Implementation
↓
Database

---

## Database

PostgreSQL

ORM:

GORM

---

## Authentication

JWT Access Token

Refresh Token

RBAC

---

## User Roles

SUPER_ADMIN

INSTITUTE_ADMIN

FACULTY

STUDENT

---

## Multi-Tenant Rule

All business data must be isolated by InstitutionID.

---

# Stage 2: Phase Breakdown

Phase 1
Project Foundation

Phase 2
Authentication & Authorization

Phase 3
Institution Management

Phase 4
User Management

Phase 5
Leave Management

Phase 6
Testing & Documentation

---

# Phase 1

## Project Foundation

Goal:

Create the project skeleton and establish application infrastructure.

No business logic.

No authentication.

No modules.

---

### Step 1.1

Project Initialization

Objective:

Create Go project structure.

Deliverables:

* go.mod
* folder structure

Review Required:

Folder naming
Architecture consistency

---

### Step 1.2

Configuration Management

Objective:

Load environment variables.

Deliverables:

configs/

.env support

Packages:

godotenv

Review Required:

Configuration design

---

### Step 1.3

Database Connectivity

Objective:

Connect PostgreSQL.

Deliverables:

database connection package

Features:

* Connection
* Health Check
* Pool Configuration

Packages:

gorm
postgres driver

Review Required:

Database initialization

---

### Step 1.4

Domain Models

Objective:

Create entities.

Deliverables:

Institution
User
LeaveBalance
LeaveRequest
RefreshToken

No repository code.

No business logic.

Review Required:

Entity structure

---

### Step 1.5

Database Migration

Objective:

Auto-migrate domain models.

Deliverables:

Migration package

Review Required:

Schema generation

---

### Step 1.6

Super Admin Seed

Objective:

Create initial Super Admin.

Deliverables:

Seed package

Review Required:

Seed strategy

---

# Phase 2

## Authentication & Authorization

Goal:

Allow secure login and protected routes.

---

### Step 2.1

Password Package

Objective:

Password hashing utilities.

Functions:

HashPassword
CheckPassword

Package:

bcrypt

Review Required:

Password security

---

### Step 2.2

JWT Package

Objective:

Token generation.

Functions:

GenerateAccessToken
ValidateAccessToken

Review Required:

Claims design

---

### Step 2.3

Refresh Token Design

Objective:

Database-backed refresh tokens.

Deliverables:

RefreshToken model usage

Review Required:

Session strategy

---

### Step 2.4

Auth Repository

Objective:

Database access for auth module.

Review Required:

Repository interfaces

---

### Step 2.5

Auth Service

Objective:

Authentication business logic.

Features:

Login
Refresh
Logout
Change Password

Review Required:

Business rules

---

### Step 2.6

Auth Handlers

Objective:

HTTP endpoints.

Endpoints:

POST /auth/login

POST /auth/change-password

POST /auth/refresh

POST /auth/logout

Review Required:

Request/Response contracts

---

### Step 2.7

Authentication Middleware

Objective:

JWT validation middleware.

Review Required:

Security

---

### Step 2.8

RBAC Middleware

Objective:

Role validation.

Review Required:

Permission enforcement

---

# Phase 3

## Institution Management

Goal:

Allow Super Admin to manage institutions.

---

### Step 3.1

Institution Repository

---

### Step 3.2

Institution Service

---

### Step 3.3

Institution Handlers

Endpoints:

POST /institutions

GET /institutions

GET /institutions/{id}

---

# Phase 4

## User Management

Goal:

Allow Institute Admin to manage Faculty and Students.

---

### Step 4.1

User Repository

---

### Step 4.2

Temporary Password Strategy

Decision Required:

Password generation format

---

### Step 4.3

User Service

Features:

Create User

Update User

Delete User

List Users

---

### Step 4.4

User Handlers

Endpoints:

POST /users

GET /users

GET /users/{id}

PUT /users/{id}

DELETE /users/{id}

---

# Phase 5

## Leave Management

Goal:

Implement leave approval workflow.

---

### Step 5.1

Leave Repository

---

### Step 5.2

Leave Balance Logic

Rules:

Approval updates balance.

---

### Step 5.3

Leave Service

Student Leave Flow

Faculty Leave Flow

---

### Step 5.4

Leave Handlers

Endpoints:

POST /leaves

GET /leaves

PUT /leaves/{id}

---

# Phase 6

## Testing & Documentation

---

### Step 6.1

Swagger Documentation

---

### Step 6.2

API Testing

Postman Collection

---

### Step 6.3

Architecture Review

Verify:

* Layer Separation
* Dependency Direction
* Scope Compliance

---

# ADR Requirements

Create:

docs/ADR.md

Initial Decisions:

ADR-001
UUID Primary Keys

ADR-002
Faculty Approves Student Leave

ADR-003
Institute Admin Approves Faculty Leave

ADR-004
Institution-Level Data Isolation

ADR-005
Super Admin Hard Seeded
