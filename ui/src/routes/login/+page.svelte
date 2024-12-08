<script lang="ts">
	import { goto } from '$app/navigation';
	import { auth } from '$lib/stores/auth';
	import { API_BASE_URL } from '$lib/config';
	import Loading from '$lib/components/Loading.svelte';
	import bcrypt from 'bcryptjs';

	let username = '';
	let password = '';
	let loading = false;
	let error = '';

	async function handleSubmit() {
		try {
			loading = true;
			error = '';

			const hashedPassword = await bcrypt.hash(password, 10);

			const response = await fetch(`${API_BASE_URL}/login`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					username,
					password: hashedPassword
				}),
				credentials: 'include' // Important for session cookies
			});

			const data = await response.json();

			if (!response.ok || data.error) {
				throw new Error(data.message || 'Login failed');
			}

			auth.login(username);
			goto('/');
		} catch (err) {
			console.error('Login error:', err);
			error =
				typeof err === 'object' && err !== null && 'message' in err
					? String(err.message)
					: 'Invalid username or password';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Login - Shorty</title>
</svelte:head>

<div class="flex min-h-screen items-center justify-center bg-gray-50">
	<form on:submit|preventDefault={handleSubmit} class="w-96 rounded-lg bg-white p-8 shadow-md">
		<h2 class="mb-6 text-center text-2xl font-bold text-gray-800">Login to Shorty</h2>

		{#if error}
			<div class="mb-4 rounded bg-red-100 p-3 text-sm text-red-700">
				{error}
			</div>
		{/if}

		<div class="space-y-4">
			<div>
				<label for="username" class="block text-sm font-medium text-gray-700">Username</label>
				<input
					id="username"
					type="text"
					bind:value={username}
					placeholder="Enter your username"
					class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 shadow-sm focus:border-blue-500 focus:outline-none focus:ring-blue-500"
					disabled={loading}
					required
				/>
			</div>

			<div>
				<label for="password" class="block text-sm font-medium text-gray-700">Password</label>
				<input
					id="password"
					type="password"
					bind:value={password}
					placeholder="Enter your password"
					class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 shadow-sm focus:border-blue-500 focus:outline-none focus:ring-blue-500"
					disabled={loading}
					required
				/>
			</div>

			<button
				type="submit"
				class="w-full rounded-md bg-blue-500 px-4 py-2 text-white hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50"
				disabled={loading}
			>
				{#if loading}
					<div class="flex items-center justify-center">
						<Loading size="w-5 h-5" />
						<span class="ml-2">Logging in...</span>
					</div>
				{:else}
					Login
				{/if}
			</button>
		</div>
	</form>
</div>
