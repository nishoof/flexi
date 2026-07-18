/**
 * This file contains functions to interact with the backend API.
 */

export class AuthError extends Error {
    constructor(message = 'Unauthorized') {
        super(message);
        this.name = 'AuthError';
    }
}

export function isAuthError(error: unknown): error is AuthError {
    return error instanceof AuthError;
}

// Term

export type Term = {
    id: number;
    name: string;
    endDate: string;
    isActive: boolean;
    daysOff: string[];
};

export type ApiTerm = {
    id: number;
    name: string;
    end_date: string;
    is_active: boolean;
    days_off: string[];
}

export async function getTerm(): Promise<Term> {
    const response = await fetchBackend('terms', 'GET');
    const data: ApiTerm[] = await response.json();
    const activeTerm = data.find((term) => term.is_active);
    if (!activeTerm) {
        throw new Error('No active term');
    }
    return {
        id: activeTerm.id,
        name: activeTerm.name,
        endDate: activeTerm.end_date,
        isActive: activeTerm.is_active,
        daysOff: activeTerm.days_off,
    };
}

type TermUpdate = {
    name: string;
    endDate: string;
    daysOff: string[];
};

export async function updateTerm(term: TermUpdate): Promise<void> {
    await fetchBackend('terms', 'PUT', {
        name: term.name,
        end_date: term.endDate,
        days_off: term.daysOff,
    });
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
        if (response.status === 401) {
            throw new AuthError();
        }
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
