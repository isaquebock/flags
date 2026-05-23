#!/bin/bash

# Generate args.json from shell environment variables
cat > azion/args.json <<EOF
{
  "GO_API_URL": "$GO_API_URL",
  "INTERNAL_TOKEN": "$INTERNAL_TOKEN",
  "API_TIMEOUT_MS": "${API_TIMEOUT_MS:-5000}",
  "CACHE_MAX_AGE": "${CACHE_MAX_AGE:-60}"
}
EOF

echo "✓ Generated azion/args.json"

# Build
npm run build:azion
echo "✓ Built Edge Function"

# Deploy via Azion CLI
azion deploy

echo "✓ Deployment complete"
