import React, { useState } from 'react';
import { LayoutDashboard, Wifi, WifiOff, MapPin, Activity, Clock, Calendar, ChevronDown, ChevronRight, Users } from 'lucide-react';

// TypeScript interfaces
interface Device {
  id: string;
  name: string;
  status: 'active' | 'inactive';
  lastData: string;
  uptime: string;
  createdAt: string;
}

interface SurveyPoint {
  id: string;
  name: string;
  location: string;
  devices: Device[];
}

interface MCU {
  id: string;
  name: string;
  surveyPoint: SurveyPoint;
}

interface Farm {
  id: number;
  name: string;
  managers: string[];
  mcu: MCU;
}

// Dữ liệu mẫu
const mockData: { farms: Farm[] } = {
  farms: [
    {
      id: 1,
      name: "Nông trại Hòa Bình",
      managers: ["Nguyễn Văn A", "Trần Thị B"],
      mcu: {
        id: "MCU001",
        name: "MCU Chính - Khu A",
        surveyPoint: {
          id: "SP001",
          name: "Điểm khảo sát 1",
          location: "Khu vực phía Đông",
          devices: [
            {
              id: "DEV001",
              name: "Cảm biến nhiệt độ",
              status: "active",
              lastData: "28.5°C",
              uptime: "48 giờ 23 phút",
              createdAt: "2024-01-15"
            },
            {
              id: "DEV002",
              name: "Cảm biến độ ẩm",
              status: "active",
              lastData: "65%",
              uptime: "48 giờ 23 phút",
              createdAt: "2024-01-15"
            },
            {
              id: "DEV003",
              name: "Cảm biến pH đất",
              status: "inactive",
              lastData: "6.8",
              uptime: "0 giờ",
              createdAt: "2024-01-16"
            }
          ]
        }
      }
    },
    {
      id: 2,
      name: "Nông trại Xanh",
      managers: ["Lê Văn C"],
      mcu: {
        id: "MCU002",
        name: "MCU Chính - Khu B",
        surveyPoint: {
          id: "SP002",
          name: "Điểm khảo sát 2",
          location: "Khu vực trung tâm",
          devices: [
            {
              id: "DEV004",
              name: "Cảm biến ánh sáng",
              status: "active",
              lastData: "12000 lux",
              uptime: "120 giờ 45 phút",
              createdAt: "2024-01-10"
            },
            {
              id: "DEV005",
              name: "Máy bơm nước",
              status: "active",
              lastData: "ON - 2.5L/phút",
              uptime: "24 giờ 12 phút",
              createdAt: "2024-01-12"
            }
          ]
        }
      }
    }
  ]
};

