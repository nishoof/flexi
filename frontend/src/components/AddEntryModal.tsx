import { useEffect, useRef } from 'react';
import { createEntry, isAuthError } from '../lib/api';
import { currentDateYYYYmmDD } from '../lib/format';

interface AddEntryModalProps {
  /** Boolean indicating if the modal is open */
  isOpen: boolean;
  /** Function to close the modal */
  close: () => void;
  /** Function to be called after an entry is added */
  onEntryAdded: () => void;
  /** Called when the session is no longer valid */
  onUnauthorized: () => void;
}

/** Modal component for adding a new entry */
export default function AddEntryModal({
  isOpen,
  close,
  onEntryAdded,
  onUnauthorized,
}: Readonly<AddEntryModalProps>) {
  const ref = useRef<HTMLDialogElement>(null);

  useEffect(() => {
    if (isOpen) {
      ref.current?.showModal();
    } else {
      ref.current?.close();
    }
  }, [isOpen]);

  async function saveEntry(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    close();

    const formData = new FormData(event.currentTarget);
    const flexiBalance = Number(formData.get('flexiBalance'));
    const date = formData.get('date') as string;

    try {
      await createEntry(flexiBalance, date);
      onEntryAdded();
    } catch (error) {
      if (isAuthError(error)) {
        onUnauthorized();
      }
    }
  }

  return (
    <dialog
      ref={ref}
      onCancel={close}
      closedby="any"
      className="w-[calc(100%-2rem)] max-w-md rounded-xl border border-(--border) bg-(--background) mt-[15vh] mx-auto backdrop:bg-black/50"
    >
      <header className="flex justify-between p-4 bg-(--background) border-b border-(--border)">
        <h2 className="font-bold">Add Entry</h2>
        <button onClick={close} aria-label="Close modal" type="button">
          X
        </button>
      </header>

      <form
        className="p-4 bg-(--background-light) space-y-4"
        onSubmit={saveEntry}
      >
        {/* Input: Flexi Balance Remaining */}
        <div className="flex flex-col gap-2">
          <label
            htmlFor="flexiBalance"
            className="font-medium"
          >
            Flexi Balance Remaining
          </label>
          <input
            type="number"
            id="flexiBalance"
            name="flexiBalance"
            step="0.01"
            className="px-3 py-2 border border-(--border) rounded-lg bg-(--background) focus:outline-none"
            placeholder="0.00"
            required
          />
        </div>

        {/* Input: Date */}
        <div className="flex flex-col gap-2">
          <label
            htmlFor="date"
            className="font-medium"
          >
            Date
          </label>
          <input
            type="date"
            id="date"
            name="date"
            defaultValue={currentDateYYYYmmDD()}
            className="px-3 py-2 border border-(--border) rounded-lg bg-(--background) focus:outline-none"
            required
          />
        </div>

        {/* Submit Button */}
        <button
          type="submit"
          className="w-full px-4 py-2 bg-(--accent) rounded-lg hover:bg-(--accent-dark) font-medium"
        >
          Save Entry
        </button>
      </form>
    </dialog>
  );
}
