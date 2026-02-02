#!/bin/sh
set -eu

API_BASE="${FLOWDB_API_BASE:-${NEXT_PUBLIC_API_BASE:-}}"
if [ -n "${API_BASE}" ]; then
  API_BASE="${API_BASE%/}"
fi

escape_json() {
  printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'
}

API_BASE_ESCAPED=$(escape_json "${API_BASE}")
cat > /app/public/flowdb-config.js <<EOF
window.__FLOWDB_CONFIG__ = { apiBase: "${API_BASE_ESCAPED}" };
EOF

exec "$@"
