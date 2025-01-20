<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { SSEHandler } from '$lib/sse';
	import type { ShortyData } from '$lib/types';
	import { API_BASE_URL } from '$lib/config';
	import { api } from '$lib/api';
	import Loading from '$lib/components/Loading.svelte';
	import FileUpload from '$lib/components/FileUpload.svelte';
	import { auth } from '$lib/stores/auth';
	import { toast, confirm, prompt } from '$lib/components/swal';

	let data: ShortyData[] = [];
	let sseHandler: SSEHandler | null = null;
	let loading = true;
	let error = '';
	let connected = false;

	// Form states
	let newUrl = '';
	// biome-ignore lint: false positive
	let customName = '';
	let formLoading = false;
	// biome-ignore lint: false positive
	let showCreateForm = false;
	// biome-ignore lint: false positive
	let showUploadForm = false;

	onMount(() => {
		sseHandler = SSEHandler.getInstance(`${API_BASE_URL}/events`);

		sseHandler.addEventListener('connected', () => {
			connected = true;
			error = '';
		});

		sseHandler.addEventListener('message', (rawData: string) => {
			try {
				const parsed = JSON.parse(rawData) as ShortyData[];
				if (Array.isArray(parsed)) {
					data = parsed;
					console.log('Data updated:', data.length, 'items');
				}
			} catch (err) {
				console.error('Failed to parse SSE data:', err);
				error = err instanceof Error ? err.message : 'Failed to parse data';
			}
			loading = false;
		});

		return () => {
			sseHandler?.close();
			sseHandler = null;
		};
	});

	onDestroy(() => {
		sseHandler?.close();
		sseHandler = null;
	});

	function handleReconnect() {
		error = '';
		loading = true;
		sseHandler?.close();
		sseHandler = SSEHandler.getInstance(`${API_BASE_URL}/events`);
	}

	async function handleCreate() {
		try {
			formLoading = true;
			await api.createShorty(newUrl, customName || undefined);
			newUrl = '';
			toast.success('Success', 'Shorty created successfully');
		} catch (err) {
			toast.error('Error', err instanceof Error ? err.message : 'Failed to create short URL');
		} finally {
			formLoading = false;
		}
	}

	async function handleDelete(shorty: string) {
		const confirmed = await confirm({
			title: 'Are you sure?',
			text: 'You want to delete this shorty?',
			icon: 'warning'
		});

		if (confirmed) {
			const promise = api.deleteShorty(shorty);
			await toast.promise(promise, {
				loading: 'Deleting shorty...',
				success: 'Shorty deleted successfully',
				error: (error: unknown) =>
					`Failed to delete: ${error instanceof Error ? error.message : 'Unknown error'}`
			});
		}
	}

	async function handleRename(oldName: string) {
		const result = await prompt({
			title: 'Enter new name',
			input: 'text',
			inputValue: oldName,
			inputValidator: (value: string) => {
				if (!value) {
					return 'You need to write something!';
				}
				return null;
			}
		});

		if (result.value) {
			const promise = api.renameShorty(oldName, result.value);
			await toast.promise(promise, {
				loading: 'Renaming shorty...',
				success: 'Shorty renamed successfully',
				error: (error: unknown) =>
					`Failed to rename: ${error instanceof Error ? error.message : 'Unknown error'}`
			});
		}
	}

	function copyToClipboard(fullUrl: string) {
		navigator.clipboard.writeText(fullUrl);
		toast.success('Copied to clipboard!');
	}

	function formatExpiry(nanoseconds: string): string {
		const secs = Number.parseInt(nanoseconds) / 1e9;

		if (secs < 60) {
			return `${Math.round(secs)}s`;
		}
		if (secs < 3600) {
			return `${Math.floor(secs / 60)}m`;
		}
		if (secs < 86400) {
			return `${Math.floor(secs / 3600)}h`;
		}
		return `${Math.floor(secs / 86400)}d`;
	}
</script>

