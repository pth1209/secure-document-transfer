import axios from 'axios';
import type { SignUpRequest, SignUpResponse, SignInRequest, SignInResponse, User } from '../types/auth';

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

export const userService = {
  searchUsers: async (query: string): Promise<User[]> => {
    const response = await api.get<User[]>('/users/search', {
      params: { q: query }
    });
    return response.data;
  },

  sendFiles: async (formData: FormData): Promise<{ message: string }> => {
    const response = await api.post<{ message: string }>('/files/send', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  },
};

export default api;

