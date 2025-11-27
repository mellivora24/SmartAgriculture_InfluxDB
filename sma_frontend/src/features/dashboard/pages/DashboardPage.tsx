import SensorCard from "../components/SensorCard";

export default function DashboardPage() {
  return (
    <div>
      <h1>Dashboard</h1>
      <SensorCard label="Nhiệt độ" value="25°C" />
    </div>
  );
}
