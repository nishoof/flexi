import React, { useEffect } from 'react';
import AddEntryModal from '../components/AddEntryModal';
import EntriesTable from '../components/EntriesTable';
import Login from '../components/Login';
import StatCard from '../components/StatCard';
import { getEntries, type Entry } from '../lib/api';

export default function OverviewPage() {
  const [isAddEntryModalOpen, setIsAddEntryModalOpen] = React.useState(false);
  const [entries, setEntries] = React.useState<Entry[]>([]);

  /** Flexi the user currently has remaining based on their latest entry. This updates automatically when entries change */
  const flexiRemaining = React.useMemo(() => {
    return entries[0]?.amountRemaining ?? 0;  // entries are sorted with newest first, so the first entry is the latest
  }, [entries]);

  /** Get the latest entries from the API and update the state */
  const refreshEntries = React.useEffectEvent(async () => {
    const entries = await getEntries();
    setEntries(entries || []);
  });

  useEffect(() => {
    refreshEntries();
  }, []);

  return (
    <div className="flex flex-col gap-4">
      <div className="max-w-3xs">
        <Login onSuccessfulLogin={refreshEntries} />
      </div>

      <div className="flex space-x-4">
        <StatCard title="Flexi Remaining" value={flexiRemaining} />
        <StatCard title="Flexi Remaining" value={flexiRemaining} />
        <StatCard title="Flexi Remaining" value={flexiRemaining} />
        <StatCard title="Flexi Remaining" value={flexiRemaining} />
      </div>

      <div className="flex flex-col gap-2">
        <h1 className="text-xl font-semibold"> Entries </h1>

        <EntriesTable entries={entries} />

        <AddEntryModal
          isOpen={isAddEntryModalOpen}
          close={() => setIsAddEntryModalOpen(false)}
          onEntryAdded={refreshEntries}
        />

        <button
          type="button"
          onClick={() => setIsAddEntryModalOpen(true)}
          className="w-full px-4 py-2 bg-(--accent) rounded-lg hover:bg-(--accent-dark) font-medium"
        >
          Add Entry
        </button>
      </div>
    </div>
  );
}
