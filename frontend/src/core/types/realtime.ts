import type { AlertSeverity } from "./common";

export type WebSocketTopic =
    | "sensor_data"
    | "control_request"
    | "control_response"
    | "alert"
    | "ping"
    | "pong";

export interface WebSocketMessage<T> {
    topic: WebSocketTopic;
    payload: T;
    timestamp?: string;
}

export interface SensorDataPayload {
    mcu_code: string;
    survey_point_id: string;
    temperature: number;
    humidity: number;
    soil_moisture: number;
    light: number;
}

export interface ControlRequestPayload {
    mcu_code: string;
    device_name: string;
    command: string;
    value?: number;
}

export interface ControlResponsePayload {
    mcu_code: string;
    device_name: string;
    command: string;
    status: "success" | "failed";
    message: string;
    executed_at: string;
}

export interface AlertPayload {
    mcu_code: string;
    title: string;
    message: string;
    severity: AlertSeverity;
    time: string;
}

export interface MQTTPublishRequest {
    topic: string;
    message: WebSocketMessage<unknown>;
}

export interface WebSocketBroadcastRequest {
    mcu_code: string;
    topic: WebSocketTopic;
    message: unknown;
}

export interface WebSocketConnectionParams {
    token: string;
    mcu_code: string;
}

export interface HealthCheckResponse {
    status: "ok" | "error";
    time: string;
    postgres: "ok" | "error";
    influxdb: "ok" | "error";
    mqtt: "ok" | "error";
    websocket: {
        connected_clients: number;
    };
}