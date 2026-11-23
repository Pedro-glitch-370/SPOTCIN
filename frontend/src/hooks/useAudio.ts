import { useCallback, useEffect, useRef, useState } from "react";

export interface AudioState {
  currentTime: number;
  duration: number;
  buffered: number; // a partir de currentTime
  playing: boolean;
  loading: boolean;
  error: string | null;
}

export function useAudio(src: string | null, initialPosition = 0) {
  const audioRef = useRef<HTMLAudioElement | null>(null);
  const [state, setState] = useState<AudioState>({
    currentTime: 0,
    duration: 0,
    buffered: 0,
    playing: false,
    loading: false,
    error: null,
  });

  // cria o <audio> uma vez só
  useEffect(() => {
    const el = new Audio();
    el.preload = "auto";
    audioRef.current = el; // usando audioref pra persistir caso renderize

    const onTime = () => {
      const buffered = el.buffered.length > 0
        ? el.buffered.end(el.buffered.length - 1)
        : 0;
      setState((s) => ({
        ...s,
        currentTime: el.currentTime,
        duration: el.duration || 0,
        buffered,
      }));
    };
    const onPlay = () => setState((s) => ({ ...s, playing: true }));
    const onPause = () => setState((s) => ({ ...s, playing: false }));
    const onWaiting = () => setState((s) => ({ ...s, loading: true }));
    const onCanPlay = () => setState((s) => ({ ...s, loading: false }));
    const onError = () => {
      const code = el.error?.code;
      let msg = "Falha ao carregar o áudio.";
      if (code === 2) msg = "Erro de rede ao carregar o áudio.";
      if (code === 3) msg = "Falha ao decodificar o áudio.";
      if (code === 4) msg = "Formato de áudio não suportado.";
      setState((s) => ({ ...s, error: msg, loading: false, playing: false }));
    };

    el.addEventListener("timeupdate", onTime);
    el.addEventListener("progress", onTime);
    el.addEventListener("play", onPlay);
    el.addEventListener("pause", onPause);
    el.addEventListener("waiting", onWaiting);
    el.addEventListener("canplay", onCanPlay);
    el.addEventListener("error", onError);

    return () => {
      el.pause();
      el.removeEventListener("timeupdate", onTime);
      el.removeEventListener("progress", onTime);
      el.removeEventListener("play", onPlay);
      el.removeEventListener("pause", onPause);
      el.removeEventListener("waiting", onWaiting);
      el.removeEventListener("canplay", onCanPlay);
      el.removeEventListener("error", onError);
      el.src = "";
    };
  }, []);

  // troca de faixa
  useEffect(() => {
    const el = audioRef.current;
    if (!el) return;
    if (!src) {
      el.removeAttribute("src");
      el.load();
      setState({ currentTime: 0, duration: 0, buffered: 0, playing: false, loading: false, error: null });
      return;
    }
    el.src = src;
    el.load();
    setState((s) => ({ ...s, error: null, loading: true }));

    // quando os metadados carregarem, faz seek para initialPosition
    const onLoaded = () => {
      if (initialPosition > 0 && Number.isFinite(el.duration)) {
        el.currentTime = Math.min(initialPosition, el.duration - 0.1);
      }
      setState((s) => ({ ...s, duration: el.duration || 0 }));
    };
    el.addEventListener("loadedmetadata", onLoaded, { once: true });
    return () => el.removeEventListener("loadedmetadata", onLoaded);
  }, [src, initialPosition]);

  const play = useCallback(() => audioRef.current?.play().catch((e) => {
    setState((s) => ({ ...s, error: `Falha ao reproduzir: ${e}` }));
  }), []);

  const pause = useCallback(() => audioRef.current?.pause(), []);

  const seek = useCallback((seconds: number) => {
    const el = audioRef.current;
    if (!el) return;
    // <audio> faz Range automaticamente pro novo offset
    el.currentTime = seconds;
  }, []);

  return { ...state, play, pause, seek };
}