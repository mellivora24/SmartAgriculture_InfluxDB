import { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { surveyPointApi } from "../api/surveyPointApi";
import type { SurveyPointListItem } from "@/core/types";

const SurveyPointListPage = () => {
    const { farmId, mcuId } = useParams<{ farmId: string; mcuId: string }>();
    const [surveyPoints, setSurveyPoints] = useState<SurveyPointListItem[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const navigate = useNavigate();

    useEffect(() => {
        const loadSurveyPoints = async () => {
            if (!mcuId) return;

            try {
                setIsLoading(true);
                setError(null);
                const response = await surveyPointApi.getSurveyPointsByMCU(mcuId);
                // FIX: Đảm bảo luôn set array, không phải null
                setSurveyPoints(response.data || []);
            } catch (err) {
                setError("Failed to load survey points");
                console.error("Error loading survey points:", err);
            } finally {
                setIsLoading(false);
            }
        };

        loadSurveyPoints();
    }, [mcuId]);

    const handleSurveyPointClick = (surveyPointId: string) => {
        navigate(`/farm/${farmId}/mcu/${mcuId}/dashboard/${surveyPointId}`);
    };

    const handleCreateSurveyPoint = () => {
        navigate(`/farm/${farmId}/mcu/${mcuId}/survey-points/create`);
    };

    const getStatusColor = (status: string) => {
        switch (status) {
            case "connected":
                return "text-green-600 bg-green-100";
            case "connecting":
                return "text-yellow-600 bg-yellow-100";
            case "disconnected":
                return "text-red-600 bg-red-100";
            default:
                return "text-gray-600 bg-gray-100";
        }
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
                        onClick={() => navigate(`/farm/${farmId}/mcus`)}
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
                        Back to MCUs
                    </button>

                    <div className="flex justify-between items-center">
                        <div>
                            <h1 className="text-3xl font-bold text-gray-900">Survey Points</h1>
                            <p className="mt-1 text-sm text-gray-500">
                                Select a survey point to view its dashboard
                            </p>
                        </div>
                        <button
                            onClick={handleCreateSurveyPoint}
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
                            Add Survey Point
                        </button>
                    </div>
                </div>
            </div>

            {/* Survey Point List */}
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
                {surveyPoints.length === 0 ? (
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
                                d="M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l4.553 2.276A1 1 0 0021 18.382V7.618a1 1 0 00-.553-.894L15 4m0 13V4m0 0L9 7"
                            />
                        </svg>
                        <h3 className="mt-2 text-sm font-medium text-gray-900">
                            No survey points
                        </h3>
                        <p className="mt-1 text-sm text-gray-500">
                            This MCU doesn't have any survey points yet.
                        </p>
                        <div className="mt-6">
                            <button
                                onClick={handleCreateSurveyPoint}
                                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
                            >
                                Add First Survey Point
                            </button>
                        </div>
                    </div>
                ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                        {surveyPoints.map((point) => (
                            <div
                                key={point.survey_point_id}
                                onClick={() => handleSurveyPointClick(point.survey_point_id)}
                                className="bg-white rounded-lg shadow hover:shadow-lg transition-shadow cursor-pointer overflow-hidden"
                            >
                                <div className="p-6">
                                    <div className="flex items-center justify-between mb-3">
                                        <h3 className="text-xl font-semibold text-gray-900">
                                            {point.survey_point_name}
                                        </h3>
                                        <span
                                            className={`px-3 py-1 rounded-full text-xs font-medium ${getStatusColor(
                                                point.status
                                            )}`}
                                        >
                                            {point.status}
                                        </span>
                                    </div>

                                    <p className="text-sm text-gray-600 mb-4">
                                        {point.description}
                                    </p>

                                    <div className="pt-4 border-t border-gray-200 space-y-2">
                                        <div className="flex items-center text-xs text-gray-500">
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
                                                    d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
                                                />
                                            </svg>
                                            Created: {new Date(point.created_at).toLocaleDateString()}
                                        </div>
                                        <div className="flex items-center text-xs text-gray-500">
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
                                                    d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                                                />
                                            </svg>
                                            Updated: {new Date(point.updated_at).toLocaleDateString()}
                                        </div>
                                    </div>

                                    <div className="mt-4">
                                        <button className="w-full px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium">
                                            View Dashboard
                                        </button>
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

export default SurveyPointListPage;