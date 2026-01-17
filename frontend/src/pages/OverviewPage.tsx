import { GoogleLogin, type CredentialResponse } from '@react-oauth/google';
import React, { useEffect } from 'react';
import AddEntryModal from '../components/AddEntryModal';
import EntriesTable from '../components/EntriesTable';
import StatCard from '../components/StatCard';

export type Entry = {
  amountRemaining: number;
  date: string;
};

function OverviewPage() {
  const [isAddEntryModalOpen, setIsAddEntryModalOpen] = React.useState(false);
  const [entries, setEntries] = React.useState<Entry[]>([]);

  /** Flexi the user currently has remaining based on their latest entry. This updates automatically when entries change */
  const flexiRemaining = React.useMemo(() => {
    if (entries.length === 0) return 0;
    const latestEntry = entries[0];
    return latestEntry.amountRemaining;
  }, [entries]);

  /** Get the latest entries from the API and update the state */
  const refreshEntries = React.useCallback(async () => {
    const entries = await getEntries();
    setEntries(entries || []);
  }, []);

  useEffect(() => {
    void refreshEntries();
  }, [refreshEntries]);

  async function handleGoogleLogin(credentialResponse: CredentialResponse) {
    const credential = credentialResponse.credential;
    if (!credential) {
      console.error('No credential received from Google Login');
      return;
    }

    try {
      const apiUrl = getApiUrl();

      const response = await fetch(`${apiUrl}/auth`, {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ credential }),
      });

      if (!response.ok) {
        throw new Error('API request failed');
      }

      await refreshEntries();
    } catch (error) {
      console.error('Error logging in:', error);
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="max-w-3xs">
        <GoogleLogin
          onSuccess={handleGoogleLogin}
          onError={() => console.log('Login Failed')}
        />
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

async function getEntries(): Promise<Entry[] | null> {
  try {
    const apiUrl = getApiUrl();

    type ApiEntry = {
      amount_remaining: number;
      date: string;
    };

    // Fetch entries from the API
    // Entries are returned in descending order by date (newest first)
    const response = await fetch(`${apiUrl}/entries`, {
      method: 'GET',
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('API request failed');
    }

    const data: ApiEntry[] = await response.json();
    return data.map((entry) => ({
      amountRemaining: entry.amount_remaining,
      date: entry.date,
    }));
  } catch (error) {
    console.error('Error fetching entries:', error);
    return null;
  }
}

function getApiUrl(): string {
  const apiUrl = import.meta.env.VITE_API_URL;
  if (typeof apiUrl !== 'string') {
    throw new TypeError('API URL is not defined in environment variables.');
  }
  return apiUrl;
}

export default OverviewPage;
