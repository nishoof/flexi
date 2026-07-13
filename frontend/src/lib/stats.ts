import type { Entry, Term } from './api';

export type Stats = {
    currentFlexiRemaining: number;
    dailySpend: number;
    endingProjection: number;
    remainingPerDay: number;
}

/**
 * Calculates the stats based on the given entries and term.
 * Entries must be in reverse chronological order (most recent first).
 */
export function calculateStats(entries: Entry[], term: Term | null): Stats {
    const numEntries = entries.length;
    const mostRecentEntry = entries[0];
    const oldestEntry = entries[numEntries - 1];

    const endDateStr = term?.endDate ?? '';
    const daysRemaining = numEntries > 0 && endDateStr
        ? calculateDateDifference(mostRecentEntry.date, endDateStr)
        : 0;

    const currentFlexiRemaining = mostRecentEntry?.amountRemaining ?? 0;

    const daysUsed = numEntries > 1 ? calculateDateDifference(oldestEntry.date, mostRecentEntry.date) : 0;
    const originalFlexi = numEntries > 0 ? oldestEntry.amountRemaining : 0;
    const dailySpend = daysUsed > 0 ? (originalFlexi - currentFlexiRemaining) / daysUsed : 0;

    const endingProjection = currentFlexiRemaining - (dailySpend * daysRemaining);

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
 * Calculates the difference in days between two date strings (in YYYY-MM-DD format).
 * For example calculateDateDifference('2026-01-01', '2026-01-02') would return 1
 */
function calculateDateDifference(startDateStr: string, endDateStr: string): number {
    const startDate = new Date(startDateStr);
    const endDate = new Date(endDateStr);
    const timeDiff = endDate.getTime() - startDate.getTime();
    const daysDiff = timeDiff / (1000 * 3600 * 24);
    return Math.ceil(daysDiff);
}
