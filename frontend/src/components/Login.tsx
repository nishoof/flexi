import { GoogleLogin, type CredentialResponse } from '@react-oauth/google';
import { useState } from 'react';
import { login } from '../lib/api';

interface LoginProps {
  /** Callback to invoke when login is successful */
  onSuccessfulLogin: () => void;
}

/** Component to handle user login using Google OAuth */
export default function Login({ onSuccessfulLogin }: Readonly<LoginProps>) {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleGoogleLogin(credentialResponse: CredentialResponse) {
    const credential = credentialResponse.credential;
    if (!credential) {
      setError('No credential received from Google. Please try again.');
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      await login(credential);
      onSuccessfulLogin();
    } catch {
      setError('Sign-in failed. Please try again.');
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <div className="flex flex-col items-center gap-3">
      {isLoading ? (
        <p className="text-sm text-(--foreground)/70">Signing in...</p>
      ) : (
        <GoogleLogin
          onSuccess={handleGoogleLogin}
          onError={() => setError('Sign-in failed. Please try again.')}
        />
      )}

      {error && (
        <p role="alert" className="text-sm text-red-400">
          {error}
        </p>
      )}
    </div>
  );
}
