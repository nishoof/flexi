import React, { useEffect, useRef } from 'react';
import { updateBudget, type Budget } from '../lib/api';
import { getCurrentDate } from '../lib/date';

interface EditBudgetModalProps {
  /** Boolean indicating if the modal is open */
  isOpen: boolean;
  /** Function to close the modal */
  close: () => void;
  /** Function to be called after the budget is updated */
  onBudgetUpdated: () => void;
  /** Initial budget data to populate the modal */
  initialBudget: Budget | null;
}

/** Modal component for editing the budget */
export default function EditBudgetModal({ isOpen, close, onBudgetUpdated, initialBudget }: Readonly<EditBudgetModalProps>) {
  const ref = useRef<HTMLDialogElement>(null);
  const [localHolidays, setLocalHolidays] = React.useState<string[]>(initialBudget?.holidays || []);

  // debug printing
  console.log('EditBudgetModal render');
  useEffect(() => {
    console.log('EditBudgetModal Mount');
    return () => {
      console.log('EditBudgetModal Unmount');
    };
  }, []);

  useEffect(() => {
    if (isOpen) {
      ref.current?.showModal();
    } else {
      ref.current?.close();
    }
  }, [isOpen]);

  async function saveBudget(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    close();

    console.log('Saving budget with holidays:', localHolidays);

    await updateBudget(localHolidays);
    onBudgetUpdated();
  }

  const holidayList = (
    <div className="bg-(--background-lightish) border border-(--border) rounded-lg overflow-y-auto">
      {localHolidays.map((holiday, index) => (
        <div
          key={index}
          className="flex justify-between items-center px-3 py-2 border-b border-(--border)"
        >
          <span>{holiday}</span>
          <button
            type="button"
            className="text-red-500"
            onClick={() => {
              const newLocalHolidays = localHolidays.filter((_, i) => i !== index);
              setLocalHolidays(newLocalHolidays);
            }}
          >
            X
          </button>
        </div>
      ))}
      <input
        type="date"
        id="newHoliday"
        name="newHoliday"
        defaultValue={getCurrentDate()}
        className="w-full px-3 py-2 bg-(--background) focus:outline-none"
        onKeyDown={(e) => {
          if (e.key !== 'Enter') return;
          e.preventDefault();
          const newHoliday = (e.target as HTMLInputElement).value.trim();
          if (!newHoliday) return;
          if (localHolidays.includes(newHoliday)) return;
          const newHolidays = [...localHolidays, newHoliday];
          setLocalHolidays(newHolidays);
          (e.target as HTMLInputElement).value = '';
        }}
      />
    </div>
  );

  return (
    <dialog
      ref={ref}
      onCancel={close}
      closedby="any"
      className="w-[calc(100%-2rem)] max-w-md rounded-xl border border-(--border) bg-(--background) mt-[15vh] mx-auto backdrop:bg-black/50"
    >
      <header className="flex justify-between p-4 bg-(--background) border-b border-(--border)">
        <div>
          <h2 className="font-bold">Edit Budget</h2>
        </div>
        <button onClick={close} aria-label="Close modal" type="button">
          X
        </button>
      </header>

      <form
        className="p-4 bg-(--background-light) space-y-4"
        onSubmit={saveBudget}
      >
        {/* Holidays */}
        <div className="flex flex-col gap-2">
          <label
            htmlFor="holidays"
            className="font-medium"
          >
            Holidays / Days Off
          </label>
          {holidayList}
        </div>

        {/* Submit Button */}
        <button
          type="submit"
          className="w-full px-4 py-2 bg-(--accent) rounded-lg hover:bg-(--accent-dark) font-medium"
        >
          Save Budget
        </button>
      </form>
    </dialog>
  );
}
