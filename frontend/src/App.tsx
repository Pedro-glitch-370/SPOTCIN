import HoraAtual from './components/Hora';
import { Navbar } from './components/Navbar';
import Cin from '../src/assets/Cin.png';
import { useEffect, useState } from 'react';
import { clearToken, getToken } from "./services/api";
import { HealthIndicator } from './components/HealthIndicator';
import { SongList } from './components/SongList';
import { Player } from './components/Player';
import { ErrorToast } from './components/ErrorToast';
import { AuthForm } from './components/AuthForm';
import type { Song } from './types/song';

export function App() {
  const [authed, setAuthed] = useState<boolean>(!!getToken());
  const [selected, setSelected] = useState<Song | null>(null);
  const [toast, setToast] = useState<string | null>(null);

  // qualquer 401 em qualquer chamada ativa o logout
  useEffect(() => {
    const onUnauthorized = () => {
      setAuthed(false);
      setSelected(null); // fecha o player
      setToast("Sessão expirada. Faça login novamente.");
    };
    window.addEventListener("spotcin:unauthorized", onUnauthorized);
    return () => window.removeEventListener("spotcin:unauthorized", onUnauthorized);
  }, []);

  if (!authed) {
    return <AuthForm onAuth={() => setAuthed(true)} />;
  }

  function handleLogout() {
    clearToken();
    setAuthed(false);
    setSelected(null);
  }

  return (
    <div className="app">
      <Navbar onLogout={handleLogout} />
      
      <article className="painel">

        <main className="main-area">
          <header className="top-bar">
            <section className="logo-message">
              <img src={Cin} alt="Logo do cin" className='logo' />
              <HoraAtual />
            </section>
          </header>

          <section className="catalog-section">
            <h2>Catálogo</h2>
            <SongList selectedId={selected?.id ?? null} onSelect={setSelected} />
          </section>
        </main>

        <aside className="sidebar">
          <HealthIndicator />
          <section className="player-section">
            <Player
              key={selected?.id}
              song={selected}
              onClose={() => setSelected(null)}
            />
          </section>
        </aside>
      </article>

      <ErrorToast message={toast} onClose={() => setToast(null)} />
    </div>
  );
}
