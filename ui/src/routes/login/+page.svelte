<script lang="ts">
	import { API_BASE_URL } from '$lib/config';
	import { toast } from '$lib/components/swal';
	import { onMount } from 'svelte';
	import { page } from '$app/stores';

	let loading = false;

	onMount(() => {
		// Check for error parameter in URL
		const error = $page.url.searchParams.get('error');
		if (error) {
			toast.error('Login Error', decodeURIComponent(error));
		}
	});

	async function handleLogin() {
		try {
			loading = true;
			const response = await fetch(`${API_BASE_URL}/auth/gitlab`);
			const data = await response.json();

			if (data.error) {
				toast.error('Login Error', data.message || 'Failed to initiate login');
				return;
			}

			if (data.message) {
				window.location.href = data.message;
			} else {
				toast.error('Login Error', 'Invalid response from server');
			}
		} catch (error) {
			toast.error(
				'Login Error',
				error instanceof Error ? error.message : 'An unexpected error occurred'
			);
		} finally {
			loading = false;
		}
	}
</script>

<div class="flex min-h-screen items-center justify-center bg-gray-50">
	<div class="w-96 rounded-lg bg-white p-8 shadow-md">
		<h2 class="mb-6 text-center text-2xl font-bold text-gray-800">Login to Shorty</h2>

		<div class="space-y-4">
			<button
				on:click={handleLogin}
				disabled={loading}
				class="flex w-full items-center justify-center rounded-md bg-[#4d7a63] px-4 py-2 text-white hover:bg-[#1f5335] focus:outline-none focus:ring-2 focus:ring-[#308149] focus:ring-offset-2 disabled:opacity-50"
			>
				{#if loading}
					<div
						class="mr-2 h-5 w-5 animate-spin rounded-full border-2 border-white border-t-transparent"
					></div>
				{:else}
					<svg class="mr-2 h-5 w-5" viewBox="0 0 586 559">
						<path
							fill="currentColor"
							d="M461.17 301.83l-18.91-58.12-37.42-115.28c-1.92-5.9-7.15-10.05-13.37-10.05s-11.45 4.15-13.37 10.05l-37.42 115.28h-126.5l-37.42-115.28c-1.92-5.9-7.15-10.05-13.37-10.05s-11.45 4.15-13.37 10.05l-37.42 115.28-18.91 58.12c-1.72 5.3.12 11.11 4.72 14.38l212.49 154.41 212.49-154.41c4.6-3.27 6.44-9.08 4.72-14.38"
						/>
					</svg>
				{/if}
				Sign in with Repo Nusatek
			</button>
		</div>
	</div>
</div>
