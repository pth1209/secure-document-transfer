-- Schema for secure document transfer application
-- This should be run on your Supabase database

-- Drop existing table and recreate to ensure clean state
DROP TABLE IF EXISTS public.users CASCADE;

-- Create the public.users table to store user encryption keys
-- Note: User authentication is handled by Supabase Auth (auth.users table)
-- This table stores additional user data including encryption keys
CREATE TABLE public.users (
    id UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
    public_key TEXT NOT NULL,
    encrypted_private_key TEXT NOT NULL,
    salt TEXT NOT NULL,
    iv TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Enable Row Level Security
ALTER TABLE public.users ENABLE ROW LEVEL SECURITY;

-- Policy: Users can read their own data
CREATE POLICY "Users can read their own data"
    ON public.users
    FOR SELECT
    USING (auth.uid()::uuid = id);

-- Policy: Users can update their own data
CREATE POLICY "Users can update their own data"
    ON public.users
    FOR UPDATE
    USING (auth.uid()::uuid = id);

-- Policy: Users can insert their own data (for signup)
CREATE POLICY "Users can insert their own data"
    ON public.users
    FOR INSERT
    WITH CHECK (auth.uid()::uuid = id);

-- Create an index on the id column for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_id ON public.users(id);

-- Create an updated_at trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON public.users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

