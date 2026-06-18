#!/bin/bash

API_URL="http://localhost:8080/api/v1"
OUTPUT_FILE="api_test_results.txt"

> "$OUTPUT_FILE"

TS=$(date +%s)

ALPHA_CODE="ALPHA_${TS}"
BETA_CODE="BETA_${TS}"

ALPHA_NAME="Alpha Institute ${TS}"
BETA_NAME="Beta Institute ${TS}"

ALPHA_ADMIN_EMAIL="admin_${TS}@alpha.com"
FACULTY_EMAIL="faculty_${TS}@alpha.com"
STUDENT_EMAIL="student_${TS}@alpha.com"
BETA_ADMIN_EMAIL="admin_${TS}@beta.com"

log() {
  echo "$1" | tee -a "$OUTPUT_FILE" >&2
}

log_request() {
  local method=$1
  local endpoint=$2
  local body=$3
  local token=$4
  local label=$5

  log "--------------------------------------------------------"
  log "TEST: $label"
  log "REQUEST: $method $API_URL$endpoint"

  local curl_cmd="curl -s -w '\nHTTP_STATUS:%{http_code}\n' -X $method $API_URL$endpoint -H 'Content-Type: application/json'"

  if [ -n "$token" ]; then
    log "AUTH: Bearer Token Provided"
    curl_cmd="$curl_cmd -H 'Authorization: Bearer $token'"
  fi

  if [ -n "$body" ]; then
    log "BODY: $body"
    curl_cmd="$curl_cmd -d '$body'"
  fi

  local response=$(eval "$curl_cmd")

  local http_status
  http_status=$(echo "$response" | grep "HTTP_STATUS:" | cut -d':' -f2)

  local resp_body
  resp_body=$(echo "$response" | sed '/HTTP_STATUS:/d')

  log "HTTP STATUS: $http_status"
  log "RESPONSE: $resp_body"
  log ""

  echo "$resp_body"
}

validate_value() {
  local value=$1
  local message=$2

  if [ "$value" = "null" ] || [ -z "$value" ]; then
    log "FAILED: $message"
    exit 1
  fi
}

echo ""
log "=================== LOGIN ==================="

SA_LOGIN_RES=$(log_request "POST" "/auth/login" '{"email":"aqumanj@gmail.com","password":"dboss000"}' "" "Super Admin Login")

SA_TOKEN=$(echo "$SA_LOGIN_RES" | jq -r '.data.access_token')

validate_value "$SA_TOKEN" "Unable to obtain Super Admin token"

echo ""
log "=================== UNAUTHENTICATED TESTS ==================="

log_request "GET" "/users" "" "" "Fail: Get Users without Token (401)"

echo ""
log "=================== INSTITUTION MANAGEMENT ==================="

INST_RES=$(log_request "POST" "/institutions" "{\"name\":\"$ALPHA_NAME\",\"code\":\"$ALPHA_CODE\",\"address\":\"123 Main St\"}" "$SA_TOKEN" "Create Alpha Institution")

INST_ID=$(echo "$INST_RES" | jq -r '.data.id')

validate_value "$INST_ID" "Institution creation failed"

echo ""
log "=================== USER MANAGEMENT (SUPER ADMIN) ==================="

IA_CREATE_RES=$(log_request "POST" "/users" "{\"name\":\"Alpha Admin\",\"email\":\"$ALPHA_ADMIN_EMAIL\",\"role\":\"institute_admin\",\"institution_id\":\"$INST_ID\"}" "$SA_TOKEN" "Create Institute Admin")

IA_ID=$(echo "$IA_CREATE_RES" | jq -r '.data.user.id')
IA_PASS=$(echo "$IA_CREATE_RES" | jq -r '.data.temporary_password')

validate_value "$IA_ID" "Institute Admin creation failed"
validate_value "$IA_PASS" "Institute Admin password missing"

log_request "POST" "/users" "{\"name\":\"Forbidden Faculty\",\"email\":\"forbidden_${TS}@alpha.com\",\"role\":\"faculty\",\"institution_id\":\"$INST_ID\"}" "$SA_TOKEN" "Fail: Super Admin creates Faculty"

echo ""
log "=================== INSTITUTE ADMIN LOGIN ==================="

IA_LOGIN_RES=$(log_request "POST" "/auth/login" "{\"email\":\"$ALPHA_ADMIN_EMAIL\",\"password\":\"$IA_PASS\"}" "" "Institute Admin Login")

IA_TOKEN=$(echo "$IA_LOGIN_RES" | jq -r '.data.access_token')

validate_value "$IA_TOKEN" "Institute Admin login failed"

echo ""
log "=================== USER MANAGEMENT (INSTITUTE ADMIN) ==================="

log_request "POST" "/users" "{\"name\":\"Rogue Admin\",\"email\":\"rogue_${TS}@alpha.com\",\"role\":\"institute_admin\"}" "$IA_TOKEN" "Fail: Institute Admin creates Institute Admin"

