import { type Entry } from '../lib/api';
import { formatDate, formatMoney } from '../lib/format';

interface EntriesTableProps {
  /** Array of entries to display. Can be an empty array */
  entries: Entry[];
}

/** Component to display the table of entries */
export default function EntriesTable({ entries }: Readonly<EntriesTableProps>) {
  const tableContent = entries.length ? entries.map((entry) => (
    <tr key={entry.date} className="border-t border-(--border)">
      <td className="px-4 py-2">{formatDate(entry.date)}</td>
      <td className="px-4 py-2">{formatMoney(entry.amountRemaining)}</td>
    </tr>
  )) : (
    <tr className="border-t border-(--border)">
      <td className="px-4 py-2 text-center text-(--foreground)/70" colSpan={2}>
        No entries yet — add your first entry below
      </td>
    </tr>
  );

  return (
    <div className="rounded-lg border border-(--border) overflow-hidden">
      <table className="w-full rounded-lg table-fixed">
        <thead className="bg-(--background-light)">
          <tr>
            <th className="px-4 py-2 text-left font-medium">Date</th>
            <th className="px-4 py-2 text-left font-medium">Flexi Remaining</th>
          </tr>
        </thead>
        <tbody className="bg-[#242424]">
          {tableContent}
        </tbody>
      </table>
    </div>
  );
}
