import { apiClient } from './client';
import { API_ENDPOINTS } from '@/shared/config/constants';
import type { LoginRequest, LoginResponse, MeResponse } from '@/shared/types/api';

export const authApi = {
  login: async (credentials: LoginRequest): Promise<LoginResponse> => {
    const response = await apiClient.post<LoginResponse>(API_ENDPOINTS.LOGIN, credentials);
    return response.data;
  },

  getMe: async (): Promise<MeResponse> => {
    const response = await apiClient.get<MeResponse>(API_ENDPOINTS.ME);
    return response.data;
  },

  logout: () => {
    // Clear tokens from storage
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
  },
};

