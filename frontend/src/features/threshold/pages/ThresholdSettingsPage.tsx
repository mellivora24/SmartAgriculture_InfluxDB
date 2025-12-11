/* eslint-disable @typescript-eslint/no-explicit-any */
import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import axios from 'axios';

interface ThresholdSettings {
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

interface AlertHistory {
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

interface AutoPumpHistory {
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

const ThresholdSettingsPage = () => {
  const { surveyPointId } = useParams<{ surveyPointId: string }>();
  const navigate = useNavigate();
  const [settings, setSettings] = useState<ThresholdSettings | null>(null);
  const [alertHistory, setAlertHistory] = useState<AlertHistory[]>([]);
  const [autoPumpHistory, setAutoPumpHistory] = useState<AutoPumpHistory[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'settings' | 'alerts' | 'pump-history'>('settings');

  const API_BASE_URL = 'http://localhost:8080/api/v1';
  const token = localStorage.getItem('token');

  useEffect(() => {
    loadThresholdSettings();
    loadAlertHistory();
    loadAutoPumpHistory();
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [surveyPointId]);

  const loadThresholdSettings = async () => {
    try {
      setIsLoading(true);
      const response = await axios.get(
        `${API_BASE_URL}/thresholds/survey-point/${surveyPointId}`,
        {
          headers: { Authorization: `Bearer ${token}` },
        }
      );
      setSettings(response.data.data);
    } catch (err: any) {
      if (err.response?.status === 404) {
        // Tạo settings mặc định nếu chưa có
        setSettings({
          id: '',
          survey_point_id: surveyPointId || '',
          auto_pump_enabled: false,
          pump_duration_seconds: 30,
          pump_cooldown_minutes: 60,
          alert_enabled: true,
          alert_cooldown_minutes: 10,
          created_at: '',
          updated_at: '',
        });
      } else {
        setError('Failed to load threshold settings');
        console.error(err);
      }
    } finally {
      setIsLoading(false);
    }
  };

  const loadAlertHistory = async () => {
    try {
      const response = await axios.get(
        `${API_BASE_URL}/thresholds/survey-point/${surveyPointId}/alerts?limit=50`,
        {
          headers: { Authorization: `Bearer ${token}` },
        }
      );
      setAlertHistory(response.data.data || []);
    } catch (err) {
      console.error('Failed to load alert history:', err);
    }
  };

  const loadAutoPumpHistory = async () => {
    try {
      const response = await axios.get(
        `${API_BASE_URL}/thresholds/survey-point/${surveyPointId}/auto-pump-history?limit=50`,
        {
          headers: { Authorization: `Bearer ${token}` },
        }
      );
      setAutoPumpHistory(response.data.data || []);
    } catch (err) {
      console.error('Failed to load auto pump history:', err);
    }
  };

  const handleSaveSettings = async () => {
    if (!settings) return;

    try {
      setIsSaving(true);
      setError(null);
      setSuccessMessage(null);

      const payload = {
        temp_min: settings.temp_min,
        temp_max: settings.temp_max,
        temp_critical_min: settings.temp_critical_min,
        temp_critical_max: settings.temp_critical_max,
        humidity_min: settings.humidity_min,
        humidity_max: settings.humidity_max,
        humidity_critical_min: settings.humidity_critical_min,
        humidity_critical_max: settings.humidity_critical_max,
        soil_moisture_min: settings.soil_moisture_min,
        soil_moisture_max: settings.soil_moisture_max,
        soil_moisture_critical_min: settings.soil_moisture_critical_min,
        soil_moisture_critical_max: settings.soil_moisture_critical_max,
        light_min: settings.light_min,
        light_max: settings.light_max,
        light_critical_min: settings.light_critical_min,
        light_critical_max: settings.light_critical_max,
        auto_pump_enabled: settings.auto_pump_enabled,
        pump_trigger_soil_moisture: settings.pump_trigger_soil_moisture,
        pump_stop_soil_moisture: settings.pump_stop_soil_moisture,
        pump_duration_seconds: settings.pump_duration_seconds,
        pump_cooldown_minutes: settings.pump_cooldown_minutes,
        alert_enabled: settings.alert_enabled,
        alert_cooldown_minutes: settings.alert_cooldown_minutes,
      };

      const response = await axios.put(
        `${API_BASE_URL}/thresholds/survey-point/${surveyPointId}`,
        payload,
        {
          headers: { Authorization: `Bearer ${token}` },
        }
      );

      setSettings(response.data.data);
      setSuccessMessage('Settings saved successfully!');
      setTimeout(() => setSuccessMessage(null), 3000);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to save settings');
      console.error(err);
    } finally {
      setIsSaving(false);
    }
  };

  const handleAcknowledgeAlert = async (alertId: string) => {
    try {
      await axios.post(
        `${API_BASE_URL}/thresholds/alerts/${alertId}/acknowledge`,
        {},
        {
          headers: { Authorization: `Bearer ${token}` },
        }
      );
      loadAlertHistory();
    } catch (err) {
      console.error('Failed to acknowledge alert:', err);
    }
  };

  const updateSetting = (field: keyof ThresholdSettings, value: any) => {
    setSettings((prev) => (prev ? { ...prev, [field]: value } : null));
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical':
        return 'bg-red-100 text-red-800 border-red-200';
      case 'warning':
        return 'bg-yellow-100 text-yellow-800 border-yellow-200';
      case 'info':
        return 'bg-blue-100 text-blue-800 border-blue-200';
      default:
        return 'bg-gray-100 text-gray-800 border-gray-200';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed':
        return 'bg-green-100 text-green-800';
      case 'running':
        return 'bg-blue-100 text-blue-800';
      case 'failed':
        return 'bg-red-100 text-red-800';
      case 'triggered':
        return 'bg-yellow-100 text-yellow-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4" />
          <p className="text-gray-600">Loading settings...</p>
        </div>
      </div>
    );
  }

  if (!settings) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <p className="text-red-600 text-lg mb-4">Failed to load settings</p>
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
            Back to Dashboard
          </button>
          <h1 className="text-3xl font-bold text-gray-900">Threshold Settings</h1>
          <p className="mt-1 text-sm text-gray-500">
            Configure alert thresholds and auto pump settings
          </p>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-sm text-red-800">{error}</p>
          </div>
        )}

