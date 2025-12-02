import { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { mcuApi } from '../api/mcuApi';
import { farmApi } from '@/features/farm/api/farmApi';
import type { MCUWithStats, FarmOverview } from '@/core/types';

const MCUListPage = () => {
  const { farmId } = useParams<{ farmId: string }>();
  const [mcus, setMcus] = useState<MCUWithStats[]>([]);
  const [farmInfo, setFarmInfo] = useState<FarmOverview | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    const loadData = async () => {
      if (!farmId) return;

      try {
        setIsLoading(true);
        setError(null);
        
        const [mcuResponse, farmResponse] = await Promise.all([
          mcuApi.getMCUsByFarm(farmId),
          farmApi.getFarmOverview(farmId),
        ]);

        setMcus(mcuResponse.data || []);
        setFarmInfo(farmResponse.data);
      } catch (err) {
        setError('Failed to load MCUs');
        console.error('Error loading MCUs:', err);
      } finally {
        setIsLoading(false);
      }
    };

    loadData();
  }, [farmId]);

  const handleMCUClick = (mcuId: string) => {
    navigate(`/farm/${farmId}/mcu/${mcuId}/survey-points`);
  };

  const handleAddMCU = () => {
    navigate(`/farm/${farmId}/mcus/create`);
  };

  const getStatusColor = (status: string) => {
    return status === 'online' ? 'text-green-600 bg-green-100' : 'text-red-600 bg-red-100';
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <p className="text-red-600 text-lg">{error}</p>
          <button
            onClick={() => window.location.reload()}
            className="mt-4 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <button
            onClick={() => navigate('/')}
            className="mb-4 flex items-center text-sm text-gray-600 hover:text-gray-900"
          >
            <svg
              className="h-4 w-4 mr-1"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M15 19l-7-7 7-7"
              />
            </svg>
            Back to Farms
          </button>
          
          <div className="flex justify-between items-center">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">
                {farmInfo?.farm_name || 'Loading...'}
              </h1>
              <p className="mt-1 text-sm text-gray-500">
                Select an MCU to view its survey points
              </p>
            </div>
            <button
              onClick={handleAddMCU}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors flex items-center"
            >
              <svg
                className="h-5 w-5 mr-2"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 4v16m8-8H4"
                />
              </svg>
              Add MCU
            </button>
          </div>

          {/* Farm Stats */}
          {farmInfo && (
            <div className="mt-6 grid grid-cols-2 md:grid-cols-4 gap-4">
              <div className="bg-blue-50 rounded-lg p-4">
                <p className="text-sm text-gray-600">Total MCUs</p>
                <p className="text-2xl font-bold text-blue-600">{farmInfo.total_mcus}</p>
              </div>
              <div className="bg-green-50 rounded-lg p-4">
                <p className="text-sm text-gray-600">Online MCUs</p>
                <p className="text-2xl font-bold text-green-600">{farmInfo.online_mcus}</p>
              </div>
              <div className="bg-purple-50 rounded-lg p-4">
                <p className="text-sm text-gray-600">Survey Points</p>
                <p className="text-2xl font-bold text-purple-600">{farmInfo.total_survey_points}</p>
              </div>
              <div className="bg-yellow-50 rounded-lg p-4">
                <p className="text-sm text-gray-600">Connected</p>
                <p className="text-2xl font-bold text-yellow-600">{farmInfo.connected_points}</p>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* MCU List */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {mcus.length === 0 ? (
          <div className="text-center py-12 bg-white rounded-lg shadow">
            <svg
              className="mx-auto h-12 w-12 text-gray-400"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"
              />
            </svg>
            <h3 className="mt-2 text-sm font-medium text-gray-900">No MCUs</h3>
            <p className="mt-1 text-sm text-gray-500">
              This farm doesn't have any MCUs yet.
            </p>
            <div className="mt-6">
              <button
                onClick={handleAddMCU}
                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
              >
                Add First MCU
              </button>
            </div>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {mcus.map((mcu) => (
              <div
                key={mcu.mcu_id}
                onClick={() => handleMCUClick(mcu.mcu_id)}
                className="bg-white rounded-lg shadow hover:shadow-lg transition-shadow cursor-pointer overflow-hidden"
              >
                <div className="p-6">
                  <div className="flex items-center justify-between mb-4">
                    <h3 className="text-xl font-semibold text-gray-900">
                      MCU {mcu.mcu_code}
                    </h3>
                    <span
                      className={`px-3 py-1 rounded-full text-xs font-medium ${getStatusColor(
                        mcu.status
                      )}`}
                    >
                      {mcu.status}
                    </span>
                  </div>

                  <div className="space-y-3">
                    <div className="flex items-center justify-between">
                      <span className="text-sm text-gray-600">Survey Points</span>
                      <span className="text-lg font-bold text-blue-600">
                        {mcu.survey_point_count}
                      </span>
                    </div>

                    <div className="pt-3 border-t border-gray-200">
                      <p className="text-xs text-gray-500">
                        Created: {new Date(mcu.created_at).toLocaleDateString()}
                      </p>
                      <p className="text-xs text-gray-500">
                        Updated: {new Date(mcu.updated_at).toLocaleDateString()}
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default MCUListPage;
