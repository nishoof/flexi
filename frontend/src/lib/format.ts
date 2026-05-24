/** Get current date (in user's timezone) in YYYY-MM-DD format */
export function currentDateYYYYmmDD() {
    return new Date().toLocaleDateString('en-CA');  // canada format is YYYY-MM-DD
}

/** Reformats a YYYY-MM-DD date string to M/D/YYYY */
export function formatDate(dateStringYYYYmmDD: string) {
    return new Date(dateStringYYYYmmDD).toLocaleDateString('en-US'); // us format is M/D/YYYY
}

/** Formats a number as a money string, e.g. 3 becomes "$3.00" and -12.5 becomes "-$12.50" */
export function formatMoney(number: number) {
    return `${number < 0 ? '-' : ''}$${Math.abs(number).toFixed(2)}`;
}
