import { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { sensorDataApi } from "../api/sensorDataApi";
import type { SensorDataPoint } from "@/core/types";

interface GroupedSensorData {
  time: string;
  temperature: number | null;
  humidity: number | null;
  soil_moisture: number | null;
  light: number | null;
  survey_point_name: string;
  mcu_code: string;
}

const SensorHistoryPage = () => {
  const { surveyPointId } = useParams<{ surveyPointId: string }>();
  const navigate = useNavigate();

  const [records, setRecords] = useState<SensorDataPoint[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  // chỉ còn limit
  const [limit, setLimit] = useState(50);

  useEffect(() => {
    if (!surveyPointId) return;

    const loadData = async () => {
      try {
        setIsLoading(true);
        const response = await sensorDataApi.getSensorData({
          survey_point_id: surveyPointId,
          limit,
        });
        setRecords(response.data || []);
      } catch (error) {
        console.error("Failed to load sensor history:", error);
      } finally {
        setIsLoading(false);
      }
    };

    loadData();
  }, [surveyPointId, limit]);

  const groupedRecords = records.reduce<Record<string, GroupedSensorData>>(
    (acc, record) => {
      const time = record._time;

      if (!acc[time]) {
        acc[time] = {
          time,
          temperature: null,
          humidity: null,
          soil_moisture: null,
          light: null,
          survey_point_name: record.survey_point_name,
          mcu_code: record.mcu_code,
        };
      }

      const field = record._field as keyof Omit<
        GroupedSensorData,
        "time" | "survey_point_name" | "mcu_code"
      >;

      acc[time][field] = record._value;

      return acc;
    },
    {}
  );

  const sortedRecords = Object.values(groupedRecords).sort(
    (a, b) => new Date(b.time).getTime() - new Date(a.time).getTime()
  );

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <button
            onClick={() => navigate(-1)}
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
            Back
          </button>

          <h1 className="text-3xl font-bold text-gray-900">
            Sensor History
          </h1>
          <p className="mt-1 text-sm text-gray-500">
            View historical sensor data for this survey point
          </p>
        </div>
      </div>

      {/* Filters */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Limit
              </label>
              <select
                value={limit}
                onChange={(e) => setLimit(Number(e.target.value))}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value={10}>10 records</option>
                <option value={50}>50 records</option>
                <option value={100}>100 records</option>
              </select>
            </div>
          </div>
        </div>

        {/* Data Table */}
        <div className="bg-white rounded-lg shadow overflow-hidden">
          {isLoading ? (
            <div className="flex items-center justify-center h-64">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600" />
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                      Timestamp
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                      Temperature (°C)
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                      Humidity (%)
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                      Soil Moisture (%)
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                      Light (lux)
                    </th>
                  </tr>
                </thead>

                <tbody className="bg-white divide-y divide-gray-200">
                  {sortedRecords.length === 0 ? (
                    <tr>
                      <td
                        colSpan={6}
                        className="px-6 py-12 text-center text-gray-500"
                      >
                        No sensor data found
                      </td>
                    </tr>
                  ) : (
                    sortedRecords.map((record, index) => (
                      <tr
                        key={index}
                        className="hover:bg-gray-50 transition-colors"
                      >
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {new Date(record.time).toLocaleString()}
                        </td>

                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {record.temperature?.toFixed(1) ?? "--"}
                        </td>

                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {record.humidity?.toFixed(1) ?? "--"}
                        </td>

                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {record.soil_moisture?.toFixed(1) ?? "--"}
                        </td>

                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {record.light?.toFixed(0) ?? "--"}
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default SensorHistoryPage;
