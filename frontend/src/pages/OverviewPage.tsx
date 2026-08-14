import React, { useEffect, useState } from "react";
import AddEntryModal from "../components/AddEntryModal";
import EditTermModal from "../components/EditTermModal";
import EntriesTable from "../components/EntriesTable";
import LoadingView from "../components/LoadingView";
import NewTermModal from "../components/NewTermModal";
import SignInScreen from "../components/SignInScreen";
import StatCard from "../components/StatCard";
import TermPicker from "../components/TermPicker";
import Toast from "../components/Toast";
import {
  activateTerm,
  createEntry,
  getEntries,
  getTerms,
  isAuthError,
  updateTerm,
  type Entry,
  type Term,
} from "../lib/api";
import { calculateStats, zeroedStats } from "../lib/stats";

type AuthStatus =
  "unauthenticated" | "loading" | "load_failed" | "authenticated";

export default function OverviewPage() {
  const [authStatus, setAuthStatus] = useState<AuthStatus>("loading");
  const [isAddEntryModalOpen, setIsAddEntryModalOpen] = useState(false);
  const [isEditTermModalOpen, setIsEditTermModalOpen] = useState(false);
  const [isNewTermModalOpen, setIsNewTermModalOpen] = useState(false);
  const [newTermModalKey, setNewTermModalKey] = useState(0);
  const [terms, setTerms] = useState<Term[]>([]);
  const [term, setTerm] = useState<Term | null>(null);
  const [entries, setEntries] = useState<Entry[]>([]);
  const [toast, setToast] = useState<{ id: number; message: string } | null>(
    null,
  );

  const stats = term ? calculateStats(entries, term) : zeroedStats;

  const handleUnauthorized = React.useEffectEvent(() => {
    // Drop any data that belonged to the previous session so the sign-in
    // screen cannot flash stale stats or entries.
    setTerms([]);
    setTerm(null);
    setEntries([]);
    setToast(null);
    setAuthStatus("unauthenticated");
  });

  // Auth errors send the user to sign-in. Everything else gets a toast.
  const handleRequestError = React.useEffectEvent(
    (error: unknown, message: string) => {
      if (isAuthError(error)) {
        handleUnauthorized();
      } else {
        setToast({ id: Date.now(), message });
      }
    },
  );

  const refreshDashboard = React.useEffectEvent(async () => {
    try {
      const [fetchedTerms, fetchedEntries] = await Promise.all([
        getTerms(),
        getEntries(),
      ]);
      const activeTerm = fetchedTerms.find((t) => t.isActive);
      if (!activeTerm) {
        throw new Error("No active term");
      }
      setTerms(fetchedTerms);
      setTerm(activeTerm);
      setEntries(fetchedEntries);
    } catch (error) {
      handleRequestError(error, "Could not refresh dashboard");
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
        throw new Error("No active term");
      }
      setTerms(fetchedTerms);
      setTerm(activeTerm);
      setEntries(fetchedEntries);
      setToast(null);
      setAuthStatus("authenticated");
    } catch (error) {
      if (isAuthError(error)) {
        handleUnauthorized();
      } else {
        // The load_failed screen has its own message — no toast needed.
        setToast(null);
        setAuthStatus("load_failed");
      }
    }
  });

  const handleSuccessfulLogin = React.useEffectEvent(async () => {
    setToast(null);
    setAuthStatus("loading");
    await initialLoad();
  });

  const handleRetryLoad = React.useEffectEvent(async () => {
    setToast(null);
    setAuthStatus("loading");
    await initialLoad();
  });

  const handleAddEntry = React.useEffectEvent(
    async (amountRemaining: number, date: string) => {
      if (entries.some((entry) => entry.date === date)) {
        setToast({ id: Date.now(), message: "Could not save entry" });
        return;
      }

      const next: Entry = { amountRemaining, date };
      setEntries((current) => insertEntrySorted(current, next));

      try {
        await createEntry(amountRemaining, date);
      } catch (error) {
        setEntries((current) =>
          current.filter((entry) => entry.date !== date),
        );
        handleRequestError(error, "Could not save entry");
      }
    },
  );

  const handleEditTerm = React.useEffectEvent(async (next: Term) => {
    const previousTerm = term;
    const previousTerms = terms;

    setTerm(next);
    setTerms((current) =>
      current.map((t) => (t.id === next.id ? next : t)),
    );

    try {
      await updateTerm({
        name: next.name,
        startDate: next.startDate,
        endDate: next.endDate,
        startingAmount: next.startingAmount,
        daysOff: next.daysOff,
      });
    } catch (error) {
      setTerm(previousTerm);
      setTerms(previousTerms);
      handleRequestError(error, "Could not save term");
    }
  });

  const handleSelectTerm = React.useEffectEvent(async (nextTerm: Term) => {
    if (nextTerm.isActive) {
      return;
    }
    try {
      await activateTerm(nextTerm.id);
      await refreshDashboard();
    } catch (error) {
      handleRequestError(error, "Could not switch term");
    }
  });

  useEffect(() => {
    initialLoad();
  }, []);

  const toastElement = toast ? (
    <Toast
      key={toast.id}
      message={toast.message}
      onDismiss={() => setToast(null)}
    />
  ) : null;

  if (authStatus === "unauthenticated") {
    return (
      <>
        <SignInScreen onSuccessfulLogin={handleSuccessfulLogin} />
        {toastElement}
      </>
    );
  }

  if (authStatus === "loading") {
    return <LoadingView />;
  }

  if (authStatus === "load_failed" || term === null) {
    // load_failed: initial fetch failed. term === null: unexpected state — retry
    // instead of spinning forever.
    return (
      <div className="flex flex-col items-center justify-center gap-4 py-24">
        <p className="text-sm text-(--foreground)/70">
          Could not load your dashboard.
        </p>
        <button
          type="button"
          onClick={handleRetryLoad}
          className="px-4 py-2 bg-(--accent) rounded-lg hover:bg-(--accent-dark) font-medium"
        >
          Retry
        </button>
      </div>
    );
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
        key={`${term.id}-${term.daysOff.join(",")}-${term.startDate}-${term.endDate}-${term.startingAmount}`} // Force remount when term settings change
        isOpen={isEditTermModalOpen}
        close={() => setIsEditTermModalOpen(false)}
        onTermUpdated={handleEditTerm}
        initialTerm={term}
      />

      <NewTermModal
        key={newTermModalKey} // Force remount so the form starts blank each open
        isOpen={isNewTermModalOpen}
        close={() => setIsNewTermModalOpen(false)}
        onTermCreated={refreshDashboard}
        onFailure={(error) =>
          handleRequestError(error, "Could not create term")
        }
      />

      <button
        type="button"
        onClick={() => setIsEditTermModalOpen(true)}
        className="px-4 py-2 bg-(--accent) rounded-lg hover:bg-(--accent-dark) font-medium"
      >
        Edit Term
      </button>

      <div className="flex flex-col gap-2">
        <h1 className="text-xl font-semibold"> Entries </h1>

        <EntriesTable entries={entries} />

        <AddEntryModal
          isOpen={isAddEntryModalOpen}
          close={() => setIsAddEntryModalOpen(false)}
          onEntryAdded={handleAddEntry}
        />

        <button
          type="button"
          onClick={() => setIsAddEntryModalOpen(true)}
          className="px-4 py-2 bg-(--accent) rounded-lg hover:bg-(--accent-dark) font-medium"
        >
          Add Entry
        </button>
      </div>

      {toastElement}
    </div>
  );
}

/** Insert an entry so the list stays reverse-chronological (newest first). */
function insertEntrySorted(entries: Entry[], next: Entry): Entry[] {
  const insertAt = entries.findIndex((entry) => entry.date < next.date);
  if (insertAt === -1) {
    return [...entries, next];
  }
  return [...entries.slice(0, insertAt), next, ...entries.slice(insertAt)];
}
