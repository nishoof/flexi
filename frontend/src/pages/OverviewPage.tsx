import React, { useEffect, useState } from 'react';
import AddEntryModal from '../components/AddEntryModal';
import EditBudgetModal from '../components/EditBudgetModal';
import EntriesTable from '../components/EntriesTable';
import LoadingView from '../components/LoadingView';
import SignInScreen from '../components/SignInScreen';
import StatCard from '../components/StatCard';
import {
  getBudget,
  getEntries,
  isAuthError,
  type Budget,
  type Entry,
} from '../lib/api';
import { calculateStats } from '../lib/stats';

type AuthStatus = 'loading' | 'authenticated' | 'unauthenticated';

export default function OverviewPage() {
  const [authStatus, setAuthStatus] = useState<AuthStatus>('loading');
  const [isAddEntryModalOpen, setIsAddEntryModalOpen] = useState(false);
  const [isEditBudgetModalOpen, setIsEditBudgetModalOpen] = useState(false);
  const [budget, setBudget] = useState<Budget | null>(null);
  const [entries, setEntries] = useState<Entry[]>([]);

  const stats = calculateStats(entries);

  const handleUnauthorized = React.useEffectEvent(() => {
    // Drop any data that belonged to the previous session so the sign-in
    // screen cannot flash stale stats or entries.
    setBudget(null);
    setEntries([]);
    setAuthStatus('unauthenticated');
  });

  const refreshEntries = React.useEffectEvent(async () => {
    try {
      const fetchedEntries = await getEntries();
      setEntries(fetchedEntries);
    } catch (error) {
      if (isAuthError(error)) {
        handleUnauthorized();
      }
    }
  });

  const refreshBudget = React.useEffectEvent(async () => {
    try {
      const fetchedBudget = await getBudget();
      setBudget(fetchedBudget);
    } catch (error) {
      if (isAuthError(error)) {
        handleUnauthorized();
      }
    }
  });

  // Fetches budget and entries in parallel, then reveals the dashboard in one step
  // so the loading spinner covers both requests (~max of the two, not their sum).
  const initialLoad = React.useEffectEvent(async () => {
    try {
      // getBudget doubles as a session probe: 401 without a cookie, and the
      // backend auto-creates a default budget for new users.
      const [fetchedBudget, fetchedEntries] = await Promise.all([
        getBudget(),
        getEntries(),
      ]);
      setBudget(fetchedBudget);
      setEntries(fetchedEntries);
      setAuthStatus('authenticated');
    } catch (error) {
      if (isAuthError(error)) {
        handleUnauthorized();
      } else {
        setAuthStatus('unauthenticated');
      }
    }
  });

  const handleSuccessfulLogin = React.useEffectEvent(async () => {
    setAuthStatus('loading');
    await initialLoad();
  });

  useEffect(() => {
    initialLoad();
  }, []);

  if (authStatus === 'loading') {
    return <LoadingView />;
  }

  if (authStatus === 'unauthenticated') {
    return <SignInScreen onSuccessfulLogin={handleSuccessfulLogin} />;
  }

  return (
    <div className="flex flex-col gap-4">
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
        onUnauthorized={handleUnauthorized}
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
          onUnauthorized={handleUnauthorized}
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