FAC_CREATE_RES=$(log_request "POST" "/users" "{\"name\":\"Alpha Faculty\",\"email\":\"$FACULTY_EMAIL\",\"role\":\"faculty\"}" "$IA_TOKEN" "Create Faculty")

FAC_ID=$(echo "$FAC_CREATE_RES" | jq -r '.data.user.id')
FAC_PASS=$(echo "$FAC_CREATE_RES" | jq -r '.data.temporary_password')

validate_value "$FAC_ID" "Faculty creation failed"
validate_value "$FAC_PASS" "Faculty password missing"

STUDENT_CREATE_RES=$(log_request "POST" "/users" "{\"name\":\"Alpha Student\",\"email\":\"$STUDENT_EMAIL\",\"role\":\"student\"}" "$IA_TOKEN" "Create Student")

STUDENT_ID=$(echo "$STUDENT_CREATE_RES" | jq -r '.data.user.id')
STUDENT_PASS=$(echo "$STUDENT_CREATE_RES" | jq -r '.data.temporary_password')

validate_value "$STUDENT_ID" "Student creation failed"
validate_value "$STUDENT_PASS" "Student password missing"

log_request "GET" "/institutions" "" "$IA_TOKEN" "Fail: Institute Admin lists Institutions"

log_request "GET" "/users" "" "$IA_TOKEN" "Institute Admin lists Users"

log_request "PUT" "/users/$FAC_ID" '{"role":"institute_admin"}' "$IA_TOKEN" "Fail: Update Role"

log_request "PUT" "/users/$FAC_ID" '{"name":"Alpha Faculty Updated"}' "$IA_TOKEN" "Update Faculty Name"

echo ""
log "=================== FACULTY FLOW ==================="

FAC_LOGIN_RES=$(log_request "POST" "/auth/login" "{\"email\":\"$FACULTY_EMAIL\",\"password\":\"$FAC_PASS\"}" "" "Faculty Login")

FAC_TOKEN=$(echo "$FAC_LOGIN_RES" | jq -r '.data.access_token')

validate_value "$FAC_TOKEN" "Faculty login failed"

log_request "POST" "/users" "{\"name\":\"Rogue Student\",\"email\":\"rogue_student_${TS}@alpha.com\",\"role\":\"student\"}" "$FAC_TOKEN" "Fail: Faculty creates User"

log_request "GET" "/users" "" "$FAC_TOKEN" "Fail: Faculty lists Users"

echo ""
log "=================== STUDENT FLOW ==================="

STUDENT_LOGIN_RES=$(log_request "POST" "/auth/login" "{\"email\":\"$STUDENT_EMAIL\",\"password\":\"$STUDENT_PASS\"}" "" "Student Login")

STUDENT_TOKEN=$(echo "$STUDENT_LOGIN_RES" | jq -r '.data.access_token')

validate_value "$STUDENT_TOKEN" "Student login failed"

log_request "GET" "/users" "" "$STUDENT_TOKEN" "Fail: Student lists Users"

log_request "POST" "/users" "{\"name\":\"Another Student\",\"email\":\"another_${TS}@alpha.com\",\"role\":\"student\"}" "$STUDENT_TOKEN" "Fail: Student creates User"

echo ""
log "=================== TENANT ISOLATION ==================="

INST2_RES=$(log_request "POST" "/institutions" "{\"name\":\"$BETA_NAME\",\"code\":\"$BETA_CODE\",\"address\":\"456 Second St\"}" "$SA_TOKEN" "Create Beta Institution")

INST2_ID=$(echo "$INST2_RES" | jq -r '.data.id')

validate_value "$INST2_ID" "Beta Institution creation failed"

IA2_RES=$(log_request "POST" "/users" "{\"name\":\"Beta Admin\",\"email\":\"$BETA_ADMIN_EMAIL\",\"role\":\"institute_admin\",\"institution_id\":\"$INST2_ID\"}" "$SA_TOKEN" "Create Beta Admin")

IA2_ID=$(echo "$IA2_RES" | jq -r '.data.user.id')

validate_value "$IA2_ID" "Beta Admin creation failed"

log_request "GET" "/users/$IA2_ID" "" "$IA_TOKEN" "Fail: Cross Tenant GET"

log_request "PUT" "/users/$IA2_ID" '{"name":"Hacked Name"}' "$IA_TOKEN" "Fail: Cross Tenant UPDATE"

log_request "DELETE" "/users/$IA2_ID" "" "$IA_TOKEN" "Fail: Cross Tenant DELETE"

echo ""
log "=================== SELF DELETE ==================="

log_request "DELETE" "/users/$IA_ID" "" "$IA_TOKEN" "Fail: Self Delete"

echo ""
log "=================== DELETE STUDENT ==================="

log_request "DELETE" "/users/$STUDENT_ID" "" "$IA_TOKEN" "Delete Student"

