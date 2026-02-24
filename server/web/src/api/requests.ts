class Requests {
  private baseURL = '';

  constructor() {
    // @ts-ignore - 忽略import.meta的类型检查
    this.baseURL = import.meta.env?.DEV ? '' : '';
  }

  get = async <T>(url: string): Promise<T> => {
    const response = await fetch(`${this.baseURL}${url}`);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    return response.json();
  };

  post = async <T>(url: string, data: any): Promise<T> => {
    const response = await fetch(`${this.baseURL}${url}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    });
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    return response.json();
  };

  postForm = async <T>(url: string, data: Record<string, any>): Promise<T> => {
    const formData = new URLSearchParams();
    Object.entries(data).forEach(([key, value]) => {
      formData.append(key, String(value));
    });
    
    const response = await fetch(`${this.baseURL}${url}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: formData,
    });
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    return response.json();
  };
}

const requests = new Requests();
export default requests;
