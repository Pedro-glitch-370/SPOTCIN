import { useAudio } from "../hooks/useAudio";
import { usePlayback } from "../hooks/usePlayback";
import type { Song } from "../types/song";
import { api } from "../services/api";

interface Props {
  song: Song | null;
  initialPosition: number; // segundos
  onClose: () => void;
}

function fmt(sec: number): string {
  if (!Number.isFinite(sec) || sec < 0) return "0:00";
  const m = Math.floor(sec / 60);
  const s = Math.floor(sec % 60);
  return `${m}:${s.toString().padStart(2, "0")}`;
}

export function Player({ song, initialPosition, onClose }: Props) {
  const src = song ? api.streamUrl(song.id) : null;
  const { currentTime, duration, buffered, playing, loading, error, play, pause, seek } =
    useAudio(src, initialPosition);

  usePlayback(song?.id ?? null, currentTime, playing);

  if (!song) {
    return <div className="player-empty">Selecione uma faixa para começar.</div>;
  }

  const bufferPct = duration > 0 ? (buffered / duration) * 100 : 0;
  //const progressPct = duration > 0 ? (currentTime / duration) * 100 : 0;

  return (
    <div className="player">
      <button className="close-btn" onClick={onClose} aria-label="Fechar player">
        ✕
      </button>
      <h2>{song.title}</h2>
      <div className="muted">
        {song.artist} · {song.mime_type}
      </div>

      {error && <div className="error">{error}</div>}

      <div className="bar-row">
        <span className="bar-label">Reprodução</span>
        <div className="bar-track">
          <input
            type="range"
            min={0}
            max={duration || 0}
            step={0.1}
            value={currentTime}
            onChange={(e) => seek(Number(e.target.value))}
            className="progress"
          />
          </div>
      </div>

      <div className="time-row">
        <span>{fmt(currentTime)}</span>
        <span>{fmt(duration)}</span>
      </div>

      <div className="bar-row">
        <span className="bar-label">Buffer</span>
        <div className="bar-track">
          <div className="buffer-bar" style={{ width: `${bufferPct}%` }} />
        </div>
      </div>
      <div className="time-row">
        <span>{fmt(0)}</span>
        <span>{fmt(buffered)}</span>
      </div>

      <div className="controls">
        <button onClick={playing ? pause : play} disabled={loading}>
          {loading ? "…" : playing ? "Pausar" : "Reproduzir"}
        </button>
        <span className="muted">
          {playing ? "tocando" : "pausado"} · buffer {fmt(buffered)}
        </span>
      </div>
    </div>
  );
}