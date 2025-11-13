<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { API_BASE_URL } from '$lib/config';

	const { children } = $props();

	async function checkAuthConfig() {
		try {
			const response = await fetch(`${API_BASE_URL}/auth/config`);
			const data = await response.json();

			// If no auth methods are configured, allow access without login
			if (!data.data?.oauth && !data.data?.userpass) {
				auth.login('Guest', data.data?.s3Enabled || false);
				return true;
			}
			return false;
		} catch (err) {
			console.error('Failed to check auth config:', err);
			return false;
		}
	}

	async function checkSession() {
		try {
			// First check if any auth is required
			const noAuthRequired = await checkAuthConfig();
			if (noAuthRequired) {
				return;
			}

			const response = await fetch(`${API_BASE_URL}/auth/check`, {
				credentials: 'include'
			});
			const data = await response.json();

			if (!data.error && data.data?.username) {
				auth.login(data.data.username, data.data.s3Enabled);
			} else {
				auth.logout();
				if (window.location.pathname !== '/login') {
					goto('/login');
				}
			}
		} catch (err) {
			console.error('Session check failed:', err);
			auth.logout();
			if (window.location.pathname !== '/login') {
				goto('/login');
			}
		}
	}

	onMount(() => {
		checkSession();
	});
</script>

{@render children()}
