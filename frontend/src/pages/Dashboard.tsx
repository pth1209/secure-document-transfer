import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { authService } from '../services/api';
import type { User } from '../types/auth';

const Dashboard: React.FC = () => {
  const navigate = useNavigate();
  const [user, setUser] = useState<User | null>(null);

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

  if (!user) {
    return null;
  }

  return (
    <div className="auth-container">
      <div className="auth-card">
        <div className="auth-header">
          <h1 className="auth-title">Welcome, {user.full_name}!</h1>
          <p className="auth-subtitle">{user.email}</p>
        </div>

        <div style={{ textAlign: 'center', padding: '2rem 0' }}>
          <p style={{ color: 'var(--text-muted)', marginBottom: '2rem' }}>
            Your secure document transfer dashboard will be here.
          </p>
          
          <button 
            onClick={handleSignOut}
            className="btn btn-secondary"
          >
            Sign Out
          </button>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;

