import { useContext } from "react";
import { AuthContext } from "@/core/contexts/AuthProvider";
import type { AuthContextType } from "@/core/contexts/AuthProvider";

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
