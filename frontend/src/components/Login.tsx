import { GoogleLogin, type CredentialResponse } from '@react-oauth/google';
import { getApiUrl } from '../lib/api';

export default function Login({ onSuccessfulLogin }: Readonly<{ onSuccessfulLogin: () => void }>) {
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

      onSuccessfulLogin();
    } catch (error) {
      console.error('Error logging in:', error);
    }
  }
  return (
    <GoogleLogin
      onSuccess={handleGoogleLogin}
      onError={() => console.log('Login Failed')}
    />
  );
}
