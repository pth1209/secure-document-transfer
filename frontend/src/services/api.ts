import axios from 'axios';
import type { SignUpRequest, SignUpResponse, SignInRequest, SignInResponse } from '../types/auth';

const api = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add token to requests if available
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export const authService = {
  signup: async (data: SignUpRequest): Promise<SignUpResponse> => {
    const response = await api.post<SignUpResponse>('/signup', data);
    return response.data;
  },

  signin: async (data: SignInRequest): Promise<SignInResponse> => {
    const response = await api.post<SignInResponse>('/signin', data);
    return response.data;
  },

  signout: async (): Promise<void> => {
    await api.post('/signout');
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
  },
};

export default api;

