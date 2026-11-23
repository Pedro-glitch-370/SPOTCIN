import { useEffect } from "react";

interface Props {
  message: string | null;
  onClose: () => void;
}

export function ErrorToast({ message, onClose }: Props) {
  useEffect(() => {
    if (!message) return;
    const t = setTimeout(onClose, 6000);
    return () => clearTimeout(t);
  }, [message, onClose]);

  if (!message) return null;
  return (
    <div className="toast" onClick={onClose} role="alert">
      {message} <span className="muted">(clique para fechar)</span>
    </div>
  );
}