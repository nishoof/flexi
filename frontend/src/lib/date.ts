/* Helper function to get current date (in user's timezone) in YYYY-MM-DD format */
export function getCurrentDate() {
    return new Date().toLocaleDateString('en-CA');
}
