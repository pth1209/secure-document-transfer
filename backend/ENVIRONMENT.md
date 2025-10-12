# Environment Variables Configuration

This document describes all required environment variables for the backend server.

## Required Variables

### Supabase Configuration

```bash
# Your Supabase project URL
SUPABASE_URL=https://your-project.supabase.co

# Supabase anonymous/public key (safe to use in frontend)
SUPABASE_ANON_KEY=your_anon_key_here

# Supabase service role key (admin key - KEEP SECRET!)
# Required for creating auto-confirmed users without verification emails
SUPABASE_SERVICE_ROLE_KEY=your_service_role_key_here

# JWT secret for verifying tokens
SUPABASE_JWT_SECRET=your_jwt_secret_here
```

### Database Configuration

```bash
# PostgreSQL connection string
DATABASE_URL=postgresql://user:password@host:port/database
```

### Server Configuration

```bash
# Port for the backend server
PORT=8080

# Frontend URL for email redirects (verification and password reset)
FRONTEND_URL=http://localhost:3000
```

## How to Get Your Supabase Keys

1. Go to [Supabase Dashboard](https://supabase.com/dashboard)
2. Select your project
3. Navigate to **Settings** > **API**
4. You'll find:
   - **Project URL** → use as `SUPABASE_URL`
   - **anon/public key** → use as `SUPABASE_ANON_KEY`
   - **service_role key** → use as `SUPABASE_SERVICE_ROLE_KEY` ⚠️ **Keep this secret!**
5. Navigate to **Settings** > **API** > **JWT Settings**
   - **JWT Secret** → use as `SUPABASE_JWT_SECRET`

## Creating Your .env File

Create a file named `.env` in the `backend/` directory:

```bash
cd backend
touch .env
```

Then copy the example above and fill in your actual values.

## Security Notes

- **NEVER commit your `.env` file to version control**
- The `.env` file is already in `.gitignore`
- The `SUPABASE_SERVICE_ROLE_KEY` has admin privileges - treat it like a password
- Only the backend server should have access to the service role key

## Why Service Role Key is Required

The service role key is needed for creating auto-confirmed users when someone receives a file share:
- Regular signup: Uses `SUPABASE_ANON_KEY` → sends verification email
- Auto-created users: Uses `SUPABASE_SERVICE_ROLE_KEY` → NO verification email, only password reset email

This allows recipients to set their password immediately without needing to verify their email first.


