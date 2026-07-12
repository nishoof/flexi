import React, { useEffect, useState } from 'react';
import AddEntryModal from '../components/AddEntryModal';
import EditTermModal from '../components/EditTermModal';
import EntriesTable from '../components/EntriesTable';
import LoadingView from '../components/LoadingView';
import SignInScreen from '../components/SignInScreen';
import StatCard from '../components/StatCard';
import {
  getEntries,
  getTerm,
  isAuthError,
  type Entry,
  type Term,
} from '../lib/api';
import { calculateStats } from '../lib/stats';

type AuthStatus = 'unauthenticated' | 'loading' | 'authenticated';

export default function OverviewPage() {
  const [authStatus, setAuthStatus] = useState<AuthStatus>('loading');
  const [isAddEntryModalOpen, setIsAddEntryModalOpen] = useState(false);
  const [isEditTermModalOpen, setIsEditTermModalOpen] = useState(false);
  const [term, setTerm] = useState<Term | null>(null);
  const [entries, setEntries] = useState<Entry[]>([]);

  const stats = calculateStats(entries);

  const handleUnauthorized = React.useEffectEvent(() => {
    // Drop any data that belonged to the previous session so the sign-in
    // screen cannot flash stale stats or entries.
    setTerm(null);
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

  const refreshTerm = React.useEffectEvent(async () => {
    try {
      const fetchedTerm = await getTerm();
      setTerm(fetchedTerm);
    } catch (error) {
      if (isAuthError(error)) {
        handleUnauthorized();
      }
    }
  });

  // Fetches term and entries in parallel, then reveals the dashboard in one step
  // so the loading spinner covers both requests (~max of the two, not their sum).
  const initialLoad = React.useEffectEvent(async () => {
    try {
      // also a session probe: 401 without a cookie
      const [fetchedTerm, fetchedEntries] = await Promise.all([
        getTerm(),
        getEntries(),
      ]);
      setTerm(fetchedTerm);
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

  if (authStatus === 'unauthenticated') {
    return <SignInScreen onSuccessfulLogin={handleSuccessfulLogin} />;
  }

  if (authStatus === 'loading') {
    return <LoadingView />;
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
        <StatCard title="Flexi Remaining" value={stats.currentFlexiRemaining} />
        <StatCard title="Avg Daily Spend" value={stats.dailySpend} />
        <StatCard title="Ending Projection" value={stats.endingProjection} />
        <StatCard title="Remaining per Day" value={stats.remainingPerDay} />
      </div>

      <EditTermModal
        key={term?.daysOff.join(',')} // Force remount when days off change
        isOpen={isEditTermModalOpen}
        close={() => setIsEditTermModalOpen(false)}
        onTermUpdated={refreshTerm}
        initialTerm={term}
        onUnauthorized={handleUnauthorized}
      />

      <button
        type="button"
        onClick={() => setIsEditTermModalOpen(true)}
        className="w-full px-4 py-2 bg-(--accent) rounded-lg hover:bg-(--accent-dark) font-medium"
      >
        Edit Term
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
