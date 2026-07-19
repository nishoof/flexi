import React, { useEffect, useState } from 'react';
import AddEntryModal from '../components/AddEntryModal';
import EditTermModal from '../components/EditTermModal';
import EntriesTable from '../components/EntriesTable';
import LoadingView from '../components/LoadingView';
import NewTermModal from '../components/NewTermModal';
import SignInScreen from '../components/SignInScreen';
import StatCard from '../components/StatCard';
import TermPicker from '../components/TermPicker';
import {
  activateTerm,
  getEntries,
  getTerms,
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
  const [isNewTermModalOpen, setIsNewTermModalOpen] = useState(false);
  const [newTermModalKey, setNewTermModalKey] = useState(0);
  const [terms, setTerms] = useState<Term[]>([]);
  const [term, setTerm] = useState<Term | null>(null);
  const [entries, setEntries] = useState<Entry[]>([]);

  const stats = calculateStats(entries, term);

  const handleUnauthorized = React.useEffectEvent(() => {
    // Drop any data that belonged to the previous session so the sign-in
    // screen cannot flash stale stats or entries.
    setTerms([]);
    setTerm(null);
    setEntries([]);
    setAuthStatus('unauthenticated');
  });

  const refreshDashboard = React.useEffectEvent(async () => {
    try {
      const [fetchedTerms, fetchedEntries] = await Promise.all([
        getTerms(),
        getEntries(),
      ]);
      const activeTerm = fetchedTerms.find((t) => t.isActive);
      if (!activeTerm) {
        throw new Error('No active term');
      }
      setTerms(fetchedTerms);
      setTerm(activeTerm);
      setEntries(fetchedEntries);
    } catch (error) {
      if (isAuthError(error)) {
        handleUnauthorized();
      }
    }
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
      const fetchedTerms = await getTerms();
      const activeTerm = fetchedTerms.find((t) => t.isActive);
      if (!activeTerm) {
        throw new Error('No active term');
      }
      setTerms(fetchedTerms);
      setTerm(activeTerm);
    } catch (error) {
      if (isAuthError(error)) {
        handleUnauthorized();
      }
    }
  });

  // Fetches terms and entries in parallel, then reveals the dashboard in one step
  // so the loading spinner covers both requests (~max of the two, not their sum).
  const initialLoad = React.useEffectEvent(async () => {
    try {
      // also a session probe: 401 without a cookie
      const [fetchedTerms, fetchedEntries] = await Promise.all([
        getTerms(),
        getEntries(),
      ]);
      const activeTerm = fetchedTerms.find((t) => t.isActive);
      if (!activeTerm) {
        throw new Error('No active term');
      }
      setTerms(fetchedTerms);
      setTerm(activeTerm);
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

  const handleSelectTerm = React.useEffectEvent(async (nextTerm: Term) => {
    if (nextTerm.isActive) {
      return;
    }
    try {
      await activateTerm(nextTerm.id);
      await refreshDashboard();
    } catch (error) {
      if (isAuthError(error)) {
        handleUnauthorized();
      }
    }
  });

  useEffect(() => {
    initialLoad();
  }, []);

  if (authStatus === 'unauthenticated') {
    return <SignInScreen onSuccessfulLogin={handleSuccessfulLogin} />;
  }

  if (authStatus === 'loading' || term === null) {
    return <LoadingView />;
  }

  return (
    <div className="flex flex-col gap-4">
      <TermPicker
        terms={terms}
        activeTerm={term}
        onSelectTerm={handleSelectTerm}
        onNewTerm={() => {
          setNewTermModalKey((key) => key + 1);
          setIsNewTermModalOpen(true);
        }}
      />

      <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
        <StatCard title="Flexi Remaining" value={stats.currentFlexiRemaining} />
        <StatCard title="Avg Daily Spend" value={stats.dailySpend} />
        <StatCard title="Ending Projection" value={stats.endingProjection} />
        <StatCard title="Remaining per Day" value={stats.remainingPerDay} />
      </div>

      <EditTermModal
        key={`${term.id}-${term.daysOff.join(',')}-${term.endDate}`} // Force remount when term settings change
        isOpen={isEditTermModalOpen}
        close={() => setIsEditTermModalOpen(false)}
        onTermUpdated={refreshTerm}
        initialTerm={term}
        onUnauthorized={handleUnauthorized}
      />

      <NewTermModal
        key={newTermModalKey} // Force remount so the form starts blank each open
        isOpen={isNewTermModalOpen}
        close={() => setIsNewTermModalOpen(false)}
        onTermCreated={refreshDashboard}
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
