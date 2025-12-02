export * from './common';
export * from './user';
export * from './farm';
export * from './mcu';
export * from './surveyPoint';
export * from './sensorData';
export * from './command';
export * from './realtime';

// WebSocket Message Types
export interface WebSocketMessage<T = unknown> {
  topic: string;
  payload: T;
  timestamp?: string;
}

// Sensor Data Types
export interface SensorDataPayload {
  mcu_code: string;
  survey_point_id: string;
  temperature?: number | null;
  humidity?: number | null;
  soil_moisture?: number | null;
  light?: number | null;
  extra?: Record<string, unknown>;
}

export interface SensorDataPoint {
  _time: string;
  _field: string;
  _value: number;
  survey_point_id: string;
  survey_point_name: string;
  mcu_code: string;
  farm_name?: string;
}

// Control Request/Response Types
export interface ControlRequestPayload {
  survey_point_id: string; // IMPORTANT: Required field
  mcu_code: string;
  device_name: string;
  command: 'on' | 'off';
  value?: unknown;
  extra?: Record<string, unknown>;
}

export interface ControlResponsePayload {
  survey_point_id: string; // IMPORTANT: Required field
  mcu_code: string;
  device_name: string;
  command: string;
  status: 'success' | 'failed';
  message?: string;
  value?: unknown;
  executed_at: string;
}

// Device Types
export interface DeviceConfig {
  id: string;
  name: string;
  device_type: string;
  mcu_code: string;
  is_active: boolean;
}

// Alert Types
export interface MQTTAlert {
  mcu_code: string;
  title: string;
  message: string;
  severity: 'info' | 'warning' | 'error' | 'critical';
  time: string;
}

// Error Types
export interface WSErrorPayload {
  code: string;
  message: string;
}

// Command History Types
export interface CommandInfo {
  command_id: string;
  survey_point_id: string;
  survey_point_name: string;
  device_name: string;
  command: string;
  status: 'pending' | 'sent' | 'success' | 'failed';
  executed_at?: string;
  created_at: string;
}

export interface CreateCommandRequest {
  survey_point_id: string;
  device_name: string;
  command: 'on' | 'off';
}

export interface CommandOperationResult {
  success: boolean;
  command_id?: string;
  message: string;
}

// Query Types
export interface QuerySensorDataRequest {
  survey_point_id?: string;
  mcu_code?: string;
  farm_id?: string;
  start_time?: string;
  end_time?: string;
  limit?: number;
}

export interface AggregationRequest {
  survey_point_id: string;
  field: 'temperature' | 'humidity' | 'soil_moisture' | 'light';
  aggregation: 'mean' | 'min' | 'max' | 'sum' | 'count';
  window: string;
  start_time: string;
  end_time: string;
}

// Farm & MCU Types
export interface Farm {
  id: string;
  name: string;
  description?: string;
  location?: string;
  created_at: string;
}

export interface MCU {
  id: string;
  mcu_code: string;
  farm_id: string;
  status: 'online' | 'offline';
  created_at: string;
  updated_at: string;
}

export interface SurveyPoint {
  id: string;
  name: string;
  description?: string;
  mcu_id: string;
  status: 'connecting' | 'connected' | 'disconnected';
  created_at: string;
  updated_at: string;
}

// API Response Types
export interface ApiResponse<T> {
  data: T;
  message?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
}