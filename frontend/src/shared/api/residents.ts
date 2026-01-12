import { apiClient } from './client';
import { API_ENDPOINTS } from '@/shared/config/constants';
import type { 
  Resident, 
  CreateResidentRequest, 
  GetResidentsRequest,
  GetResidentsResponse 
} from '@/shared/types/api';

export const residentsApi = {
  /**
   * Получить список резидентов с фильтрами
   */
  async getAll(params?: GetResidentsRequest): Promise<Resident[]> {
    const response = await apiClient.get<GetResidentsResponse>(API_ENDPOINTS.RESIDENTS, {
      params,
    });
    return response.data.residents;
  },

  /**
   * Создать нового резидента
   */
  async create(data: CreateResidentRequest): Promise<Resident> {
    const response = await apiClient.post<Resident>(API_ENDPOINTS.RESIDENTS, data);
    return response.data;
  },

  /**
   * Массовое создание резидентов
   */
  async createBulk(data: CreateResidentRequest[]): Promise<{
    created: number;
    residents: Resident[];
    errors: any[];
  }> {
    const response = await apiClient.post(API_ENDPOINTS.RESIDENTS_BULK, data);
    return response.data;
  },

  /**
   * Импорт резидентов из CSV файла
   */
  async importFromCSV(file: File, buildingId: number): Promise<{
    imported: number;
    errors: any[];
  }> {
    const formData = new FormData();
    formData.append('file', file);
    
    const response = await apiClient.post(API_ENDPOINTS.RESIDENTS_IMPORT, formData, {
      params: {
        building_id: buildingId,
      },
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  },

  /**
   * Удалить резидента
   */
  async delete(id: number): Promise<{ message: string }> {
    const response = await apiClient.delete<{ message: string }>(`${API_ENDPOINTS.RESIDENTS}/${id}`);
    return response.data;
  },
};

