/**
 * This file contains functions to interact with the backend API.
 */

// Budget

export type Budget = {
    holidays: string[];
}

export async function getBudget(): Promise<Budget> {
    console.log('Fetching budget from API');

    const response = await fetchBackend('budget', 'GET');
    const data: Budget[] = await response.json();

    console.log('Received budget data:', data);

    return { holidays: data[0].holidays };
}

export async function updateBudget(holidays: string[]): Promise<void> {
    const budget: Budget = { holidays };
    await fetchBackend('budget', 'PUT', budget);
}

// Entry

export type Entry = {
    amountRemaining: number;
    date: string;
};

type ApiEntry = {
    amount_remaining: number;
    date: string;
};

export async function getEntries(): Promise<Entry[]> {
    const response = await fetchBackend('entries', 'GET');

    const data: ApiEntry[] = await response.json();
    return data.map((entry) => ({
        amountRemaining: entry.amount_remaining,
        date: entry.date,
    }));
}

export async function createEntry(amountRemaining: number, date: string): Promise<void> {
    const entry: ApiEntry = {
        amount_remaining: amountRemaining,
        date
    };
    await fetchBackend('entries', 'POST', entry);
}

// Auth

export async function login(credential: string): Promise<void> {
    await fetchBackend('auth', 'POST', { credential });
}

// Helpers

async function fetchBackend(endpoint: string, method: string, body?: unknown): Promise<Response> {
    const response = await fetch(`${getApiUrl()}/${endpoint}`, {
        method: method,
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: body ? JSON.stringify(body) : undefined,
    });

    if (!response.ok) {
        throw new Error('API request failed');
    }

    return response;
}

function getApiUrl(): string {
    const apiUrl = import.meta.env.VITE_API_URL;
    if (typeof apiUrl !== 'string') {
        throw new TypeError('API URL is not defined in environment variables.');
    }
    return apiUrl;
}
