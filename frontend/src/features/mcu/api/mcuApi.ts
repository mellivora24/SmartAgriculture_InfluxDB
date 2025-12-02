import axios from '@/core/api/axios';
import type {
  ApiResponse,
  MCU,
  CreateMCURequest,
  UpdateMCUStatusRequest,
  MCUWithStats,
  SuccessResponse,
} from '@/core/types';

export const mcuApi = {
  // Create new MCU
  createMCU: async (data: CreateMCURequest) => {
    const response = await axios.post<ApiResponse<MCU>>('/mcus', data);
    return response.data;
  },

  // Get MCU by ID
  getMCUById: async (mcuId: string) => {
    const response = await axios.get<ApiResponse<MCU>>(`/mcus/${mcuId}`);
    return response.data;
  },

  // Update MCU status by code
  updateMCUStatus: async (mcuCode: string, data: UpdateMCUStatusRequest) => {
    const response = await axios.put<ApiResponse<SuccessResponse>>(
      `/mcus/code/${mcuCode}/status`,
      data
    );
    return response.data;
  },

  // Delete MCU
  deleteMCU: async (mcuId: string) => {
    const response = await axios.delete<{ message: string }>(`/mcus/${mcuId}`);
    return response.data;
  },

  // Get MCUs by farm ID
  getMCUsByFarm: async (farmId: string) => {
    const response = await axios.get<ApiResponse<MCUWithStats[]>>(`/mcus/farm/${farmId}`);
    return response.data;
  },

  // Get MCUs by status with pagination
  getMCUsByStatus: async (status: 'online' | 'offline', limit = 10, offset = 0) => {
    const response = await axios.get<ApiResponse<MCU[]>>('/mcus/status', {
      params: { status, limit, offset },
    });
    return response.data;
  },
};
