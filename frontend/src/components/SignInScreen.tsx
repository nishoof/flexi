import Login from './Login';

interface SignInScreenProps {
  onSuccessfulLogin: () => void;
}

/** Screen shown when the user is not authenticated */
export default function SignInScreen({ onSuccessfulLogin }: Readonly<SignInScreenProps>) {
  return (
    <div className="flex flex-col items-center justify-center gap-6 py-24">
      <div className="flex flex-col items-center gap-2 text-center max-w-sm">
        <h1 className="text-3xl font-semibold">Flexi</h1>
        <p className="text-(--foreground)/70">
          Track your flexi balance and see spending projections.
        </p>
      </div>

      <Login onSuccessfulLogin={onSuccessfulLogin} />
    </div>
  );
}
