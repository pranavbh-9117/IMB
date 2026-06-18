# Architecture Decision Records

---

## ADR-001 — UUID Primary Keys

**Status:** Accepted

All entities use UUID (`uuid.UUID`) as their primary key instead of auto-increment integers.

**Rationale:** Prevents enumeration attacks, supports distributed generation without coordination, and aligns with multi-tenant data isolation requirements.

---

## ADR-002 — One Quiz Attempt Per Student

**Status:** Accepted

A student may submit exactly one attempt per quiz.

**Rationale:** Prevents result manipulation and simplifies scoring logic. Re-attempt capability is out of scope for this version.

---

## ADR-003 — Faculty Approves Student Leave

**Status:** Accepted

Leave requests submitted by students are reviewed and approved or rejected by faculty.

**Rationale:** Faculty are the direct academic supervisors of students and are best positioned to assess leave impact.

---

## ADR-004 — Institute Admin Approves Faculty Leave

**Status:** Accepted

Leave requests submitted by faculty are reviewed and approved or rejected by the Institute Admin.

**Rationale:** Faculty leave affects institute-level scheduling and must be governed at the administrative level.

---

## ADR-005 — Institution-Level Data Isolation

**Status:** Accepted

All business entities (users, quizzes, leave records) carry an `InstitutionID` foreign key. All queries are scoped by `InstitutionID`.

**Rationale:** The platform is multi-tenant. Data leakage across institution boundaries is a hard security requirement.

---

## ADR-006 — Super Admin Hard Seeded

**Status:** Accepted

The Super Admin account is created once at application startup via a seed function. Credentials are sourced from environment variables. The seed is idempotent — it does nothing if the account already exists.

**Rationale:** Bootstrapping the system requires at least one privileged account before any UI or API flow can be used.