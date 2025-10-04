import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { authService, userService } from '../services/api';
import type { User } from '../types/auth';

const Dashboard: React.FC = () => {
  const navigate = useNavigate();
  const [user, setUser] = useState<User | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<User[]>([]);
  const [selectedUsers, setSelectedUsers] = useState<User[]>([]);
  const [files, setFiles] = useState<File[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const [uploading, setUploading] = useState(false);

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

  const handleSelectUser = (user: User) => {
    if (!selectedUsers.some(u => u.id === user.id)) {
      setSelectedUsers(prev => [...prev, user]);
    }
    setSearchQuery('');
    setSearchResults([]);
  };

  const handleRemoveUser = (userId: string) => {
    setSelectedUsers(prev => prev.filter(u => u.id !== userId));
  };

  const handleSendFiles = async () => {
    if (selectedUsers.length === 0 || files.length === 0) {
      alert('Please select at least one recipient and attach at least one file');
      return;
    }

    try {
      setUploading(true);
      const formData = new FormData();
      files.forEach(file => {
        formData.append('files', file);
      });
      // Send recipient IDs as comma-separated string or JSON array
      selectedUsers.forEach(user => {
        formData.append('recipient_ids[]', user.id);
      });

      await userService.sendFiles(formData);
      
      // Reset form
      setFiles([]);
      setSelectedUsers([]);
      setSearchQuery('');
      setSearchResults([]);
      alert('Files sent successfully!');
    } catch (err) {
      console.error('Upload error:', err);
      alert('Failed to send files. Please try again.');
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
              Search for users to send documents
            </label>
            <div style={{ position: 'relative' }}>
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Search by email or name..."
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
            {searchResults.length > 0 && (
              <div style={{
                marginTop: '0.5rem',
                border: '1px solid var(--border-color)',
                borderRadius: '8px',
                maxHeight: '200px',
                overflowY: 'auto',
                backgroundColor: 'var(--background)'
              }}>
                {searchResults.map((result) => (
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
                ))}
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

