import { writable } from 'svelte/store';

interface AuthStore {
	isAuthenticated: boolean;
	username: string | null;
	s3Enabled: boolean;
}

const initialState: AuthStore = {
	isAuthenticated: false,
	username: null,
	s3Enabled: false
};

function createAuthStore() {
	const { subscribe, set, update } = writable<AuthStore>(initialState);

	return {
		subscribe,
		login: (username: string, s3Enabled: boolean) =>
			set({
				isAuthenticated: true,
				username,
				s3Enabled
			}),
		logout: () =>
			set({
				isAuthenticated: false,
				username: null,
				s3Enabled: false
			}),
		updateS3Status: (status: boolean) => update((state) => ({ ...state, s3Enabled: status }))
	};
}

export const auth = createAuthStore();
