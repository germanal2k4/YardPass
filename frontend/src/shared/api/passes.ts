import { apiClient } from './client';
import { API_ENDPOINTS } from '@/shared/config/constants';
import type {
  ValidatePassRequest,
  ValidatePassResponse,
  GetActivePassesResponse,
  RevokePassResponse,
} from '@/shared/types/api';

export const passesApi = {
  validate: async (params: { qr_uuid?: string; car_plate?: string }): Promise<ValidatePassResponse> => {
    const response = await apiClient.post<ValidatePassResponse>(
      API_ENDPOINTS.VALIDATE_PASS,
      params as ValidatePassRequest
    );
    return response.data;
  },

  getActive: async (apartmentId: number): Promise<GetActivePassesResponse> => {
    const response = await apiClient.get<GetActivePassesResponse>(
      `${API_ENDPOINTS.ACTIVE_PASSES}?apartment_id=${apartmentId}`
    );
    return response.data;
  },

  revoke: async (passId: string): Promise<RevokePassResponse> => {
    const response = await apiClient.post<RevokePassResponse>(
      API_ENDPOINTS.REVOKE_PASS(passId)
    );
    return response.data;
  },
};

