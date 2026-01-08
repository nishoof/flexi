import { GoogleLogin } from '@react-oauth/google';
import React from 'react';
import AddEntryModal from '../components/AddEntryModal';
import StatCard from '../components/StatCard';

function OverviewPage() {
  const [isAddEntryModalOpen, setIsAddEntryModalOpen] = React.useState(false);
  const apiUrl = import.meta.env.VITE_API_URL;

  const handleGoogleLogin = async (credential: string) => {
    try {
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

      console.log('Logged in successfully');
    } catch (error) {
      console.error('Error logging in:', error);
    }
  };

  return (
    <>
      <div className="mb-4 max-w-3xs">
        <GoogleLogin
          onSuccess={(credentialResponse) => {
            if (credentialResponse.credential) {
              console.log('Google Credential:', credentialResponse.credential);
              handleGoogleLogin(credentialResponse.credential);
            }
          }}
          onError={() => console.log('Login Failed')}
        />
      </div>

      <div className="flex space-x-4">
        <StatCard title="Flexi Remaining" value={67.67} />
        <StatCard title="Flexi Remaining" value={67.67} />
        <StatCard title="Flexi Remaining" value={67.67} />
        <StatCard title="Flexi Remaining" value={67.67} />
      </div>

      <AddEntryModal
        isOpen={isAddEntryModalOpen}
        close={() => setIsAddEntryModalOpen(false)}
      />
      <button onClick={() => setIsAddEntryModalOpen(true)} >
        Open modal
      </button>
    </>
  );
}

export default OverviewPage;
