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
};

function mapApiTerm(data: ApiTerm): Term {
    return {
        id: data.id,
        name: data.name,
        endDate: data.end_date,
        isActive: data.is_active,
        daysOff: data.days_off,
    };
}

/** All terms for the current user. */
export async function getTerms(): Promise<Term[]> {
    const response = await fetchBackend('terms', 'GET');
    const data: ApiTerm[] = await response.json();
    return data.map(mapApiTerm);
}

/** Active term for the dashboard. */
export async function getTerm(): Promise<Term> {
    const terms = await getTerms();
    const activeTerm = terms.find((term) => term.isActive);
    if (!activeTerm) {
        throw new Error('No active term');
    }
    return activeTerm;
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

type TermCreate = {
    name: string;
    endDate: string;
    daysOff?: string[];
};

/** Create a new (inactive) term. */
export async function createTerm(term: TermCreate): Promise<Term> {
    const response = await fetchBackend('terms', 'POST', {
        name: term.name,
        end_date: term.endDate,
        days_off: term.daysOff ?? [],
    });
    const data: ApiTerm = await response.json();
    return mapApiTerm(data);
}

/** Set a term as the active term. */
export async function activateTerm(termId: number): Promise<void> {
    await fetchBackend(`terms/${termId}/activate`, 'POST');
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
