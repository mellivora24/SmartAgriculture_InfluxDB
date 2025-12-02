import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { farmApi } from '../api/farmApi';
import type { MyFarm } from '@/core/types';

const HomePage = () => {
  const [farms, setFarms] = useState<MyFarm[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    const loadFarms = async () => {
      try {
        setIsLoading(true);
        setError(null);
        const response = await farmApi.getMyFarms();
        setFarms(response.data || []);
      } catch (err) {
        setError('Failed to load farms');
        console.error('Error loading farms:', err);
      } finally {
        setIsLoading(false);
      }
    };

    loadFarms();
  }, []);

  const handleFarmClick = (farmId: string) => {
    navigate(`/farm/${farmId}/mcus`);
  };

  const handleCreateFarm = () => {
    navigate('/farms/create');
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
          <div className="flex justify-between items-center">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">My Farms</h1>
              <p className="mt-1 text-sm text-gray-500">
                Select a farm to view its MCUs and survey points
              </p>
            </div>
            <button
              onClick={handleCreateFarm}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
            >
              Add New Farm
            </button>
          </div>
        </div>
      </div>

      {/* Farm List */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {farms.length === 0 ? (
          <div className="text-center py-12">
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
                d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"
              />
            </svg>
            <h3 className="mt-2 text-sm font-medium text-gray-900">No farms</h3>
            <p className="mt-1 text-sm text-gray-500">
              Get started by creating a new farm.
            </p>
            <div className="mt-6">
              <button
                onClick={handleCreateFarm}
                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
              >
                Create Farm
              </button>
            </div>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {farms.map((farm) => (
              <div
                key={farm.farm_id}
                onClick={() => handleFarmClick(farm.farm_id)}
                className="bg-white rounded-lg shadow hover:shadow-lg transition-shadow cursor-pointer overflow-hidden"
              >
                <div className="p-6">
                  <h3 className="text-xl font-semibold text-gray-900 mb-2">
                    {farm.farm_name}
                  </h3>
                  <p className="text-sm text-gray-600 mb-4">
                    {farm.farm_description}
                  </p>
                  <div className="flex items-center text-sm text-gray-500 mb-2">
                    <svg
                      className="h-4 w-4 mr-2"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"
                      />
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"
                      />
                    </svg>
                    {farm.farm_location}
                  </div>
                  
                  <div className="grid grid-cols-3 gap-4 mt-4 pt-4 border-t border-gray-200">
                    <div className="text-center">
                      <p className="text-2xl font-bold text-blue-600">
                        {farm.mcu_count}
                      </p>
                      <p className="text-xs text-gray-500">MCUs</p>
                    </div>
                    <div className="text-center">
                      <p className="text-2xl font-bold text-green-600">
                        {farm.online_mcu_count}
                      </p>
                      <p className="text-xs text-gray-500">Online</p>
                    </div>
                    <div className="text-center">
                      <p className="text-2xl font-bold text-purple-600">
                        {farm.survey_point_count}
                      </p>
                      <p className="text-xs text-gray-500">Points</p>
                    </div>
                  </div>

                  <div className="mt-4 flex items-center justify-between text-sm">
                    <span className="text-gray-500">
                      Role: <span className="font-medium text-gray-700">{farm.user_role}</span>
                    </span>
                    <span className="text-gray-500">
                      {new Date(farm.created_at).toLocaleDateString()}
                    </span>
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

export default HomePage;
