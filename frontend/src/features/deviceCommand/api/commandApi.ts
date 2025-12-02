import apiClient from '@/core/api/axios';
import type { 
  CommandInfo, 
  CommandOperationResult, 
  CreateCommandRequest,
  ApiResponse 
} from '@/core/types';

export const commandApi = {
  /**
   * Get command history with filters
   */
  getCommandHistory: async (params?: {
    survey_point_id?: string;
    device_name?: string;
    limit?: number;
  }): Promise<ApiResponse<CommandInfo[]>> => {
    const response = await apiClient.get('/commands/history', { params });
    return response.data;
  },

  /**
   * Get pending commands
   */
  getPendingCommands: async (limit: number = 100): Promise<ApiResponse<CommandInfo[]>> => {
    const response = await apiClient.get('/commands/pending', {
      params: { limit }
    });
    return response.data;
  },

  /**
   * Get command by ID
   */
  getCommandById: async (commandId: string): Promise<ApiResponse<CommandInfo>> => {
    const response = await apiClient.get(`/commands/${commandId}`);
    return response.data;
  },

  /**
   * Create a new command
   */
  createCommand: async (request: CreateCommandRequest): Promise<ApiResponse<CommandOperationResult>> => {
    const response = await apiClient.post('/commands', request);
    return response.data;
  },

  /**
   * Update command status (usually called by MCU)
   */
  updateCommandStatus: async (
    commandId: string,
    status: 'pending' | 'sent' | 'success' | 'failed'
  ): Promise<ApiResponse<CommandOperationResult>> => {
    const response = await apiClient.put(`/commands/${commandId}/status`, { status });
    return response.data;
  }
};
