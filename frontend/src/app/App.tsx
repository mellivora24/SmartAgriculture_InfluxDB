import { RouterProvider } from "react-router-dom";
import AuthProvider from "@/core/contexts/AuthProvider";
import { router } from "./router";
import { ErrorBoundary } from "@/shared/components/ErrorBoundary";

export default function App() {
  return (
    <ErrorBoundary>
      <AuthProvider>
        <RouterProvider router={router} />
      </AuthProvider>
    </ErrorBoundary>
  );
}
