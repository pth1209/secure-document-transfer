# Secure Document Transfer

A secure document transfer application with user authentication, built with Go, React, TypeScript, and Supabase Auth.

## Project Structure

```
secure-document-transfer/
├── backend/
│   ├── main.go           # Application entry point and routing
│   ├── auth.go           # Supabase Auth client and JWT middleware
│   ├── database.go       # Database connection and setup
│   ├── handlers.go       # HTTP request handlers
│   ├── models.go         # Data models and validation
│   ├── go.mod            # Go dependencies
│   └── .env              # Environment variables (create this)
├── frontend/
│   ├── src/
│   │   ├── pages/        # Page components (SignUp, Login, Dashboard)
│   │   ├── services/     # API services
│   │   ├── types/        # TypeScript types
│   │   ├── App.tsx       # Main app component
│   │   ├── App.css       # Global styles
│   │   └── main.tsx      # Entry point
│   ├── package.json      # Frontend dependencies
│   └── vite.config.ts    # Vite configuration
└── README.md             # This file
```

## Quick Start

### Prerequisites

- **Go 1.21+** - [Install Go](https://golang.org/dl/)
- **Node.js 18+** - [Install Node.js](https://nodejs.org/)
- **Supabase Account** - [Sign up free](https://supabase.com)

### 1. Set Up Supabase

1. Create an account at [supabase.com](https://supabase.com)
2. Create a new project
3. **Enable Email Authentication:**
   - Go to **Authentication** → **Providers**
   - Enable **Email** provider
   - (Optional for development) Disable email confirmation:
     - Go to **Authentication** → **Settings**
     - Toggle off **Enable email confirmations**
4. Get your credentials from **Project Settings** → **API**:
   - Project URL
   - Anon/Public key
   - Database URL (from **Settings** → **Database**)

### 2. Configure Environment

Create a `.env` file in the `backend/` directory:

```env
SUPABASE_URL=https://xxxxxxxxxxxxx.supabase.co
SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
DATABASE_URL=postgresql://postgres:password@db.project.supabase.co:5432/postgres
PORT=8080
```

### 3. Install Dependencies

**Backend:**
```bash
cd backend
go mod download
```

**Frontend:**
```bash
cd frontend
npm install
```

### 4. Run the Application

**Terminal 1 - Start Backend:**
```bash
cd backend
go run .
```

You should see:
```
Server starting on port 8080...
Supabase Auth integration enabled
```

**Terminal 2 - Start Frontend:**
```bash
cd frontend
npm run dev
```

The application will be available at **http://localhost:3000**

### 5. Use the Application

Open your browser and navigate to `http://localhost:3000`. You'll see the beautiful login page. You can:

1. Click "Create Account" to sign up
2. Fill in your details (name, email, password)
3. Sign in after creating your account
4. Access your dashboard

### 6. Test the API (Optional)

**Health check:**
```bash
curl http://localhost:8080/api/health
```

**Sign up a user:**
```bash
curl -X POST http://localhost:8080/api/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"SecurePass123","full_name":"Test User"}'
```

**Sign in:**
```bash
curl -X POST http://localhost:8080/api/signin \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"SecurePass123"}'
```

## API Documentation

### Authentication Flow

1. **Sign Up** - Creates user in Supabase Auth
2. **Sign In** - Returns access token (JWT) and refresh token
3. **Protected Routes** - Middleware validates JWT on each request
4. **Sign Out** - Invalidates current session

### Endpoints

#### Public Endpoints (No Authentication)

##### `GET /api/health`
Health check endpoint

**Response:**
```json
{
  "status": "healthy"
}
```

##### `POST /api/signup`
Register a new user

**Request:**
```json
{
  "email": "user@example.com",
  "password": "securepass123",
  "full_name": "John Doe"
}
```

**Success Response (201):**
```json
{
  "message": "User registered successfully. Please check your email to verify your account.",
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid-here",
    "email": "user@example.com",
    "full_name": "John Doe"
  }
}
```

**Validation Rules:**
- Email: Required, valid email format
- Password: Required, minimum 8 characters
- Full Name: Optional, maximum 255 characters

##### `POST /api/signin`
Login with email and password

**Request:**
```json
{
  "email": "user@example.com",
  "password": "securepass123"
}
```

**Success Response (200):**
```json
{
  "message": "Login successful",
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "refresh-token-here",
  "user": {
    "id": "uuid-here",
    "email": "user@example.com",
    "full_name": "John Doe"
  }
}
```

#### Protected Endpoints (Authentication Required)

For protected endpoints, include the access token in the Authorization header:
```
Authorization: Bearer <access_token>
```

##### `GET /api/profile`
Get authenticated user's profile

**Request:**
```bash
curl http://localhost:8080/api/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Success Response (200):**
```json
{
  "id": "uuid-here",
  "email": "user@example.com",
  "full_name": "John Doe"
}
```

##### `POST /api/signout`
Sign out and invalidate session

**Request:**
```bash
curl -X POST http://localhost:8080/api/signout \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Success Response (200):**
```json
{
  "message": "Signed out successfully"
}
```

### Error Responses

All endpoints may return error responses in this format:
```json
{
  "error": "Error message",
  "details": "Additional details (optional)"
}
```

Common HTTP status codes:
- `400` - Bad Request (validation errors, invalid input)
- `401` - Unauthorized (invalid credentials, missing/invalid token)
- `500` - Internal Server Error

## Testing with cURL

Complete testing flow:

```bash
# 1. Check server health
curl http://localhost:8080/api/health

# 2. Sign up a new user
curl -X POST http://localhost:8080/api/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"SecurePass123","full_name":"John Doe"}'

# 3. Sign in (save the access_token from response)
curl -X POST http://localhost:8080/api/signin \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"SecurePass123"}'

# 4. Get profile (replace YOUR_TOKEN with access_token from step 3)
curl http://localhost:8080/api/profile \
  -H "Authorization: Bearer YOUR_TOKEN"

# 5. Sign out
curl -X POST http://localhost:8080/api/signout \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Frontend Features

The frontend is a modern, beautifully designed React + TypeScript application with:

- **🎨 Modern UI Design** - Inspired by leading SaaS platforms like Paraform and Krew
- **🌈 Gradient Effects** - Beautiful animated gradient backgrounds
- **⚡ Fast Performance** - Built with Vite for lightning-fast development and builds
- **📱 Responsive Design** - Works perfectly on desktop, tablet, and mobile
- **🎭 Smooth Animations** - Fade-in effects and hover animations
- **🔒 Secure Authentication** - JWT token management with localStorage
- **♿ Accessible Forms** - Proper form labels and validation
- **🎯 TypeScript** - Full type safety throughout the application

### Pages

- **Sign Up** (`/signup`) - Beautiful registration form with validation
- **Login** (`/login`) - Secure login page with error handling
- **Dashboard** (`/dashboard`) - Protected route showing user information

### Tech Stack

- React 18 with TypeScript
- Vite (build tool)
- React Router (routing)
- Axios (API calls)
- CSS3 with CSS Variables

## Security Features

- ✅ **Supabase Auth** - Industry-standard authentication with JWT
- ✅ **Token Verification** - Automatic JWT validation on protected routes
- ✅ **Password Security** - Secure password hashing managed by Supabase
- ✅ **SQL Injection Prevention** - Parameterized queries
- ✅ **Input Validation** - Email format and password strength validation
- ✅ **Environment Variables** - Secure credential management
- ✅ **CORS Support** - Configurable cross-origin requests
- ✅ **SSL/TLS** - Encrypted database connections

## Configuration

### Environment Variables

Create `backend/.env`:

```env
# Required: Supabase credentials (from Project Settings → API)
SUPABASE_URL=https://xxxxxxxxxxxxx.supabase.co
SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

# Required: Database connection (from Project Settings → Database)
DATABASE_URL=postgresql://postgres:password@db.project.supabase.co:5432/postgres

# Optional: Server port (default: 8080)
PORT=8080
```

### CORS Configuration

The server is configured to accept requests from any origin. To restrict access, modify the `enableCORS` function in `backend/main.go`:

```go
w.Header().Set("Access-Control-Allow-Origin", "https://yourdomain.com")
```

## Dependencies

- **[gorilla/mux](https://github.com/gorilla/mux)** - HTTP router
- **[lib/pq](https://github.com/lib/pq)** - PostgreSQL driver
- **[godotenv](https://github.com/joho/godotenv)** - Environment loader
- **[supabase-go](https://github.com/supabase-community/supabase-go)** - Supabase client
- **[jwt](https://github.com/golang-jwt/jwt/v5)** - JWT token handling

## Development

### Code Structure

**Handler Pattern:**
```go
func SignUpHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Handler logic using SupabaseClient
    }
}
```

**Route Setup with Middleware:**
```go
// Public routes
api.HandleFunc("/signup", SignUpHandler()).Methods("POST")

// Protected routes (with authentication middleware)
api.HandleFunc("/profile", AuthMiddleware(GetProfileHandler())).Methods("GET")
```

**Accessing User Info in Protected Routes:**
```go
userID := r.Context().Value("user_id")
userEmail := r.Context().Value("user_email")
```

### Adding New Endpoints

1. Create handler in `backend/handlers.go`
2. Add route in `backend/main.go`:
   ```go
   // Public route
   api.HandleFunc("/my-endpoint", MyHandler()).Methods("GET")
   
   // Protected route
   api.HandleFunc("/my-protected", AuthMiddleware(MyHandler())).Methods("GET")
   ```

### Database Usage

The database connection is available for application-specific data. User authentication is handled entirely by Supabase Auth (`auth.users` table).

Always use parameterized queries:
```go
db.QueryRow("SELECT * FROM documents WHERE user_id = $1", userID)

db.QueryRow(fmt.Sprintf("SELECT * FROM documents WHERE user_id = '%s'", userID))
```

## Building for Production

```bash
cd backend

# Build binary
go build -o server

# Build with optimizations (smaller binary)
go build -ldflags="-s -w" -o server

# Run binary
./server
```

## Troubleshooting

**"Failed to initialize Supabase client"**
- Check `SUPABASE_URL` and `SUPABASE_ANON_KEY` in `.env`
- Verify credentials in Supabase dashboard (Project Settings → API)
- Remove extra spaces or quotes from values

**"Invalid or expired token"**
- JWT tokens expire (default: 1 hour)
- Sign in again to get a new access token
- Check if email confirmation is required

**"Failed to connect to database"**
- Verify `DATABASE_URL` in `.env`
- Ensure Supabase project is fully provisioned (wait 2-3 minutes after creation)
- Test connection in Supabase SQL Editor

**"Authentication not working"**
- Ensure Email provider is enabled in Supabase (Authentication → Providers)
- Check if email confirmation is required (Authentication → Settings)
- For development, consider disabling email confirmation

**"Port already in use"**
- Change `PORT` in `.env` to a different port
- Or kill the process:
  ```bash
  lsof -ti:8080 | xargs kill -9
  ```

**"Invalid request body" with curl**
- Use single-line JSON in curl commands
- Avoid pretty-printed JSON with newlines

## Roadmap

### Completed
- [x] User registration (sign up)
- [x] User login (sign in)
- [x] JWT authentication middleware
- [x] User profile endpoint
- [x] Sign out functionality
- [x] Modern React + TypeScript frontend
- [x] Beautiful UI with animations
- [x] Responsive design

### In Progress
- [ ] Password reset
- [ ] Email verification
- [ ] OAuth providers (Google, GitHub)
- [ ] Document upload/download
- [ ] File encryption
- [ ] Rate limiting
- [ ] Request logging
- [ ] Unit and integration tests

## Learn More

- [Supabase Docs](https://supabase.com/docs) - Official documentation
- [Supabase Auth](https://supabase.com/docs/guides/auth) - Authentication guide
- [Go Documentation](https://golang.org/doc/) - Go language docs
- [Gorilla Mux](https://github.com/gorilla/mux) - Router documentation

## License

MIT License - feel free to use this project for learning or production!

## Contributing

Contributions welcome! Feel free to open issues or submit pull requests.

---

Built with Go and Supabase
