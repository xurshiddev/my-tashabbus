import React from 'react';
import ReactDOM from 'react-dom/client';
import './styles.css';

const cards = ['API Status', 'MFY Dashboard', 'Streets', 'Users', 'Reports'];

function App() {
  return (
    <main className="shell">
      <section className="intro">
        <p className="eyebrow">Stage 0 Foundation</p>
        <h1>My Tashabbus Admin</h1>
        <p>MFY monitoring dashboard foundation</p>
      </section>

      <section className="grid" aria-label="Admin dashboard placeholders">
        {cards.map((card) => (
          <article className="card" key={card}>
            <h2>{card}</h2>
            <p>Ready for future API-backed workflows.</p>
          </article>
        ))}
      </section>
    </main>
  );
}

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
