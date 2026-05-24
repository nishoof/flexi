import React, { useEffect } from 'react';
import AddEntryModal from '../components/AddEntryModal';
import EditBudgetModal from '../components/EditBudgetModal';
import EntriesTable from '../components/EntriesTable';
import Login from '../components/Login';
import StatCard from '../components/StatCard';
import { getBudget, getEntries, type Budget, type Entry } from '../lib/api';
import { calculateStats } from '../lib/stats';

export default function OverviewPage() {
  const [isAddEntryModalOpen, setIsAddEntryModalOpen] = React.useState(false);
  const [isEditBudgetModalOpen, setIsEditBudgetModalOpen] = React.useState(false);
  const [budget, setBudget] = React.useState<Budget | null>(null);
  const [entries, setEntries] = React.useState<Entry[]>([]);

  const stats = calculateStats(entries);

  /** Gets the latest entries from the API and updates the state */
  const refreshEntries = React.useEffectEvent(async () => {
    const entries = await getEntries();
    setEntries(entries || []);
  });

  /** Gets the latest budget from the API and updates the state */
  const refreshBudget = React.useEffectEvent(async () => {
    const budget = await getBudget();
    setBudget(budget);
  });

  /** On component mount, fetch the initial data from the API */
  useEffect(() => {
    refreshEntries();
    refreshBudget();
  }, []);

  return (
    <div className="flex flex-col gap-4">
      <div className="max-w-3xs">
        <Login onSuccessfulLogin={() => {
          refreshEntries();
          refreshBudget();
        }} />
      </div>

      <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
        <StatCard title="Flexi Remaining" value={stats.currentFlexiRemaining} />
        <StatCard title="Avg Daily Spend" value={stats.dailySpend} />
        <StatCard title="Ending Projection" value={stats.endingProjection} />
        <StatCard title="Remaining per Day" value={stats.remainingPerDay} />
      </div>

      <EditBudgetModal
        key={budget?.holidays.join(',')} // Force remount when holidays change
        isOpen={isEditBudgetModalOpen}
        close={() => setIsEditBudgetModalOpen(false)}
        onBudgetUpdated={refreshBudget}
        initialBudget={budget}
      />

      <button
        type="button"
        onClick={() => setIsEditBudgetModalOpen(true)}
        className="w-full px-4 py-2 bg-(--accent) rounded-lg hover:bg-(--accent-dark) font-medium"
      >
        Edit Budget
      </button>

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
