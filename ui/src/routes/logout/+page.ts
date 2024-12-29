import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';
import { auth } from '$lib/stores/auth';
import { API_BASE_URL } from '$lib/config';

export const load: PageLoad = async () => {
	try {
		await fetch(`${API_BASE_URL}/logout`, {
			credentials: 'include'
		});
	} finally {
		auth.logout();
		redirect(302, '/login');
	}
};
