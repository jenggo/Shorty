<script lang="ts">
	import { API_BASE_URL } from '$lib/config';
	import { toast } from '$lib/components/swal';
	import { onMount } from 'svelte';
	import { page } from '$app/stores';

	let loading = false;
	let username = '';
	let password = '';
	let authMethod: 'oauth' | 'userpass' | 'none' = 'none';

	onMount(async () => {
		// Check for error parameter in URL
		const error = $page.url.searchParams.get('error');
		if (error) {
			toast.error('Login Error', decodeURIComponent(error));
		}

		// Fetch auth configuration
		try {
			const response = await fetch(`${API_BASE_URL}/auth/config`);
			const data = await response.json();

			if (!data.error && data.data) {
				if (data.data.oauth) {
					authMethod = 'oauth';
				} else if (data.data.userpass) {
					authMethod = 'userpass';
				} else {
					authMethod = 'none';
				}
			}
		} catch (error) {
			console.error('Failed to fetch auth config:', error);
			authMethod = 'none';
		}
	});

	async function handleOAuthLogin() {
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

	async function handleUserPassLogin() {
		if (!username || !password) {
			toast.error('Validation Error', 'Please enter username and password');
			return;
		}

		try {
			loading = true;
			const response = await fetch(`${API_BASE_URL}/auth/login`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				credentials: 'include',
				body: JSON.stringify({
					username,
					password
				})
			});

			const data = await response.json();

			if (data.error) {
				toast.error('Login Error', data.message || 'Login failed');
				return;
			}

			// Redirect to home page on successful login
			window.location.href = '/';
		} catch (error) {
			toast.error(
				'Login Error',
				error instanceof Error ? error.message : 'An unexpected error occurred'
			);
		} finally {
			loading = false;
		}
	}

	function handleKeyPress(event: KeyboardEvent) {
		if (event.key === 'Enter' && authMethod === 'userpass') {
			handleUserPassLogin();
		}
	}
</script>

<div class="flex min-h-screen items-center justify-center bg-gray-50">
	<div class="w-96 rounded-lg bg-white p-8 shadow-md">
		{#if authMethod === 'none'}
			<!-- No authentication configured -->
			<h1 class="mb-6 text-center text-3xl font-bold text-gray-800">Shorty</h1>
			<p class="text-center text-gray-600">Welcome to Shorty</p>
		{:else if authMethod === 'oauth'}
			<!-- OAuth Login -->
			<h2 class="mb-6 text-center text-2xl font-bold text-gray-800">Login to Shorty</h2>

			<div class="space-y-4">
				<button
					on:click={handleOAuthLogin}
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
		{:else if authMethod === 'userpass'}
			<!-- Username/Password Login -->
			<h2 class="mb-6 text-center text-2xl font-bold text-gray-800">Login to Shorty</h2>

			<form on:submit|preventDefault={handleUserPassLogin} class="space-y-4">
				<div>
					<label for="username" class="block text-sm font-medium text-gray-700">Username</label>
					<input
						id="username"
						type="text"
						bind:value={username}
						on:keypress={handleKeyPress}
						disabled={loading}
						class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 shadow-sm focus:border-blue-500 focus:outline-none focus:ring-blue-500 disabled:bg-gray-100"
						placeholder="Enter your username"
						required
					/>
				</div>

				<div>
					<label for="password" class="block text-sm font-medium text-gray-700">Password</label>
					<input
						id="password"
						type="password"
						bind:value={password}
						on:keypress={handleKeyPress}
						disabled={loading}
						class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 shadow-sm focus:border-blue-500 focus:outline-none focus:ring-blue-500 disabled:bg-gray-100"
						placeholder="Enter your password"
						required
					/>
				</div>

				<button
					type="submit"
					disabled={loading}
					class="flex w-full items-center justify-center rounded-md bg-blue-600 px-4 py-2 text-white hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50"
				>
					{#if loading}
						<div
							class="mr-2 h-5 w-5 animate-spin rounded-full border-2 border-white border-t-transparent"
						></div>
					{/if}
					Login
				</button>
			</form>
		{/if}
	</div>
</div>
