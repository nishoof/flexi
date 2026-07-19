import { useEffect, useRef, useState } from 'react';
import type { Term } from '../lib/api';
import { formatDate } from '../lib/format';

interface TermPickerProps {
  /** Terms available to the user */
  terms: Term[];
  /** Currently selected (active) term */
  activeTerm: Term;
  /** Called when the user picks a different term */
  onSelectTerm: (term: Term) => void;
  /** Called when the user chooses to create a new term */
  onNewTerm: () => void;
}

/** Dropdown to switch terms or start a new one */
export default function TermPicker({
  terms,
  activeTerm,
  onSelectTerm,
  onNewTerm,
}: Readonly<TermPickerProps>) {
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!isOpen) return;

    function handlePointerDown(event: MouseEvent) {
      if (!containerRef.current?.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        setIsOpen(false);
      }
    }

    document.addEventListener('mousedown', handlePointerDown);
    document.addEventListener('keydown', handleKeyDown);
    return () => {
      document.removeEventListener('mousedown', handlePointerDown);
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [isOpen]);

  function closeAnd(action: () => void) {
    setIsOpen(false);
    action();
  }

  return (
    <div ref={containerRef} className="relative self-start">
      <button
        type="button"
        aria-expanded={isOpen}
        onClick={() => setIsOpen((open) => !open)}
        className="text-xl font-semibold"
      >
        {activeTerm.name} ▾
      </button>

      {isOpen && (
        <div className="absolute mt-2 min-w-64 overflow-hidden rounded-xl border border-(--border) bg-(--background-light)">
          {terms.map((term) => {
            const isActiveTerm = term.id === activeTerm.id;
            return (
              <button
                key={term.id}
                type="button"
                className="flex w-full justify-between px-4 py-3 text-left hover:bg-(--background-lightish)"
                onClick={() => {
                  if (!isActiveTerm) closeAnd(() => onSelectTerm(term));
                  else setIsOpen(false);
                }}
              >
                <span>
                  <span className="block font-medium">{term.name}</span>
                  <span className="block text-sm opacity-60">
                    Ends {formatDate(term.endDate)}
                  </span>
                </span>
                {isActiveTerm && <span aria-hidden>✓</span>}
              </button>
            );
          })}

          <button
            type="button"
            className="w-full border-t border-(--border) px-4 py-3 text-left hover:bg-(--background-lightish)"
            onClick={() => closeAnd(onNewTerm)}
          >
            + New term
          </button>
        </div>
      )}
    </div>
  );
}
