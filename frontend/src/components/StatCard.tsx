export default function StatCard({ title, value }: Readonly<{ title: string; value: number }>) {
  const displayValue = `${value < 0 ? '-' : ''}$${Math.abs(value).toFixed(2)}`;
  return (
    <div className="flex-1 grow aspect-2/1 bg-(--background-light) p-4 rounded-lg border border-(--border)">
      <h2 className="text-lg font-semibold mb-2">{title}</h2>
      <p className="text-3xl font-bold">{displayValue}</p>
    </div>
  );
}
