import { useEffect, useRef } from "react";
import { api } from "../services/api";
import { ApiError } from "../types/song";

const PUT_INTERVAL_MS = 5000;

export function usePlayback(songId: string | null, currentTime: number, playing: boolean) {
  const lastPutRef = useRef(0);

  useEffect(() => {
    if (!songId || !playing) return;
    const now = Date.now();
    if (now - lastPutRef.current < PUT_INTERVAL_MS) return;
    lastPutRef.current = now;

    api.putPlayback(songId, currentTime).catch((e) => {
      if (e instanceof ApiError && e.kind === "network") {
        console.warn("PUT playback falhou (rede):", e.message);
      } else {
        console.warn("PUT playback falhou:", e);
      }
    });
  }, [songId, currentTime, playing]);

  // PUT final quando sair ou trocar de faixa
  useEffect(() => {
    return () => {
      if (songId) {
        api.putPlayback(songId, currentTime).catch(() => {});
      }
    };
  }, [songId]);
}