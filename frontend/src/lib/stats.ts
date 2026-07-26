import type { Entry, Term } from "./api";

export type Stats = {
  currentFlexiRemaining: number;
  dailySpend: number;
  endingProjection: number;
  remainingPerDay: number;
};

export const zeroedStats: Stats = {
  currentFlexiRemaining: 0,
  dailySpend: 0,
  endingProjection: 0,
  remainingPerDay: 0,
};

/**
 * Calculates the stats based on the given entries and term.
 * Entries must be in reverse chronological order (most recent first).
 * With no entries, uses the term's start date and starting amount as the
 * initial snapshot.
 */
export function calculateStats(entries: Entry[], term: Term): Stats {
  const mostRecentEntry = entries[0];
  const todayDateStr = mostRecentEntry?.date ?? term.startDate;
  const currentFlexiRemaining =
    mostRecentEntry?.amountRemaining ?? term.startingAmount;

  const { daysUsed, daysRemaining } = calculateDaysUsedAndRemaining(
    term.startDate,
    todayDateStr,
    term.endDate,
    term.daysOff,
  );

  const dailySpend =
    daysUsed > 0 ? (term.startingAmount - currentFlexiRemaining) / daysUsed : 0;

  const endingProjection =
    daysRemaining > 0
      ? currentFlexiRemaining - dailySpend * daysRemaining
      : currentFlexiRemaining;

  const remainingPerDay =
    daysRemaining > 0 ? currentFlexiRemaining / daysRemaining : 0;

  return {
    currentFlexiRemaining,
    dailySpend,
    endingProjection,
    remainingPerDay,
  };
}

/**
 * Calculates the number of days used and the number of days remaining.
 * Days used: from the start date to today, not including today.
 * Days remaining: from today to the end date, not including today.
 * Takes into account any days off within those windows.
 */
function calculateDaysUsedAndRemaining(
  startDateStr: string,
  todayDateStr: string,
  endDateStr: string,
  daysOffDateStrs: string[],
): {
  daysUsed: number;
  daysRemaining: number;
} {
  let daysRemaining = calculateDateDifference(todayDateStr, endDateStr);
  let daysUsed = calculateDateDifference(startDateStr, todayDateStr);

  for (const dayOffStr of daysOffDateStrs) {
    // Past window: start <= dayOff < today
    if (dayOffStr >= startDateStr && dayOffStr < todayDateStr) {
      daysUsed--;
      // Future window: today < dayOff <= end
    } else if (dayOffStr > todayDateStr && dayOffStr <= endDateStr) {
      daysRemaining--;
    }
  }

  return {
    daysUsed: Math.max(0, daysUsed),
    daysRemaining: Math.max(0, daysRemaining),
  };
}

/**
 * Calculates the difference in days between two date strings (in YYYY-MM-DD format).
 * Returns a positive number if the end date is after the start date,
 * and a negative number if the end date is before the start date.
 * For example calculateDateDifference('2026-01-01', '2026-01-02') would return 1
 */
function calculateDateDifference(
  startDateStr: string,
  endDateStr: string,
): number {
  const startDate = new Date(startDateStr);
  const endDate = new Date(endDateStr);
  const timeDiff = endDate.getTime() - startDate.getTime();
  const daysDiff = timeDiff / (1000 * 3600 * 24);
  return Math.ceil(daysDiff);
}
