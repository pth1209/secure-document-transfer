# Secure Document Transfer - Backend

A secure document transfer backend built with Go, featuring end-to-end encryption and Supabase authentication.

## Project Structure

```
backend/
├── cmd/
│   └── server/          # Main application entry point
│       └── main.go      # Server initialization and routing
├── internal/            # Private application code
│   ├── config/          # Configuration and external services
│   │   └── supabase.go  # Supabase client initialization
│   ├── crypto/          # Cryptography utilities
│   │   ├── encryption.go      # RSA key generation and AES encryption
│   │   └── encryption_test.go # Encryption tests
│   ├── database/        # Database layer
│   │   ├── db.go        # Database connection
│   │   └── users.go     # User-related database queries
│   ├── handlers/        # HTTP request handlers
│   │   ├── auth.go      # Authentication handlers (signup, signin, signout)
│   │   ├── user.go      # User handlers (profile, search, public keys)
│   │   └── helpers.go   # Response helper functions
│   ├── middleware/      # HTTP middleware
│   │   └── auth.go      # Authentication middleware
│   └── models/          # Data models and validation
│       └── user.go      # User models and validation logic
├── sql/                 # SQL scripts
│   └── schema.sql       # Database schema
├── go.mod              # Go module definition
├── go.sum              # Go module checksums
└── .gitignore          # Git ignore rules
```

## Architecture Overview

### Separation of Concerns

- **cmd/**: Contains the application entry points (main packages)
- **internal/**: Private application code that cannot be imported by other projects
  - **config/**: External service configurations (Supabase client)
  - **crypto/**: Cryptography operations (RSA key generation, AES encryption)
  - **database/**: Database connection and queries
  - **handlers/**: HTTP request handling logic
  - **middleware/**: HTTP middleware (authentication, CORS, etc.)
  - **models/**: Data structures and business logic

### Key Features

1. **Supabase Integration**: Uses Supabase for authentication and PostgreSQL database
2. **End-to-End Encryption**: Generates RSA key pairs for users, encrypts private keys with password-derived keys
3. **RESTful API**: Clean HTTP API with proper error handling
4. **Middleware**: JWT authentication middleware for protected routes
5. **Type Safety**: Strongly typed models with validation

## Building and Running

### Build the server

```bash
go build -o bin/server ./cmd/server
```

### Run the server

```bash
./bin/server
```

Or directly with Go:

```bash
go run ./cmd/server
```

### Run tests

```bash
go test ./...
```

### Run specific package tests

```bash
go test ./internal/crypto -v
```

## Environment Variables

Create a `.env` file in the backend directory:

```env
# Supabase Configuration
SUPABASE_URL=your_supabase_url
SUPABASE_ANON_KEY=your_supabase_anon_key
SUPABASE_JWT_SECRET=your_jwt_secret

# Database Configuration
DATABASE_URL=your_database_url

# Server Configuration
PORT=8080
```

## API Endpoints

### Public Endpoints

- `GET /api/health` - Health check
- `POST /api/signup` - User registration
- `POST /api/signin` - User login

### Protected Endpoints (require authentication)

- `GET /api/profile` - Get user profile
- `POST /api/signout` - Sign out user
- `GET /api/users/search?q=query` - Search for users
- `GET /api/users/public-key?user_id=id` - Get user's public key

## Development Guidelines

### Adding New Features

1. **Models**: Add data structures to `internal/models/`
2. **Database**: Add queries to appropriate file in `internal/database/`
3. **Handlers**: Add HTTP handlers to `internal/handlers/`
4. **Routes**: Register routes in `cmd/server/main.go`

### Code Organization

- Keep packages focused and cohesive
- Use dependency injection where possible
- Write tests for critical functionality
- Follow Go naming conventions

### Testing

- Unit tests should be in the same package as the code they test
- Use descriptive test names: `TestFunctionName_Scenario_ExpectedBehavior`
- Mock external dependencies when necessary

## Security Notes

- Private keys are encrypted with AES-256-GCM before storage
- Passwords are used to derive encryption keys via PBKDF2 (100,000 iterations)
- JWT tokens are verified on every protected route
- CORS is configured for frontend integration

