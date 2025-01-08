<script lang="ts">
	import { API_BASE_URL } from '$lib/config';
	import { onDestroy } from 'svelte';

	interface UploadResponse {
		error: boolean;
		message: string;
	}

	let files: FileList | null = null;
	let uploading = false;
	let progress = 0;
	let error = '';
	let success = '';

	// Define max file size (100MB in bytes)
	const MAX_FILE_SIZE = 100 * 1024 * 1024; // 100MB

	// Notification timeout (in milliseconds)
	const NOTIFICATION_TIMEOUT = 5000; // 5 seconds

	let notificationTimer: number;

	// Clear notifications after timeout
	function clearNotificationsAfterDelay() {
		clearTimeout(notificationTimer);
		notificationTimer = setTimeout(() => {
			error = '';
			success = '';
		}, NOTIFICATION_TIMEOUT);
	}

	// Cleanup on component destroy
	onDestroy(() => {
		clearTimeout(notificationTimer);
	});

	// Reset form
	function resetForm() {
		files = null;
		progress = 0;
		error = '';
		success = '';
	}

	async function handleUpload() {
		if (!files?.[0]) return;

		// Check file size before uploading
		if (files[0].size > MAX_FILE_SIZE) {
			error = 'File size exceeds 100MB limit';
			files = null;
			clearNotificationsAfterDelay();
			return;
		}

		uploading = true;
		progress = 0;
		error = '';
		success = '';

		const formData = new FormData();
		formData.append('file', files[0]);

		try {
			const xhr = new XMLHttpRequest();

			xhr.upload.onprogress = (event) => {
				if (event.lengthComputable) {
					progress = Math.round((event.loaded / event.total) * 100);
				}
			};

			const uploadPromise = new Promise<UploadResponse>((resolve, reject) => {
				xhr.onload = () => {
					if (xhr.status === 200) {
						resolve(JSON.parse(xhr.response) as UploadResponse);
					} else {
						reject(new Error('Upload failed'));
					}
				};
				xhr.onerror = () => reject(new Error('Upload failed'));
			});

			xhr.open('POST', `${API_BASE_URL}/upload`);
			xhr.withCredentials = true;
			xhr.send(formData);

			const response = await uploadPromise;
			success = response.message;
			resetForm();
			clearNotificationsAfterDelay();
		} catch (err) {
			if (err instanceof Error) {
				error = err.message;
			} else {
				error = 'An unknown error occurred';
			}
			clearNotificationsAfterDelay();
		} finally {
			uploading = false;
		}
	}

	// Watch for file changes and trigger upload automatically
	$: if (files?.[0] && !uploading) {
		handleUpload();
	}

	// Helper function to format file size
	function formatFileSize(bytes: number): string {
		if (bytes === 0) return '0 Bytes';
		const k = 1024;
		const sizes = ['Bytes', 'KB', 'MB', 'GB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
	}
</script>

<div class="rounded-lg bg-white p-6 shadow-md">
	<h2 class="mb-4 text-xl font-semibold">Upload File</h2>

	<p class="mb-4 text-sm text-gray-600">Maximum file size: {formatFileSize(MAX_FILE_SIZE)}</p>

	{#if error}
		<div class="mb-4 rounded bg-red-100 p-3 text-red-700">
			{error}
			<button class="ml-2 text-red-500" on:click={() => (error = '')}>✕</button>
		</div>
	{/if}

	{#if success}
		<div class="mb-4 rounded bg-green-100 p-3 text-green-700">
			{success}
			<button class="ml-2 text-green-500" on:click={() => (success = '')}>✕</button>
		</div>
	{/if}

	<div class="mb-4">
		<input
			type="file"
			bind:files
			disabled={uploading}
			class="block w-full text-sm text-gray-500
            file:mr-4 file:cursor-pointer file:rounded-full
            file:border-0 file:bg-blue-50 file:px-4
            file:py-2 file:text-sm
            file:font-semibold
            file:text-blue-700
            file:transition-colors
            file:duration-200
            hover:cursor-pointer
            hover:file:bg-blue-100"
		/>
	</div>

	{#if uploading}
		<div class="mb-4">
			<div class="h-2 w-full rounded-full bg-gray-200">
				<div
					class="h-2 rounded-full bg-blue-600 transition-all duration-300"
					style="width: {progress}%"
				></div>
			</div>
			<p class="mt-1 text-sm text-gray-600">{progress}% uploaded</p>
		</div>
	{/if}
</div>
