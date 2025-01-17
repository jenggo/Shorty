import type { SSECallback } from './types';

export class SSEHandler {
	private eventSource: EventSource | null = null;
	private static instance: SSEHandler | null = null;
	private url = ''; // Removed type annotation since it's inferred
	private reconnectAttempts = 0;
	private maxReconnectAttempts = 3;
	private callbacks = new Map<string, SSECallback[]>();

	private constructor(url: string) {
		this.url = url;
		this.connect();
	}

	static getInstance(url: string): SSEHandler {
		if (!SSEHandler.instance) {
			SSEHandler.instance = new SSEHandler(url);
		} else if (SSEHandler.instance.url !== url) {
			SSEHandler.instance.close();
			SSEHandler.instance = new SSEHandler(url);
		}
		return SSEHandler.instance;
	}

	private connect() {
		if (this.eventSource) {
			this.eventSource.close();
		}

		this.eventSource = new EventSource(this.url, { withCredentials: true });
		this.reconnectAttempts = 0;

		// Add connected event handler
		this.eventSource.addEventListener('connected', () => {
			console.log('SSE connected successfully');
			this.reconnectAttempts = 0;
		});

		this.eventSource.onopen = () => {
			console.log('SSE connection opened');
			this.reconnectAttempts = 0;

			for (const [event, handlers] of this.callbacks) {
				for (const callback of handlers) {
					this.attachEventListener(event, callback);
				}
			}
		};

		this.eventSource.onerror = (err) => {
			console.error('SSE connection error:', err);
			this.reconnectAttempts++;

			if (this.reconnectAttempts >= this.maxReconnectAttempts) {
				this.showErrorNotification();
				this.close();
			} else {
				// Try to reconnect
				setTimeout(() => this.connect(), 1000);
			}
		};
	}

	private attachEventListener(event: string, callback: SSECallback) {
		this.eventSource?.addEventListener(event, (e: MessageEvent) => {
			try {
				callback(e.data);
			} catch (err) {
				console.error(`Error handling ${event}:`, err);
			}
		});
	}

	addEventListener(event: string, callback: SSECallback) {
		if (!this.callbacks.has(event)) {
			this.callbacks.set(event, []);
		}
		this.callbacks.get(event)?.push(callback);

		if (this.eventSource?.readyState === EventSource.OPEN) {
			this.attachEventListener(event, callback);
		}
	}

	private showErrorNotification() {
		const toast = document.createElement('div');
		toast.className = 'fixed top-4 right-4 bg-red-500 text-white px-6 py-3 rounded shadow-lg z-50';
		toast.innerHTML = `
      <div class="flex items-center">
        <span class="mr-2">⚠️</span>
        <span>Connection lost. Please refresh the page or try logging in again.</span>
        <button class="ml-4 hover:text-red-200" onclick="this.parentElement.parentElement.remove()">✕</button>
      </div>
    `;
		document.body.appendChild(toast);
		setTimeout(() => toast.remove(), 10000);
	}

	close() {
		if (this.eventSource) {
			this.eventSource.close();
			this.eventSource = null;
		}
		this.callbacks.clear();
		SSEHandler.instance = null;
	}
}
