import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { authService } from '../services/api';
import type { SignInRequest } from '../types/auth';

const Login: React.FC = () => {
  const navigate = useNavigate();
  const [formData, setFormData] = useState<SignInRequest>({
    email: '',
    password: '',
  });
  const [error, setError] = useState<string>('');
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
    setLoading(true);

    try {
      const response = await authService.signin(formData);
      
      // Store tokens
      localStorage.setItem('access_token', response.access_token);
      localStorage.setItem('refresh_token', response.refresh_token);
      
      // Store user info
      localStorage.setItem('user', JSON.stringify(response.user));
      
      // Redirect to dashboard
      navigate('/dashboard');
    } catch (err: any) {
      const errorMessage = err.response?.data?.error || 'Invalid credentials. Please try again.';
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
            <Link to="/signup" className="btn btn-primary">
              Sign Up
            </Link>
          </div>
        </div>
      </nav>
      <div className="auth-container">
      <div className="auth-card">
        <div className="auth-header">
          <h1 className="auth-title">Welcome Back</h1>
          <p className="auth-subtitle">Sign in to your account</p>
        </div>

        {error && <div className="error-message">{error}</div>}

        <form className="form" onSubmit={handleSubmit}>
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
              placeholder="Enter your password"
              value={formData.password}
              onChange={handleChange}
              required
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
                <span style={{ marginLeft: '0.5rem' }}>Signing In...</span>
              </>
            ) : (
              'Sign In'
            )}
          </button>
        </form>

        <div className="auth-footer">
          Don't have an account?{' '}
          <Link to="/signup" className="auth-link">
            Create Account
          </Link>
        </div>
      </div>
    </div>
    </>
  );
};

export default Login;

