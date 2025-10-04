import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { authService } from '../services/api';
import type { SignUpRequest } from '../types/auth';

const SignUp: React.FC = () => {
  const navigate = useNavigate();
  const [formData, setFormData] = useState<SignUpRequest>({
    email: '',
    password: '',
    full_name: '',
  });
  const [error, setError] = useState<string>('');
  const [success, setSuccess] = useState<string>('');
  const [loading, setLoading] = useState(false);

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
    setLoading(true);

    try {
      const response = await authService.signup(formData);
      setSuccess(response.message);
      
      // If we get an access token, store it and redirect
      if (response.access_token) {
        localStorage.setItem('access_token', response.access_token);
        setTimeout(() => {
          navigate('/dashboard');
        }, 1500);
      } else {
        // Otherwise, redirect to login after showing success message
        setTimeout(() => {
          navigate('/login');
        }, 2000);
      }
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || 'Failed to create account. Please try again.';
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
          <h1 className="auth-title">Create Account</h1>
          <p className="auth-subtitle">Start transferring documents securely</p>
        </div>

        {error && <div className="error-message">{error}</div>}
        {success && <div className="success-message">{success}</div>}

        <form className="form" onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="full_name" className="form-label">
              Full Name
            </label>
            <input
              type="text"
              id="full_name"
              name="full_name"
              className="form-input"
              placeholder="John Doe"
              value={formData.full_name}
              onChange={handleChange}
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="email" className="form-label">
              Email Address
            </label>
            <input
              type="email"
              id="email"
              name="email"
              className="form-input"
              placeholder="you@example.com"
              value={formData.email}
              onChange={handleChange}
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="password" className="form-label">
              Password
            </label>
            <input
              type="password"
              id="password"
              name="password"
              className="form-input"
              placeholder="At least 8 characters"
              value={formData.password}
              onChange={handleChange}
              required
              minLength={8}
            />
          </div>

          <button 
            type="submit" 
            className="btn btn-primary submit-btn"
            disabled={loading}
          >
            {loading ? (
              <>
                <span className="loading-spinner"></span>
                <span style={{ marginLeft: '0.5rem' }}>Creating Account...</span>
              </>
            ) : (
              'Create Account'
            )}
          </button>
        </form>

        <div className="auth-footer">
          Already have an account?{' '}
          <Link to="/login" className="auth-link">
            Sign In
          </Link>
        </div>
      </div>
    </div>
    </>
  );
};

export default SignUp;

