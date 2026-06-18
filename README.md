# Institute Management Backend (IMB)

The Institute Management Backend (IMB) is a robust, production-ready REST API designed to handle multi-tenant institutional operations. Built with Go, it strictly adheres to **Clean Architecture** principles, ensuring that core business logic is entirely decoupled from external frameworks and database drivers.

## Core Features

- **Multi-Tenant Isolation**: Rigid boundaries isolate data by `InstitutionID`. Users and their operations are cryptographically bound to their institution via JWT payload injection, preventing horizontal privilege escalation and cross-tenant data leakage.
- **Role-Based Access Control (RBAC)**: Strict hierarchical authorization supporting `SUPER_ADMIN`, `INSTITUTE_ADMIN`, `FACULTY`, and `STUDENT` roles.
- **Top-Down Provisioning**: Prevents unauthorized self-signups. Super Admins provision Institutes and Institute Admins. Institute Admins provision Faculty and Students. Temporary cryptographic passwords are automatically generated and enforced.
- **Hierarchical Leave Management**: 
  - Students apply for leaves → Approved by Faculty.
  - Faculty apply for leaves → Approved by Institute Admins.
  - Robust leave balance tracking mapped atomically to the approval lifecycle.
- **Structured Observability**: Built-in HTTP request tracing, latency tracking, and context-aware structured JSON logging utilizing `log/slog`.
- **Automated API Documentation**: Self-hosting, interactive OpenAPI 2.0 (Swagger) documentation generated directly from handler annotations.

## Technology Stack

- **Language**: Go 1.22+
- **HTTP Framework**: [Gin](https://gin-gonic.com/)
- **ORM**: [GORM](https://gorm.io/)
- **Database**: PostgreSQL
- **Authentication**: JWT (JSON Web Tokens) with HttpOnly Refresh Cookies
- **Documentation**: [swaggo/gin-swagger](https://github.com/swaggo/gin-swagger)

## Project Structure

The project directory follows a strict layered architectural pattern:

```text
IMB/
├── cmd/
│   └── server/             # Application entrypoint & dependency injection wiring
├── configs/                # Environment variables and configuration files
├── docs/                   # Swagger artifacts, Postman collections, and ADRs
├── internal/               # Core application modules
│   ├── auth/               # Authentication, login, and refresh token flows
│   ├── domain/             # Pure Go enterprise entities (User, Institution, Leave)
│   ├── institution/        # Tenant management logic
│   ├── leave/              # Leave requests, hierarchical approvals, and balances
│   ├── middleware/         # Security, Context, and Logging middlewares
│   ├── migration/          # GORM Auto-Migration configurations
│   ├── seed/               # Bootstrapping logic for the initial Super Admin
│   └── user/               # User provisioning and identity management
└── pkg/                    # Shared utility libraries (logger, passwords, HTTP responses)
```

## Getting Started

### Prerequisites
- Go 1.22 or higher
- PostgreSQL 15+

### Installation & Setup

1. **Clone the repository:**
   ```bash
   git clone <repository_url>
   cd IMB
   ```

2. **Configure Environment:**
   Copy the example configuration and update the PostgreSQL credentials.
   ```bash
   cp configs/.env.example .env
   ```

3. **Install Dependencies:**
   ```bash
   go mod tidy
   ```

4. **Run the Server:**
   The application will automatically connect to PostgreSQL, run any pending auto-migrations, seed the default Super Admin, and start the Gin server.
   ```bash
   go run cmd/server/main.go
   ```

## API Documentation & Testing

### Interactive Swagger UI
Once the server is running, the interactive Swagger documentation is accessible via the browser:
```text
http://localhost:8080/swagger/index.html
```

