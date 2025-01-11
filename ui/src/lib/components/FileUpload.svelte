<script lang="ts">
	import { API_BASE_URL } from '$lib/config';
	import { toast } from 'svelte-sonner';

	interface UploadResponse {
		error: boolean;
		message: string;
	}

	let files: FileList | null = null;
	let uploading = false;
	let progress = 0;
	let fileInput: HTMLInputElement;

	// Define max file size (100MB in bytes)
	const MAX_FILE_SIZE = 100 * 1024 * 1024; // 100MB

	// Helper function to format file size
	function formatFileSize(bytes: number): string {
		if (bytes === 0) return '0 Bytes';
		const k = 1024;
		const sizes = ['Bytes', 'KB', 'MB', 'GB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return `${Number.parseFloat((bytes / k ** i).toFixed(2))} ${sizes[i]}`;
	}

	async function handleUpload() {
		if (!files?.[0] || uploading) return;

		// Check file size before uploading
		if (files[0].size > MAX_FILE_SIZE) {
			toast.error('File size exceeds limit');
			resetUpload();
			return;
		}

		uploading = true;
		progress = 0;

		const formData = new FormData();
		formData.append('file', files[0]);

		try {
			const xhr = new XMLHttpRequest();
			xhr.timeout = 30 * 60 * 1000; // 30 minutes timeout

			xhr.upload.onprogress = (event) => {
				if (event.lengthComputable) {
					progress = Math.round((event.loaded / event.total) * 100);
				}
			};

			const uploadPromise = new Promise<UploadResponse>((resolve, reject) => {
				xhr.onload = () => {
					if (xhr.status === 200) {
						resolve(JSON.parse(xhr.response));
					} else {
						try {
							const errorResponse = JSON.parse(xhr.response);
							reject(new Error(errorResponse.message || `Upload failed with status ${xhr.status}`));
						} catch {
							reject(new Error(`Upload failed with status ${xhr.status}`));
						}
					}
				};
				xhr.onerror = () => reject(new Error('Network error occurred'));
				xhr.ontimeout = () => reject(new Error('Upload timed out'));
				xhr.onabort = () => reject(new Error('Upload was aborted'));
			});

			xhr.open('POST', `${API_BASE_URL}/upload`);
			xhr.withCredentials = true;
			xhr.send(formData);

			const response = await uploadPromise;
			toast.success('Upload successful', {
				description: response.message
			});
		} catch (error) {
			console.error('Upload error:', error);
			toast.error('Upload failed', {
				description: error instanceof Error ? error.message : 'Unknown error occurred'
			});
		} finally {
			resetUpload();
		}
	}

	function resetUpload() {
		uploading = false;
		progress = 0;
		files = null;
		if (fileInput) {
			fileInput.value = ''; // Clear the input
		}
	}

	function handleFileSelect() {
		if (files?.[0] && !uploading) {
			handleUpload();
		}
	}
</script>

<div class="rounded-lg bg-white p-6 shadow-md">
	<h2 class="mb-4 text-xl font-semibold">Upload File</h2>

	<p class="mb-4 text-sm text-gray-600">Maximum file size: {formatFileSize(MAX_FILE_SIZE)}</p>

	<div class="mb-4">
		<input
			type="file"
			bind:files
			bind:this={fileInput}
			on:change={handleFileSelect}
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
