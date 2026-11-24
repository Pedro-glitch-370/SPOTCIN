import { useEffect, useState } from "react";
import { api } from "../services/api";
import type { HealthReady } from "../types/song";

export function HealthIndicator() {
  const [health, setHealth] = useState<HealthReady | null>(null);
  const [offline, setOffline] = useState(false);

  useEffect(() => {
    let cancelled = false;

    const check = async () => {
      try {
        const h = await api.healthReady();
        if (!cancelled) {
          setHealth(h);
          setOffline(false);
        }
      } catch {
        if (!cancelled) {
          setHealth(null);
          setOffline(true);
        }
      }
    };

    check();
    const t = setInterval(check, 10000);
    return () => {
      cancelled = true;
      clearInterval(t);
    };
  }, []);

  const ok = health?.status === "ok";

  return (
    <div className={`health ${ok ? "ok" : "bad"}`}>
      <div className="health-main">
        <span className="dot" />
        <span className="health-text">
          {offline ? "backend offline" : ok ? "Sistema Ok" : "Sistema Degradado"}
        </span>
      </div>
      {health?.checks && !ok && (
        <ul className="checks">
          {Object.entries(health.checks).map(([k, v]) => (
            <li key={k}>
              {k}: {v}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}