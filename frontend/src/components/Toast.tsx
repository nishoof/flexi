import React, { useEffect } from "react";

const DISMISS_MS = 4 * 1000;

interface ToastProps {
  /** Message shown to the user */
  message: string;
  /** Called when the toast is dismissed */
  onDismiss: () => void;
}

/** Temporary error notification */
export default function Toast({ message, onDismiss }: Readonly<ToastProps>) {
  const dismiss = React.useEffectEvent(onDismiss);

  useEffect(() => {
    const timeoutId = window.setTimeout(dismiss, DISMISS_MS);
    return () => window.clearTimeout(timeoutId);
  }, []);

  return (
    <div
      role="alert"
      className="fixed bottom-6 left-1/2 z-50 flex w-[calc(100%-2rem)] max-w-md -translate-x-1/2 items-center gap-3 rounded-lg border border-(--border) bg-(--background-light) px-4 py-3"
    >
      <p className="grow text-sm text-(--red)">{message}</p>
      <button type="button" aria-label="Dismiss" onClick={onDismiss}>
        X
      </button>
    </div>
  );
}