log_request "GET" "/users/$STUDENT_ID" "" "$IA_TOKEN" "Verify Student Deleted"

log_request "POST" "/auth/login" "{\"email\":\"$STUDENT_EMAIL\",\"password\":\"$STUDENT_PASS\"}" "" "Verify Deleted Student Login"

echo ""
log "=================== EMAIL REUSE ==================="

log_request "POST" "/users" "{\"name\":\"Student Reused\",\"email\":\"$STUDENT_EMAIL\",\"role\":\"student\"}" "$IA_TOKEN" "Fail: Reuse Deleted Email"

echo ""
log "=================== LEAVE MANAGEMENT END-TO-END ==================="

# 1. Create a fresh Student
NEW_STUDENT_EMAIL="new_student_${TS}@alpha.com"
NEW_STUDENT_RES=$(log_request "POST" "/users" "{\"name\":\"Leave Student\",\"email\":\"$NEW_STUDENT_EMAIL\",\"role\":\"student\"}" "$IA_TOKEN" "Create Student for Leaves")
NEW_STUDENT_ID=$(echo "$NEW_STUDENT_RES" | jq -r '.data.user.id')
NEW_STUDENT_PASS=$(echo "$NEW_STUDENT_RES" | jq -r '.data.temporary_password')
validate_value "$NEW_STUDENT_ID" "New Student creation failed"

# Login as Student
STUDENT_LOGIN_RES=$(log_request "POST" "/auth/login" "{\"email\":\"$NEW_STUDENT_EMAIL\",\"password\":\"$NEW_STUDENT_PASS\"}" "" "New Student Login")
STUDENT_TOKEN=$(echo "$STUDENT_LOGIN_RES" | jq -r '.data.access_token')
validate_value "$STUDENT_TOKEN" "New Student login failed"

# 2. Student Applies for Leave
TOMORROW=$(date -d "+1 day" -u +"%Y-%m-%dT00:00:00Z" 2>/dev/null || date -v+1d -u +"%Y-%m-%dT00:00:00Z") # GNU or BSD date
DAY_AFTER=$(date -d "+3 days" -u +"%Y-%m-%dT00:00:00Z" 2>/dev/null || date -v+3d -u +"%Y-%m-%dT00:00:00Z")
LEAVE_APPLY_RES=$(log_request "POST" "/leaves" "{\"start_date\":\"$TOMORROW\",\"end_date\":\"$DAY_AFTER\",\"reason\":\"Sick leave\"}" "$STUDENT_TOKEN" "Student Applies for Leave")
LEAVE_ID=$(echo "$LEAVE_APPLY_RES" | jq -r '.data.id')
validate_value "$LEAVE_ID" "Leave application failed"

# 3. RBAC Checks
log_request "PUT" "/leaves/$LEAVE_ID" '{"status":"approved","note":"self approve"}' "$STUDENT_TOKEN" "Fail: Student Self Approve"
log_request "PUT" "/leaves/$LEAVE_ID" '{"status":"approved","note":"admin approve"}' "$IA_TOKEN" "Fail: Inst Admin Approves Student Leave"

# 4. Faculty Approves Student Leave
log_request "PUT" "/leaves/$LEAVE_ID" '{"status":"approved","note":"approved by faculty"}' "$FAC_TOKEN" "Faculty Approves Student Leave"

# 5. Overlap Check
log_request "POST" "/leaves" "{\"start_date\":\"$TOMORROW\",\"end_date\":\"$DAY_AFTER\",\"reason\":\"Double booking\"}" "$STUDENT_TOKEN" "Fail: Overlapping Leave"

# 6. Faculty Leave Flow
FAC_LEAVE_RES=$(log_request "POST" "/leaves" "{\"start_date\":\"$TOMORROW\",\"end_date\":\"$DAY_AFTER\",\"reason\":\"Conference\"}" "$FAC_TOKEN" "Faculty Applies for Leave")
FAC_LEAVE_ID=$(echo "$FAC_LEAVE_RES" | jq -r '.data.id')
log_request "PUT" "/leaves/$FAC_LEAVE_ID" '{"status":"approved","note":"admin approves faculty"}' "$IA_TOKEN" "Inst Admin Approves Faculty Leave"

# 7. Cancellations
CANCEL_LEAVE_RES=$(log_request "POST" "/leaves" "{\"start_date\":\"2026-12-01T00:00:00Z\",\"end_date\":\"2026-12-02T00:00:00Z\",\"reason\":\"To be cancelled\"}" "$STUDENT_TOKEN" "Student Applies for Future Leave")
CANCEL_ID=$(echo "$CANCEL_LEAVE_RES" | jq -r '.data.id')
log_request "DELETE" "/leaves/$CANCEL_ID" "" "$STUDENT_TOKEN" "Student Cancels Leave"

echo ""
log "=================== TEST RUN COMPLETE ==================="
log "Results saved to $OUTPUT_FILE"