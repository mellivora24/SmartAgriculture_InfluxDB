import { createBrowserRouter } from "react-router-dom";

import LoginPage from "@/features/auth/pages/LoginPage";
import RegisterPage from "@/features/auth/pages/RegisterPage";
import DashboardPage from "@/features/dashboard/pages/DashboardPage";
import SurveyListPage from "@/features/surveyPoint/pages/SurveyListPage";
import SwitchHistoryPage from "@/features/deviceCommand/pages/SwitchHistoryPage";
import SensorHistoryPage from "@/features/sensorData/pages/SensorHistoryPage";
import OnboardingPage from "@/features/onboarding/pages/OnboardingPage";
import AccountSettingsPage from "@/features/account/pages/AccountSettingsPage";


export const router = createBrowserRouter([
  { path: "/", element: <DashboardPage /> },
  { path: "/survey", element: <SurveyListPage /> },
  { path: "/history/switch", element: <SwitchHistoryPage /> },
  { path: "/history/sensor", element: <SensorHistoryPage /> },
  { path: "/login", element: <LoginPage /> },
  { path: "/register", element: <RegisterPage /> },
  { path: "/onboarding", element: <OnboardingPage /> },
  { path: "/account/settings", element: <AccountSettingsPage /> }
]);
