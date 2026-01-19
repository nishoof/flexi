import { GoogleLogin, type CredentialResponse } from '@react-oauth/google';
import { login } from '../lib/api';

interface LoginProps {
  /** Callback to invoke when login is successful */
  onSuccessfulLogin: () => void;
}

/** Component to handle user login using Google OAuth */
export default function Login({ onSuccessfulLogin }: Readonly<LoginProps>) {
  async function handleGoogleLogin(credentialResponse: CredentialResponse) {
    const credential = credentialResponse.credential;
    if (!credential) {
      console.error('No credential received from Google Login');
      return;
    }

    await login(credential);
    onSuccessfulLogin();
  }

  return (
    <GoogleLogin
      onSuccess={handleGoogleLogin}
      onError={() => console.log('Login Failed')}
    />
  );
}
