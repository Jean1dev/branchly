#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="${RUNNER_ENV_FILE:-${ROOT_DIR}/.env}"

if [[ -f "${ENV_FILE}" ]]; then
  set -a
  source "${ENV_FILE}"
  set +a
fi

if [[ -z "${EMAIL_API_URL:-}" ]]; then
  echo "EMAIL_API_URL is required"
  exit 1
fi

EMAIL_FROM="${EMAIL_FROM:-noreply@branchly.com}"
BASE_URL="${EMAIL_API_URL%/}"
TEMPLATES_DIR="${ROOT_DIR}/internal/notifier/templates"

register_template() {
  local slug="$1"
  local name="$2"
  local file_path="${TEMPLATES_DIR}/${slug}.html"

  if [[ ! -f "${file_path}" ]]; then
    echo "template file not found: ${file_path}"
    return 1
  fi

  local put_payload
  put_payload="$(python3 - "$name" "$EMAIL_FROM" "$file_path" <<'PY'
import json
import pathlib
import sys

name = sys.argv[1]
from_email = sys.argv[2]
file_path = pathlib.Path(sys.argv[3])
html = file_path.read_text(encoding="utf-8")
print(json.dumps({
    "name": name,
    "fromEmail": from_email,
    "htmlTemplate": html,
}))
PY
)"

  local response_file
  response_file="$(mktemp)"

  local put_status
  put_status="$(
    curl -sS -o "${response_file}" -w "%{http_code}" \
      -X PUT "${BASE_URL}/v2/email/templates/${slug}" \
      -H "Content-Type: application/json" \
      --data "${put_payload}"
  )"

  if [[ "${put_status}" == "404" ]]; then
    local post_payload
    post_payload="$(python3 - "$slug" "$name" "$EMAIL_FROM" "$file_path" <<'PY'
import json
import pathlib
import sys

slug = sys.argv[1]
name = sys.argv[2]
from_email = sys.argv[3]
file_path = pathlib.Path(sys.argv[4])
html = file_path.read_text(encoding="utf-8")
print(json.dumps({
    "slug": slug,
    "name": name,
    "fromEmail": from_email,
    "htmlTemplate": html,
}))
PY
)"

    local post_status
    post_status="$(
      curl -sS -o "${response_file}" -w "%{http_code}" \
        -X POST "${BASE_URL}/v2/email/templates" \
        -H "Content-Type: application/json" \
        --data "${post_payload}"
    )"
    if [[ "${post_status}" != "200" && "${post_status}" != "201" ]]; then
      echo "failed to create template ${slug} (status ${post_status})"
      cat "${response_file}"
      rm -f "${response_file}"
      return 1
    fi
    echo "created template: ${slug}"
    rm -f "${response_file}"
    return 0
  fi

  if [[ "${put_status}" != "200" && "${put_status}" != "204" ]]; then
    echo "failed to upsert template ${slug} (status ${put_status})"
    cat "${response_file}"
    rm -f "${response_file}"
    return 1
  fi

  echo "updated template: ${slug}"
  rm -f "${response_file}"
}

register_template "branchly-job-completed" "Branchly - Job Completed"
register_template "branchly-job-failed" "Branchly - Job Failed"
register_template "branchly-pr-opened" "Branchly - Pull Request Opened"

echo "templates registered successfully"
