import React, { useEffect, useState } from 'react';
import ReactDOM from 'react-dom/client';
import {
  assignChairman,
  assignResponsible,
  assignStreetLeader,
  createHousehold,
  createMFY,
  createStreet,
  devLoginAsSuperAdmin,
  fetchCurrentUser,
  listHouseholds,
  listMFYs,
  listResponsibleAssignments,
  listStreets,
  type Household,
  type MFY,
  type ResponsibleAssignment,
  type Street,
  type User,
} from './api';
import './styles.css';

const cards = ['API Status', 'MFY Dashboard', 'Streets', 'Users', 'Reports'];

function App() {
  const [token, setToken] = useState(() => localStorage.getItem('my_tashabbus_admin_token') ?? '');
  const [user, setUser] = useState<User | null>(null);
  const [status, setStatus] = useState('Not authenticated');
  const [mfys, setMFYs] = useState<MFY[]>([]);
  const [streets, setStreets] = useState<Street[]>([]);
  const [households, setHouseholds] = useState<Household[]>([]);
  const [responsibleAssignments, setResponsibleAssignments] = useState<ResponsibleAssignment[]>([]);
  const [mfyID, setMFYID] = useState('');
  const [streetID, setStreetID] = useState('');
  const [chairmanUserID, setChairmanUserID] = useState('');
  const [leaderUserID, setLeaderUserID] = useState('');
  const [responsibleUserID, setResponsibleUserID] = useState('');
  const [stageStatus, setStageStatus] = useState('Stage 3 tools ready');

  useEffect(() => {
    if (!token) {
      return;
    }
    fetchCurrentUser(token)
      .then((currentUser) => {
        setUser(currentUser);
        setStatus('Authenticated');
      })
      .catch(() => {
        setUser(null);
        setStatus('Session needs refresh');
      });
  }, [token]);

  async function handleDevLogin() {
    setStatus('Signing in...');
    try {
      const result = await devLoginAsSuperAdmin();
      localStorage.setItem('my_tashabbus_admin_token', result.access_token);
      setToken(result.access_token);
      setUser(result.user);
      setStatus('Authenticated');
    } catch {
      setStatus('Dev login failed');
    }
  }

  async function handleCreateMFY(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    try {
      const targetVotes = Number(form.get('target_votes') || 0);
      const mfy = await createMFY(token, {
        name: String(form.get('name') ?? ''),
        region: String(form.get('region') ?? '') || null,
        district: String(form.get('district') ?? '') || null,
        target_votes: Number.isNaN(targetVotes) ? null : targetVotes,
        season: String(form.get('season') ?? '') || null,
      });
      setMFYID(mfy.id);
      setStageStatus('MFY created');
    } catch {
      setStageStatus('MFY create failed');
    }
  }

  async function handleListMFYs() {
    try {
      setMFYs(await listMFYs(token));
      setStageStatus('MFYs loaded');
    } catch {
      setStageStatus('MFY list failed');
    }
  }

  async function handleCreateStreet(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    try {
      const street = await createStreet(token, mfyID, {
        name: String(form.get('name') ?? ''),
        planned_households_count: Number(form.get('planned_households_count') || 0),
        notes: String(form.get('notes') ?? '') || null,
      });
      setStreetID(street.id);
      setStageStatus('Street created');
    } catch {
      setStageStatus('Street create failed');
    }
  }

  async function handleListStreets() {
    try {
      setStreets(await listStreets(token, mfyID));
      setStageStatus('Streets loaded');
    } catch {
      setStageStatus('Street list failed');
    }
  }

  async function handleAssignChairman() {
    try {
      await assignChairman(token, mfyID, chairmanUserID);
      setStageStatus('Chairman assigned');
    } catch {
      setStageStatus('Chairman assignment failed');
    }
  }

  async function handleAssignLeader() {
    try {
      await assignStreetLeader(token, streetID, leaderUserID);
      setStageStatus('Street leader assigned');
    } catch {
      setStageStatus('Street leader assignment failed');
    }
  }

  async function handleCreateHousehold(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    try {
      const household = await createHousehold(token, streetID, {
        house_number: String(form.get('house_number') ?? ''),
        total_numbers: Number(form.get('total_numbers') || 0),
        contacted_numbers: Number(form.get('contacted_numbers') || 0),
        voted_numbers: Number(form.get('voted_numbers') || 0),
        status: String(form.get('status') ?? 'NEW'),
        notes: String(form.get('notes') ?? '') || null,
      });
      setHouseholds((items) => [household, ...items]);
      setStageStatus('Household created');
    } catch {
      setStageStatus('Household create failed');
    }
  }

  async function handleListHouseholds() {
    try {
      setHouseholds(await listHouseholds(token, streetID));
      setStageStatus('Households loaded');
    } catch {
      setStageStatus('Household list failed');
    }
  }

  async function handleAssignResponsible(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    try {
      const assignment = await assignResponsible(token, streetID, {
        user_id: responsibleUserID,
        from_house_number: String(form.get('from_house_number') ?? ''),
        to_house_number: String(form.get('to_house_number') ?? ''),
      });
      setResponsibleAssignments((items) => [assignment, ...items]);
      setStageStatus('Responsible person assigned');
    } catch {
      setStageStatus('Responsible assignment failed');
    }
  }

  async function handleListResponsibleAssignments() {
    try {
      setResponsibleAssignments(await listResponsibleAssignments(token, streetID));
      setStageStatus('Responsible assignments loaded');
    } catch {
      setStageStatus('Responsible assignment list failed');
    }
  }

  return (
    <main className="shell">
      <section className="intro">
        <p className="eyebrow">Stage 0 Foundation</p>
        <h1>My Tashabbus Admin</h1>
        <p>MFY monitoring dashboard foundation</p>
      </section>

      <section className="auth-panel" aria-label="Development authentication">
        <div>
          <h2>Authentication</h2>
          <p>{user ? `${user.full_name} - ${user.role}` : status}</p>
        </div>
        <button type="button" onClick={handleDevLogin}>
          Dev Login as Super Admin
        </button>
      </section>

      <section className="grid" aria-label="Admin dashboard placeholders">
        {cards.map((card) => (
          <article className="card" key={card}>
            <h2>{card}</h2>
            <p>Ready for future API-backed workflows.</p>
          </article>
        ))}
      </section>

      <section className="stage-grid" aria-label="Stage 3 management">
        <article className="tool-panel">
          <h2>MFY Management</h2>
          <p>{stageStatus}</p>
          <form onSubmit={handleCreateMFY}>
            <input name="name" placeholder="MFY name" />
            <input name="region" placeholder="Region" />
            <input name="district" placeholder="District" />
            <input name="target_votes" placeholder="Target votes" type="number" min="0" />
            <input name="season" placeholder="Season" />
            <button type="submit">Create MFY</button>
          </form>
          <button type="button" onClick={handleListMFYs}>List MFYs</button>
          <div className="result-list">
            {mfys.map((mfy) => <button type="button" key={mfy.id} onClick={() => setMFYID(mfy.id)}>{mfy.name}</button>)}
          </div>
        </article>

        <article className="tool-panel">
          <h2>Street Management</h2>
          <input value={mfyID} onChange={(event) => setMFYID(event.target.value)} placeholder="MFY ID" />
          <form onSubmit={handleCreateStreet}>
            <input name="name" placeholder="Street name" />
            <input name="planned_households_count" placeholder="Planned households" type="number" min="0" />
            <input name="notes" placeholder="Notes" />
            <button type="submit">Create Street</button>
          </form>
          <button type="button" onClick={handleListStreets}>List Streets</button>
          <div className="result-list">
            {streets.map((street) => <button type="button" key={street.id} onClick={() => setStreetID(street.id)}>{street.name}</button>)}
          </div>
        </article>

        <article className="tool-panel">
          <h2>Assign Chairman</h2>
          <input value={mfyID} onChange={(event) => setMFYID(event.target.value)} placeholder="MFY ID" />
          <input value={chairmanUserID} onChange={(event) => setChairmanUserID(event.target.value)} placeholder="User ID" />
          <button type="button" onClick={handleAssignChairman}>Assign Chairman</button>
        </article>

        <article className="tool-panel">
          <h2>Assign Street Leader</h2>
          <input value={streetID} onChange={(event) => setStreetID(event.target.value)} placeholder="Street ID" />
          <input value={leaderUserID} onChange={(event) => setLeaderUserID(event.target.value)} placeholder="User ID" />
          <button type="button" onClick={handleAssignLeader}>Assign Street Leader</button>
        </article>

        <article className="tool-panel">
          <h2>Household Management</h2>
          <input value={streetID} onChange={(event) => setStreetID(event.target.value)} placeholder="Street ID" />
          <form onSubmit={handleCreateHousehold}>
            <input name="house_number" placeholder="House number" />
            <input name="total_numbers" placeholder="Total numbers" type="number" min="0" />
            <input name="contacted_numbers" placeholder="Contacted numbers" type="number" min="0" />
            <input name="voted_numbers" placeholder="Voted numbers" type="number" min="0" />
            <select name="status" defaultValue="NEW">
              <option value="NEW">NEW</option>
              <option value="VISITED">VISITED</option>
              <option value="EXPLAINED">EXPLAINED</option>
              <option value="PARTIALLY_VOTED">PARTIALLY_VOTED</option>
              <option value="FULLY_VOTED">FULLY_VOTED</option>
              <option value="NOT_HOME">NOT_HOME</option>
              <option value="CALLBACK_NEEDED">CALLBACK_NEEDED</option>
              <option value="REFUSED">REFUSED</option>
              <option value="INVALID_INFO">INVALID_INFO</option>
            </select>
            <input name="notes" placeholder="Notes" />
            <button type="submit">Create Household</button>
          </form>
          <button type="button" onClick={handleListHouseholds}>List Households</button>
          <div className="result-list">
            {households.map((household) => (
              <button type="button" key={household.id} onClick={() => setStreetID(household.street_id)}>
                {household.house_number} - {household.status}
              </button>
            ))}
          </div>
        </article>

        <article className="tool-panel">
          <h2>Responsible Assignment</h2>
          <input value={streetID} onChange={(event) => setStreetID(event.target.value)} placeholder="Street ID" />
          <input value={responsibleUserID} onChange={(event) => setResponsibleUserID(event.target.value)} placeholder="Responsible user ID" />
          <form onSubmit={handleAssignResponsible}>
            <input name="from_house_number" placeholder="From house number" />
            <input name="to_house_number" placeholder="To house number" />
            <button type="submit">Assign Responsible</button>
          </form>
          <button type="button" onClick={handleListResponsibleAssignments}>List Responsible Assignments</button>
          <div className="result-list">
            {responsibleAssignments.map((assignment) => (
              <button type="button" key={assignment.id} onClick={() => setResponsibleUserID(assignment.responsible_user_id)}>
                {assignment.from_house_number}-{assignment.to_house_number}
              </button>
            ))}
          </div>
        </article>
      </section>
    </main>
  );
}

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
