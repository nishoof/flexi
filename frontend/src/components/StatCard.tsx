function StatCard({ title, value }: { title: string; value: number }) {
  return (
    <div className="grow aspect-2/1 bg-(--background-light) p-4 rounded-lg">
      <h2 className="text-lg font-semibold mb-2">{title}</h2>
      <p className="text-3xl font-bold">${value}</p>
    </div>
  );
}

export default StatCard;
