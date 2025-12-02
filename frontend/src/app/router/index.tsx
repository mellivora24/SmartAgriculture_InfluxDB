import { createBrowserRouter, Navigate } from "react-router-dom";

import LoginPage from "@/features/auth/pages/LoginPage";
import RegisterPage from "@/features/auth/pages/RegisterPage";
import DashboardPage from "@/features/dashboard/pages/DashboardPage";
import SurveyListPage from "@/features/surveyPoint/pages/SurveyListPage";
import SwitchHistoryPage from "@/features/deviceCommand/pages/SwitchHistoryPage";
import SensorHistoryPage from "@/features/sensorData/pages/SensorHistoryPage";
import OnboardingPage from "@/features/onboarding/pages/OnboardingPage";
import AccountSettingsPage from "@/features/account/pages/AccountSettingsPage";
import ProtectedRoute from "@/shared/components/ProtectedRoute";
import GuestRoute from "@/shared/components/GuestRoute";

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
        <DashboardPage />
      </ProtectedRoute>
    ),
  },
  {
    path: "/survey",
    element: (
      <ProtectedRoute>
        <SurveyListPage />
      </ProtectedRoute>
    ),
  },
  {
    path: "/history/switch",
    element: (
      <ProtectedRoute>
        <SwitchHistoryPage />
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
    path: "/onboarding",
    element: (
      <ProtectedRoute>
        <OnboardingPage />
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
    path: "*",
    element: <Navigate to="/" replace />,
  },
]);
