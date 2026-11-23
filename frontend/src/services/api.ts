import { ApiError, type HealthReady, type Playback, type Song } from "../types/song";

const BASE_URL = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

// user_id ainda provisório, o UUID fica persistido no localStorage
// se tiver uma auth real, aí vai vir do token
export function getUserId(): string {
  const key = "spotcin_user_id";
  let id = localStorage.getItem(key);
  if (!id) {
    id = crypto.randomUUID();
    localStorage.setItem(key, id);
  }
  return id; // retora um UUID v4
}

async function request<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const url = `${BASE_URL}${path}`;
  let resp: Response;
  try {
    resp = await fetch(url, {
      ...options,
      headers: {
        ...(options.body ? { "Content-Type": "application/json" } : {}),
        ...options.headers,
      },
    });
  } catch (e) {
    throw new ApiError(
      "network",
      "Não foi possível conectar ao servidor. Verifique se o backend está rodando.",
      undefined,
      String(e),
    );
  }

  if (!resp.ok) {
    const detail = await resp.text().catch(() => "");
    if (resp.status === 404) {
      throw new ApiError("not_found", "Faixa ou posição não encontrada.", 404, detail);
    }
    if (resp.status === 416) {
      throw new ApiError("range", "Posição inválida para esta faixa.", 416, detail);
    }
    if (resp.status === 422) {
      throw new ApiError("validation", "Dados inválidos.", 422, detail);
    }
    if (resp.status === 503) {
      throw new ApiError(
        "unavailable",
        "Sistema temporariamente indisponível (Postgres/Redis/Garage).",
        503,
        detail,
      );
    }
    throw new ApiError("unknown", `Erro ${resp.status}.`, resp.status, detail);
  }

  if (resp.status === 204) {
    return undefined as T;
  }
  return (await resp.json()) as T;
}

export const api = {
  listSongs: () => request<Song[]>("/api/songs"),

  getSong: (id: string) => request<Song>(`/api/songs/${id}`),

  getPlayback: async (id: string): Promise<Playback | null> => {
    try {
      return await request<Playback>(`/api/playback/${id}`, {
        headers: { "X-User-ID": getUserId() },
      });
    } catch (e) {
      if (e instanceof ApiError && e.kind === "not_found") {
        return null; // sem posição salva, aí começa do zero
      }
      throw e;
    }
  },

  putPlayback: (id: string, posicaoSeg: number) =>
    request<void>(`/api/playback/${id}`, {
      method: "PUT",
      headers: { "X-User-ID": getUserId() },
      body: JSON.stringify({ posicao_seg: posicaoSeg }),
    }),

  healthReady: () => request<HealthReady>("/health/ready"),

  // URL do stream pro <audio src=...>
  streamUrl: (id: string) => `${BASE_URL}/api/stream/${id}`,
};