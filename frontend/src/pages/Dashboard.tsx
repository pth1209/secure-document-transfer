import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { authService, userService } from '../services/api';
import type { User } from '../types/auth';
import type { FileChunk, ChunkedFile } from '../types/file';

// Configuration for file chunking
const CHUNK_SIZE = 1024 * 1024 * 2; // 2MB chunks

const Dashboard: React.FC = () => {
  const navigate = useNavigate();
  const [user, setUser] = useState<User | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<User[]>([]);
  const [selectedUsers, setSelectedUsers] = useState<User[]>([]);
  const [files, setFiles] = useState<File[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState<{
    currentFile: string;
    currentChunk: number;
    totalChunks: number;
    fileIndex: number;
    totalFiles: number;
  } | null>(null);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (!token) {
      navigate('/login');
      return;
    }

    const userStr = localStorage.getItem('user');
    if (userStr) {
      setUser(JSON.parse(userStr));
    }
  }, [navigate]);

  useEffect(() => {
    const delaySearch = setTimeout(() => {
      if (searchQuery.trim().length > 0) {
        handleSearch();
      } else {
        setSearchResults([]);
      }
    }, 300);

    return () => clearTimeout(delaySearch);
  }, [searchQuery, selectedUsers]);

  const handleSearch = async () => {
    try {
      setIsSearching(true);
      const results = await userService.searchUsers(searchQuery);
      // Filter out already selected users
      const filteredResults = results.filter(
        result => !selectedUsers.some(selected => selected.id === result.id)
      );
      setSearchResults(filteredResults);
    } catch (err) {
      console.error('Search error:', err);
      setSearchResults([]);
    } finally {
      setIsSearching(false);
    }
  };

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      const newFiles = Array.from(e.target.files);
      setFiles(prev => [...prev, ...newFiles]);
    }
  };

  const handleRemoveFile = (index: number) => {
    setFiles(prev => prev.filter((_, i) => i !== index));
  };

  const isValidEmail = (email: string): boolean => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  };

  const handleSelectUser = (user: User) => {
    if (!selectedUsers.some(u => u.id === user.id)) {
      setSelectedUsers(prev => [...prev, user]);
    }
    setSearchQuery('');
    setSearchResults([]);
  };

  const handleAddEmailRecipient = (email: string) => {
    const trimmedEmail = email.trim().toLowerCase();
    
    if (!isValidEmail(trimmedEmail)) {
      alert('Please enter a valid email address');
      return;
    }

    // Check if email already exists in selected users
    if (selectedUsers.some(u => u.email.toLowerCase() === trimmedEmail)) {
      alert('This email is already in the recipient list');
      return;
    }

    // Create a user object with the email
    const newRecipient: User = {
      id: `email-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`,
      email: trimmedEmail,
      full_name: trimmedEmail, // Use email as display name
    };

    setSelectedUsers(prev => [...prev, newRecipient]);
    setSearchQuery('');
    setSearchResults([]);
  };

  const handleSearchKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      if (searchQuery.trim()) {
        handleAddEmailRecipient(searchQuery);
      }
    }
  };

  const handleRemoveUser = (userId: string) => {
    setSelectedUsers(prev => prev.filter(u => u.id !== userId));
  };

  // Generate a unique file ID
  const generateFileId = (): string => {
    return `${Date.now()}-${Math.random().toString(36).substring(2, 15)}`;
  };

  // Chunk a single file into smaller segments
  const chunkFile = async (file: File): Promise<ChunkedFile> => {
    const fileId = generateFileId();
    const chunks: FileChunk[] = [];
    const totalChunks = Math.ceil(file.size / CHUNK_SIZE);

    for (let i = 0; i < totalChunks; i++) {
      const start = i * CHUNK_SIZE;
      const end = Math.min(start + CHUNK_SIZE, file.size);
      const chunkBlob = file.slice(start, end);

      chunks.push({
        file_id: fileId,
        chunk_index: i,
        total_chunks: totalChunks,
        chunk_data: chunkBlob,
        original_filename: file.name,
        file_size: file.size,
        chunk_size: chunkBlob.size,
      });
    }

    return {
      file_id: fileId,
      original_file: file,
      chunks,
      total_chunks: totalChunks,
    };
  };

  const handleSendFiles = async () => {
    if (selectedUsers.length === 0 || files.length === 0) {
      alert('Please select at least one recipient and attach at least one file');
      return;
    }

    try {
      setUploading(true);
      const recipientIds = selectedUsers.map(u => u.id);

      // Chunk all files
      const chunkedFiles: ChunkedFile[] = [];
      for (const file of files) {
        const chunked = await chunkFile(file);
        console.log(`File "${file.name}" chunked:`, {
          file_id: chunked.file_id,
          total_chunks: chunked.total_chunks,
          original_size: file.size,
          chunk_size: CHUNK_SIZE,
          chunks: chunked.chunks.map(c => ({
            chunk_index: c.chunk_index,
            chunk_size: c.chunk_size,
            file_id: c.file_id
          }))
        });
        chunkedFiles.push(chunked);
      }

      // Send chunks for all files
      for (let fileIndex = 0; fileIndex < chunkedFiles.length; fileIndex++) {
        const chunkedFile = chunkedFiles[fileIndex];
        
        for (let chunkIndex = 0; chunkIndex < chunkedFile.chunks.length; chunkIndex++) {
          const chunk = chunkedFile.chunks[chunkIndex];
          
          // Update progress
          setUploadProgress({
            currentFile: chunk.original_filename,
            currentChunk: chunkIndex + 1,
            totalChunks: chunk.total_chunks,
            fileIndex: fileIndex + 1,
            totalFiles: chunkedFiles.length,
          });

          // Send the chunk
          await userService.sendFileChunk(chunk, recipientIds);
        }
      }
      
      // Reset form and progress
      setUploadProgress(null);
      setFiles([]);
      setSelectedUsers([]);
      setSearchQuery('');
      setSearchResults([]);
      alert('Files sent successfully!');
    } catch (err) {
      console.error('Upload error:', err);
      alert('Failed to send files. Please try again.');
      setUploadProgress(null);
    } finally {
      setUploading(false);
    }
  };

  const handleSignOut = async () => {
    try {
      await authService.signout();
      localStorage.clear();
      navigate('/login');
    } catch (err) {
      console.error('Sign out error:', err);
      localStorage.clear();
      navigate('/login');
    }
  };

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
  };

  if (!user) {
    return null;
  }

  return (
    <div className="auth-container">
      <div className="auth-card" style={{ maxWidth: '800px' }}>
        {/* Header */}
        <div className="auth-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', textAlign: 'left' }}>
          <div>
            <h1 className="auth-title" style={{ marginBottom: '0.25rem' }}>Welcome, {user.full_name}!</h1>
            <p className="auth-subtitle" style={{ margin: 0 }}>{user.email}</p>
          </div>
          <button 
            onClick={handleSignOut}
            className="btn btn-secondary"
            style={{ height: 'fit-content' }}
          >
            Sign Out
          </button>
        </div>

        <div style={{ padding: '2rem 0' }}>
          {/* User Search */}
          <div style={{ marginBottom: '2rem' }}>
            <label style={{ 
              display: 'block', 
              marginBottom: '0.5rem', 
              fontWeight: 500,
              color: 'var(--text-primary)'
            }}>
              Add recipients
            </label>
            <div style={{ position: 'relative' }}>
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                onKeyDown={handleSearchKeyDown}
                placeholder="Type email and press Enter, or search users..."
                className="input"
                style={{ width: '100%' }}
              />
              {isSearching && (
                <div style={{
                  position: 'absolute',
                  right: '1rem',
                  top: '50%',
                  transform: 'translateY(-50%)',
                  color: 'var(--text-muted)'
                }}>
                  Searching...
                </div>
              )}
            </div>

            {/* Search Results */}
            {searchQuery.trim().length > 0 && (
              <div style={{
                marginTop: '0.5rem',
                border: '1px solid var(--border-color)',
                borderRadius: '8px',
                maxHeight: '200px',
                overflowY: 'auto',
                backgroundColor: 'var(--background)'
              }}>
                {searchResults.length > 0 ? (
                  searchResults.map((result) => (
                    <div
                      key={result.id}
                      onClick={() => handleSelectUser(result)}
                      style={{
                        padding: '1rem',
                        borderBottom: '1px solid var(--border-color)',
                        cursor: 'pointer',
                        transition: 'background-color 0.2s'
                      }}
                      onMouseEnter={(e) => e.currentTarget.style.backgroundColor = 'var(--surface)'}
                      onMouseLeave={(e) => e.currentTarget.style.backgroundColor = 'var(--background)'}
                    >
                      <div style={{ fontWeight: 500, color: 'var(--text-primary)' }}>
                        {result.full_name}
                      </div>
                      <div style={{ fontSize: '0.875rem', color: 'var(--text-muted)' }}>
                        {result.email}
                      </div>
                    </div>
                  ))
                ) : (
                  !isSearching && (
                    <div
                      onClick={() => handleAddEmailRecipient(searchQuery)}
                      style={{
                        padding: '1rem',
                        cursor: 'pointer',
                        transition: 'background-color 0.2s',
                        textAlign: 'center'
                      }}
                      onMouseEnter={(e) => e.currentTarget.style.backgroundColor = 'var(--surface)'}
                      onMouseLeave={(e) => e.currentTarget.style.backgroundColor = 'var(--background)'}
                    >
                      <div style={{ color: 'var(--primary-color)', fontWeight: 500 }}>
                        {isValidEmail(searchQuery.trim()) 
                          ? `Add "${searchQuery.trim()}" as recipient`
                          : 'Press Enter to add email'}
                      </div>
                    </div>
                  )
                )}
              </div>
            )}

            {/* Selected Users */}
            {selectedUsers.length > 0 && (
              <div style={{ marginTop: '1rem' }}>
                <div style={{ 
                  fontWeight: 500, 
                  marginBottom: '0.5rem',
                  color: 'var(--text-primary)'
                }}>
                  Sending to ({selectedUsers.length} {selectedUsers.length === 1 ? 'recipient' : 'recipients'})
                </div>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                  {selectedUsers.map((selectedUser) => (
                    <div
                      key={selectedUser.id}
                      style={{
                        padding: '0.75rem',
                        backgroundColor: 'var(--surface)',
                        borderRadius: '6px',
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center',
                        border: '1px solid var(--border-color)'
                      }}
                    >
                      <div>
                        <div style={{ fontWeight: 500, color: 'var(--text-primary)' }}>
                          {selectedUser.full_name}
                        </div>
                        <div style={{ fontSize: '0.875rem', color: 'var(--text-muted)' }}>
                          {selectedUser.email}
                        </div>
                      </div>
                      <button
                        onClick={() => handleRemoveUser(selectedUser.id)}
                        style={{
                          background: 'none',
                          border: 'none',
                          color: 'var(--text-muted)',
                          cursor: 'pointer',
                          fontSize: '1.25rem',
                          padding: '0.25rem 0.5rem'
                        }}
                      >
                        ✕
                      </button>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>

          {/* File Upload Area */}
          <div style={{ marginBottom: '2rem' }}>
            <label style={{ 
              display: 'block', 
              marginBottom: '0.5rem', 
              fontWeight: 500,
              color: 'var(--text-primary)'
            }}>
              Attach files
            </label>
            
            {/* File Input */}
            <div style={{
              border: '2px dashed var(--border-color)',
              borderRadius: '8px',
              padding: '2rem',
              textAlign: 'center',
              backgroundColor: 'var(--surface)',
              cursor: 'pointer',
              transition: 'all 0.2s'
            }}
            onDragOver={(e) => {
              e.preventDefault();
              e.currentTarget.style.borderColor = 'var(--primary-color)';
              e.currentTarget.style.backgroundColor = 'var(--background)';
            }}
            onDragLeave={(e) => {
              e.currentTarget.style.borderColor = 'var(--border-color)';
              e.currentTarget.style.backgroundColor = 'var(--surface)';
            }}
            onDrop={(e) => {
              e.preventDefault();
              e.currentTarget.style.borderColor = 'var(--border-color)';
              e.currentTarget.style.backgroundColor = 'var(--surface)';
              const droppedFiles = Array.from(e.dataTransfer.files);
              setFiles(prev => [...prev, ...droppedFiles]);
            }}
            onClick={() => document.getElementById('file-input')?.click()}
            >
              <div style={{ fontSize: '2rem', marginBottom: '0.5rem' }}>📎</div>
              <div style={{ color: 'var(--text-primary)', marginBottom: '0.25rem' }}>
                Click to browse or drag and drop files here
              </div>
              <div style={{ fontSize: '0.875rem', color: 'var(--text-muted)' }}>
                You can attach multiple files
              </div>
              <input
                id="file-input"
                type="file"
                multiple
                onChange={handleFileSelect}
                style={{ display: 'none' }}
              />
            </div>

            {/* Attached Files List */}
            {files.length > 0 && (
              <div style={{ marginTop: '1rem' }}>
                <div style={{ 
                  fontWeight: 500, 
                  marginBottom: '0.5rem',
                  color: 'var(--text-primary)'
                }}>
                  Attached files ({files.length})
                </div>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                  {files.map((file, index) => (
                    <div
                      key={index}
                      style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center',
                        padding: '0.75rem',
                        backgroundColor: 'var(--background)',
                        borderRadius: '6px',
                        border: '1px solid var(--border-color)'
                      }}
                    >
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <div style={{
                          fontWeight: 500,
                          color: 'var(--text-primary)',
                          whiteSpace: 'nowrap',
                          overflow: 'hidden',
                          textOverflow: 'ellipsis'
                        }}>
                          {file.name}
                        </div>
                        <div style={{ fontSize: '0.875rem', color: 'var(--text-muted)' }}>
                          {formatFileSize(file.size)}
                        </div>
                      </div>
                      <button
                        onClick={() => handleRemoveFile(index)}
                        style={{
                          background: 'none',
                          border: 'none',
                          color: 'var(--text-muted)',
                          cursor: 'pointer',
                          fontSize: '1.25rem',
                          padding: '0.25rem 0.5rem',
                          marginLeft: '1rem'
                        }}
                      >
                        ✕
                      </button>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>

          {/* Upload Progress */}
          {uploadProgress && (
            <div style={{ 
              marginBottom: '2rem', 
              padding: '1.5rem',
              backgroundColor: 'var(--surface)',
              borderRadius: '8px',
              border: '1px solid var(--border-color)'
            }}>
              <div style={{ 
                fontWeight: 500, 
                marginBottom: '1rem',
                color: 'var(--text-primary)'
              }}>
                Uploading files...
              </div>
              
              <div style={{ marginBottom: '0.5rem', color: 'var(--text-primary)' }}>
                File {uploadProgress.fileIndex} of {uploadProgress.totalFiles}: {uploadProgress.currentFile}
              </div>
              
              <div style={{ marginBottom: '0.75rem', fontSize: '0.875rem', color: 'var(--text-muted)' }}>
                Chunk {uploadProgress.currentChunk} of {uploadProgress.totalChunks}
              </div>
              
              {/* Progress Bar */}
              <div style={{
                width: '100%',
                height: '8px',
                backgroundColor: 'var(--background)',
                borderRadius: '4px',
                overflow: 'hidden'
              }}>
                <div style={{
                  height: '100%',
                  backgroundColor: 'var(--primary-color)',
                  width: `${(uploadProgress.currentChunk / uploadProgress.totalChunks) * 100}%`,
                  transition: 'width 0.3s ease'
                }} />
              </div>
            </div>
          )}

          {/* Send Button */}
          <button 
            onClick={handleSendFiles}
            className="btn btn-primary"
            disabled={selectedUsers.length === 0 || files.length === 0 || uploading}
            style={{ width: '100%' }}
          >
            {uploading ? 'Sending...' : 'Send Files'}
          </button>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;

