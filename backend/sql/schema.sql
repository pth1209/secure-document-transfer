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

-- ============================================================================
-- FILE MANAGEMENT TABLES
-- ============================================================================

-- Drop existing file-related tables if they exist
DROP TABLE IF EXISTS public.file_recipients CASCADE;
DROP TABLE IF EXISTS public.file_chunks CASCADE;
DROP TABLE IF EXISTS public.file_metadata CASCADE;

-- Create the file_metadata table to track files being transferred
CREATE TABLE public.file_metadata (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id TEXT NOT NULL UNIQUE, -- Client-generated file ID
    sender_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    original_filename TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    total_chunks INTEGER NOT NULL,
    mime_type TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

-- Create index for faster lookups
CREATE INDEX idx_file_metadata_file_id ON public.file_metadata(file_id);
CREATE INDEX idx_file_metadata_sender_id ON public.file_metadata(sender_id);

-- Create the file_chunks table to track individual chunks
CREATE TABLE public.file_chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id TEXT NOT NULL REFERENCES public.file_metadata(file_id) ON DELETE CASCADE,
    chunk_index INTEGER NOT NULL,
    chunk_size BIGINT NOT NULL,
    storage_path TEXT NOT NULL, -- Path in Supabase Storage
    encryption_iv TEXT NOT NULL, -- IV used for chunk encryption (base64)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(file_id, chunk_index)
);

-- Create indexes for efficient chunk retrieval
CREATE INDEX idx_file_chunks_file_id ON public.file_chunks(file_id);
CREATE INDEX idx_file_chunks_file_id_index ON public.file_chunks(file_id, chunk_index);

-- Create the file_recipients table to track encrypted keys for each recipient
-- This implements hybrid encryption: file encrypted with AES, AES key encrypted with each recipient's RSA public key
CREATE TABLE public.file_recipients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id TEXT NOT NULL REFERENCES public.file_metadata(file_id) ON DELETE CASCADE,
    recipient_id UUID REFERENCES auth.users(id) ON DELETE CASCADE,
    recipient_email TEXT NOT NULL, -- Store email for recipients who don't have accounts yet
    encrypted_file_key TEXT NOT NULL, -- AES key encrypted with recipient's public key (base64)
    downloaded_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    -- Ensure each recipient can only be added once per file (for concurrent upload safety)
    UNIQUE(file_id, recipient_email)
);

-- Create indexes for recipient lookups
CREATE INDEX idx_file_recipients_file_id ON public.file_recipients(file_id);
CREATE INDEX idx_file_recipients_recipient_id ON public.file_recipients(recipient_id);
CREATE INDEX idx_file_recipients_recipient_email ON public.file_recipients(recipient_email);

-- ============================================================================
-- ROW LEVEL SECURITY POLICIES FOR FILE TABLES
-- ============================================================================

-- Enable RLS on all file tables
ALTER TABLE public.file_metadata ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.file_chunks ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.file_recipients ENABLE ROW LEVEL SECURITY;

-- file_metadata policies
CREATE POLICY "Users can insert their own files"
    ON public.file_metadata
    FOR INSERT
    WITH CHECK (auth.uid()::uuid = sender_id);

CREATE POLICY "Users can view files they sent or received"
    ON public.file_metadata
    FOR SELECT
    USING (
        auth.uid()::uuid = sender_id 
        OR EXISTS (
            SELECT 1 FROM public.file_recipients 
            WHERE file_recipients.file_id = file_metadata.file_id 
            AND (file_recipients.recipient_id = auth.uid()::uuid 
                 OR file_recipients.recipient_email = (SELECT email FROM auth.users WHERE id = auth.uid()::uuid))
        )
    );

-- file_chunks policies
CREATE POLICY "Senders can insert chunks for their files"
    ON public.file_chunks
    FOR INSERT
    WITH CHECK (
        EXISTS (
            SELECT 1 FROM public.file_metadata 
            WHERE file_metadata.file_id = file_chunks.file_id 
            AND file_metadata.sender_id = auth.uid()::uuid
        )
    );

CREATE POLICY "Users can view chunks for files they have access to"
    ON public.file_chunks
    FOR SELECT
    USING (
        EXISTS (
            SELECT 1 FROM public.file_metadata 
            WHERE file_metadata.file_id = file_chunks.file_id 
            AND (file_metadata.sender_id = auth.uid()::uuid
                 OR EXISTS (
                     SELECT 1 FROM public.file_recipients 
                     WHERE file_recipients.file_id = file_chunks.file_id 
                     AND (file_recipients.recipient_id = auth.uid()::uuid
                          OR file_recipients.recipient_email = (SELECT email FROM auth.users WHERE id = auth.uid()::uuid))
                 ))
        )
    );

-- file_recipients policies
CREATE POLICY "Senders can insert recipients for their files"
    ON public.file_recipients
    FOR INSERT
    WITH CHECK (
        EXISTS (
            SELECT 1 FROM public.file_metadata 
            WHERE file_metadata.file_id = file_recipients.file_id 
            AND file_metadata.sender_id = auth.uid()::uuid
        )
    );

CREATE POLICY "Users can view recipient records for their files"
    ON public.file_recipients
    FOR SELECT
    USING (
        EXISTS (
            SELECT 1 FROM public.file_metadata 
            WHERE file_metadata.file_id = file_recipients.file_id 
            AND file_metadata.sender_id = auth.uid()::uuid
        )
        OR recipient_id = auth.uid()::uuid
        OR recipient_email = (SELECT email FROM auth.users WHERE id = auth.uid()::uuid)
    );

CREATE POLICY "Recipients can update their download status"
    ON public.file_recipients
    FOR UPDATE
    USING (
        recipient_id = auth.uid()::uuid
        OR recipient_email = (SELECT email FROM auth.users WHERE id = auth.uid()::uuid)
    );

-- ============================================================================
-- STORAGE BUCKET POLICIES
-- ============================================================================
-- Note: These policies control access to the 'encrypted-files' storage bucket
-- The bucket must be created manually in Supabase Dashboard or via API:
--   1. Go to Storage → New bucket
--   2. Name: encrypted-files
--   3. Public: false (private bucket)
--   4. Enable RLS

-- Storage policies use the storage.objects table
-- Path format: encrypted-files/{file_id}/chunk_{index}.enc

-- Simplified storage policies to avoid infinite recursion with RLS
-- We rely on application-level checks in the backend for fine-grained access control

-- Policy: Allow authenticated users to upload to encrypted-files bucket
-- The backend ensures users only upload chunks for their own files
CREATE POLICY "Authenticated users can upload encrypted files"
    ON storage.objects
    FOR INSERT
    WITH CHECK (
        bucket_id = 'encrypted-files' 
        AND auth.uid() IS NOT NULL
    );

-- Policy: Allow authenticated users to download from encrypted-files bucket
-- The backend ensures users only download chunks they have access to
CREATE POLICY "Authenticated users can download encrypted files"
    ON storage.objects
    FOR SELECT
    USING (
        bucket_id = 'encrypted-files'
        AND auth.uid() IS NOT NULL
    );

-- Policy: Allow authenticated users to delete from encrypted-files bucket
-- The backend ensures users only delete their own files
CREATE POLICY "Authenticated users can delete encrypted files"
    ON storage.objects
    FOR DELETE
    USING (
        bucket_id = 'encrypted-files'
        AND auth.uid() IS NOT NULL
    );

