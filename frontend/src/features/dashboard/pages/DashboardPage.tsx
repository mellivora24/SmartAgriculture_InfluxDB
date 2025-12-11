/* eslint-disable @typescript-eslint/no-explicit-any */
import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useWebSocket } from '@/core/hooks/useWebSocket';
import { sensorDataApi } from '@/features/sensorData/api/sensorDataApi';
import SensorCard from '../components/SensorCard';
import type { SensorDataPoint, WebSocketMessage, SensorDataPayload, ControlRequestPayload, ControlResponsePayload, MQTTAlert } from '@/core/types';

interface SensorReading {
  temperature: number | null;
  humidity: number | null;
  soil_moisture: number | null;
  light: number | null;
  timestamp: string | null;
  survey_point_name: string | null;
  mcu_code: string | null;
}

interface PumpStatus {
  status: 'on' | 'off' | 'pending' | 'unknown';
  lastCommand: string | null;
  lastUpdate: string | null;
}

const DashboardPage = () => {
  const { surveyPointId } = useParams<{ surveyPointId: string }>();
  const navigate = useNavigate();
  const [sensorData, setSensorData] = useState<SensorReading>({
    temperature: null,
    humidity: null,
    soil_moisture: null,
    light: null,
    timestamp: null,
    survey_point_name: null,
    mcu_code: null,
  });
  const [pumpStatus, setPumpStatus] = useState<PumpStatus>({
    status: 'unknown',
    lastCommand: null,
    lastUpdate: null,
  });
  const [plantAlert, setPlantAlert] = useState<MQTTAlert | null>(null);
  const [alertHistory, setAlertHistory] = useState<MQTTAlert[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [commandStatus, setCommandStatus] = useState<string>('');
  const [mcuCodeReady, setMcuCodeReady] = useState<string | undefined>(undefined);

  useEffect(() => {
    const loadInitialData = async () => {
      if (!surveyPointId) return;

      try {
        setIsLoading(true);
        setError(null);
        console.log('Loading sensor data for survey point:', surveyPointId);

        try {
          const surveyPointResponse = await fetch(
            `http://localhost:8080/api/v1/survey-points/${surveyPointId}`,
            {
              headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`,
              },
            }
          );

          if (surveyPointResponse.ok) {
            const surveyPointData = await surveyPointResponse.json();
            console.log('Survey Point data:', surveyPointData);

            if (surveyPointData.data?.mcu_id) {
              const mcuResponse = await fetch(
                `http://localhost:8080/api/v1/mcus/${surveyPointData.data.mcu_id}`,
                {
                  headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`,
                  },
                }
              );

              if (mcuResponse.ok) {
                const mcuData = await mcuResponse.json();
                console.log('MCU data:', mcuData);

                if (mcuData.data?.mcu_code) {
                  setSensorData((prev) => ({
                    ...prev,
                    mcu_code: mcuData.data.mcu_code,
                    survey_point_name: surveyPointData.data.name,
                  }));
                  setMcuCodeReady(mcuData.data.mcu_code);
                  console.log('MCU Code set:', mcuData.data.mcu_code);
                } else {
                  setSensorData((prev) => ({
                    ...prev,
                    survey_point_name: surveyPointData.data.name,
                  }));
                }
              } else {
                setSensorData((prev) => ({
                  ...prev,
                  survey_point_name: surveyPointData.data.name,
                }));
              }
            } else {
              setSensorData((prev) => ({
                ...prev,
                survey_point_name: surveyPointData.data.name,
              }));
            }
          }
        } catch (err) {
          console.error('Failed to load survey point/MCU:', err);
        }

        const response = await sensorDataApi.getLatestData(surveyPointId, 1);
        console.log('Sensor data response:', response);

        if (response.data && response.data.length > 0) {
          const mcuCodeFromData = response.data[0].mcu_code;

          if (mcuCodeFromData && !mcuCodeReady) {
            setMcuCodeReady(mcuCodeFromData);
          }

          const latestData: Partial<SensorReading> = {
            survey_point_name: response.data[0].survey_point_name,
            mcu_code: mcuCodeFromData || sensorData.mcu_code,
            timestamp: response.data[0]._time,
          };

          console.log('MCU Code from API:', mcuCodeFromData);

          response.data.forEach((item: SensorDataPoint) => {
            if (item._field === 'temperature') latestData.temperature = item._value;
            if (item._field === 'humidity') latestData.humidity = item._value;
            if (item._field === 'soil_moisture') latestData.soil_moisture = item._value;
            if (item._field === 'light') latestData.light = item._value;
          });

          setSensorData((prev) => ({ ...prev, ...latestData }));
          console.log('Sensor data loaded successfully');
        } else {
          console.warn('No sensor data found');
        }
      } catch (err) {
        setError('Failed to load sensor data');
        console.error('Error loading sensor data:', err);
      } finally {
        setIsLoading(false);
      }
    };

    loadInitialData();
  }, [mcuCodeReady, sensorData.mcu_code, surveyPointId]);

  const { isConnected, sendMessage, connectionError } = useWebSocket({
    mcuCode: mcuCodeReady || undefined,
    onMessage: (message: WebSocketMessage<unknown>) => {
      console.log('WebSocket message received:', message);

      if (message.topic === 'sensor_data') {
        const payload = message.payload as SensorDataPayload;
        console.log('Sensor data payload:', payload);

        if (payload.survey_point_id === surveyPointId) {
          setSensorData((prev) => ({
            ...prev,
            temperature: payload.temperature ?? prev.temperature,
            humidity: payload.humidity ?? prev.humidity,
            soil_moisture: payload.soil_moisture ?? prev.soil_moisture,
            light: payload.light ?? prev.light,
            timestamp: message.timestamp || new Date().toISOString(),
          }));
          console.log('Sensor data updated');
        } else {
          console.log('Sensor data for different survey point, skipping');
        }
      } else if (message.topic === 'alert') {
        const payload = message.payload as MQTTAlert;
        console.log('Alert received:', payload);

        if (payload.mcu_code === mcuCodeReady) {
          setPlantAlert(payload);
          setAlertHistory((prev) => [payload, ...prev.slice(0, 9)]);
          console.log('Plant alert displayed');

          if ('Notification' in window && Notification.permission === 'granted') {
            new Notification(payload.title, {
              body: payload.message,
            });
          }
        }
      } else if (message.topic === 'control_response') {
        type ControlResponseWithPending = Omit<ControlResponsePayload, 'status'> & {
          command_id?: string;
          survey_point_id?: string;
          mcu_code?: string;
          status: 'success' | 'failed' | 'pending';
        };
        const payload = message.payload as ControlResponseWithPending;
        console.log('Control response payload:', payload);

        const isForCurrentDevice =
          payload.survey_point_id === surveyPointId ||
          payload.mcu_code === mcuCodeReady;

        if (!isForCurrentDevice) {
          console.log('Control response for different device, skipping');
          return;
        }

        if (payload.status === 'pending' && payload.command_id) {
          setCommandStatus(`Command sent to device (ID: ${payload.command_id}). Waiting for ESP8266...`);
          setPumpStatus((prev) => ({
            ...prev,
            status: 'pending',
            lastCommand: payload.command || prev.lastCommand,
          }));
          console.log('Command pending, waiting for ESP8266 response');
        } else {
          if ((window as any).__pumpControlTimeout) {
            clearTimeout((window as any).__pumpControlTimeout);
            (window as any).__pumpControlTimeout = null;
          }

          console.log('Final response from ESP8266:', payload);

          const command = payload.command?.toLowerCase() || '';
          const isOn = command.includes('on') || command === 'turn_on';

          const isSuccess = !payload.status || payload.status === 'success';

          setPumpStatus({
            status: isSuccess ? (isOn ? 'on' : 'off') : 'off',
            lastCommand: payload.command,
            lastUpdate: message.timestamp || new Date().toISOString(),
          });

          if (isSuccess) {
            setCommandStatus(`Pump ${isOn ? 'turned ON' : 'turned OFF'} successfully!`);
          } else {
            setCommandStatus(`Failed to ${isOn ? 'turn ON' : 'turn OFF'} pump: ${payload.message || 'Unknown error'}`);
          }

          setTimeout(() => setCommandStatus(''), 3000);
        }
      } else if (message.topic === 'error') {
        const errorPayload = message.payload as { code: string; message: string };
        console.error('WebSocket error:', errorPayload);
        setCommandStatus(`Error: ${errorPayload.message}`);
        setTimeout(() => setCommandStatus(''), 5000);
      }
    },
  });

  useEffect(() => {
    if ('Notification' in window && Notification.permission === 'default') {
      Notification.requestPermission();
    }
  }, []);

  const handlePumpControl = async (command: 'on' | 'off') => {
    if (!isConnected || !mcuCodeReady || !surveyPointId) {
      alert('Not connected or missing required data. Please wait...');
      return;
    }

    try {
      console.log('Sending pump control command:', { surveyPointId, mcuCodeReady, command });

      const controlRequest: WebSocketMessage<ControlRequestPayload> = {
        topic: 'control_request',
        payload: {
          survey_point_id: surveyPointId,
          mcu_code: mcuCodeReady,
          device_name: 'pump',
          command,
        },
      };

      sendMessage(controlRequest);
      setCommandStatus(`Sending ${command} command to pump...`);

      setPumpStatus((prev) => ({
        ...prev,
        status: 'pending',
        lastCommand: command,
      }));

      console.log('Control request sent via WebSocket');

      const timeoutId = setTimeout(() => {
        if (pumpStatus.status === 'pending') {
          setCommandStatus('Timeout: No response from device. Please try again.');
          setPumpStatus((prev) => ({
            ...prev,
            status: prev.status === 'pending' ? 'unknown' : prev.status,
          }));
          setTimeout(() => setCommandStatus(''), 5000);
        }
      }, 30000);

      (window as any).__pumpControlTimeout = timeoutId;
    } catch (err) {
      console.error('Failed to send command:', err);
      setCommandStatus('Failed to send command');
      setTimeout(() => setCommandStatus(''), 3000);
    }
  };

  const dismissAlert = () => {
    setPlantAlert(null);
  };

  const getAlertStyles = (severity: string) => {
    switch (severity) {
      case 'critical':
        return 'bg-red-100 border-red-500 text-red-900';
      case 'error':
        return 'bg-red-50 border-red-400 text-red-800';
      case 'warning':
        return 'bg-yellow-50 border-yellow-400 text-yellow-800';
      case 'info':
        return 'bg-blue-50 border-blue-400 text-blue-800';
      default:
        return 'bg-gray-50 border-gray-400 text-gray-800';
    }
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4" />
          <p className="text-gray-600">Loading dashboard...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <p className="text-red-600 text-lg mb-4">{error}</p>
          <button
            onClick={() => window.location.reload()}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  const getPumpStatusColor = () => {
    switch (pumpStatus.status) {
      case 'on':
        return 'bg-green-100 text-green-800';
      case 'off':
        return 'bg-gray-100 text-gray-800';
      case 'pending':
        return 'bg-yellow-100 text-yellow-800 animate-pulse';
      default:
        return 'bg-gray-100 text-gray-500';
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <button
            onClick={() => navigate(-1)}
            className="mb-4 flex items-center text-sm text-gray-600 hover:text-gray-900"
          >
            <svg className="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
            Back
          </button>
          <div className="flex justify-between items-center">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">
                {sensorData.survey_point_name || 'Dashboard'}
              </h1>
              <p className="mt-1 text-sm text-gray-500">
                Real-time sensor monitoring - MCU: {sensorData.mcu_code || 'N/A'} - Survey Point: {surveyPointId?.slice(0, 8)}...
              </p>
            </div>
            <div className="flex items-center space-x-3">
              <div className="flex items-center">
                <div
                  className={`h-3 w-3 rounded-full ${isConnected ? 'bg-green-500' : 'bg-red-500'
                    } animate-pulse`}
                />
                <span className="ml-2 text-sm font-medium text-gray-700">
                  {isConnected ? 'Connected' : 'Disconnected'}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {plantAlert && (
          <div className={`mb-6 p-6 rounded-lg border-l-4 shadow-lg ${getAlertStyles(plantAlert.severity)} animate-pulse`}>
            <div className="flex items-start justify-between">
              <div className="flex items-start space-x-3 flex-1">
                <div className="flex-1">
                  <h3 className="text-lg font-bold mb-1">{plantAlert.title}</h3>
                  <p className="text-base mb-2">{plantAlert.message}</p>
                  <p className="text-xs opacity-75">
                    {new Date(plantAlert.time).toLocaleString()} - MCU: {plantAlert.mcu_code}
                  </p>
                </div>
              </div>
              <button
                onClick={dismissAlert}
                className="ml-4 text-gray-500 hover:text-gray-700 text-2xl"
              >
                ×
              </button>
            </div>
          </div>
        )}

        {commandStatus && (
          <div className={`mb-6 p-4 rounded-lg border ${commandStatus.includes('successfully')
              ? 'bg-green-50 border-green-200'
              : commandStatus.includes('Failed') || commandStatus.includes('Timeout')
                ? 'bg-red-50 border-red-200'
                : 'bg-blue-50 border-blue-200'
            }`}>
            <p className={`text-sm font-medium ${commandStatus.includes('successfully')
                ? 'text-green-800'
                : commandStatus.includes('Failed') || commandStatus.includes('Timeout')
                  ? 'text-red-800'
                  : 'text-blue-800'
              }`}>
              {commandStatus}
            </p>
          </div>
        )}

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <SensorCard
            title="Temperature"
            value={sensorData.temperature}
            unit="°C"
            icon=""
            color="text-red-600"
            bgColor="bg-red-50"
          />
          <SensorCard
            title="Humidity"
            value={sensorData.humidity}
            unit="%"
            icon=""
            color="text-blue-600"
            bgColor="bg-blue-50"
          />
          <SensorCard
            title="Soil Moisture"
            value={sensorData.soil_moisture}
            unit="%"
            icon=""
            color="text-green-600"
            bgColor="bg-green-50"
          />
          <SensorCard
            title="Light"
            value={sensorData.light}
            unit="lux"
            icon=""
            color="text-yellow-600"
            bgColor="bg-yellow-50"
          />
        </div>

        {alertHistory.length > 0 && (
          <div className="mb-8 bg-white rounded-lg shadow p-6">
            <h2 className="text-xl font-bold text-gray-900 mb-4">Alert History</h2>
            <div className="space-y-3 max-h-64 overflow-y-auto">
              {alertHistory.map((alert, index) => (
                <div
                  key={index}
                  className={`p-3 rounded border-l-4 text-sm ${getAlertStyles(alert.severity)}`}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-2">
                      <span className="font-semibold">{alert.title}</span>
                    </div>
                    <span className="text-xs opacity-75">
                      {new Date(alert.time).toLocaleTimeString()}
                    </span>
                  </div>
                  <p className="mt-1 ml-6">{alert.message}</p>
                </div>
              ))}
            </div>
          </div>
        )}

        <div className="mb-8 bg-white rounded-lg shadow-lg p-8">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-2xl font-bold text-gray-900">Water Pump Control</h2>
            <div className={`px-4 py-2 rounded-full text-sm font-semibold ${getPumpStatusColor()}`}>
              {pumpStatus.status.toUpperCase()}
            </div>
          </div>

          <div className="flex items-center justify-center mb-6">
            <div className="text-center">
              <p className="text-gray-600">
                {pumpStatus.status === 'on' && 'Pump is running'}
                {pumpStatus.status === 'off' && 'Pump is stopped'}
                {pumpStatus.status === 'pending' && 'Processing command...'}
                {pumpStatus.status === 'unknown' && 'Status unknown'}
              </p>
              {pumpStatus.lastUpdate && (
                <p className="text-xs text-gray-500 mt-2">
                  Last update: {new Date(pumpStatus.lastUpdate).toLocaleTimeString()}
                </p>
              )}
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4 max-w-md mx-auto">
            <button
              onClick={() => handlePumpControl('on')}
              disabled={!isConnected || pumpStatus.status === 'pending' || !surveyPointId}
              className="px-6 py-4 bg-green-500 hover:bg-green-600 text-white rounded-lg font-semibold text-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed shadow-lg hover:shadow-xl"
            >
              Turn ON
            </button>
            <button
              onClick={() => handlePumpControl('off')}
              disabled={!isConnected || pumpStatus.status === 'pending' || !surveyPointId}
              className="px-6 py-4 bg-red-500 hover:bg-red-600 text-white rounded-lg font-semibold text-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed shadow-lg hover:shadow-xl"
            >
              Turn OFF
            </button>
          </div>

          {!isConnected && (
            <div className="mt-4 p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
              <p className="text-sm text-yellow-800 text-center">
                {connectionError || 'WebSocket disconnected. Reconnecting...'}
              </p>
              <div className="mt-2 text-xs text-gray-600 text-center space-y-1">
                <p>MCU Code: {sensorData.mcu_code || 'Not loaded yet'}</p>
                <p>Survey Point ID: {surveyPointId || 'Missing'}</p>
                <p>Token: {localStorage.getItem('token') ? 'Present' : 'Missing'}</p>
              </div>
            </div>
          )}
        </div>

        {sensorData.timestamp && (
          <div className="bg-white rounded-lg shadow p-4 text-center mb-8">
            <p className="text-sm text-gray-500">
              Last sensor update: {new Date(sensorData.timestamp).toLocaleString()}
            </p>
          </div>
        )}

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <button
            onClick={() => navigate(`/dashboard/${surveyPointId}/sensor-history`)}
            className="bg-white rounded-lg shadow p-6 hover:shadow-lg transition-shadow text-left"
          >
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-lg font-semibold text-gray-900">Sensor History</h3>
                <p className="text-sm text-gray-600">View historical sensor data</p>
              </div>
              <svg className="h-6 w-6 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
                />
              </svg>
            </div>
          </button>

          <button
            onClick={() => navigate(`/dashboard/${surveyPointId}/control-history`)}
            className="bg-white rounded-lg shadow p-6 hover:shadow-lg transition-shadow text-left"
          >
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-lg font-semibold text-gray-900">Control History</h3>
                <p className="text-sm text-gray-600">View pump control history</p>
              </div>
              <svg className="h-6 w-6 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4"
                />
              </svg>
            </div>
          </button>

          <button
            onClick={() => navigate(`/dashboard/${surveyPointId}/threshold-settings`)}
            className="bg-white rounded-lg shadow p-6 hover:shadow-lg transition-shadow text-left border-2 border-orange-200"
          >
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-lg font-semibold text-gray-900">Threshold Settings</h3>
                <p className="text-sm text-gray-600">Configure alerts & auto pump</p>
              </div>
              <svg className="h-6 w-6 text-orange-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
                />
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                />
              </svg>
            </div>
          </button>
        </div>
      </div>
    </div>
  );
};

export default DashboardPage;
