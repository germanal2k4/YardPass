import { apiClient } from './client';
import { API_ENDPOINTS } from '@/shared/config/constants';
import type { ListUsersResponse, RegisterUserRequest, UpdateUserCredentialsRequest, User } from '@/shared/types/api';

export const usersApi = {
  register: async (data: RegisterUserRequest): Promise<User> => {
    const response = await apiClient.post<User>(API_ENDPOINTS.USERS, data);
    return response.data;
  },

  listGuards: async (): Promise<ListUsersResponse> => {
    const response = await apiClient.get<ListUsersResponse>(`${API_ENDPOINTS.USERS}?role=guard`);
    return response.data;
  },

  updateCredentials: async (userId: number, data: UpdateUserCredentialsRequest): Promise<User> => {
    const response = await apiClient.put<User>(API_ENDPOINTS.USER_CREDENTIALS(userId), data);
    return response.data;
  },
};
