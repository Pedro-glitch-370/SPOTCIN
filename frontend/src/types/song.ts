export interface Song {
  id: string;
  title: string;
  artist: string;
  duration: number; // segundos
  size: number;     // bytes
  mime_type: string;
}

export interface Playback {
  track_id: string;
  posicao_seg: number;
  atualizado_em: string;
}

export interface HealthReady {
  status: "ok" | "degraded";
  checks: Record<string, string>;
}

export type ApiErrorKind =
  | "network"
  | "not_found"
  | "range"
  | "unavailable"
  | "validation"
  | "unauthorized"
  | "unknown";

export class ApiError extends Error {
  kind: ApiErrorKind;
  status?: number;
  detail?: string;

  constructor(kind: ApiErrorKind, message: string, status?: number, detail?: string) {
    super(message);
    this.kind = kind;
    this.status = status;
    this.detail = detail;
  }
}