# Secure Document Transfer - Frontend

Modern React + TypeScript frontend for secure document transfer application.

## Features

- 🎨 Beautiful, modern UI inspired by Paraform and Krew
- 🔐 Secure authentication (signup/login)
- ⚡ Fast and responsive
- 🎭 Smooth animations and transitions
- 📱 Mobile-friendly design

## Tech Stack

- React 18
- TypeScript
- Vite
- React Router
- Axios
- CSS3 with custom properties

## Getting Started

### Prerequisites

- Node.js 18+ and npm

### Installation

```bash
cd frontend
npm install
```

### Development

```bash
npm run dev
```

The app will run on [http://localhost:3000](http://localhost:3000)

### Build

```bash
npm run build
```

## Environment Setup

Make sure your backend is running on `http://localhost:8080` for the API proxy to work correctly.

## Project Structure

```
frontend/
├── src/
│   ├── components/      # Reusable components
│   ├── pages/          # Page components (SignUp, Login, Dashboard)
│   ├── services/       # API services
│   ├── types/          # TypeScript types
│   ├── App.tsx         # Main app component
│   ├── App.css         # Global styles
│   └── main.tsx        # Entry point
├── index.html
├── package.json
└── vite.config.ts
```

## API Integration

The frontend communicates with the backend API:

- `POST /api/signup` - User registration
- `POST /api/signin` - User login
- `POST /api/signout` - User logout

Authentication tokens are stored in localStorage and automatically included in API requests.

