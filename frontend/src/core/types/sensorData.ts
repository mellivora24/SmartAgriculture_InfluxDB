export interface WriteSensorDataRequest {
    survey_point_id: string;
    survey_point_name: string;
    mcu_code: string;
    farm_name: string;
    temperature: number;
    humidity: number;
    soil_moisture: number;
    light: number;
    timestamp: string;
    metadata?: Record<string, unknown>;
}

export interface SensorDataPoint {
    _field: string;
    _measurement: string;
    _start: string;
    _stop: string;
    _time: string;
    _value: number;
    farm_name: string;
    mcu_code: string;
    result: string;
    survey_point_id: string;
    survey_point_name: string;
    table: number;
}

export interface QuerySensorDataParams {
    survey_point_id: string;
    mcu_code?: string;
    limit?: number;
}

export interface SensorAggregationRequest {
    survey_point_id: string;
    field: 'temperature' | 'humidity' | 'soil_moisture' | 'light';
    aggregation: 'mean' | 'min' | 'max' | 'sum';
    window: string; // e.g., "1h", "5m", "1d"
    start_time: string;
    end_time: string;
}

export interface ProcessedSensorData {
    timestamp: string;
    temperature?: number;
    humidity?: number;
    soil_moisture?: number;
    light?: number;
}

export interface LatestSensorReadings {
    survey_point_id: string;
    survey_point_name: string;
    temperature: number | null;
    humidity: number | null;
    soil_moisture: number | null;
    light: number | null;
    last_updated: string;
}
