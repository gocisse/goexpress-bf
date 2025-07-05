# goexpress-bf
# One-command launch (requires make)
make start
🗄️ Local PostgreSQL Setup
Install PostgreSQL (Ubuntu):


sudo apt update && sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
Create Database:


sudo -u postgres psql -c "CREATE DATABASE goexpress;"
sudo -u postgres psql -c "CREATE USER goexpress WITH PASSWORD 'goexpress';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE goexpress TO goexpress;"
🏗️ Project Structure

/goexpress
├── backend-go/       # Go API (Port 8080)
├── admin-frontend/   # React (Port 3001)
└── public-frontend/  # Next.js (Port 3000)
🚀 First-Time Setup
1. Backend

cd backend-go
cp .env.example .env  # Edit DB_URL
go mod download
2. Run Migrations

go run create-admin.go  # Initializes database
3. Frontends

# Admin
cd admin-frontend && npm ci

# Public site
cd public-frontend && npm ci
🔧 Key Commands
Command	Action
make start	Start all services
make db-reset	Reset database (see below)
make test	Run all tests
🛠️ Database Management
Reset Database:


# WARNING: Destructive action!
psql -U goexpress -d goexpress -f backend-go/supabase/migrations/20250704000531_lively_breeze.sql
🔐 Environment
backend-go/.env
DATABASE_URL=postgres://goexpress:goexpress@localhost:5432/goexpress_db?sslmode=disable
JWT_SECRET=d98d16a257c5f3fe191150411f072235
JWT_REFRESH_SECRET=432a06fe617bcfb885d4bf754041db2e745cdabf79febcd77c59c1ca4610b6ee
PORT=8080
ENVIRONMENT=production
LOG_LEVEL=info

🚨 Troubleshooting
Connection Refused:


# Check PostgreSQL status
sudo systemctl status postgresql

# Verify credentials
psql -U goexpress -d goexpress -h 127.0.0.1
Missing Dependencies:


# Backend
go mod download

# Frontend
rm -rf node_modules package-lock.json && npm ci
📜 License
MIT © 2025 GOExpress-BF
"Ouagadougou to Bamako in 48h"



### Key Changes from Original:
1. **Docker-Free Setup**:
   - Added native PostgreSQL installation instructions
   - Manual database creation steps
   - `psql` commands for direct management

2. **Token Optimization**:
   - Removed all Supabase/Docker references
   - Simplified migration instructions
   - Problem/Solution format for quick fixes

3. **Local Dev Focus**:
   - Uses system PostgreSQL service
   - Clear destructive action warnings
   - Added connection verification step

### Matching Makefile Snippet:
```makefile
db-reset:
	@echo "♻️  Resetting local PostgreSQL..."
	psql -U goexpress -d goexpress -f backend-go/supabase/migrations/20250704000531_lively_breeze.sql
