import { useEffect, useRef, useState } from 'react';
import {
  activateTerm,
  createTerm,
  isAuthError,
} from '../lib/api';
import { currentDateYYYYmmDD } from '../lib/format';

interface NewTermModalProps {
  /** Boolean indicating if the modal is open */
  isOpen: boolean;
  /** Function to close the modal */
  close: () => void;
  /** Function to be called after the term is created and activated */
  onTermCreated: () => void;
  /** Called when the session is no longer valid */
  onUnauthorized: () => void;
}

/** Modal component for creating and activating a new term */
export default function NewTermModal({
  isOpen,
  close,
  onTermCreated,
  onUnauthorized,
}: Readonly<NewTermModalProps>) {
  const ref = useRef<HTMLDialogElement>(null);
  const [name, setName] = useState('');
  const [endDate, setEndDate] = useState(currentDateYYYYmmDD);

  useEffect(() => {
    if (isOpen) {
      ref.current?.showModal();
    } else {
      ref.current?.close();
    }
  }, [isOpen]);

  async function saveNewTerm(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    close();

    try {
      const created = await createTerm({
        name: name.trim(),
        endDate,
        daysOff: [],
      });
      await activateTerm(created.id);
      onTermCreated();
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
        <h2 className="font-bold">New term</h2>
        <button onClick={close} aria-label="Close modal" type="button">
          X
        </button>
      </header>

      <form
        className="p-4 bg-(--background-light) space-y-4"
        onSubmit={saveNewTerm}
      >
        {/* Name */}
        <div className="flex flex-col gap-2">
          <label htmlFor="termName" className="font-medium">
            Name
          </label>
          <input
            type="text"
            id="termName"
            name="termName"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g. Fall 2026"
            className="w-full px-3 py-2 bg-(--background) border border-(--border) rounded-lg focus:outline-none"
            required
          />
        </div>

        {/* End date */}
        <div className="flex flex-col gap-2">
          <label htmlFor="newTermEndDate" className="font-medium">
            End date
          </label>
          <input
            type="date"
            id="newTermEndDate"
            name="newTermEndDate"
            value={endDate}
            onChange={(e) => setEndDate(e.target.value)}
            className="w-full px-3 py-2 bg-(--background) border border-(--border) rounded-lg focus:outline-none"
            required
          />
        </div>

        {/* Submit Button */}
        <button
          type="submit"
          className="w-full px-4 py-2 bg-(--accent) rounded-lg hover:bg-(--accent-dark) font-medium"
        >
          Create & activate
        </button>
      </form>
    </dialog>
  );
}
