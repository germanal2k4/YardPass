import { apiClient } from './client';
import { API_ENDPOINTS } from '@/shared/config/constants';
import type {
  LoginRequest,
  LoginResponse,
  MeResponse,
  PurchaseSubscriptionRequest,
  PurchaseSubscriptionResponse,
} from '@/shared/types/api';

export const authApi = {
  login: async (credentials: LoginRequest): Promise<LoginResponse> => {
    const response = await apiClient.post<LoginResponse>(API_ENDPOINTS.LOGIN, credentials);
    return response.data;
  },

  getMe: async (): Promise<MeResponse> => {
    const response = await apiClient.get<MeResponse>(API_ENDPOINTS.ME);
    return response.data;
  },

  purchaseSubscription: async (
    payload: PurchaseSubscriptionRequest
  ): Promise<PurchaseSubscriptionResponse> => {
    const response = await apiClient.post<PurchaseSubscriptionResponse>(
      API_ENDPOINTS.PURCHASE_SUBSCRIPTION,
      payload
    );
    return response.data;
  },

  logout: async (): Promise<void> => {
    await apiClient.post(API_ENDPOINTS.LOGOUT, {});
    // Keep local cleanup for backward compatibility with old flows.
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
  },
};
