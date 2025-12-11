// src/app/router/index.tsx
import { createBrowserRouter, Navigate } from "react-router-dom";

import LoginPage from "@/features/auth/pages/LoginPage";
import RegisterPage from "@/features/auth/pages/RegisterPage";
import HomePage from "@/features/farm/pages/HomePage";
import MCUListPage from "@/features/mcu/pages/MCUListPage";
import SurveyPointListPage from "@/features/surveyPoint/pages/SurveyPointListPage";
import DashboardPage from "@/features/dashboard/pages/DashboardPage";
import SensorHistoryPage from "@/features/sensorData/pages/SensorHistoryPage";
import AccountSettingsPage from "@/features/account/pages/AccountSettingsPage";
import ProtectedRoute from "@/shared/components/ProtectedRoute";
import GuestRoute from "@/shared/components/GuestRoute";
import CreateFarmPage from '@/features/farm/pages/CreateFarmPage';
import CreateMCUPage from '@/features/mcu/pages/CreateMCUPage';
import CreateSurveyPointPage from '@/features/surveyPoint/pages/CreateSurveyPointPage';
import CommandHistoryPage from '@/features/deviceCommand/pages/CommandHistoryPage';
import ThresholdSettingsPage from "@/features/threshold/pages/ThresholdSettingsPage";

export const router = createBrowserRouter([
  {
    path: "/login",
    element: (
      <GuestRoute>
        <LoginPage />
      </GuestRoute>
    ),
  },
  {
    path: "/register",
    element: (
      <GuestRoute>
        <RegisterPage />
      </GuestRoute>
    ),
  },
  {
    path: "/",
    element: (
      <ProtectedRoute>
        <HomePage />
      </ProtectedRoute>
    ),
  },
  {
    path: "/farm/:farmId/mcus",
    element: (
      <ProtectedRoute>
        <MCUListPage />
      </ProtectedRoute>
    ),
  },
  {
    path: '/farm/:farmId/mcus/create',
    element: (
      <ProtectedRoute>
        <CreateMCUPage />
      </ProtectedRoute>
    ),
  },
  {
    path: '/farms/create',
    element: (
      <ProtectedRoute>
        <CreateFarmPage />
      </ProtectedRoute>
    ),
  },
  {
    path: "/farm/:farmId/mcu/:mcuId/survey-points",
    element: (
      <ProtectedRoute>
        <SurveyPointListPage />
      </ProtectedRoute>
    ),
  },
  {
    path: '/farm/:farmId/mcu/:mcuId/survey-points/create',
    element: (
      <ProtectedRoute>
        <CreateSurveyPointPage />
      </ProtectedRoute>
    ),
  },
  {
    path: "/farm/:farmId/mcu/:mcuId/dashboard/:surveyPointId",
    element: (
      <ProtectedRoute>
        <DashboardPage />
      </ProtectedRoute>
    ),
  },
  {
    path: "/dashboard/:surveyPointId/sensor-history",
    element: (
      <ProtectedRoute>
        <SensorHistoryPage />
      </ProtectedRoute>
    ),
  },
  {
    path: "/history/sensor",
    element: (
      <ProtectedRoute>
        <SensorHistoryPage />
      </ProtectedRoute>
    ),
  },
  {
    path: "/account/settings",
    element: (
      <ProtectedRoute>
        <AccountSettingsPage />
      </ProtectedRoute>
    ),
  },
  {
    path: '/dashboard/:surveyPointId/control-history',
    element: (
      <ProtectedRoute>
        <CommandHistoryPage />
      </ProtectedRoute>
    ),
  },
  {
    path: "*",
    element: <Navigate to="/" replace />,
  },
  {
    path: '/dashboard/:surveyPointId/threshold-settings',
    element: (
      <ProtectedRoute>
        <ThresholdSettingsPage />
      </ProtectedRoute>
    ),
  }
]);
