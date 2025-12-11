import apiClient from './api-client';
import { User } from '@/types';

// Auth-related API calls
export class AuthService {
  // Login user
  async login(email: string, password: string) {
    return apiClient.post<{ token: string; user: User }>('/api/v1/auth/login', {
      email,
      password,
    });
  }

  // Register new user
  async register(userData: {
    name: string;
    email: string;
    password: string;
    displayName?: string;
    phoneNumber?: string;
  }) {
    return apiClient.post<{ user: User }>('/api/v1/auth/register', userData);
  }

  // Get current user profile
  async getCurrentUser() {
    return apiClient.get<User>('/api/v1/auth/me');
  }

  // Update user profile
  async updateProfile(userData: {
    displayName?: string;
    phoneNumber?: string;
    avatarUrl?: string;
  }) {
    return apiClient.put<User>('/api/v1/auth/profile', userData);
  }

  // Refresh token
  async refreshToken(refreshToken: string) {
    return apiClient.post<{ token: string; refreshToken: string }>('/api/v1/auth/refresh', {
      refreshToken,
    });
  }

  // Logout user
  async logout() {
    return apiClient.post<void>('/api/v1/auth/logout');
  }

  // Verify token
  async verifyToken(token: string) {
    return apiClient.post<{ valid: boolean; userId?: string }>('/api/v1/auth/verify', {
      token,
    });
  }
}

// Export singleton instance
export const authService = new AuthService();