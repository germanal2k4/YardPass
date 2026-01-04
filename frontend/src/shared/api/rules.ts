import { apiClient } from './client';
import { API_ENDPOINTS } from '@/shared/config/constants';
import type { Rule, UpdateRuleRequest } from '@/shared/types/api';

export const rulesApi = {
  get: async (buildingId: number): Promise<Rule> => {
    const response = await apiClient.get<Rule>(
      `${API_ENDPOINTS.RULES}?building_id=${buildingId}`
    );
    return response.data;
  },

  update: async (buildingId: number, data: UpdateRuleRequest): Promise<Rule> => {
    const response = await apiClient.put<Rule>(
      `${API_ENDPOINTS.RULES}?building_id=${buildingId}`,
      data
    );
    return response.data;
  },
};

