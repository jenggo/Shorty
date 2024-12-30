import { writable } from 'svelte/store';

interface AuthStore {
	isAuthenticated: boolean;
	username: string | null;
}

const createAuthStore = () => {
	const { subscribe, set } = writable<AuthStore>({
		isAuthenticated: false,
		username: null
	});

	return {
		subscribe,
		login: (username: string) => set({ isAuthenticated: true, username }),
		logout: () => set({ isAuthenticated: false, username: null })
	};
};

export const auth = createAuthStore();
