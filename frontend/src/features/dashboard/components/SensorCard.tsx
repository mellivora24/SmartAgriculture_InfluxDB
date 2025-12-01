export default function SensorCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="card">
      <strong>{label}</strong>
      <p>{value}</p>
    </div>
  );
}
