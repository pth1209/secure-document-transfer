export interface User {
  id: string;
  email: string;
  full_name: string;
}

export interface SignUpRequest {
  email: string;
  password: string;
  full_name: string;
}

export interface SignUpResponse {
  message: string;
  access_token?: string;
  user: User;
}

export interface SignInRequest {
  email: string;
  password: string;
}

export interface SignInResponse {
  message: string;
  access_token: string;
  refresh_token: string;
  user: User;
}

export interface ErrorResponse {
  error: string;
  details?: string;
}

