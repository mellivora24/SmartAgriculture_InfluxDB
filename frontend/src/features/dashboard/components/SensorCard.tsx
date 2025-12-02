interface SensorCardProps {
  title: string;
  value: number | null;
  unit: string;
  icon: string;
  color: string;
  bgColor: string;
}

const SensorCard = ({
  title,
  value,
  unit,
  icon,
  color,
  bgColor,
}: SensorCardProps) => {
  const displayValue = value !== null ? value.toFixed(1) : '--';
  const percentage = value !== null ? Math.min((value / 100) * 100, 100) : 0;

  return (
    <div className="bg-white rounded-lg shadow-md overflow-hidden hover:shadow-lg transition-shadow duration-300">
      <div className={`${bgColor} px-6 py-4`}>
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <span className="text-3xl">{icon}</span>
            <h3 className="text-lg font-semibold text-gray-900">{title}</h3>
          </div>
        </div>
      </div>
      <div className="px-6 py-6">
        <div className="flex items-baseline">
          <span className={`text-4xl font-bold ${color}`}>
            {displayValue}
          </span>
          <span className="ml-2 text-xl text-gray-600">{unit}</span>
        </div>
        <div className="mt-4">
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div
              className={`${bgColor} h-2 rounded-full transition-all duration-500`}
              style={{ width: `${percentage}%` }}
            />
          </div>
        </div>
      </div>
    </div>
  );
};

export default SensorCard;
