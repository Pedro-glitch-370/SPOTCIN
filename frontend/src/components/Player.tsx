import { useAudio } from "../hooks/useAudio";
import type { Song } from "../types/song";
import { api } from "../services/api";
import { useState } from "react";

interface Props {
  song: Song | null;
  onClose: () => void;
}

function fmt(sec: number): string {
  if (!Number.isFinite(sec) || sec < 0) return "0:00";
  const m = Math.floor(sec / 60);
  const s = Math.floor(sec % 60);
  return `${m}:${s.toString().padStart(2, "0")}`;
}

export function Player({ song, onClose }: Props) {
  const [savingPos, setSavingPos] = useState(false);
  const [resumeMsg, setResumeMsg] = useState<string | null>(null);

  const src = song ? api.streamUrl(song.id) : null;
  const { currentTime, duration, buffered, playing, loading, error, play, pause, seek } =
    useAudio(src);

  if (!song) {
    return <div className="player-empty">Selecione uma faixa para começar.</div>;
  }

  async function handleSave() {
    if (!song) return;
    setSavingPos(true);
    try {
      await api.putPlayback(song.id, currentTime);
      setResumeMsg(`Posição ${fmt(currentTime)} salva.`);
    } catch {
      setResumeMsg("Falha ao salvar posição.");
    } finally {
      setSavingPos(false);
      setTimeout(() => setResumeMsg(null), 3000);
    }
  }

  async function handleResume() {
    if (!song) return;
    try {
      const p = await api.getPlayback(song.id);
      if (!p || p.posicao_seg <= 0) {
        setResumeMsg("Nenhuma posição salva.");
      } else {
        seek(p.posicao_seg);
        setResumeMsg(`Retomado em ${fmt(p.posicao_seg)}.`);
      }
    } catch {
      setResumeMsg("Falha ao buscar posição.");
    } finally {
      setTimeout(() => setResumeMsg(null), 3000);
    }
  }

  const bufferPct = duration > 0 ? (buffered / duration) * 100 : 0;

  return (
    <div className="player-conteiner">
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
          <span>{fmt(buffered)}</span>
          <span>{fmt(duration)}</span>
        </div>

        <div className="controls">
          <button onClick={playing ? pause : play} disabled={loading}>
            {loading ? "…" : playing ? "Pausar" : "Reproduzir"}
          </button>
          <button onClick={handleSave} disabled={savingPos}>
            {savingPos ? "Salvando…" : "Salvar posição"}
          </button>
          <button onClick={handleResume}>Retomar</button>
        </div>
      </div>
      {resumeMsg && <div className="health">{resumeMsg}</div>}
    </div>
  );
}