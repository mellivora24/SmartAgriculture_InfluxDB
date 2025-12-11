import axios from 'axios';

const API_BASE_URL = 'http://localhost:8080/api/v1';

export interface ThresholdSettings {
  id: string;
  survey_point_id: string;
  
  // Temperature thresholds
  temp_min?: number;
  temp_max?: number;
  temp_critical_min?: number;
  temp_critical_max?: number;
  
  // Humidity thresholds
  humidity_min?: number;
  humidity_max?: number;
  humidity_critical_min?: number;
  humidity_critical_max?: number;
  
  // Soil moisture thresholds
  soil_moisture_min?: number;
  soil_moisture_max?: number;
  soil_moisture_critical_min?: number;
  soil_moisture_critical_max?: number;
  
  // Light thresholds
  light_min?: number;
  light_max?: number;
  light_critical_min?: number;
  light_critical_max?: number;
  
  // Auto pump settings
  auto_pump_enabled: boolean;
  pump_trigger_soil_moisture?: number;
  pump_stop_soil_moisture?: number;
  pump_duration_seconds: number;
  pump_cooldown_minutes: number;
  
  // Alert settings
  alert_enabled: boolean;
  alert_cooldown_minutes: number;
  
  created_at: string;
  updated_at: string;
}

export interface UpdateThresholdRequest {
  temp_min?: number;
  temp_max?: number;
  temp_critical_min?: number;
  temp_critical_max?: number;
  
  humidity_min?: number;
  humidity_max?: number;
  humidity_critical_min?: number;
  humidity_critical_max?: number;
  
  soil_moisture_min?: number;
  soil_moisture_max?: number;
  soil_moisture_critical_min?: number;
  soil_moisture_critical_max?: number;
  
  light_min?: number;
  light_max?: number;
  light_critical_min?: number;
  light_critical_max?: number;
  
  auto_pump_enabled?: boolean;
  pump_trigger_soil_moisture?: number;
  pump_stop_soil_moisture?: number;
  pump_duration_seconds?: number;
  pump_cooldown_minutes?: number;
  
  alert_enabled?: boolean;
  alert_cooldown_minutes?: number;
}

export interface AlertHistory {
  id: string;
  survey_point_id: string;
  alert_type: string;
  severity: string;
  sensor_value: number;
  threshold_value: number;
  message: string;
  acknowledged: boolean;
  acknowledged_at?: string;
  acknowledged_by?: string;
  created_at: string;
}

export interface AutoPumpHistory {
  id: string;
  survey_point_id: string;
  command_id?: string;
  trigger_soil_moisture: number;
  target_soil_moisture: number;
  pump_duration_seconds: number;
  status: string;
  started_at: string;
  completed_at?: string;
  notes?: string;
}

const getAuthHeader = () => {
  const token = localStorage.getItem('token');
  return {
    Authorization: `Bearer ${token}`,
  };
};

export const thresholdApi = {
  // Get threshold settings by survey point
  getSettings: async (surveyPointId: string) => {
    const response = await axios.get<{ data: ThresholdSettings }>(
      `${API_BASE_URL}/thresholds/survey-point/${surveyPointId}`,
      { headers: getAuthHeader() }
    );
    return response.data;
  },

  // Update threshold settings
  updateSettings: async (surveyPointId: string, data: UpdateThresholdRequest) => {
    const response = await axios.put<{ data: ThresholdSettings }>(
      `${API_BASE_URL}/thresholds/survey-point/${surveyPointId}`,
      data,
      { headers: getAuthHeader() }
    );
    return response.data;
  },

  // Get alert history
  getAlertHistory: async (surveyPointId: string, limit: number = 50) => {
    const response = await axios.get<{ data: AlertHistory[] }>(
      `${API_BASE_URL}/thresholds/survey-point/${surveyPointId}/alerts`,
      {
        params: { limit },
        headers: getAuthHeader(),
      }
    );
    return response.data;
  },

  // Acknowledge alert
  acknowledgeAlert: async (alertId: string) => {
    const response = await axios.post<{ data: { acknowledged: boolean } }>(
      `${API_BASE_URL}/thresholds/alerts/${alertId}/acknowledge`,
      {},
      { headers: getAuthHeader() }
    );
    return response.data;
  },

  // Get auto pump history
  getAutoPumpHistory: async (surveyPointId: string, limit: number = 50) => {
    const response = await axios.get<{ data: AutoPumpHistory[] }>(
      `${API_BASE_URL}/thresholds/survey-point/${surveyPointId}/auto-pump-history`,
      {
        params: { limit },
        headers: getAuthHeader(),
      }
    );
    return response.data;
  },
};
