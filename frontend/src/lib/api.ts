export function getApiUrl(): string {
    const apiUrl = import.meta.env.VITE_API_URL;
    if (typeof apiUrl !== 'string') {
        throw new TypeError('API URL is not defined in environment variables.');
    }
    return apiUrl;
}
