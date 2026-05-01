import { apiClient } from './client';
import { API_ENDPOINTS } from '@/shared/config/constants';
import type { Building, UpdateApartmentCountRequest } from '@/shared/types/api';

export const buildingsApi = {
  getById: async (buildingId: number): Promise<Building> => {
    const response = await apiClient.get<Building>(API_ENDPOINTS.BUILDING_BY_ID(buildingId));
    return response.data;
  },

  updateApartmentCount: async (buildingId: number, data: UpdateApartmentCountRequest): Promise<Building> => {
    const response = await apiClient.put<Building>(
      API_ENDPOINTS.BUILDING_APARTMENT_COUNT(buildingId),
      data
    );
    return response.data;
  },
};

