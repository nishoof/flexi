import { useEffect, useRef } from 'react';

function AddEntryModal({ isOpen, close }: { isOpen: boolean; close: () => void; }) {
  const ref = useRef<HTMLDialogElement>(null);

  useEffect(() => {
    if (isOpen) {
      ref.current?.showModal();
    } else {
      ref.current?.close();
    }
  }, [isOpen]);

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
        onSubmit={(event) => saveEntry(event, close)}
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
            defaultValue={getCurrentDate()}
            className="px-3 py-2 border border-(--border) rounded-lg bg-(--background) focus:outline-none"
            required
          />
        </div>
        {/* Submit Button */}
        <button
          type="submit"
          className="w-full px-4 py-2 bg-(--accent) text-white rounded-lg hover:bg-(--accent-dark) font-medium"
        >
          Save Entry
        </button>
      </form>
    </dialog>
  );
}

/* Helper function to get current date (in user's timezone) in YYYY-MM-DD format */
function getCurrentDate() {
  return new Date().toLocaleDateString('en-CA');
}

async function saveEntry(event: React.FormEvent<HTMLFormElement>, close: () => void) {
  event.preventDefault();
  close();

  const formData = new FormData(event.currentTarget);
  const flexiBalance = Number(formData.get('flexiBalance'));
  const date = formData.get('date') as string;

  const apiUrl = import.meta.env.VITE_API_URL;
  if (typeof apiUrl !== 'string') {
    console.error('API URL is not defined in environment variables.');
    return;
  }

  console.log(JSON.stringify({ amount_remaining: flexiBalance, date }));

  try {
    const response = await fetch(`${apiUrl}/entries`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ amount_remaining: flexiBalance, date }),
    });

    if (!response.ok) {
      throw new Error('Network response was not ok');
    }
  } catch (error) {
    console.error('Error saving entry:', error);
  }
}

export default AddEntryModal;
