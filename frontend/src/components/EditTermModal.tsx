import React, { useEffect, useRef } from "react";
import { isAuthError, updateTerm, type Term } from "../lib/api";
import { currentDateYYYYmmDD, formatDate } from "../lib/format";

interface EditTermModalProps {
  /** Boolean indicating if the modal is open */
  isOpen: boolean;
  /** Function to close the modal */
  close: () => void;
  /** Function to be called after the term is updated */
  onTermUpdated: () => void;
  /** Initial term data to populate the modal */
  initialTerm: Term;
  /** Called when the session is no longer valid */
  onUnauthorized: () => void;
}

/** Modal component for editing the term */
export default function EditTermModal({
  isOpen,
  close,
  onTermUpdated,
  initialTerm,
  onUnauthorized,
}: Readonly<EditTermModalProps>) {
  const ref = useRef<HTMLDialogElement>(null);
  const [localDaysOff, setLocalDaysOff] = React.useState<string[]>(
    initialTerm.daysOff,
  );
  const [localStartDate, setLocalStartDate] = React.useState(
    initialTerm.startDate,
  );
  const [localEndDate, setLocalEndDate] = React.useState(initialTerm.endDate);
  const [localStartingAmount, setLocalStartingAmount] = React.useState(
    String(initialTerm.startingAmount),
  );

  useEffect(() => {
    if (isOpen) {
      ref.current?.showModal();
    } else {
      ref.current?.close();
    }
  }, [isOpen]);

  async function saveTerm(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    close();

    try {
      await updateTerm({
        name: initialTerm.name,
        startDate: localStartDate,
        endDate: localEndDate,
        startingAmount: Number(localStartingAmount),
        daysOff: localDaysOff,
      });
      onTermUpdated();
    } catch (error) {
      if (isAuthError(error)) {
        onUnauthorized();
      }
    }
  }

  const daysOffList = (
    <div className="bg-(--background-lightish) border border-(--border) rounded-lg overflow-y-auto">
      {localDaysOff.map((dayOff, index) => (
        <div
          key={index}
          className="flex justify-between items-center px-3 py-2 border-b border-(--border)"
        >
          <span>{formatDate(dayOff)}</span>
          <button
            type="button"
            className="text-(--red)"
            onClick={() => {
              setLocalDaysOff(localDaysOff.filter((_, i) => i !== index));
            }}
          >
            X
          </button>
        </div>
      ))}
      <input
        type="date"
        id="newDayOff"
        name="newDayOff"
        defaultValue={currentDateYYYYmmDD()}
        min={localStartDate}
        max={localEndDate}
        className="w-full px-3 py-2 bg-(--background) focus:outline-none"
        onKeyDown={(e) => {
          if (e.key !== "Enter") return;
          e.preventDefault();
          const newDayOff = (e.target as HTMLInputElement).value.trim();
          if (!newDayOff) return;
          if (localDaysOff.includes(newDayOff)) return;
          if (newDayOff < localStartDate || newDayOff > localEndDate) return;
          setLocalDaysOff([...localDaysOff, newDayOff]);
          (e.target as HTMLInputElement).value = "";
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
        <h2 className="font-bold">Term Settings</h2>
        <button onClick={close} aria-label="Close modal" type="button">
          X
        </button>
      </header>

      <form
        className="p-4 bg-(--background-light) space-y-4"
        onSubmit={saveTerm}
      >
        {/* Start date */}
        <div className="flex flex-col gap-2">
          <label htmlFor="startDate" className="font-medium">
            Start Date
          </label>
          <input
            type="date"
            id="startDate"
            name="startDate"
            value={localStartDate}
            onChange={(e) => setLocalStartDate(e.target.value)}
            max={localEndDate}
            className="px-3 py-2 bg-(--background) border border-(--border) rounded-lg focus:outline-none"
            required
          />
        </div>

        {/* End date */}
        <div className="flex flex-col gap-2">
          <label htmlFor="endDate" className="font-medium">
            End Date
          </label>
          <input
            type="date"
            id="endDate"
            name="endDate"
            value={localEndDate}
            onChange={(e) => setLocalEndDate(e.target.value)}
            min={localStartDate}
            className="px-3 py-2 bg-(--background) border border-(--border) rounded-lg focus:outline-none"
            required
          />
        </div>

        {/* Starting amount */}
        <div className="flex flex-col gap-2">
          <label htmlFor="startingAmount" className="font-medium">
            Starting Amount
          </label>
          <input
            type="number"
            id="startingAmount"
            name="startingAmount"
            value={localStartingAmount}
            onChange={(e) => setLocalStartingAmount(e.target.value)}
            step="0.01"
            min="0"
            className="px-3 py-2 bg-(--background) border border-(--border) rounded-lg focus:outline-none"
            placeholder="0.00"
            required
          />
        </div>

        {/* Days off */}
        <div className="flex flex-col gap-2">
          <label htmlFor="daysOff" className="font-medium">
            Days Off Campus
          </label>
          {daysOffList}
        </div>

        {/* Submit Button */}
        <button
          type="submit"
          className="w-full px-4 py-2 bg-(--accent) rounded-lg hover:bg-(--accent-dark) font-medium"
        >
          Save
        </button>
      </form>
    </dialog>
  );
}
