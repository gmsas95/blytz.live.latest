import { ApiResponse, PaginatedResponse } from '@/types';

// Base API configuration
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8092';
const API_TIMEOUT = 10000; // 10 seconds

// Generic API response wrapper from backend
interface BackendResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

// HTTP client for backend API communication
class ApiClient {
  private baseUrl: string;
  private defaultHeaders: Record<string, string>;

  constructor(baseUrl: string = API_BASE_URL) {
    this.baseUrl = baseUrl;
    this.defaultHeaders = {
      'Content-Type': 'application/json',
    };
  }

  private getStoredToken(): string | null {
    if (typeof window === 'undefined') return null;

    // Try localStorage first
    const localStorageToken = localStorage.getItem('blytz_auth_token');
    if (localStorageToken) return localStorageToken;

    // Try cookies
    const cookieToken = this.getCookie('blytz_auth_token');
    if (cookieToken) return cookieToken;

    return null;
  }

  private getCookie(name: string): string | null {
    if (typeof window === 'undefined') return null;

    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);

    if (parts.length === 2) {
      return parts.pop()?.split(';').shift() || null;
    }

    return null;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    try {
      const token = this.getStoredToken();
      const headers: HeadersInit = {
        ...this.defaultHeaders,
        ...options.headers,
      };

      // Add auth header if token is available
      if (token) {
        headers.Authorization = `Bearer ${token}`;
      }

      // Add correlation ID for tracking
      headers['X-Correlation-ID'] = this.generateCorrelationId();

      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), API_TIMEOUT);

      const response = await fetch(`${this.baseUrl}${endpoint}`, {
        ...options,
        headers,
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      if (!response.ok) {
        let errorMessage = `HTTP ${response.status}: ${response.statusText}`;

        try {
          const errorData = await response.json() as BackendResponse;
          if (errorData.error) {
            errorMessage = errorData.error;
          } else if (errorData.message) {
            errorMessage = errorData.message;
          }
        } catch {
          // Use default error message if response is not JSON
        }

        return { success: false, error: errorMessage };
      }

      const data = await response.json() as BackendResponse<T>;
      
      if (data.success && data.data !== undefined) {
        return { success: true, data: data.data };
      } else if (data.success) {
        return { success: true, data: data as T };
      } else {
        return { success: false, error: data.error || 'Request failed' };
      }
    } catch (error) {
      if (error instanceof Error) {
        if (error.name === 'AbortError') {
          return { success: false, error: 'Request timeout' };
        }
        return { success: false, error: error.message };
      }
      return { success: false, error: 'Network error occurred' };
    }
  }

  private generateCorrelationId(): string {
    return `corr_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  // HTTP method helpers
  async get<T>(endpoint: string): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { method: 'GET' });
  }

  async post<T>(endpoint: string, data?: any): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  async put<T>(endpoint: string, data?: any): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  async delete<T>(endpoint: string): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { method: 'DELETE' });
  }

  // Utility method for file uploads
  async upload<T>(endpoint: string, file: File, additionalData?: Record<string, any>): Promise<ApiResponse<T>> {
    const formData = new FormData();
    formData.append('file', file);

    if (additionalData) {
      Object.entries(additionalData).forEach(([key, value]) => {
        formData.append(key, String(value));
      });
    }

    const token = this.getStoredToken();
    const headers: HeadersInit = {};
    
    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }

    try {
      const response = await fetch(`${this.baseUrl}${endpoint}`, {
        method: 'POST',
        body: formData,
        headers,
      });

      if (!response.ok) {
        const errorData = await response.json() as BackendResponse;
        return { success: false, error: errorData.error || 'Upload failed' };
      }

      const data = await response.json() as BackendResponse<T>;
      
      if (data.success && data.data !== undefined) {
        return { success: true, data: data.data };
      } else {
        return { success: false, error: data.error || 'Upload failed' };
      }
    } catch (error) {
      return { success: false, error: error instanceof Error ? error.message : 'Upload error' };
    }
  }
}

// Create singleton instance
const apiClient = new ApiClient();

export default apiClient;