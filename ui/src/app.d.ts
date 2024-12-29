// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare module 'bcryptjs';
declare namespace svelteHTML {
	interface HTMLAttributes {
		[key: string]: string | boolean | number | undefined;
	}
}
declare global {
	namespace App {
		// interface Error {}
		// interface Locals {}
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}
}

export {};
