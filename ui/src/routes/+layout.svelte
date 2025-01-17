<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { auth } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { API_BASE_URL } from '$lib/config';

	const { children } = $props();

	async function checkSession() {
		try {
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
