# AAA Backend

## 📋 Table of Contents

- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [API Endpoints](#api-endpoints)
- [API Testing with Insomnia](#API-Testing-with-Insomnia)

## ️ Tech Stack

- **Language**: Go
- **Database**: PostgreSQL
- **Authentication**: JWT with RSA-2048 keys
- **Containerization**: Docker & Docker Compose
- **API Documentation**: Insomnia

## Project Structure

```
.
├── cmd/
│   ├── main.go              # Application entry point
│   └── seed/
│       └── main.go          # Database seeding script
── internal/
│   ├── auth/
│   │   ├── jwt.go           # JWT token handling
│   │   └── password.go      # Password hashing utilities
│   ├── db/
│   │   ├── db.go            # Database connection pool
│   │   └── init.sql         # Database schema
│   ├── handler/
│   │   ├── auth.go          # Login/Register handlers
│   │   ├── compliance.go    # PDP-related handlers
│   │   └── exhibition.go    # Project & grading handlers
│   ├── lmodel/              # Data models
│   └── middlewares/
│       ├── auth.go          # Authentication middleware
│       └── audit.go         # Audit logging middleware
├── keys/                    # RSA key pairs (generated)
├── docker-compose.yml       # Docker services configuration
├── go.mod                   # Go module file
├── Makefile                 # Build automation
└── README.md
```

## Prerequisites

- Go 1.21 or higher
- Docker & Docker Compose
- OpenSSL (for key generation)
- Make (optional, but recommended)

## 🚀 Installation

### 1. Clone the Repository

```bash
git clone <repository-url>
cd aaa_bekend
```

### 2. Generate RSA Keys

```bash
make keys
```

This creates a `keys/` directory with:
- `private.pem` - Private key for signing JWT tokens
- `public.pem` - Public key for verifying JWT tokens

### 3. Start the Database

```bash
make up
```

This starts PostgreSQL using Docker Compose.

### 4. Initialize the Database

```bash
make seed
```

This runs the database migrations and seeds initial data.

### 5. Run the Server

```bash
make run
```

The server will start on `http://localhost:8080`

### Database Connection

The application uses a connection pool managed in `internal/db/db.go`. Default configuration:
- Host: localhost
- Port: 5432
- Database: exhibition_db

##  API Endpoint

### Authentication

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/v1/register` | Register a new user | No |
| POST | `/api/v1/login` | Login and get JWT token | No |

### Projects

| Method | Endpoint | Description | Auth Required | Role |
|--------|----------|-------------|---------------|------|
| GET | `/api/v1/projects` | List all projects | Yes | All |
| POST | `/api/v1/projects` | Submit a new project | Yes | Student |
| GET | `/api/v1/projects/{id}/grades` | Get grades for a project | Yes | All |
| POST | `/api/v1/projects/{id}/grade` | Grade a project | Yes | Judge, Admin |
| GET | `/api/v1/projects/grade` | Get all project grades | Yes | Judge, Admin |

### User & Privacy (PDP)

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/v1/users/me/data` | Get personal data | Yes |
| PUT | `/api/v1/users/me/data` | Update personal data | Yes |
| GET | `/api/v1/users/me/export` | Export personal data | Yes |
| GET | `/api/v1/users/me/audit-log` | Get personal audit log | Yes |
| DELETE | `/api/v1/users/me` | Delete account | Yes |

### Request Headers

All authenticated requests must include:
```
Authorization: Bearer <jwt_token>
```

## API-Testing-with-Insomnia

This project includes a pre-configured **Insomnia Collection** to quickly test all endpoints and verify AAA + UU PDP compliance.

### Import Instructions

1. Open **Insomnia** (Desktop or Web).
2. Click **Application** (top-left corner) → **Preferences** → **Data**.
3. Click **Import Data** → select imsomnia.yaml.
4. Click **Scan & Import**.
5. A new workspace named **"AAA + UU PDP Grading"** will appear. Open it.

### Setup Before Testing

Make sure your environment is running:

```bash
make up       # Start PostgreSQL database
make keys     # Generate RSA keys (first time only)
make seed     # Populate database with test data
make run      # Start the Go server on http://localhost:8080
```

### Managing Tokens
After running the login requests you have to configure these token value:

- `update_token`
- `student_token`
- `judge_token`
- `delete_token`

If you need to refresh a token manually:
1. Open **Manage Environments** (top-left dropdown).
2. Select **"Base Environment"**.
3. Paste the new token value into the appropriate variable.