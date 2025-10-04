import React from 'react';
import { Link } from 'react-router-dom';

const Home: React.FC = () => {
  return (
    <div className="home">
      {/* Navigation */}
      <nav className="nav">
        <div className="nav-container">
          <Link to="/" className="logo">
            SecureTransfer
          </Link>
          <div className="nav-links">
            <Link to="/login" className="btn btn-secondary">
              Sign In
            </Link>
            <Link to="/signup" className="btn btn-primary">
              Get Started
            </Link>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <section className="hero">
        <div className="hero-content">
          <div className="hero-badge">
            🔒 Enterprise-grade security
          </div>
          <h1 className="hero-title">
            Share Documents
            <br />
            <span className="gradient-text">Securely & Instantly</span>
          </h1>
          <p className="hero-subtitle">
            Transfer sensitive files with end-to-end encryption, 
            secure authentication, and complete peace of mind. 
            Built for teams that value privacy.
          </p>
          <div className="hero-cta">
            <Link to="/signup" className="btn btn-primary btn-large">
              Start Transferring Free
            </Link>
            <Link to="/login" className="btn btn-secondary btn-large">
              Sign In
            </Link>
          </div>
          <div className="hero-stats">
            <div className="stat">
              <div className="stat-value">256-bit</div>
              <div className="stat-label">Encryption</div>
            </div>
            <div className="stat">
              <div className="stat-value">100%</div>
              <div className="stat-label">Secure</div>
            </div>
            <div className="stat">
              <div className="stat-value">Instant</div>
              <div className="stat-label">Transfer</div>
            </div>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="features">
        <div className="features-container">
          <div className="section-header">
            <h2 className="section-title">Built for Security</h2>
            <p className="section-subtitle">
              Enterprise-grade features that keep your documents safe
            </p>
          </div>

          <div className="features-grid">
            <div className="feature-card">
              <div className="feature-icon">🔐</div>
              <h3 className="feature-title">End-to-End Encryption</h3>
              <p className="feature-description">
                Your files are encrypted before leaving your device. 
                Nobody can access them without your permission.
              </p>
            </div>

            <div className="feature-card">
              <div className="feature-icon">⚡</div>
              <h3 className="feature-title">Lightning Fast</h3>
              <p className="feature-description">
                Transfer files of any size instantly. No waiting, 
                no compromises on speed or security.
              </p>
            </div>

            <div className="feature-card">
              <div className="feature-icon">👥</div>
              <h3 className="feature-title">Secure Authentication</h3>
              <p className="feature-description">
                JWT-based authentication with Supabase ensures 
                only authorized users can access your documents.
              </p>
            </div>

            <div className="feature-card">
              <div className="feature-icon">🌐</div>
              <h3 className="feature-title">Access Anywhere</h3>
              <p className="feature-description">
                Share documents from any device, anywhere in the world. 
                Your data is always accessible.
              </p>
            </div>

            <div className="feature-card">
              <div className="feature-icon">🎯</div>
              <h3 className="feature-title">Simple & Intuitive</h3>
              <p className="feature-description">
                Beautiful, modern interface that makes secure file 
                sharing as easy as a few clicks.
              </p>
            </div>

            <div className="feature-card">
              <div className="feature-icon">🛡️</div>
              <h3 className="feature-title">Privacy First</h3>
              <p className="feature-description">
                We never store your encryption keys. Your privacy 
                is guaranteed by design.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* How It Works Section */}
      <section className="how-it-works">
        <div className="how-it-works-container">
          <div className="section-header">
            <h2 className="section-title">How It Works</h2>
            <p className="section-subtitle">
              Secure file transfer in three simple steps
            </p>
          </div>

          <div className="steps">
            <div className="step">
              <div className="step-number">1</div>
              <h3 className="step-title">Create Your Account</h3>
              <p className="step-description">
                Sign up in seconds with just your email. 
                No credit card required.
              </p>
            </div>

            <div className="step-connector"></div>

            <div className="step">
              <div className="step-number">2</div>
              <h3 className="step-title">Upload Your Files</h3>
              <p className="step-description">
                Drag and drop your documents. They're encrypted 
                automatically before upload.
              </p>
            </div>

            <div className="step-connector"></div>

            <div className="step">
              <div className="step-number">3</div>
              <h3 className="step-title">Share Securely</h3>
              <p className="step-description">
                Generate secure links and share directly with team members. 
                Full control over access.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="cta-section">
        <div className="cta-content">
          <h2 className="cta-title">Ready to secure your documents?</h2>
          <p className="cta-subtitle">
            Join teams that trust SecureTransfer for their most sensitive files.
          </p>
          <Link to="/signup" className="btn btn-primary btn-large">
            Get Started Free
          </Link>
        </div>
      </section>

      {/* Footer */}
      <footer className="footer">
        <div className="footer-content">
          <div className="footer-brand">
            <div className="logo">SecureTransfer</div>
            <p className="footer-tagline">Secure document transfer made simple</p>
          </div>
          <div className="footer-links">
            <div className="footer-column">
              <h4 className="footer-heading">Product</h4>
              <a href="#features" className="footer-link">Features</a>
              <a href="#security" className="footer-link">Security</a>
              <a href="#pricing" className="footer-link">Pricing</a>
            </div>
            <div className="footer-column">
              <h4 className="footer-heading">Company</h4>
              <a href="#about" className="footer-link">About</a>
              <a href="#contact" className="footer-link">Contact</a>
              <a href="#careers" className="footer-link">Careers</a>
            </div>
            <div className="footer-column">
              <h4 className="footer-heading">Legal</h4>
              <a href="#privacy" className="footer-link">Privacy</a>
              <a href="#terms" className="footer-link">Terms</a>
              <a href="#security" className="footer-link">Security</a>
            </div>
          </div>
        </div>
        <div className="footer-bottom">
          <p>&copy; 2024 SecureTransfer. Built with security in mind.</p>
        </div>
      </footer>
    </div>
  );
};

export default Home;

