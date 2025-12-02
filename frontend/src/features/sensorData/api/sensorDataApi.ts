import axios from "@/core/api/axios";
import type {
  ApiResponse,
  SensorDataPoint,
  QuerySensorDataParams,
  SensorAggregationRequest,
} from "@/core/types";

export const sensorDataApi = {
  getSensorData: async (params: QuerySensorDataParams) => {
    const response = await axios.get<ApiResponse<SensorDataPoint[]>>(
      "/sensor-data",
      { params }
    );
    return response.data;
  },
  getLatestData: async (surveyPointId: string, limit: number = 5) => {
    const response = await axios.get<ApiResponse<SensorDataPoint[]>>(
      `/sensor-data/latest/${surveyPointId}`,
      { params: { limit } }
    );
    return response.data;
  },
  getAggregation: async (query: SensorAggregationRequest) => {
    const response = await axios.post<ApiResponse<SensorDataPoint[]>>(
      "/sensor-data/aggregation",
      query
    );
    return response.data;
  },
};
