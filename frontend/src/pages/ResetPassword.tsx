import React, { useState, useEffect } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { authService } from '../services/api';
import type { PasswordResetConfirm } from '../types/auth';

const ResetPassword: React.FC = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [formData, setFormData] = useState({
    new_password: '',
    confirm_password: '',
  });
  const [token, setToken] = useState<string>('');
  const [error, setError] = useState<string>('');
  const [success, setSuccess] = useState<string>('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    // Extract token from URL hash (Supabase sends it as #access_token=...)
    const hash = window.location.hash;
    if (hash) {
      const params = new URLSearchParams(hash.substring(1));
      const accessToken = params.get('access_token');
      if (accessToken) {
        setToken(accessToken);
      } else {
        setError('Invalid or missing reset token. Please request a new password reset.');
      }
    } else {
      // Try getting from query params as fallback
      const tokenParam = searchParams.get('token');
      if (tokenParam) {
        setToken(tokenParam);
      } else {
        setError('Invalid or missing reset token. Please request a new password reset.');
      }
    }
  }, [searchParams]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    });
    setError('');
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    // Validate passwords match
    if (formData.new_password !== formData.confirm_password) {
      setError('Passwords do not match');
      return;
    }

    // Validate password length
    if (formData.new_password.length < 8) {
      setError('Password must be at least 8 characters long');
      return;
    }

    if (!token) {
      setError('Invalid or missing reset token. Please request a new password reset.');
      return;
    }

    setLoading(true);

    try {
      const resetData: PasswordResetConfirm = {
        token: token,
        new_password: formData.new_password,
      };
      const response = await authService.resetPassword(resetData);
      setSuccess(response.message);
      
      // Redirect to login after a short delay
      setTimeout(() => {
        navigate('/login');
      }, 2000);
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || 'Failed to reset password. The link may have expired.';
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <nav className="nav">
        <div className="nav-container">
          <Link to="/" className="logo">
            SecureTransfer
          </Link>
          <div className="nav-links">
            <Link to="/login" className="btn btn-secondary">
              Sign In
            </Link>
          </div>
        </div>
      </nav>
      <div className="auth-container">
        <div className="auth-card">
          <div className="auth-header">
            <h1 className="auth-title">Set New Password</h1>
            <p className="auth-subtitle">
              Enter your new password below
            </p>
          </div>

          {error && <div className="error-message">{error}</div>}
          {success && <div className="success-message">{success}</div>}

          <form className="form" onSubmit={handleSubmit}>
            <div className="form-group">
              <label htmlFor="new_password" className="form-label">
                New Password
              </label>
              <input
                type="password"
                id="new_password"
                name="new_password"
                className="form-input"
                placeholder="At least 8 characters"
                value={formData.new_password}
                onChange={handleChange}
                required
                minLength={8}
                disabled={!token || !!success}
              />
            </div>

            <div className="form-group">
              <label htmlFor="confirm_password" className="form-label">
                Confirm New Password
              </label>
              <input
                type="password"
                id="confirm_password"
                name="confirm_password"
                className="form-input"
                placeholder="Re-enter your password"
                value={formData.confirm_password}
                onChange={handleChange}
                required
                minLength={8}
                disabled={!token || !!success}
              />
            </div>

            <button 
              type="submit" 
              className="btn btn-primary submit-btn"
              disabled={loading || !token || !!success}
            >
              {loading ? (
                <>
                  <span className="loading-spinner"></span>
                  <span style={{ marginLeft: '0.5rem' }}>Resetting Password...</span>
                </>
              ) : (
                'Reset Password'
              )}
            </button>
          </form>

          <div className="auth-footer">
            Remember your password?{' '}
            <Link to="/login" className="auth-link">
              Sign In
            </Link>
          </div>
        </div>
      </div>
    </>
  );
};

export default ResetPassword;

