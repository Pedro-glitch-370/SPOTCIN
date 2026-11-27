import { useEffect, useState } from "react";
import { api } from "../services/api";
import { ApiError, type Song } from "../types/song";

interface Props {
  selectedId: string | null;
  onSelect: (song: Song) => void;
}

export function SongList({ selectedId, onSelect }: Props) {
  const [songs, setSongs] = useState<Song[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    api
      .listSongs()
      .then((s) => {
        if (!cancelled) setSongs(s);
      })
      .catch((e) => {
        if (!cancelled) {
          setError(
            e instanceof ApiError ? e.message : "Erro ao carregar catálogo.",
          );
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  if (loading) return <p className="muted">Carregando catálogo…</p>;
  if (error) return <p className="error">Catálogo: {error}</p>;
  if (songs.length === 0) return <p className="muted">Catálogo vazio.</p>;

  return (
    <div className="song-grid">
      {songs.map((s) => (
        <div
          key={s.id}
          className={`song-card ${s.id === selectedId ? "selected" : ""}`}
          onClick={() => onSelect(s)}
        >
          <div className="song-title">{s.title}</div>
          <div className="song-meta">
            {Math.floor(s.duration)}s
          </div>
        </div>
      ))}
    </div>
  );
}