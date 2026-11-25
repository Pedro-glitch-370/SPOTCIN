import { ApiError, type HealthReady, type Playback, type Song } from "../types/song";

const BASE_URL = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

const TOKEN_KEY = "spotcin_token";

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY);
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const url = `${BASE_URL}${path}`;
  const token = getToken();

  let resp: Response;
  try {
    resp = await fetch(url, {
      ...options,
      headers: {
        ...(options.body ? { "Content-Type": "application/json" } : {}),
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...options.headers,
      },
    });
  } catch (e) {
    throw new ApiError("network", "Não foi possível conectar ao servidor.", undefined, String(e));
  }

  if (resp.status === 401) {
    clearToken();
    throw new ApiError("unauthorized", "Sessão expirada. Faça login novamente.", 401);
  }

  if (!resp.ok) {
    const detail = await resp.text().catch(() => "");
    if (resp.status === 404) throw new ApiError("not_found", "Faixa ou posição não encontrada.", 404, detail);
    if (resp.status === 416) throw new ApiError("range", "Posição inválida para esta faixa.", 416, detail);
    if (resp.status === 422) throw new ApiError("validation", "Dados inválidos.", 422, detail);
    if (resp.status === 503) throw new ApiError("unavailable", "Sistema temporariamente indisponível.", 503, detail);
    throw new ApiError("unknown", `Erro ${resp.status}.`, resp.status, detail);
  }

  if (resp.status === 204) return undefined as T;
  return (await resp.json()) as T;
}

export const api = {
  // auth
  register: (email: string, password: string) =>
    request<{ id: string; email: string }>("/api/auth/register", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),

  login: (email: string, password: string) =>
    request<{ token: string }>("/api/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),

  me: () => request<{ id: string; email: string }>("/api/auth/me"),
  
  // catálogo e stream
  listSongs: () => request<Song[]>("/api/songs"),
  getSong: (id: string) => request<Song>(`/api/songs/${id}`),
  streamUrl: (id: string) => `${BASE_URL}/api/stream/${id}`,
  
  // playback
  getPlayback: async (id: string): Promise<Playback | null> => {
    try {
      return await request<Playback>(`/api/playback/${id}`);
    } catch (e) {
      if (e instanceof ApiError && e.kind === "not_found") return null;
      throw e;
    }
  },

  putPlayback: (id: string, posicaoSeg: number) =>
    request<void>(`/api/playback/${id}`, {
      method: "PUT",
      body: JSON.stringify({ posicao_seg: posicaoSeg }),
    }),

  // health
  healthReady: () => request<HealthReady>("/health/ready"),
};