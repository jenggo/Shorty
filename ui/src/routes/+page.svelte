<script lang="ts">
	import { onMount } from 'svelte';
	import { WS_BASE_URL, API_BASE_URL } from '$lib/config';
	import { api } from '$lib/api';
	import Loading from '$lib/components/Loading.svelte';
	import FileUpload from '$lib/components/FileUpload.svelte';
	import { auth } from '$lib/stores/auth';
	import Swal from 'sweetalert2';

	interface ShortyData {
		shorty: string;
		file: string;
		url: string;
		expired: string;
	}

	let data: ShortyData[] = [];
	let ws: WebSocket;
	let loading = true;
	let error = '';

	// Form states
	let newUrl = '';
	let customName = '';
	let showCreateForm = false;
	let showUploadForm = false;
	let formLoading = false;

	onMount(() => {
		const unsubscribe = auth.subscribe(($auth) => {
			if ($auth.isAuthenticated) {
				connectWebSocket();
			} else {
				loading = false;
				ws?.close();
			}
		});

		// Return cleanup function
		return () => {
			unsubscribe();
			ws?.close();
		};
	});

	function connectWebSocket() {
		try {
			ws = new WebSocket(`${WS_BASE_URL}/ws`);

			ws.onopen = () => {
				console.log('WebSocket connected');
				loading = false;
			};

			ws.onmessage = (event) => {
				try {
					data = JSON.parse(event.data);
				} catch (err) {
					console.error('Failed to parse WebSocket data:', err);
				}
			};

			ws.onerror = (event) => {
				console.error('WebSocket error:', event);
				error = 'WebSocket connection error';
				loading = false;
			};

			ws.onclose = () => {
				console.log('WebSocket disconnected');
				Swal.fire({
					title: 'Connection Lost',
					text: 'Your session has been disconnected.',
					icon: 'warning',
					showConfirmButton: false,
					timer: 5000,
					timerProgressBar: true
				}).then(() => {
					location.assign(`${API_BASE_URL}`);
				});
			};
		} catch (err) {
			error = 'Failed to connect to WebSocket: ' + err;
			loading = false;
		}
	}

	async function handleCreate() {
		try {
			formLoading = true;
			await api.createShorty(newUrl, customName || undefined);
			newUrl = '';
			customName = '';
			showCreateForm = false;
		} catch (err) {
			Swal.fire({
				icon: 'error',
				title: 'Error',
				text: err instanceof Error ? err.message : 'Failed to create short URL'
			});
		} finally {
			formLoading = false;
		}
	}

	async function handleDelete(shorty: string) {
		const result = await Swal.fire({
			title: 'Are you sure?',
			text: "You won't be able to revert this!",
			icon: 'warning',
			showCancelButton: true,
			confirmButtonColor: '#3085d6',
			cancelButtonColor: '#d33',
			confirmButtonText: 'Yes, delete it!'
		});

		if (result.isConfirmed) {
			try {
				await api.deleteShorty(shorty);
				Swal.fire('Deleted!', 'Your shorty has been deleted.', 'success');
			} catch (err) {
				error = 'Failed to delete short URL: ' + err;
				Swal.fire('Error!', 'Failed to delete short URL.', 'error');
			}
		}
	}

	async function handleRename(oldName: string) {
		const result = await Swal.fire({
			title: 'Rename Shorty',
			input: 'text',
			inputLabel: 'New name',
			inputValue: oldName,
			showCancelButton: true,
			inputValidator: (value) => {
				if (!value) {
					return 'You need to provide a name!';
				}
			}
		});

		if (result.isConfirmed) {
			try {
				await api.renameShorty(oldName, result.value);
				Swal.fire('Renamed!', 'Your shorty has been renamed.', 'success');
			} catch (err) {
				error = 'Failed to rename short URL: ' + err;
				Swal.fire('Error!', 'Failed to rename short URL.', 'error');
			}
		}
	}

	function formatExpiry(nanoseconds: string): string {
		const secs = parseInt(nanoseconds) / 1e9;

		if (secs < 60) {
			return `${Math.round(secs)}s`;
		} else if (secs < 3600) {
			return `${Math.floor(secs / 60)}m`;
		} else if (secs < 86400) {
			return `${Math.floor(secs / 3600)}h`;
		} else {
			return `${Math.floor(secs / 86400)}d`;
		}
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
				{showCreateForm ? 'Cancel' : 'Create New'}
			</button>
			{#if $auth.s3Enabled}
				<button
					class="rounded bg-green-600 px-4 py-2 text-white hover:bg-green-700"
					on:click={() => (showUploadForm = !showUploadForm)}
				>
					{showUploadForm ? 'Cancel' : 'Upload File'}
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
	{#if error}
		<div class="mb-4 rounded bg-red-100 p-4 text-red-700">
			{error}
			<button class="ml-2 text-red-500" on:click={() => (error = '')}>✕</button>
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
										navigator.clipboard.writeText(fullUrl);
										Swal.fire({
											toast: true,
											position: 'top-end',
											showConfirmButton: false,
											timer: 2000,
											icon: 'success',
											title: 'Copied to clipboard!'
										});
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
