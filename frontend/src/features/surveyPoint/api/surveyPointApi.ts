import axios from '@/core/api/axios';
import type {
  ApiResponse,
  SurveyPoint,
  CreateSurveyPointRequest,
  UpdateSurveyPointRequest,
  UpdateSurveyPointStatusRequest,
  SurveyPointListItem,
  SuccessResponse,
  PaginationParams,
  SurveyPointStatus,
} from '@/core/types';

export const surveyPointApi = {
  // Create new survey point
  createSurveyPoint: async (data: CreateSurveyPointRequest) => {
    const response = await axios.post<ApiResponse<SurveyPoint>>('/survey-points', data);
    return response.data;
  },

  // Get survey point by ID
  getSurveyPointById: async (surveyPointId: string) => {
    const response = await axios.get<ApiResponse<SurveyPoint>>(`/survey-points/${surveyPointId}`);
    return response.data;
  },

  // Update survey point
  updateSurveyPoint: async (surveyPointId: string, data: UpdateSurveyPointRequest) => {
    const response = await axios.put<ApiResponse<SurveyPoint>>(
      `/survey-points/${surveyPointId}`,
      data
    );
    return response.data;
  },

  // Update survey point status
  updateSurveyPointStatus: async (
    surveyPointId: string,
    data: UpdateSurveyPointStatusRequest
  ) => {
    const response = await axios.put<ApiResponse<SuccessResponse>>(
      `/survey-points/${surveyPointId}/status`,
      data
    );
    return response.data;
  },

  // Delete survey point
  deleteSurveyPoint: async (surveyPointId: string) => {
    const response = await axios.delete<{ message: string }>(`/survey-points/${surveyPointId}`);
    return response.data;
  },

  // Get survey points by MCU ID
  getSurveyPointsByMCU: async (mcuId: string) => {
    const response = await axios.get<ApiResponse<SurveyPointListItem[]>>(
      `/survey-points/mcu/${mcuId}`
    );
    return response.data;
  },

  // List survey points with optional MCU filter
  listSurveyPoints: async (mcuId?: string) => {
    const response = await axios.get<ApiResponse<SurveyPoint[]>>('/survey-points/list', {
      params: mcuId ? { mcu_id: mcuId } : undefined,
    });
    return response.data;
  },

  // Get survey points by status with pagination
  getSurveyPointsByStatus: async (
    status: SurveyPointStatus,
    params?: PaginationParams
  ) => {
    const response = await axios.get<ApiResponse<SurveyPoint[]>>('/survey-points/status', {
      params: { status, ...params },
    });
    return response.data;
  },
};
