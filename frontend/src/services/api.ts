import axios from 'axios';
import type { SignUpRequest, SignUpResponse, SignInRequest, SignInResponse, User } from '../types/auth';
import type { FileChunk } from '../types/file';

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

  sendFileChunk: async (chunk: FileChunk, recipientIds: string[]): Promise<{ message: string }> => {
    const formData = new FormData();
    formData.append('chunk', chunk.chunk_data);
    formData.append('file_id', chunk.file_id);
    formData.append('chunk_index', chunk.chunk_index.toString());
    formData.append('total_chunks', chunk.total_chunks.toString());
    formData.append('original_filename', chunk.original_filename);
    formData.append('file_size', chunk.file_size.toString());
    formData.append('chunk_size', chunk.chunk_size.toString());
    
    recipientIds.forEach(id => {
      formData.append('recipient_ids[]', id);
    });

    const response = await api.post<{ message: string }>('/files/send-chunk', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  },
};

export default api;

