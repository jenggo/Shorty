import Swal from 'sweetalert2';

export const toast = {
	success: (title: string, message?: string) => {
		return Swal.fire({
			title,
			text: message,
			icon: 'success',
			toast: true,
			position: 'top-end',
			showConfirmButton: false,
			timer: 3000,
			timerProgressBar: true
		});
	},
	error: (title: string, message?: string) => {
		return Swal.fire({
			title,
			text: message,
			icon: 'error',
			toast: true,
			position: 'top-end',
			showConfirmButton: false,
			timer: 3000,
			timerProgressBar: true
		});
	},
	warning: (title: string, message?: string) => {
		return Swal.fire({
			title,
			text: message,
			icon: 'warning',
			toast: true,
			position: 'top-end',
			showConfirmButton: false,
			timer: 3000,
			timerProgressBar: true
		});
	},
	promise: async <T>(
		promise: Promise<T>,
		messages: { loading: string; success: string; error: string | ((error: unknown) => string) }
	) => {
		try {
			Swal.fire({
				title: messages.loading,
				allowOutsideClick: false,
				didOpen: () => {
					Swal.showLoading();
				}
			});

			const result = await promise;

			Swal.fire({
				icon: 'success',
				title: messages.success,
				timer: 3000,
				timerProgressBar: true,
				showConfirmButton: false
			});

			return result;
		} catch (error) {
			const errorMessage =
				typeof messages.error === 'function' ? messages.error(error) : messages.error;

			Swal.fire({
				icon: 'error',
				title: errorMessage,
				timer: 3000,
				timerProgressBar: true,
				showConfirmButton: false
			});

			throw error;
		}
	}
};

export const confirm = async (options: {
	title: string;
	text?: string;
	icon?: 'warning' | 'error' | 'success' | 'info' | 'question';
	confirmButtonText?: string;
	cancelButtonText?: string;
}) => {
	const result = await Swal.fire({
		title: options.title,
		text: options.text,
		icon: options.icon || 'warning',
		showCancelButton: true,
		confirmButtonColor: '#3085d6',
		cancelButtonColor: '#d33',
		confirmButtonText: options.confirmButtonText || 'Yes',
		cancelButtonText: options.cancelButtonText || 'Cancel'
	});

	return result.isConfirmed;
};

export const prompt = async (options: {
	title: string;
	input?: 'text' | 'textarea' | 'password' | 'email';
	inputValue?: string;
	inputValidator?: (value: string) => string | null;
	showCancelButton?: boolean;
}) => {
	const result = await Swal.fire({
		title: options.title,
		input: options.input || 'text',
		inputValue: options.inputValue || '',
		showCancelButton: options.showCancelButton ?? true,
		inputValidator: options.inputValidator,
		confirmButtonColor: '#3085d6',
		cancelButtonColor: '#d33'
	});

	return result;
};