<nav class="bg-gray-800 p-4">
	<div class="container mx-auto flex items-center justify-between">
		<h1 class="text-xl font-bold text-white">
			Hello {$auth.isAuthenticated ? $auth.username : 'Guest'}
		</h1>
		<div class="flex gap-4">
			<button
				class="rounded bg-blue-600 px-4 py-2 text-white hover:bg-blue-700"
				on:click={() => (showCreateForm = !showCreateForm)}
			>
				{showCreateForm ? 'Close' : 'Create New'}
			</button>
			{#if $auth.s3Enabled}
				<button
					class="rounded bg-green-600 px-4 py-2 text-white hover:bg-green-700"
					on:click={() => (showUploadForm = !showUploadForm)}
				>
					{showUploadForm ? 'Close' : 'Upload File'}
				</button>
			{/if}
			<button
				class="rounded bg-red-600 px-4 py-2 text-white hover:bg-red-700"
				on:click={() => (window.location.href = '/logout')}
			>
				Logout
			</button>
		</div>
	</div>
</nav>

<main class="container mx-auto p-4">
	{#if !$auth.isAuthenticated}
		<div class="p-4 text-center">
			<p>Please log in to access this page</p>
			<a href="/login" class="text-blue-600 hover:text-blue-800">Login</a>
		</div>
	{/if}
	{#if error || !connected}
		<div class="mb-4 rounded bg-red-100 p-4 text-red-700">
			{#if !connected}
				<span>Connection lost - attempting to reconnect...</span>
			{:else}
				{error}
			{/if}
			<button class="ml-2 text-red-500 hover:text-red-700" on:click={handleReconnect}>
				Retry
			</button>
		</div>
	{/if}

	{#if showCreateForm}
		<div class="mb-6 rounded-lg bg-white p-4 shadow">
			<form on:submit|preventDefault={handleCreate} class="space-y-4">
				<div>
					<label for="url" class="block text-sm font-medium text-gray-700">URL</label>
					<input
						name="url"
						type="url"
						bind:value={newUrl}
						required
						class="mt-1 block w-full rounded-md border-gray-300 shadow-sm"
						placeholder="https://example.com"
					/>
				</div>
				<div>
					<label for="customName" class="block text-sm font-medium text-gray-700"
						>Custom Name (Optional)</label
					>
					<input
						name="customName"
						type="text"
						bind:value={customName}
						class="mt-1 block w-full rounded-md border-gray-300 shadow-sm"
						placeholder="my-custom-url"
					/>
				</div>
				<div class="flex justify-end gap-2">
					<button
						type="button"
						class="rounded border px-4 py-2 text-gray-700 hover:bg-gray-50"
						on:click={() => (showCreateForm = false)}
					>
						Cancel
					</button>
					<button
						type="submit"
						class="rounded bg-blue-600 px-4 py-2 text-white hover:bg-blue-700 disabled:opacity-50"
						disabled={formLoading}
					>
						{#if formLoading}
							<Loading size="w-5 h-5" />
						{:else}
							Create
						{/if}
					</button>
				</div>
			</form>
		</div>
	{/if}

	{#if showUploadForm && $auth.s3Enabled}
		<div class="mb-6">
			<FileUpload />
		</div>
	{/if}

	{#if loading}
		<div class="flex justify-center p-8">
			<Loading size="w-8 h-8" />
		</div>
	{:else}
		<div class="overflow-x-auto">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th
							class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500"
						>
							Shorty
						</th>
						<th
							class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500"
						>
							File
						</th>
						<th
							class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500"
						>
							Url
						</th>
						<th
							class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500"
						>
							Expired
						</th>
						<th
							class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500"
						>
							Actions
						</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-200 bg-white">
					{#each data as row}
						<tr>
							<td class="whitespace-nowrap px-6 py-4">
								<button
									class="flex items-center gap-2 text-blue-600 hover:text-blue-800"
									on:click={() => {
										const fullUrl = `${API_BASE_URL}/${row.shorty}`;
										copyToClipboard(fullUrl);
									}}
								>
									{row.shorty}
									<svg
										xmlns="http://www.w3.org/2000/svg"
										class="h-4 w-4"
										fill="none"
										viewBox="0 0 24 24"
										stroke="currentColor"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
										/>
									</svg>
								</button>
							</td>
							<td class="whitespace-nowrap px-6 py-4">{row.file}</td>
							<td class="max-w-xs px-6 py-4">
								<a
									href={row.url}
									class="block truncate text-blue-500 hover:underline"
									target="_blank"
									title={row.url}
								>
									{row.url}
								</a>
							</td>
							<td class="whitespace-nowrap px-6 py-4">{formatExpiry(row.expired)}</td>
							<td class="whitespace-nowrap px-6 py-4">
								<div class="flex gap-2">
									<button
										class="rounded-md bg-blue-600 px-3 py-1 text-sm font-medium text-white hover:bg-blue-700"
										on:click={() => handleRename(row.shorty)}
									>
										Rename
									</button>
									<button
										class="rounded-md bg-red-600 px-3 py-1 text-sm font-medium text-white hover:bg-red-700"
										on:click={() => handleDelete(row.shorty)}
									>
										Delete
									</button>
								</div>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</main>
