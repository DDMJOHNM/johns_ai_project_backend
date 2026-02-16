#!/bin/bash

# Helper script for Docker debugging

echo "🐳 Starting Docker containers with debugger..."
docker-compose up -d --build

echo ""
echo "⏳ Waiting for services to be ready..."
sleep 5

echo ""
echo "✅ Services are running:"
docker-compose ps

echo ""
echo "📝 To debug:"
echo "   1. Set breakpoints in your Go code (e.g., internal/handler/auth_handler.go)"
echo "   2. In VS Code, press F5 and select 'Attach to Docker Container'"
echo "   3. The server will start once debugger attaches"
echo "   4. Trigger your frontend login route"
echo "   5. The debugger will pause at your breakpoints!"
echo ""
echo "📋 Useful commands:"
echo "   View logs:     docker-compose logs -f backend"
echo "   Stop services: docker-compose down"
echo "   Restart:       docker-compose restart backend"
echo "   Rebuild:       docker-compose up -d --build backend"
echo ""
echo "🎯 Debugger is listening on localhost:2345"
echo "🌐 Backend will listen on localhost:8080 (after debugger attaches)"

