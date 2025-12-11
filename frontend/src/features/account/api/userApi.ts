import apiClient from '@/core/api/axios';
import type { User, UpdateUserRequest } from '@/core/types/user';
import type { ApiResponse } from '@/core/types/common';

export const userApi = {
  getUserById: async (userId: string): Promise<ApiResponse<User>> => {
    const response = await apiClient.get(`/users/${userId}`);
    return response.data;
  },

  getCurrentUser: async (): Promise<ApiResponse<User>> => {
    const token = localStorage.getItem('token');
    if (!token) {
      throw new Error('No token found');
    }
    
    // Parse JWT payload to get user_id
    const payload = JSON.parse(atob(token.split('.')[1]));
    const userId = payload.user_id;
    
    if (!userId) {
      throw new Error('Invalid token: missing user_id');
    }
    
    return userApi.getUserById(userId);
  },

  updateUser: async (userId: string, data: UpdateUserRequest): Promise<ApiResponse<User>> => {
    const response = await apiClient.put(`/users/${userId}`, data);
    return response.data;
  },

  deleteUser: async (userId: string): Promise<ApiResponse<{ message: string }>> => {
    const response = await apiClient.delete(`/users/${userId}`);
    return response.data;
  }
};
