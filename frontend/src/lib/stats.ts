import type { Entry, Term } from './api';

export type Stats = {
    currentFlexiRemaining: number;
    dailySpend: number;
    endingProjection: number;
    remainingPerDay: number;
}

export const zeroedStats: Stats = {
    currentFlexiRemaining: 0,
    dailySpend: 0,
    endingProjection: 0,
    remainingPerDay: 0,
};

/**
 * Calculates the stats based on the given entries and term.
 * Entries must be in reverse chronological order (most recent first).
 */
export function calculateStats(entries: Entry[], term: Term): Stats {
    const numEntries = entries.length;
    if (numEntries === 0) {
        return zeroedStats;
    }

    const mostRecentEntry = entries[0];
    const oldestEntry = entries[numEntries - 1];

    const endDateStr = term.endDate;
    const startDateStr = oldestEntry.date;
    const todayDateStr = mostRecentEntry.date;

    const { daysUsed, daysRemaining } = calculateDaysUsedAndRemaining(startDateStr, todayDateStr, endDateStr, term.daysOff);

    const currentFlexiRemaining = mostRecentEntry.amountRemaining;

    const originalFlexi = oldestEntry.amountRemaining;
    const dailySpend = daysUsed > 0 ? (originalFlexi - currentFlexiRemaining) / daysUsed : 0;

    const endingProjection = daysRemaining > 0 ? currentFlexiRemaining - (dailySpend * daysRemaining) : currentFlexiRemaining;

    const remainingPerDay = daysRemaining > 0 ? currentFlexiRemaining / daysRemaining : 0;

    const stats: Stats = {
        currentFlexiRemaining,
        dailySpend,
        endingProjection,
        remainingPerDay,
    };
    return stats;
}

/**
 * Calculates the number of days used (from the start date to today, not
 * including today) and the number of days remaining (from today to the
 * end date, not including today). Takes into account any days off.
 */
function calculateDaysUsedAndRemaining(
    startDateStr: string,
    todayDateStr: string, 
    endDateStr: string, 
    daysOffDateStrs: string[],
): { 
    daysUsed: number,
    daysRemaining: number,
} {
    let daysRemaining = calculateDateDifference(todayDateStr, endDateStr);
    let daysUsed = calculateDateDifference(startDateStr, todayDateStr);

    for (const dayOffStr of daysOffDateStrs) {
        const diff = calculateDateDifference(todayDateStr, dayOffStr);
        if (diff > 0) {
            daysRemaining--;
        } else if (diff < 0) {
            daysUsed--;
        }
    }

    return { daysUsed, daysRemaining };
}

/**
 * Calculates the difference in days between two date strings (in YYYY-MM-DD format).
 * Returns a positive number if the end date is after the start date,
 * and a negative number if the end date is before the start date.
 * For example calculateDateDifference('2026-01-01', '2026-01-02') would return 1
 */
function calculateDateDifference(startDateStr: string, endDateStr: string): number {
    const startDate = new Date(startDateStr);
    const endDate = new Date(endDateStr);
    const timeDiff = endDate.getTime() - startDate.getTime();
    const daysDiff = timeDiff / (1000 * 3600 * 24);
    return Math.ceil(daysDiff);
}
