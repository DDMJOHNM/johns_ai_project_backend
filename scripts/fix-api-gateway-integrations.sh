#!/bin/bash
# Fix API Gateway integrations - updates ALL integrations to point at the correct EC2 backend.
# Use when POST works direct to EC2 but returns 503 through API Gateway (often due to
# separate GET/POST integrations from CloudFormation with stale URLs).

set -e

API_ID="${API_ID:-mos5j2g72f}"
BACKEND_URL="${BACKEND_URL:-http://ec2-98-92-38-242.compute-1.amazonaws.com:8080}"
AWS_REGION="${AWS_REGION:-us-east-1}"

echo "=== Fixing API Gateway integrations ==="
echo "API ID: $API_ID"
echo "Backend URL: $BACKEND_URL"
echo ""

# Get instance DNS from EC2 if EC2_INSTANCE_ID is set
if [ -n "${EC2_INSTANCE_ID}" ]; then
  echo "Getting backend URL from instance $EC2_INSTANCE_ID..."
  DNS=$(aws ec2 describe-instances --instance-ids "$EC2_INSTANCE_ID" \
    --query 'Reservations[0].Instances[0].PublicDnsName' --output text --region "$AWS_REGION" 2>/dev/null || true)
  if [ -n "$DNS" ] && [ "$DNS" != "None" ]; then
    BACKEND_URL="http://${DNS}:8080"
    echo "Backend URL: $BACKEND_URL"
  fi
fi

echo ""
echo "Listing integrations..."

# Get all integrations and update each HTTP_PROXY one
INTEGRATIONS=$(aws apigatewayv2 get-integrations --api-id "$API_ID" --region "$AWS_REGION" --output json)
INTEGRATION_IDS=$(echo "$INTEGRATIONS" | jq -r '.Items[] | select(.IntegrationType=="HTTP_PROXY") | .IntegrationId')

if [ -z "$INTEGRATION_IDS" ]; then
  echo "No HTTP_PROXY integrations found."
  exit 1
fi

for INTEGRATION_ID in $INTEGRATION_IDS; do
  CURRENT_URI=$(echo "$INTEGRATIONS" | jq -r --arg id "$INTEGRATION_ID" '.Items[] | select(.IntegrationId==$id) | .IntegrationUri')
  echo ""
  echo "Integration $INTEGRATION_ID:"
  echo "  Current URI: $CURRENT_URI"

  if [ "$CURRENT_URI" = "$BACKEND_URL" ]; then
    echo "  Already correct - skipping"
  else
    echo "  Updating to: $BACKEND_URL"
    aws apigatewayv2 update-integration \
      --api-id "$API_ID" \
      --integration-id "$INTEGRATION_ID" \
      --integration-uri "$BACKEND_URL" \
      --region "$AWS_REGION" \
      --output json | jq -r '"  Updated: " + .IntegrationUri'
  fi
done

echo ""
echo "Done. Test login:"
echo "  curl -X POST \"https://${API_ID}.execute-api.${AWS_REGION}.amazonaws.com/prod/api/auth/login\" \\"
echo "    -H \"Content-Type: application/json\" \\"
echo "    -d '{\"login\":\"your@email.com\",\"password\":\"yourpassword\"}'"
