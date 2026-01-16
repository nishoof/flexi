import { GoogleLogin, type CredentialResponse } from '@react-oauth/google';
import React, { useEffect } from 'react';
import AddEntryModal from '../components/AddEntryModal';
import StatCard from '../components/StatCard';

function OverviewPage() {
  const [isAddEntryModalOpen, setIsAddEntryModalOpen] = React.useState(false);
  const [flexi, setFlexi] = React.useState<number>(0);

  /** Get the latest flexi remaining value from the API and update the state */
  const refreshFlexi = React.useCallback(async () => {
    const remaining = await getFlexiRemaining();
    if (typeof remaining === 'number') {
      setFlexi(remaining);
    }
  }, []);

  useEffect(() => {
    void refreshFlexi();
  }, [refreshFlexi]);

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

      await refreshFlexi();
    } catch (error) {
      console.error('Error logging in:', error);
    }
  }

  return (
    <>
      <div className="mb-4 max-w-3xs">
        <GoogleLogin
          onSuccess={handleGoogleLogin}
          onError={() => console.log('Login Failed')}
        />
      </div>

      <div className="flex space-x-4">
        <StatCard title="Flexi Remaining" value={flexi} />
        <StatCard title="Flexi Remaining" value={flexi} />
        <StatCard title="Flexi Remaining" value={flexi} />
        <StatCard title="Flexi Remaining" value={flexi} />
      </div>

      <AddEntryModal
        isOpen={isAddEntryModalOpen}
        close={() => setIsAddEntryModalOpen(false)}
        onEntryAdded={refreshFlexi}
      />

      <button onClick={() => setIsAddEntryModalOpen(true)} >
        Open modal
      </button>
    </>
  );
}

async function getFlexiRemaining(): Promise<number | null> {
  try {
    const apiUrl = getApiUrl();

    const response = await fetch(`${apiUrl}/entries`, {
      method: 'GET',
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('API request failed');
    }

    type Entry = {
      amount_remaining: number;
      date: string;
    };

    const data: Entry[] = await response.json();
    const numEntries = data.length;
    if (numEntries === 0) {
      return 0;
    }
    const lastEntry = data[numEntries - 1];
    const flexiRemaining = lastEntry.amount_remaining;
    return flexiRemaining;
  } catch (error) {
    console.error('Error fetching flexi remaining:', error);
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