const Dashboard: React.FC = () => {
  const [selectedFarm, setSelectedFarm] = useState<number | null>(null);
  const [expandedFarms, setExpandedFarms] = useState<Record<number, boolean>>({});

  const toggleFarm = (farmId: number): void => {
    setExpandedFarms(prev => ({
      ...prev,
      [farmId]: !prev[farmId]
    }));
  };

  const getStatusColor = (status: 'active' | 'inactive'): string => {
    return status === 'active' ? 'text-green-500' : 'text-red-500';
  };

  const getStatusBg = (status: 'active' | 'inactive'): string => {
    return status === 'active' ? 'bg-green-100' : 'bg-red-100';
  };

  // Thống kê tổng quan
  const totalFarms = mockData.farms.length;
  const totalDevices = mockData.farms.reduce((sum, farm) => 
    sum + farm.mcu.surveyPoint.devices.length, 0
  );
  const activeDevices = mockData.farms.reduce((sum, farm) => 
    sum + farm.mcu.surveyPoint.devices.filter(d => d.status === 'active').length, 0
  );

  return (
    <div className="min-h-screen bg-gradient-to-br from-green-50 to-blue-50">
      {/* Header */}
      <div className="bg-white shadow-md border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <LayoutDashboard className="w-8 h-8 text-green-600" />
              <h1 className="text-2xl font-bold text-gray-800">Quản Lý Nông Trại</h1>
            </div>
            <div className="text-sm text-gray-600">
              {new Date().toLocaleDateString('vi-VN', { 
                weekday: 'long', 
                year: 'numeric', 
                month: 'long', 
                day: 'numeric' 
              })}
            </div>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-6 py-6">
        {/* Thống kê tổng quan */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
          <div className="bg-white rounded-lg shadow-md p-6 border-l-4 border-green-500">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-gray-600 text-sm font-medium">Tổng số nông trại</p>
                <p className="text-3xl font-bold text-gray-800 mt-2">{totalFarms}</p>
              </div>
              <div className="bg-green-100 rounded-full p-3">
                <LayoutDashboard className="w-8 h-8 text-green-600" />
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow-md p-6 border-l-4 border-blue-500">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-gray-600 text-sm font-medium">Tổng thiết bị</p>
                <p className="text-3xl font-bold text-gray-800 mt-2">{totalDevices}</p>
              </div>
              <div className="bg-blue-100 rounded-full p-3">
                <Activity className="w-8 h-8 text-blue-600" />
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow-md p-6 border-l-4 border-green-500">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-gray-600 text-sm font-medium">Thiết bị hoạt động</p>
                <p className="text-3xl font-bold text-gray-800 mt-2">
                  {activeDevices}/{totalDevices}
                </p>
              </div>
              <div className="bg-green-100 rounded-full p-3">
                <Wifi className="w-8 h-8 text-green-600" />
              </div>
            </div>
          </div>
        </div>

        {/* Danh sách nông trại */}
        <div className="bg-white rounded-lg shadow-md p-6">
          <h2 className="text-xl font-bold text-gray-800 mb-4 flex items-center gap-2">
            <LayoutDashboard className="w-6 h-6 text-green-600" />
            Danh sách nông trại
          </h2>

          <div className="space-y-4">
            {mockData.farms.map(farm => (
              <div key={farm.id} className="border border-gray-200 rounded-lg overflow-hidden">
                {/* Farm Header */}
                <div 
                  className="bg-gradient-to-r from-green-50 to-blue-50 p-4 cursor-pointer hover:bg-green-100 transition-colors"
                  onClick={() => toggleFarm(farm.id)}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      {expandedFarms[farm.id] ? 
                        <ChevronDown className="w-5 h-5 text-gray-600" /> : 
                        <ChevronRight className="w-5 h-5 text-gray-600" />
                      }
                      <h3 className="text-lg font-bold text-gray-800">{farm.name}</h3>
                    </div>
                    <div className="flex items-center gap-2 text-sm text-gray-600">
                      <Users className="w-4 h-4" />
                      <span>{farm.managers.join(", ")}</span>
                    </div>
                  </div>
                </div>

                {/* Farm Content */}
                {expandedFarms[farm.id] && (
                  <div className="p-4 bg-white">
                    {/* MCU Info */}
                    <div className="mb-4 p-3 bg-gray-50 rounded-lg">
                      <h4 className="font-semibold text-gray-700 mb-2">MCU Chính: {farm.mcu.name}</h4>
                      <div className="flex items-center gap-2 text-sm text-gray-600">
                        <MapPin className="w-4 h-4" />
                        <span>{farm.mcu.surveyPoint.name} - {farm.mcu.surveyPoint.location}</span>
                      </div>
                    </div>

                    {/* Devices Table */}
                    <div className="overflow-x-auto">
                      <table className="w-full">
                        <thead>
                          <tr className="bg-gray-100 border-b border-gray-200">
                            <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700">Tên thiết bị</th>
                            <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700">Trạng thái</th>
                            <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700">Dữ liệu cuối</th>
                            <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700">Thời gian hoạt động</th>
                            <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700">Ngày thêm</th>
                          </tr>
                        </thead>
                        <tbody>
                          {farm.mcu.surveyPoint.devices.map(device => (
                            <tr key={device.id} className="border-b border-gray-100 hover:bg-gray-50 transition-colors">
                              <td className="px-4 py-3">
                                <div className="font-medium text-gray-800">{device.name}</div>
                                <div className="text-xs text-gray-500">{device.id}</div>
                              </td>
                              <td className="px-4 py-3">
                                <div className="flex items-center gap-2">
                                  {device.status === 'active' ? 
                                    <Wifi className={`w-4 h-4 ${getStatusColor(device.status)}`} /> :
                                    <WifiOff className={`w-4 h-4 ${getStatusColor(device.status)}`} />
                                  }
                                  <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusBg(device.status)} ${getStatusColor(device.status)}`}>
                                    {device.status === 'active' ? 'Hoạt động' : 'Ngưng'}
                                  </span>
                                </div>
                              </td>
                              <td className="px-4 py-3">
                                <span className="font-medium text-gray-800">{device.lastData}</span>
                              </td>
                              <td className="px-4 py-3">
                                <div className="flex items-center gap-2 text-gray-700">
                                  <Clock className="w-4 h-4" />
                                  <span>{device.uptime}</span>
                                </div>
                              </td>
                              <td className="px-4 py-3">
                                <div className="flex items-center gap-2 text-gray-700">
                                  <Calendar className="w-4 h-4" />
                                  <span>{new Date(device.createdAt).toLocaleDateString('vi-VN')}</span>
                                </div>
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
