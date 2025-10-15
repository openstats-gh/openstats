type SessionSummary = {
  displayName: string;
  slug: string;
};

export interface Api {
  with(fetchShim: typeof fetch): Api;
  getSessionSummary(): Promise<SessionSummary | null>;
}

export class StubbedApi implements Api {
  with(): Api {
    return this;
  }

  getSessionSummary(): Promise<SessionSummary | null> {
    return Promise.reject();
  }
}

export class LiveApi implements Api {
  private readonly _fetch: typeof fetch;
  constructor(fetchShim: typeof fetch) {
    this._fetch = fetchShim;
  }

  with(fetchShim: typeof fetch): Api {
    return new LiveApi(fetchShim);
  }

  async getSessionSummary(): Promise<SessionSummary | null> {
    const response = await this._fetch(`/api/internal/session/`);
    if (!response.ok) {
      if (response.status === 401) {
        return null;
      }
      throw await response.json();
    }

    // noinspection UnnecessaryLocalVariableJS
    const result: SessionSummary = await response.json();
    return result;
  }
}

export const api: Api = new LiveApi(fetch);
// export const api: Api = new StubbedApi()
