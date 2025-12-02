import { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { mcuApi } from '../api/mcuApi';
import type { CreateMCURequest, ApiErrorResponse } from '@/core/types';

const CreateMCUPage = () => {
  const { farmId } = useParams<{ farmId: string }>();
  const navigate = useNavigate();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [formData, setFormData] = useState<CreateMCURequest>({
    farm_id: farmId || '',
    mcu_code: '',
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!farmId) {
      setError('Farm ID is required');
      return;
    }

    try {
      setIsLoading(true);
      setError(null);
      
      await mcuApi.createMCU(formData);
      
      // Navigate back to MCU list
      navigate(`/farm/${farmId}/mcus`, { replace: true });
    } catch (err) {
      const error = err as ApiErrorResponse;
      setError(error.response?.data?.message || error.message || 'Failed to create MCU');
      console.error('Error creating MCU:', err);
    } finally {
      setIsLoading(false);
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    });
  };

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="bg-white rounded-lg shadow p-6">
          <div className="mb-6">
            <button
              onClick={() => navigate(`/farm/${farmId}/mcus`)}
              className="flex items-center text-gray-600 hover:text-gray-900"
            >
              <svg className="w-5 h-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
              </svg>
              Back to MCU List
            </button>
          </div>

          <h1 className="text-2xl font-bold text-gray-900 mb-6">Add New MCU</h1>
          
          {error && (
            <div className="mb-4 p-4 bg-red-50 border border-red-200 rounded-lg">
              <p className="text-red-600">{error}</p>
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-6">
            <div>
              <label htmlFor="mcu_code" className="block text-sm font-medium text-gray-700 mb-2">
                MCU Code *
              </label>
              <input
                type="text"
                id="mcu_code"
                name="mcu_code"
                required
                value={formData.mcu_code}
                onChange={handleChange}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="Enter MCU code (e.g., 123456)"
              />
              <p className="mt-1 text-sm text-gray-500">
                Enter the unique code from your MCU device
              </p>
            </div>

            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
              <div className="flex">
                <svg
                  className="h-5 w-5 text-blue-400 mt-0.5"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
                <div className="ml-3">
                  <h3 className="text-sm font-medium text-blue-800">Before adding MCU</h3>
                  <div className="mt-2 text-sm text-blue-700">
                    <ul className="list-disc list-inside space-y-1">
                      <li>Make sure your MCU device is powered on</li>
                      <li>The MCU code must match the device configuration</li>
                      <li>Each MCU code can only be added once</li>
                    </ul>
                  </div>
                </div>
              </div>
            </div>

            <div className="flex gap-4">
              <button
                type="button"
                onClick={() => navigate(`/farm/${farmId}/mcus`)}
                className="flex-1 px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
                disabled={isLoading}
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={isLoading}
                className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isLoading ? 'Adding...' : 'Add MCU'}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
};

export default CreateMCUPage;
