import { apiClient } from './client';
import { API_ENDPOINTS } from '@/shared/config/constants';
import type { ListUsersResponse, RegisterUserRequest, User } from '@/shared/types/api';

export const usersApi = {
  register: async (data: RegisterUserRequest): Promise<User> => {
    const response = await apiClient.post<User>(API_ENDPOINTS.USERS, data);
    return response.data;
  },

  listGuards: async (): Promise<ListUsersResponse> => {
    const response = await apiClient.get<ListUsersResponse>(`${API_ENDPOINTS.USERS}?role=guard`);
    return response.data;
  },
};
