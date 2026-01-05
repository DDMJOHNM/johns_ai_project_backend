# Debugging Guide

This document explains the debugging features and how to troubleshoot the 404 issue with `POST /api/clients/add`.

## Changes Made

### 1. Fixed Route Mismatch
- **Problem**: API Gateway was configured for `POST /api/client/add` (singular), but the catch-all route `/api/clients/` could interfere
- **Solution**: 
  - Added explicit route for `POST /api/clients/add` (plural, RESTful convention) in both router and API Gateway
  - Kept legacy route `POST /api/client/add` (singular) for backward compatibility
  - Added comprehensive logging to track which route handles each request

### 2. Enhanced API Gateway Logging
- **Access Logs**: Already configured, now with enhanced error information
- **Execution Logs**: Enabled with `DataTraceEnabled: true` and `LoggingLevel: INFO`
  - Shows detailed request/response data
  - Logs integration errors and backend responses
- **Log Groups**:
  - `/aws/apigateway/{API_ID}/access` - Access logs
  - `/aws/apigateway/{API_ID}/execution` - Execution logs (detailed)

### 3. CloudWatch Logs for EC2 Backend
- **CloudWatch Agent**: Installed and configured on EC2 instance
- **Log Forwarding**: Application logs from `journalctl` are forwarded to CloudWatch via rsyslog
- **Log Group**: `/aws/ec2/john-ai-backend/application`
- **Retention**: 7 days

### 4. Enhanced Router Logging
- Added detailed logging at multiple points:
  - Middleware: Logs all incoming requests with headers
  - Route handlers: Logs when routes are matched
  - Catch-all handler: Logs when requests fall through to catch-all

## How to Debug

### View API Gateway Logs

```bash
# View API Gateway access logs
./scripts/view-cloudwatch-logs.sh api-gateway

# Or view in AWS Console
# Navigate to: CloudWatch > Log groups > /aws/apigateway/{API_ID}/access
```

**What to look for:**
- `routeKey`: Should show `POST /api/clients/add` or `POST /api/client/add`
- `status`: HTTP status code (404 indicates route not found)
- `error`: Any error messages from API Gateway or backend

### View EC2 Backend Logs

```bash
# View EC2 application logs
./scripts/view-cloudwatch-logs.sh ec2

# Or view via SSM (direct journalctl)
./scripts/view-backend-logs.sh
```

**What to look for:**
- `[MIDDLEWARE]`: Shows incoming request path and method
- `[ROUTER]`: Shows which route handler is being called
- `POST /api/clients/add`: Should appear when the route is matched

### View All Logs

```bash
# View both API Gateway and EC2 logs
./scripts/view-cloudwatch-logs.sh all
```

## Testing the Fix

### 1. Test via API Gateway

```bash
# Get API Gateway URL
API_URL=$(make get-api-url | grep -v "Stack not found")

# Test POST /api/clients/add
curl -v -X POST $API_URL/api/clients/add \
  -H "Content-Type: application/json" \
  -d '{"first_name":"Test","last_name":"User","email":"test@example.com"}'
```

### 2. Test Directly Against Backend

```bash
# Get backend URL
BACKEND_URL=$(aws cloudformation describe-stacks \
  --stack-name johns-ai-backend-ec2 \
  --query 'Stacks[0].Outputs[?OutputKey==`BackendURL`].OutputValue' \
  --output text)

# Test POST /api/clients/add
curl -v -X POST $BACKEND_URL/api/clients/add \
  -H "Content-Type: application/json" \
  -d '{"first_name":"Test","last_name":"User","email":"test@example.com"}'
```

## Common Issues

### Issue: 404 from API Gateway
**Check:**
1. API Gateway logs - verify route is registered
2. Route key matches exactly: `POST /api/clients/add` (with leading slash)
3. Integration is configured correctly

### Issue: 404 from Backend
**Check:**
1. EC2 logs - see which route handler is being called
2. Middleware logs - verify path after stage prefix stripping
3. Router logs - verify route matching logic

### Issue: No Logs Appearing
**Check:**
1. CloudWatch agent is running: `systemctl status amazon-cloudwatch-agent`
2. Log groups exist in CloudWatch Console
3. IAM permissions for CloudWatch Logs
4. Service is running: `systemctl status john-ai-backend.service`

## Route Configuration

### Supported Routes

| Method | Path | Handler | Notes |
|--------|------|---------|-------|
| GET | `/health` | Health check | |
| GET | `/api/clients` | Get all clients | |
| POST | `/api/clients/add` | Create client | **Primary route** |
| POST | `/api/client/add` | Create client | Legacy (backward compat) |
| GET | `/api/clients/active` | Get active clients | |
| GET | `/api/clients/inactive` | Get inactive clients | |
| GET | `/api/clients/{id}` | Get client by ID | Catch-all route |

### Route Matching Order

1. Specific routes (registered first):
   - `/api/clients/active`
   - `/api/clients/inactive`
   - `/api/clients/add` ← **This should match POST requests**
   - `/api/client/add` (legacy)
   - `/api/clients` (exact match, GET only)

2. Catch-all route (registered last):
   - `/api/clients/` - Matches `/api/clients/{id}` for GET requests only

## Next Steps

1. **Deploy the changes:**
   ```bash
   # Rebuild and redeploy backend
   make build-server
   ./scripts/deploy-backend.sh
   
   # Update API Gateway
   make deploy-api-gateway BACKEND_URL=http://your-backend-url:8080
   ```

2. **Test the endpoint:**
   ```bash
   curl -X POST $API_URL/api/clients/add \
     -H "Content-Type: application/json" \
     -d '{"first_name":"Test","last_name":"User","email":"test@example.com"}'
   ```

3. **Monitor logs:**
   ```bash
   # Watch logs in real-time
   ./scripts/view-cloudwatch-logs.sh all
   ```

4. **If still getting 404:**
   - Check API Gateway logs for route matching
   - Check EC2 logs for backend route matching
   - Verify the request path exactly matches the route key
   - Check for any path rewriting issues

