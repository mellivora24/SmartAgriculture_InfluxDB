import axios from '@/core/api/axios';
import type {
  ApiResponse,
  Farm,
  MyFarm,
  CreateFarmRequest,
  CreateFarmResponse,
  UpdateFarmRequest,
  FarmOverview,
  FarmStructure,
  AddUserToFarmRequest,
  SuccessResponse,
} from '@/core/types';

export const farmApi = {
  // Get my farms (farms where user is owner/member)
  getMyFarms: async () => {
    const response = await axios.get<ApiResponse<MyFarm[]>>('/farms/my-farms');
    return response.data;
  },

  // Get all farms (admin only)
  getAllFarms: async () => {
    const response = await axios.get<ApiResponse<Farm[]>>('/farms');
    return response.data;
  },

  // Get farm by ID
  getFarmById: async (farmId: string) => {
    const response = await axios.get<ApiResponse<Farm>>(`/farms/${farmId}`);
    return response.data;
  },

  // Create new farm
  createFarm: async (data: CreateFarmRequest) => {
    const response = await axios.post<ApiResponse<CreateFarmResponse>>('/farms', data);
    return response.data;
  },

  // Update farm
  updateFarm: async (farmId: string, data: UpdateFarmRequest) => {
    const response = await axios.put<ApiResponse<Farm>>(`/farms/${farmId}`, data);
    return response.data;
  },

  // Get farm overview (statistics)
  getFarmOverview: async (farmId: string) => {
    const response = await axios.get<ApiResponse<FarmOverview>>(`/farms/${farmId}/overview`);
    return response.data;
  },

  // Get farm structure (farms with MCUs and survey points)
  getFarmStructure: async (farmId: string) => {
    const response = await axios.get<ApiResponse<FarmStructure[]>>(`/farms/${farmId}/structure`);
    return response.data;
  },

  // Add user to farm
  addUserToFarm: async (farmId: string, data: AddUserToFarmRequest) => {
    const response = await axios.post<ApiResponse<SuccessResponse>>(
      `/farms/${farmId}/users`,
      data
    );
    return response.data;
  },

  // Remove user from farm
  removeUserFromFarm: async (farmId: string, userId: string) => {
    const response = await axios.delete<ApiResponse<SuccessResponse>>(
      `/farms/${farmId}/users/${userId}`
    );
    return response.data;
  },
};
