# Swadiq Schools Management System

A comprehensive school management system built with Go and Fiber for Swadiq Schools in Uganda.

## Features

- **Authentication System**: Login/logout with session management
- **Role-based Access Control**: Admin, Head Teacher, Class Teacher, Subject Teacher
- **Password Management**: Change password functionality
- **Responsive UI**: Built with Tailwind CSS
- **PostgreSQL Integration**: Async queries with PostgreSQL 18

## Project Structure

```
swadiq-schools/
├── go.mod
├── main.go
├── schema.sql
└── app/
    ├── config/
    │   └── config.go          # Database configuration
    ├── models/
    │   └── user.go           # User and session models
    ├── database/
    │   └── queries.go        # Database queries
    ├── routes/
    │   └── auth/             # Authentication module
    │       ├── API.go        # API endpoints
    │       ├── routes.go     # Route handlers
    │       └── utils.go      # Helper functions
    └── templates/
        └── auth/             # Authentication templates
            ├── login.html
            └── profile.html
```

## Setup Instructions

### 1. Install Go
```bash
# Download and install Go 1.21 or later
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

### 2. Install Dependencies
```bash
cd swadiq-schools
go mod tidy
```

### 3. Database Setup
```bash
# Connect to PostgreSQL and run the schema
psql -h 129.80.199.242 -p 5432 -U imaad -d swadiq -f schema.sql
```

### 4. Run the Application
```bash
go run main.go
```

The application will start on `http://localhost:3000`

## Default Login Credentials

- **Admin**: imaad.ssebintu@gmail.com / Ertdfgx@0
- **Head Teacher**: headteacher@swadiq.com / password123
- **Class Teacher**: teacher1@swadiq.com / password123
- **Subject Teacher**: teacher2@swadiq.com / password123

## API Endpoints

### Authentication
- `GET /auth/login` - Login page
- `POST /auth/login` - Login API
- `POST /auth/logout` - Logout API
- `GET /auth/profile` - Profile page (protected)
- `POST /auth/change-password` - Change password API (protected)

## Database Schema

### Users Table
- `id` - Primary key
- `email` - Unique email address
- `password` - Bcrypt hashed password
- `first_name` - User's first name
- `last_name` - User's last name
- `role` - User role (admin, head_teacher, class_teacher, subject_teacher)
- `is_active` - Account status
- `created_at` - Creation timestamp
- `updated_at` - Last update timestamp

### Sessions Table
- `id` - Session UUID
- `user_id` - Foreign key to users table
- `expires_at` - Session expiration time
- `created_at` - Session creation time

## Next Steps

To extend the system, follow the same modular structure:

1. Create new module folder in `app/routes/`
2. Add `API.go`, `routes.go`, and `utils.go`
3. Create corresponding templates in `app/templates/`
4. Register routes in `main.go`

Example modules to implement:
- Teachers management
- Classes management
- Students management
- Reports and analytics
- Attendance tracking
- Grade management

## Security Features

- Password hashing with bcrypt
- Session-based authentication
- Role-based access control
- CSRF protection ready
- SQL injection prevention with prepared statements
- XSS protection with proper template escaping