        {successMessage && (
          <div className="mb-6 p-4 bg-green-50 border border-green-200 rounded-lg">
            <p className="text-sm text-green-800">{successMessage}</p>
          </div>
        )}

        {/* Tabs */}
        <div className="mb-6 border-b border-gray-200">
          <nav className="-mb-px flex space-x-8">
            <button
              onClick={() => setActiveTab('settings')}
              className={`py-4 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'settings'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Settings
            </button>
            <button
              onClick={() => setActiveTab('alerts')}
              className={`py-4 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'alerts'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Alert History ({alertHistory.filter((a) => !a.acknowledged).length})
            </button>
            <button
              onClick={() => setActiveTab('pump-history')}
              className={`py-4 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'pump-history'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Auto Pump History
            </button>
          </nav>
        </div>

        {/* Settings Tab */}
        {activeTab === 'settings' && (
          <div className="space-y-6">
            {/* Alert Settings */}
            <div className="bg-white rounded-lg shadow p-6">
              <h2 className="text-xl font-bold text-gray-900 mb-4">Alert Settings</h2>
              <div className="space-y-4">
                <div className="flex items-center">
                  <input
                    type="checkbox"
                    id="alert_enabled"
                    checked={settings.alert_enabled}
                    onChange={(e) => updateSetting('alert_enabled', e.target.checked)}
                    className="h-4 w-4 text-blue-600 rounded"
                  />
                  <label htmlFor="alert_enabled" className="ml-2 text-sm text-gray-700">
                    Enable Alerts
                  </label>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Alert Cooldown (minutes)
                  </label>
                  <input
                    type="number"
                    value={settings.alert_cooldown_minutes}
                    onChange={(e) => updateSetting('alert_cooldown_minutes', parseInt(e.target.value))}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                    min="1"
                  />
                  <p className="text-xs text-gray-500 mt-1">
                    Minimum time between alerts of the same type
                  </p>
                </div>
              </div>
            </div>

            {/* Temperature Thresholds */}
            <div className="bg-white rounded-lg shadow p-6">
              <h2 className="text-xl font-bold text-gray-900 mb-4">Temperature Thresholds (°C)</h2>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Warning Min
                  </label>
                  <input
                    type="number"
                    step="0.1"
                    value={settings.temp_min || ''}
                    onChange={(e) => updateSetting('temp_min', e.target.value ? parseFloat(e.target.value) : undefined)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                    placeholder="e.g., 15"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Warning Max
                  </label>
                  <input
                    type="number"
                    step="0.1"
                    value={settings.temp_max || ''}
                    onChange={(e) => updateSetting('temp_max', e.target.value ? parseFloat(e.target.value) : undefined)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                    placeholder="e.g., 35"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-red-700 mb-2">
                    Critical Min
                  </label>
                  <input
                    type="number"
                    step="0.1"
                    value={settings.temp_critical_min || ''}
                    onChange={(e) => updateSetting('temp_critical_min', e.target.value ? parseFloat(e.target.value) : undefined)}
                    className="w-full px-3 py-2 border border-red-300 rounded-lg"
                    placeholder="e.g., 10"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-red-700 mb-2">
                    Critical Max
                  </label>
                  <input
                    type="number"
                    step="0.1"
                    value={settings.temp_critical_max || ''}
                    onChange={(e) => updateSetting('temp_critical_max', e.target.value ? parseFloat(e.target.value) : undefined)}
                    className="w-full px-3 py-2 border border-red-300 rounded-lg"
                    placeholder="e.g., 40"
                  />
                </div>
              </div>
            </div>

            {/* Humidity Thresholds */}
            <div className="bg-white rounded-lg shadow p-6">
              <h2 className="text-xl font-bold text-gray-900 mb-4">Humidity Thresholds (%)</h2>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Warning Min
                  </label>
                  <input
                    type="number"
                    step="0.1"
                    value={settings.humidity_min || ''}
                    onChange={(e) => updateSetting('humidity_min', e.target.value ? parseFloat(e.target.value) : undefined)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                    placeholder="e.g., 40"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Warning Max
                  </label>
                  <input
                    type="number"
                    step="0.1"
                    value={settings.humidity_max || ''}
                    onChange={(e) => updateSetting('humidity_max', e.target.value ? parseFloat(e.target.value) : undefined)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                    placeholder="e.g., 80"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-red-700 mb-2">
                    Critical Min
                  </label>
                  <input
                    type="number"
                    step="0.1"
                    value={settings.humidity_critical_min || ''}
                    onChange={(e) => updateSetting('humidity_critical_min', e.target.value ? parseFloat(e.target.value) : undefined)}
                    className="w-full px-3 py-2 border border-red-300 rounded-lg"
                    placeholder="e.g., 30"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-red-700 mb-2">
                    Critical Max
                  </label>
                  <input
                    type="number"
                    step="0.1"
                    value={settings.humidity_critical_max || ''}
                    onChange={(e) => updateSetting('humidity_critical_max', e.target.value ? parseFloat(e.target.value) : undefined)}
                    className="w-full px-3 py-2 border border-red-300 rounded-lg"
                    placeholder="e.g., 90"
                  />
                </div>
              </div>
            </div>

            {/* Soil Moisture Thresholds */}
            <div className="bg-white rounded-lg shadow p-6">
              <h2 className="text-xl font-bold text-gray-900 mb-4">Soil Moisture Thresholds (%)</h2>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Warning Min
                  </label>
                  <input
                    type="number"
                    step="0.1"
                    value={settings.soil_moisture_min || ''}
                    onChange={(e) => updateSetting('soil_moisture_min', e.target.value ? parseFloat(e.target.value) : undefined)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                    placeholder="e.g., 30"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Warning Max
                  </label>
                  <input
                    type="number"
                    step="0.1"
                    value={settings.soil_moisture_max || ''}
                    onChange={(e) => updateSetting('soil_moisture_max', e.target.value ? parseFloat(e.target.value) : undefined)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                    placeholder="e.g., 80"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-red-700 mb-2">
                    Critical Min
                  </label>
                  <input
                    type="number"
                    step="0.1"
                    value={settings.soil_moisture_critical_min || ''}
                    onChange={(e) => updateSetting('soil_moisture_critical_min', e.target.value ? parseFloat(e.target.value) : undefined)}
                    className="w-full px-3 py-2 border border-red-300 rounded-lg"
                    placeholder="e.g., 20"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-red-700 mb-2">
                    Critical Max
                  </label>
                  <input
                    type="number"
                    step="0.1"
                    value={settings.soil_moisture_critical_max || ''}
                    onChange={(e) => updateSetting('soil_moisture_critical_max', e.target.value ? parseFloat(e.target.value) : undefined)}
                    className="w-full px-3 py-2 border border-red-300 rounded-lg"
                    placeholder="e.g., 90"
                  />
                </div>
              </div>
            </div>

            {/* Light Thresholds */}
            <div className="bg-white rounded-lg shadow p-6">
              <h2 className="text-xl font-bold text-gray-900 mb-4">Light Thresholds (lux)</h2>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Warning Min
                  </label>
                  <input
                    type="number"
                    step="1"
                    value={settings.light_min || ''}
                    onChange={(e) => updateSetting('light_min', e.target.value ? parseFloat(e.target.value) : undefined)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                    placeholder="e.g., 1000"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Warning Max
                  </label>
                  <input
                    type="number"
                    step="1"
                    value={settings.light_max || ''}
                    onChange={(e) => updateSetting('light_max', e.target.value ? parseFloat(e.target.value) : undefined)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                    placeholder="e.g., 50000"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-red-700 mb-2">
                    Critical Min
                  </label>
                  <input
                    type="number"
                    step="1"
                    value={settings.light_critical_min || ''}
                    onChange={(e) => updateSetting('light_critical_min', e.target.value ? parseFloat(e.target.value) : undefined)}
                    className="w-full px-3 py-2 border border-red-300 rounded-lg"
                    placeholder="e.g., 500"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-red-700 mb-2">
                    Critical Max
                  </label>
                  <input
                    type="number"
                    step="1"
                    value={settings.light_critical_max || ''}
                    onChange={(e) => updateSetting('light_critical_max', e.target.value ? parseFloat(e.target.value) : undefined)}
                    className="w-full px-3 py-2 border border-red-300 rounded-lg"
                    placeholder="e.g., 70000"
                  />
                </div>
              </div>
            </div>

            {/* Auto Pump Settings */}
            <div className="bg-white rounded-lg shadow p-6">
              <h2 className="text-xl font-bold text-gray-900 mb-4">Auto Pump Settings</h2>
              <div className="space-y-4">
                <div className="flex items-center">
                  <input
                    type="checkbox"
                    id="auto_pump_enabled"
                    checked={settings.auto_pump_enabled}
                    onChange={(e) => updateSetting('auto_pump_enabled', e.target.checked)}
                    className="h-4 w-4 text-blue-600 rounded"
                  />
                  <label htmlFor="auto_pump_enabled" className="ml-2 text-sm text-gray-700">
                    Enable Auto Pump
                  </label>
                </div>
                {settings.auto_pump_enabled && (
                  <>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Trigger Soil Moisture (%)
                      </label>
                      <input
                        type="number"
                        step="0.1"
                        value={settings.pump_trigger_soil_moisture || ''}
                        onChange={(e) => updateSetting('pump_trigger_soil_moisture', e.target.value ? parseFloat(e.target.value) : undefined)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                        placeholder="e.g., 30"
                      />
                      <p className="text-xs text-gray-500 mt-1">
                        Pump will turn on when soil moisture drops below this value
                      </p>
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Target Soil Moisture (%)
                      </label>
                      <input
                        type="number"
                        step="0.1"
                        value={settings.pump_stop_soil_moisture || ''}
                        onChange={(e) => updateSetting('pump_stop_soil_moisture', e.target.value ? parseFloat(e.target.value) : undefined)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                        placeholder="e.g., 60"
                      />
                      <p className="text-xs text-gray-500 mt-1">
                        Target moisture level after pumping
                      </p>
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Pump Duration (seconds)
                      </label>
                      <input
                        type="number"
                        value={settings.pump_duration_seconds}
                        onChange={(e) => updateSetting('pump_duration_seconds', parseInt(e.target.value))}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                        min="1"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Cooldown Period (minutes)
                      </label>
                      <input
                        type="number"
                        value={settings.pump_cooldown_minutes}
                        onChange={(e) => updateSetting('pump_cooldown_minutes', parseInt(e.target.value))}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg"
                        min="1"
                      />
                      <p className="text-xs text-gray-500 mt-1">
                        Minimum time between auto pump activations
                      </p>
                    </div>
                  </>
                )}
              </div>
            </div>

            {/* Save Button */}
            <div className="flex justify-end">
              <button
                onClick={handleSaveSettings}
                disabled={isSaving}
                className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed font-semibold"
              >
                {isSaving ? 'Saving...' : 'Save Settings'}
              </button>
            </div>
          </div>
        )}

        {/* Alert History Tab */}
        {activeTab === 'alerts' && (
          <div className="bg-white rounded-lg shadow">
            <div className="p-6">
              <h2 className="text-xl font-bold text-gray-900 mb-4">Alert History</h2>
              {alertHistory.length === 0 ? (
                <p className="text-gray-500 text-center py-8">No alerts recorded</p>
              ) : (
                <div className="space-y-3">
                  {alertHistory.map((alert) => (
                    <div
                      key={alert.id}
                      className={`p-4 rounded-lg border ${getSeverityColor(alert.severity)} ${
                        alert.acknowledged ? 'opacity-50' : ''
                      }`}
                    >
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <div className="flex items-center space-x-2 mb-2">
                            <span className="font-semibold capitalize">{alert.alert_type.replace('_', ' ')}</span>
                            <span className="px-2 py-1 text-xs font-semibold rounded-full bg-white">
                              {alert.severity}
                            </span>
                          </div>
                          <p className="text-sm mb-2">{alert.message}</p>
                          <div className="flex items-center space-x-4 text-xs text-gray-600">
                            <span>Value: {alert.sensor_value.toFixed(2)}</span>
                            <span>Threshold: {alert.threshold_value.toFixed(2)}</span>
                            <span>{new Date(alert.created_at).toLocaleString()}</span>
                          </div>
                          {alert.acknowledged && (
                            <p className="text-xs text-gray-500 mt-2">
                              ✓ Acknowledged at {new Date(alert.acknowledged_at!).toLocaleString()}
                            </p>
                          )}
                        </div>
                        {!alert.acknowledged && (
                          <button
                            onClick={() => handleAcknowledgeAlert(alert.id)}
                            className="ml-4 px-3 py-1 text-sm bg-white border border-gray-300 rounded hover:bg-gray-50"
                          >
                            Acknowledge
                          </button>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        )}

        {/* Auto Pump History Tab */}
        {activeTab === 'pump-history' && (
          <div className="bg-white rounded-lg shadow">
            <div className="p-6">
              <h2 className="text-xl font-bold text-gray-900 mb-4">Auto Pump History</h2>
              {autoPumpHistory.length === 0 ? (
                <p className="text-gray-500 text-center py-8">No auto pump activities recorded</p>
              ) : (
                <div className="space-y-3">
                  {autoPumpHistory.map((history) => (
                    <div key={history.id} className="p-4 rounded-lg border border-gray-200">
                      <div className="flex items-start justify-between mb-2">
                        <div className="flex items-center space-x-2">
                          <span className="font-semibold">Auto Pump Activation</span>
                          <span className={`px-2 py-1 text-xs font-semibold rounded-full ${getStatusColor(history.status)}`}>
                            {history.status}
                          </span>
                        </div>
                        <span className="text-xs text-gray-500">
                          {new Date(history.started_at).toLocaleString()}
                        </span>
                      </div>
                      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                        <div>
                          <p className="text-gray-600">Trigger Moisture</p>
                          <p className="font-semibold">{history.trigger_soil_moisture.toFixed(1)}%</p>
                        </div>
                        <div>
                          <p className="text-gray-600">Target Moisture</p>
                          <p className="font-semibold">{history.target_soil_moisture.toFixed(1)}%</p>
                        </div>
                        <div>
                          <p className="text-gray-600">Duration</p>
                          <p className="font-semibold">{history.pump_duration_seconds}s</p>
                        </div>
                        <div>
                          <p className="text-gray-600">Completed</p>
                          <p className="font-semibold">
                            {history.completed_at ? new Date(history.completed_at).toLocaleTimeString() : 'In progress'}
                          </p>
                        </div>
                      </div>
                      {history.notes && (
                        <p className="text-sm text-gray-600 mt-2">
                          <span className="font-medium">Notes:</span> {history.notes}
                        </p>
                      )}
                      {history.command_id && (
                        <p className="text-xs text-gray-500 mt-2">
                          Command ID: {history.command_id}
                        </p>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default ThresholdSettingsPage;
