#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Monopoly M&M - Full Stack Startup${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Step 1: Check and setup environment
echo -e "${YELLOW}[1/4] Checking environment setup...${NC}"
if [ ! -f ".internal.env" ]; then
    echo -e "${YELLOW}Creating .internal.env...${NC}"
    postgres_pass=$(openssl rand -base64 24 | tr -dc 'A-Za-z0-9' | head -c 24)
    touch .internal.env
    echo "POSTGRES_PASSWORD=$postgres_pass" >> .internal.env
    echo "POSTGRES_PORT=1357" >> .internal.env
    echo -e "${GREEN}✓ .internal.env created${NC}"
else
    echo -e "${GREEN}✓ .internal.env exists${NC}"
fi
echo ""

# Step 2: Check Docker and start postgres if needed
echo -e "${YELLOW}[2/4] Setting up PostgreSQL Docker container...${NC}"

# Source the env file to get the port and password
source .internal.env

# Check if docker is installed
if ! command -v docker &> /dev/null; then
    echo -e "${RED}✗ Docker is not installed. Please install Docker first.${NC}"
    exit 1
fi

# Check if container exists
if docker ps -a --format '{{.Names}}' | grep -q "^monopoly-postgres-e$"; then
    if docker ps --format '{{.Names}}' | grep -q "^monopoly-postgres-e$"; then
        echo -e "${GREEN}✓ PostgreSQL container is already running${NC}"
    else
        echo -e "${YELLOW}Starting PostgreSQL container...${NC}"
        docker start monopoly-postgres-e
        echo -e "${GREEN}✓ PostgreSQL container started${NC}"
    fi
else
    echo -e "${YELLOW}Creating and starting PostgreSQL container...${NC}"
    docker run --name monopoly-postgres-e \
        -e POSTGRES_PASSWORD=$POSTGRES_PASSWORD \
        -d \
        -p $POSTGRES_PORT:5432 \
        postgres
    echo -e "${GREEN}✓ PostgreSQL container created and started${NC}"
fi

sleep 2
echo ""

# Step 3: Check frontend dependencies
echo -e "${YELLOW}[3/4] Checking frontend dependencies...${NC}"
if [ ! -d "frontend/node_modules" ]; then
    echo -e "${YELLOW}Installing frontend dependencies...${NC}"
    cd frontend
    npm install
    cd ..
    echo -e "${GREEN}✓ Frontend dependencies installed${NC}"
else
    echo -e "${GREEN}✓ Frontend dependencies exist${NC}"
fi
echo ""

# Step 4: Start servers
echo -e "${YELLOW}[4/4] Starting servers...${NC}"
echo ""

# Start backend in background
echo -e "${BLUE}Starting Backend Server (port 9876)...${NC}"
cd backend
go run main.go &
BACKEND_PID=$!
cd ..

sleep 3

# Start frontend
echo -e "${BLUE}Starting Frontend Dev Server (port 3000)...${NC}"
cd frontend
npm run dev &
FRONTEND_PID=$!
cd ..

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  ✓ All systems running!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${BLUE}Web UI:${NC}      http://localhost:3000"
echo -e "${BLUE}Backend API:${NC}  http://localhost:9876"
echo -e "${BLUE}Database:${NC}     localhost:$POSTGRES_PORT"
echo ""
echo -e "${YELLOW}To stop all services:${NC}"
echo "  Press Ctrl+C"
echo ""
echo -e "${YELLOW}To stop only the Docker database:${NC}"
echo "  docker stop monopoly-postgres-e"
echo ""
echo -e "${YELLOW}Process IDs:${NC}"
echo "  Backend:  $BACKEND_PID"
echo "  Frontend: $FRONTEND_PID"
echo ""

# Wait for processes
wait
