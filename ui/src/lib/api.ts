import { API_BASE_URL } from '$lib/config';
class ApiClient {
  private getHeaders() {
    return {
      'Content-Type': 'application/json'
    };
  }

  private async fetchWithCredentials(url: string, options: RequestInit = {}) {
    const response = await fetch(url, {
      ...options,
      credentials: 'include',
      headers: {
        ...this.getHeaders(),
        ...options.headers
      }
    });

    if (!response.ok) {
      throw new Error('API request failed');
    }

    return response;
  }

  async createShorty(url: string, customName?: string) {
    const response = await this.fetchWithCredentials(`${API_BASE_URL}/shorty`, {
      method: 'POST',
      body: JSON.stringify({ url, custom_name: customName })
    });
    return response.json();
  }

  async deleteShorty(shorty: string) {
    await this.fetchWithCredentials(`${API_BASE_URL}/${shorty}`, {
      method: 'DELETE'
    });
  }

  async renameShorty(oldName: string, newName: string) {
    await this.fetchWithCredentials(`${API_BASE_URL}/${oldName}/${newName}`, {
      method: 'PATCH'
    });
  }
}

export const api = new ApiClient();
