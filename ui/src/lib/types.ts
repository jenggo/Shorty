export interface ShortyData {
	shorty: string;
	file: string;
	url: string;
	expired: string;
}

export type SSECallback = (data: string) => void;
