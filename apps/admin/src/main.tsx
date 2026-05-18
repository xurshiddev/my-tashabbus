import React, { useEffect, useMemo, useState } from 'react';
import ReactDOM from 'react-dom/client';
import {
  apiBaseUrl,
  assignChairman,
  assignResponsible,
  assignStreetLeader,
  bindTelegram,
  createHousehold,
  createMFY,
  createStreet,
  createUser,
  deactivateResponsibleAssignment,
  devLoginAsSuperAdmin,
  fetchCurrentUser,
  friendlyError,
  listHouseholdLogs,
  listHouseholds,
  listMFYs,
  listResponsibleAssignments,
  listStreets,
  listUsers,
  tokenStore,
  updateHousehold,
  type Household,
  type HouseholdLog,
  type HouseholdStatus,
  type MFY,
  type ResponsibleAssignment,
  type Role,
  type Street,
  type User,
} from './api';
import './styles.css';

type Tab = 'overview' | 'users' | 'mfys' | 'streets' | 'households' | 'assignments' | 'devtools';
type Notice = { tone: 'info' | 'success' | 'error'; text: string };

const roles: Role[] = ['SUPER_ADMIN', 'MFY_CHAIRMAN', 'STREET_LEADER', 'RESPONSIBLE_PERSON'];
const statuses: HouseholdStatus[] = [
  'NEW',
  'VISITED',
  'EXPLAINED',
  'PARTIALLY_VOTED',
  'FULLY_VOTED',
  'NOT_HOME',
  'CALLBACK_NEEDED',
  'REFUSED',
  'INVALID_INFO',
];
const tabs: { id: Tab; label: string }[] = [
  { id: 'overview', label: 'Overview' },
  { id: 'users', label: 'Users' },
  { id: 'mfys', label: 'MFYs' },
  { id: 'streets', label: 'Streets' },
  { id: 'households', label: 'Households' },
  { id: 'assignments', label: 'Assignments' },
  { id: 'devtools', label: 'Dev Tools' },
];

const store = tokenStore();

function App() {
  const [activeTab, setActiveTab] = useState<Tab>('overview');
  const [token, setToken] = useState(store.get);
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [notice, setNotice] = useState<Notice>({ tone: 'info', text: 'Dev login required' });
  const [busyAction, setBusyAction] = useState('');
  const [users, setUsers] = useState<User[]>([]);
  const [mfys, setMFYs] = useState<MFY[]>([]);
  const [streets, setStreets] = useState<Street[]>([]);
  const [households, setHouseholds] = useState<Household[]>([]);
  const [responsibleAssignments, setResponsibleAssignments] = useState<ResponsibleAssignment[]>([]);
  const [householdLogs, setHouseholdLogs] = useState<HouseholdLog[]>([]);
  const [selectedMFYID, setSelectedMFYID] = useState('');
  const [selectedStreetID, setSelectedStreetID] = useState('');
  const [selectedHouseholdID, setSelectedHouseholdID] = useState('');
  const [selectedUserID, setSelectedUserID] = useState('');

  const selectedHousehold = households.find((item) => item.id === selectedHouseholdID) ?? null;
  const chairmen = useMemo(() => users.filter((item) => item.role === 'MFY_CHAIRMAN'), [users]);
  const streetLeaders = useMemo(() => users.filter((item) => item.role === 'STREET_LEADER'), [users]);
  const responsiblePeople = useMemo(() => users.filter((item) => item.role === 'RESPONSIBLE_PERSON'), [users]);

  useEffect(() => {
    if (!token) {
      setCurrentUser(null);
      return;
    }
    run('me', () => fetchCurrentUser(token), {
      silent: true,
      onSuccess: (user) => {
        setCurrentUser(user);
        setNotice({ tone: 'success', text: 'Authenticated' });
      },
      onError: () => {
        store.clear();
        setToken('');
        setCurrentUser(null);
        setNotice({ tone: 'error', text: 'Sessiya tugagan. Qayta kiring.' });
      },
    });
  }, [token]);

  async function run<T>(
    action: string,
    task: () => Promise<T>,
    options: {
      success?: string;
      silent?: boolean;
      onSuccess?: (result: T) => void;
      onError?: (error: unknown) => void;
    } = {},
  ) {
    setBusyAction(action);
    if (!options.silent) {
      setNotice({ tone: 'info', text: 'Loading...' });
    }
    try {
      const result = await task();
      options.onSuccess?.(result);
      if (options.success) {
        setNotice({ tone: 'success', text: options.success });
      }
    } catch (error) {
      options.onError?.(error);
      if (!options.silent) {
        setNotice({ tone: 'error', text: friendlyError(error) });
      }
    } finally {
      setBusyAction('');
    }
  }

  function requireToken(): string | null {
    if (!token) {
      setNotice({ tone: 'error', text: 'Avval tizimga kiring.' });
      return null;
    }
    return token;
  }

  function logout() {
    store.clear();
    setToken('');
    setCurrentUser(null);
    setNotice({ tone: 'info', text: 'Logged out' });
  }

  function login() {
    run('login', devLoginAsSuperAdmin, {
      success: 'Dev Super Admin logged in',
      onSuccess: (result) => {
        store.set(result.access_token);
        setToken(result.access_token);
        setCurrentUser(result.user);
      },
    });
  }

  function refreshUsers() {
    const auth = requireToken();
    if (!auth) return;
    run('users:list', () => listUsers(auth), {
      success: 'Users loaded',
      onSuccess: setUsers,
    });
  }

  function refreshMFYs() {
    const auth = requireToken();
    if (!auth) return;
    run('mfys:list', () => listMFYs(auth), {
      success: 'MFYs loaded',
      onSuccess: (items) => {
        setMFYs(items);
        if (!selectedMFYID && items[0]) setSelectedMFYID(items[0].id);
      },
    });
  }

  function refreshStreets(mfyID = selectedMFYID) {
    const auth = requireToken();
    if (!auth || !mfyID) {
      setNotice({ tone: 'error', text: 'MFY tanlang.' });
      return;
    }
    run('streets:list', () => listStreets(auth, mfyID), {
      success: 'Streets loaded',
      onSuccess: (items) => {
        setStreets(items);
        if (!selectedStreetID && items[0]) setSelectedStreetID(items[0].id);
      },
    });
  }

  function refreshHouseholds(streetID = selectedStreetID) {
    const auth = requireToken();
    if (!auth || !streetID) {
      setNotice({ tone: 'error', text: 'Street tanlang.' });
      return;
    }
    run('households:list', () => listHouseholds(auth, streetID), {
      success: 'Households loaded',
      onSuccess: (items) => {
        setHouseholds(items);
        if (!selectedHouseholdID && items[0]) setSelectedHouseholdID(items[0].id);
      },
    });
  }

  function refreshResponsibleAssignments(streetID = selectedStreetID) {
    const auth = requireToken();
    if (!auth || !streetID) {
      setNotice({ tone: 'error', text: 'Street tanlang.' });
      return;
    }
    run('responsibles:list', () => listResponsibleAssignments(auth, streetID), {
      success: 'Assignments loaded',
      onSuccess: setResponsibleAssignments,
    });
  }

  function handleCreateUser(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const auth = requireToken();
    if (!auth) return;
    const form = new FormData(event.currentTarget);
    const fullName = text(form, 'full_name');
    if (!fullName) {
      setNotice({ tone: 'error', text: 'Full name required' });
      return;
    }
    run(
      'users:create',
      () =>
        createUser(auth, {
          full_name: fullName,
          phone: nullableText(form, 'phone'),
          role: text(form, 'role') as Role,
          telegram_id: nullableNumber(form, 'telegram_id'),
          telegram_username: nullableText(form, 'telegram_username'),
          mfy_id: nullableText(form, 'mfy_id'),
        }),
      {
        success: 'User created',
        onSuccess: (user) => {
          setUsers((items) => [user, ...items]);
          setSelectedUserID(user.id);
        },
      },
    );
  }

  function handleBindTelegram(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const auth = requireToken();
    if (!auth) return;
    const form = new FormData(event.currentTarget);
    const userID = text(form, 'user_id') || selectedUserID;
    const telegramID = nullableNumber(form, 'telegram_id');
    if (!userID || telegramID === null) {
      setNotice({ tone: 'error', text: 'User ID and Telegram ID required' });
      return;
    }
    run('users:telegram', () => bindTelegram(auth, userID, {
      telegram_id: telegramID,
      telegram_username: nullableText(form, 'telegram_username'),
    }), {
      success: 'Telegram identity bound',
      onSuccess: (updated) => setUsers((items) => upsert(items, updated)),
    });
  }

  function handleCreateMFY(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const auth = requireToken();
    if (!auth) return;
    const form = new FormData(event.currentTarget);
    const name = text(form, 'name');
    const targetVotes = nullableNumber(form, 'target_votes');
    if (!name) {
      setNotice({ tone: 'error', text: 'MFY name required' });
      return;
    }
    if (targetVotes !== null && targetVotes < 0) {
      setNotice({ tone: 'error', text: 'Target votes cannot be negative' });
      return;
    }
    run('mfys:create', () => createMFY(auth, {
      name,
      region: nullableText(form, 'region'),
      district: nullableText(form, 'district'),
      target_votes: targetVotes,
      season: nullableText(form, 'season'),
    }), {
      success: 'MFY created',
      onSuccess: (mfy) => {
        setMFYs((items) => [mfy, ...items]);
        setSelectedMFYID(mfy.id);
      },
    });
  }

  function handleAssignChairman(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const auth = requireToken();
    if (!auth) return;
    const form = new FormData(event.currentTarget);
    const mfyID = text(form, 'mfy_id') || selectedMFYID;
    const userID = text(form, 'user_id');
    if (!mfyID || !userID) {
      setNotice({ tone: 'error', text: 'MFY and chairman required' });
      return;
    }
    run('mfys:chairman', () => assignChairman(auth, mfyID, userID), {
      success: 'Chairman assigned',
      onSuccess: (updated) => setUsers((items) => upsert(items, updated)),
    });
  }

  function handleCreateStreet(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const auth = requireToken();
    if (!auth) return;
    const form = new FormData(event.currentTarget);
    const mfyID = text(form, 'mfy_id') || selectedMFYID;
    const name = text(form, 'name');
    const planned = numberValue(form, 'planned_households_count');
    if (!mfyID || !name) {
      setNotice({ tone: 'error', text: 'MFY and street name required' });
      return;
    }
    if (planned < 0) {
      setNotice({ tone: 'error', text: 'Planned households cannot be negative' });
      return;
    }
    run('streets:create', () => createStreet(auth, mfyID, {
      name,
      planned_households_count: planned,
      notes: nullableText(form, 'notes'),
    }), {
      success: 'Street created',
      onSuccess: (street) => {
        setStreets((items) => [street, ...items]);
        setSelectedStreetID(street.id);
      },
    });
  }

  function handleAssignStreetLeader(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const auth = requireToken();
    if (!auth) return;
    const form = new FormData(event.currentTarget);
    const streetID = text(form, 'street_id') || selectedStreetID;
    const userID = text(form, 'user_id');
    if (!streetID || !userID) {
      setNotice({ tone: 'error', text: 'Street and leader required' });
      return;
    }
    run('streets:leader', () => assignStreetLeader(auth, streetID, userID), {
      success: 'Street leader assigned',
    });
  }

  function householdPayload(form: FormData): Record<string, unknown> | null {
    const houseNumber = text(form, 'house_number');
    const total = numberValue(form, 'total_numbers');
    const contacted = numberValue(form, 'contacted_numbers');
    const voted = numberValue(form, 'voted_numbers');
    const status = text(form, 'status') as HouseholdStatus;
    if (!houseNumber) {
      setNotice({ tone: 'error', text: 'House number required' });
      return null;
    }
    if (total < 0 || contacted < 0 || voted < 0) {
      setNotice({ tone: 'error', text: 'Counts cannot be negative' });
      return null;
    }
    if (contacted > total || voted > total) {
      setNotice({ tone: 'error', text: 'Contacted and voted counts cannot exceed total' });
      return null;
    }
    if (!statuses.includes(status)) {
      setNotice({ tone: 'error', text: 'Invalid status' });
      return null;
    }
    return {
      house_number: houseNumber,
      total_numbers: total,
      contacted_numbers: contacted,
      voted_numbers: voted,
      status,
      notes: nullableText(form, 'notes'),
    };
  }

  function handleCreateHousehold(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const auth = requireToken();
    if (!auth) return;
    const form = new FormData(event.currentTarget);
    const streetID = text(form, 'street_id') || selectedStreetID;
    const payload = householdPayload(form);
    if (!streetID || !payload) {
      if (!streetID) setNotice({ tone: 'error', text: 'Street required' });
      return;
    }
    run('households:create', () => createHousehold(auth, streetID, payload), {
      success: 'Household created',
      onSuccess: (household) => {
        setHouseholds((items) => [household, ...items]);
        setSelectedHouseholdID(household.id);
      },
    });
  }

  function handleUpdateHousehold(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const auth = requireToken();
    if (!auth) return;
    const form = new FormData(event.currentTarget);
    const householdID = text(form, 'household_id') || selectedHouseholdID;
    const payload = householdPayload(form);
    if (!householdID || !payload) {
      if (!householdID) setNotice({ tone: 'error', text: 'Household required' });
      return;
    }
    run('households:update', () => updateHousehold(auth, householdID, payload), {
      success: 'Household updated',
      onSuccess: (household) => setHouseholds((items) => upsert(items, household)),
    });
  }

  function handleLoadLogs() {
    const auth = requireToken();
    if (!auth || !selectedHouseholdID) {
      setNotice({ tone: 'error', text: 'Household tanlang.' });
      return;
    }
    run('households:logs', () => listHouseholdLogs(auth, selectedHouseholdID), {
      success: 'Logs loaded',
      onSuccess: setHouseholdLogs,
    });
  }

  function handleAssignResponsible(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const auth = requireToken();
    if (!auth) return;
    const form = new FormData(event.currentTarget);
    const streetID = text(form, 'street_id') || selectedStreetID;
    const userID = text(form, 'user_id');
    const from = text(form, 'from_house_number');
    const to = text(form, 'to_house_number');
    if (!streetID || !userID || !from || !to) {
      setNotice({ tone: 'error', text: 'Street, responsible user, and range required' });
      return;
    }
    run('responsibles:create', () => assignResponsible(auth, streetID, {
      user_id: userID,
      from_house_number: from,
      to_house_number: to,
    }), {
      success: 'Responsible person assigned',
      onSuccess: (assignment) => setResponsibleAssignments((items) => [assignment, ...items]),
    });
  }

  function handleDeactivateAssignment(id: string) {
    const auth = requireToken();
    if (!auth) return;
    run('responsibles:deactivate', () => deactivateResponsibleAssignment(auth, id), {
      success: 'Assignment deactivated',
      onSuccess: (updated) => setResponsibleAssignments((items) => upsert(items, updated)),
    });
  }

  return (
    <main className="app-shell">
      <aside className="sidebar">
        <div>
          <p className="eyebrow">Internal ops</p>
          <h1>My Tashabbus Admin</h1>
        </div>
        <nav className="tabs" aria-label="Admin sections">
          {tabs.map((tab) => (
            <button
              className={activeTab === tab.id ? 'tab active' : 'tab'}
              key={tab.id}
              type="button"
              onClick={() => setActiveTab(tab.id)}
            >
              {tab.label}
            </button>
          ))}
        </nav>
      </aside>

      <section className="workspace">
        <header className="topbar">
          <div>
            <p className="section-kicker">Stage UX-1</p>
            <h2>{tabs.find((tab) => tab.id === activeTab)?.label}</h2>
          </div>
          <div className="identity">
            {currentUser ? (
              <>
                <strong>{currentUser.full_name}</strong>
                <span>{currentUser.role}</span>
              </>
            ) : (
              <span>Not authenticated</span>
            )}
            {currentUser ? (
              <button className="ghost" type="button" onClick={logout}>Logout</button>
            ) : (
              <button type="button" onClick={login} disabled={busyAction === 'login'}>Dev Login as Super Admin</button>
            )}
          </div>
        </header>

        <div className={`notice ${notice.tone}`}>{notice.text}</div>

        {activeTab === 'overview' && (
          <section className="content-grid">
            <InfoCard title="Operational Boundary" value="Internal tracking only" detail="No official voting, SMS code collection, or citizen impersonation." />
            <InfoCard title="Users Loaded" value={String(users.length)} detail="Load users before assigning roles to MFYs or streets." />
            <InfoCard title="MFYs Loaded" value={String(mfys.length)} detail={selectedMFYID || 'No selected MFY'} />
            <InfoCard title="Households Loaded" value={String(households.length)} detail={selectedStreetID || 'No selected street'} />
          </section>
        )}

        {activeTab === 'users' && (
          <section className="two-column">
            <Panel title="Create User">
              <form className="form-grid" onSubmit={handleCreateUser}>
                <input name="full_name" placeholder="Full name" required />
                <input name="phone" placeholder="Phone" />
                <select name="role" defaultValue="RESPONSIBLE_PERSON">
                  {roles.map((role) => <option key={role} value={role}>{role}</option>)}
                </select>
                <input name="telegram_id" placeholder="Telegram ID optional" type="number" />
                <input name="telegram_username" placeholder="Telegram username optional" />
                <input name="mfy_id" placeholder="MFY ID optional" />
                <button type="submit" disabled={busyAction === 'users:create'}>Create user</button>
              </form>
            </Panel>
            <Panel title="Bind Telegram">
              <form className="form-grid" onSubmit={handleBindTelegram}>
                <select name="user_id" value={selectedUserID} onChange={(event) => setSelectedUserID(event.target.value)}>
                  <option value="">Select loaded user</option>
                  {users.map((item) => <option key={item.id} value={item.id}>{item.full_name} - {item.role}</option>)}
                </select>
                <input name="telegram_id" placeholder="Telegram ID" type="number" required />
                <input name="telegram_username" placeholder="Telegram username" />
                <button type="submit" disabled={busyAction === 'users:telegram'}>Bind Telegram</button>
              </form>
            </Panel>
            <Panel title="Users" wide>
              <button type="button" onClick={refreshUsers} disabled={busyAction === 'users:list'}>Load users</button>
              <div className="table-list">
                {users.map((item) => (
                  <article className="row-card" key={item.id}>
                    <div>
                      <strong>{item.full_name}</strong>
                      <span>{item.role} · {item.is_active ? 'active' : 'inactive'}</span>
                      <small>TG: {item.telegram_id ?? 'not bound'} · MFY: {item.mfy_id ?? 'none'}</small>
                    </div>
                    <button className="ghost" type="button" onClick={() => copyID(item.id, setNotice)}>Copy ID</button>
                  </article>
                ))}
                {users.length === 0 && <EmptyState text="Users not loaded yet." />}
              </div>
            </Panel>
          </section>
        )}

        {activeTab === 'mfys' && (
          <section className="two-column">
            <Panel title="Create MFY">
              <form className="form-grid" onSubmit={handleCreateMFY}>
                <input name="name" placeholder="MFY name" required />
                <input name="region" placeholder="Region" />
                <input name="district" placeholder="District" />
                <input name="target_votes" placeholder="Target votes" type="number" min="0" />
                <input name="season" placeholder="Season" />
                <button type="submit" disabled={busyAction === 'mfys:create'}>Create MFY</button>
              </form>
            </Panel>
            <Panel title="Assign Chairman">
              <form className="form-grid" onSubmit={handleAssignChairman}>
                <MFYSelect mfys={mfys} value={selectedMFYID} onChange={setSelectedMFYID} name="mfy_id" />
                <UserSelect users={chairmen} name="user_id" label="Select chairman" />
                <button type="submit" disabled={busyAction === 'mfys:chairman'}>Assign chairman</button>
              </form>
            </Panel>
            <Panel title="MFYs" wide>
              <button type="button" onClick={refreshMFYs} disabled={busyAction === 'mfys:list'}>Load MFYs</button>
              <div className="table-list">
                {mfys.map((item) => (
                  <article className={selectedMFYID === item.id ? 'row-card selected' : 'row-card'} key={item.id}>
                    <button className="link-button" type="button" onClick={() => setSelectedMFYID(item.id)}>
                      <strong>{item.name}</strong>
                      <span>{item.region ?? 'No region'} · {item.district ?? 'No district'}</span>
                      <small>Target: {item.target_votes ?? 'none'} · Season: {item.season ?? 'none'}</small>
                    </button>
                    <button className="ghost" type="button" onClick={() => copyID(item.id, setNotice)}>Copy ID</button>
                  </article>
                ))}
                {mfys.length === 0 && <EmptyState text="MFYs not loaded yet." />}
              </div>
            </Panel>
          </section>
        )}

        {activeTab === 'streets' && (
          <section className="two-column">
            <Panel title="Create Street">
              <form className="form-grid" onSubmit={handleCreateStreet}>
                <MFYSelect mfys={mfys} value={selectedMFYID} onChange={setSelectedMFYID} name="mfy_id" />
                <input name="name" placeholder="Street name" required />
                <input name="planned_households_count" placeholder="Planned households" type="number" min="0" defaultValue="0" />
                <input name="notes" placeholder="Notes" />
                <button type="submit" disabled={busyAction === 'streets:create'}>Create street</button>
              </form>
            </Panel>
            <Panel title="Assign Street Leader">
              <form className="form-grid" onSubmit={handleAssignStreetLeader}>
                <StreetSelect streets={streets} value={selectedStreetID} onChange={setSelectedStreetID} name="street_id" />
                <UserSelect users={streetLeaders} name="user_id" label="Select street leader" />
                <button type="submit" disabled={busyAction === 'streets:leader'}>Assign leader</button>
              </form>
            </Panel>
            <Panel title="Streets" wide>
              <div className="inline-actions">
                <MFYSelect mfys={mfys} value={selectedMFYID} onChange={setSelectedMFYID} />
                <button type="button" onClick={() => refreshStreets()} disabled={busyAction === 'streets:list'}>Load streets</button>
              </div>
              <div className="table-list">
                {streets.map((item) => (
                  <article className={selectedStreetID === item.id ? 'row-card selected' : 'row-card'} key={item.id}>
                    <button className="link-button" type="button" onClick={() => setSelectedStreetID(item.id)}>
                      <strong>{item.name}</strong>
                      <span>{item.planned_households_count} planned households · {item.is_active ? 'active' : 'inactive'}</span>
                      <small>{item.notes ?? 'No notes'}</small>
                    </button>
                    <button className="ghost" type="button" onClick={() => copyID(item.id, setNotice)}>Copy ID</button>
                  </article>
                ))}
                {streets.length === 0 && <EmptyState text="Select an MFY and load streets." />}
              </div>
            </Panel>
          </section>
        )}

        {activeTab === 'households' && (
          <section className="two-column">
            <Panel title="Create Household">
              <HouseholdForm
                action="Create household"
                busy={busyAction === 'households:create'}
                streetID={selectedStreetID}
                streets={streets}
                onStreetChange={setSelectedStreetID}
                onSubmit={handleCreateHousehold}
              />
            </Panel>
            <Panel title="Update Household">
              <HouseholdForm
                action="Update household"
                busy={busyAction === 'households:update'}
                household={selectedHousehold}
                householdID={selectedHouseholdID}
                households={households}
                onHouseholdChange={setSelectedHouseholdID}
                onSubmit={handleUpdateHousehold}
              />
            </Panel>
            <Panel title="Households" wide>
              <div className="inline-actions">
                <StreetSelect streets={streets} value={selectedStreetID} onChange={setSelectedStreetID} />
                <button type="button" onClick={() => refreshHouseholds()} disabled={busyAction === 'households:list'}>Load households</button>
              </div>
              <div className="table-list">
                {households.map((item) => (
                  <article className={selectedHouseholdID === item.id ? 'row-card selected' : 'row-card'} key={item.id}>
                    <button className="link-button" type="button" onClick={() => setSelectedHouseholdID(item.id)}>
                      <strong>House {item.house_number}</strong>
                      <span>{item.status} · voted {item.voted_numbers}/{item.total_numbers} · contacted {item.contacted_numbers}/{item.total_numbers}</span>
                      <small>Responsible: {item.assigned_responsible_user_id ?? 'not assigned'} · {item.notes ?? 'No notes'}</small>
                    </button>
                    <button className="ghost" type="button" onClick={() => copyID(item.id, setNotice)}>Copy ID</button>
                  </article>
                ))}
                {households.length === 0 && <EmptyState text="Select a street and load households." />}
              </div>
            </Panel>
            <Panel title="Household Logs" wide>
              <button type="button" onClick={handleLoadLogs} disabled={busyAction === 'households:logs'}>Load selected household logs</button>
              <div className="table-list compact">
                {householdLogs.map((log) => (
                  <article className="row-card" key={log.id}>
                    <div>
                      <strong>{log.field_name}</strong>
                      <span>{log.old_value ?? 'empty'} {'->'} {log.new_value ?? 'empty'}</span>
                      <small>{new Date(log.created_at).toLocaleString()}</small>
                    </div>
                  </article>
                ))}
                {householdLogs.length === 0 && <EmptyState text="No logs loaded." />}
              </div>
            </Panel>
          </section>
        )}

        {activeTab === 'assignments' && (
          <section className="two-column">
            <Panel title="Assign Responsible Person">
              <form className="form-grid" onSubmit={handleAssignResponsible}>
                <StreetSelect streets={streets} value={selectedStreetID} onChange={setSelectedStreetID} name="street_id" />
                <UserSelect users={responsiblePeople} name="user_id" label="Select responsible person" />
                <input name="from_house_number" placeholder="From house number" required />
                <input name="to_house_number" placeholder="To house number" required />
                <button type="submit" disabled={busyAction === 'responsibles:create'}>Assign responsible</button>
              </form>
            </Panel>
            <Panel title="Responsible Assignments" wide>
              <div className="inline-actions">
                <StreetSelect streets={streets} value={selectedStreetID} onChange={setSelectedStreetID} />
                <button type="button" onClick={() => refreshResponsibleAssignments()} disabled={busyAction === 'responsibles:list'}>Load assignments</button>
              </div>
              <div className="table-list">
                {responsibleAssignments.map((item) => (
                  <article className="row-card" key={item.id}>
                    <div>
                      <strong>{userName(users, item.responsible_user_id)}</strong>
                      <span>Range {item.from_house_number}-{item.to_house_number} · {item.is_active ? 'active' : 'inactive'}</span>
                      <small>{item.responsible_user_id}</small>
                    </div>
                    {item.is_active && (
                      <button className="ghost danger" type="button" onClick={() => handleDeactivateAssignment(item.id)}>
                        Deactivate
                      </button>
                    )}
                  </article>
                ))}
                {responsibleAssignments.length === 0 && <EmptyState text="No assignments loaded." />}
              </div>
            </Panel>
          </section>
        )}

        {activeTab === 'devtools' && (
          <section className="content-grid">
            <InfoCard title="API Base URL" value={apiBaseUrl} detail="Public frontend value; no secrets." />
            <InfoCard title="Token" value={token ? 'present' : 'missing'} detail="Stored in localStorage for Stage UX-1 development." />
            <InfoCard title="Selected MFY ID" value={selectedMFYID || 'none'} detail="Used by street and chairman forms." />
            <InfoCard title="Selected Street ID" value={selectedStreetID || 'none'} detail="Used by household and assignment forms." />
            <InfoCard title="Selected Household ID" value={selectedHouseholdID || 'none'} detail="Used by update/log forms." />
          </section>
        )}
      </section>
    </main>
  );
}

function Panel(props: { title: string; children: React.ReactNode; wide?: boolean }) {
  return (
    <article className={props.wide ? 'panel wide' : 'panel'}>
      <h3>{props.title}</h3>
      {props.children}
    </article>
  );
}

function InfoCard(props: { title: string; value: string; detail: string }) {
  return (
    <article className="info-card">
      <span>{props.title}</span>
      <strong>{props.value}</strong>
      <p>{props.detail}</p>
    </article>
  );
}

function EmptyState(props: { text: string }) {
  return <p className="empty">{props.text}</p>;
}

function MFYSelect(props: { mfys: MFY[]; value: string; onChange: (value: string) => void; name?: string }) {
  return (
    <select name={props.name} value={props.value} onChange={(event) => props.onChange(event.target.value)}>
      <option value="">Select MFY</option>
      {props.mfys.map((mfy) => <option key={mfy.id} value={mfy.id}>{mfy.name}</option>)}
    </select>
  );
}

function StreetSelect(props: { streets: Street[]; value: string; onChange: (value: string) => void; name?: string }) {
  return (
    <select name={props.name} value={props.value} onChange={(event) => props.onChange(event.target.value)}>
      <option value="">Select street</option>
      {props.streets.map((street) => <option key={street.id} value={street.id}>{street.name}</option>)}
    </select>
  );
}

function UserSelect(props: { users: User[]; name: string; label: string }) {
  return (
    <select name={props.name} defaultValue="">
      <option value="">{props.label}</option>
      {props.users.map((user) => <option key={user.id} value={user.id}>{user.full_name}</option>)}
    </select>
  );
}

function HouseholdForm(props: {
  action: string;
  busy: boolean;
  streetID?: string;
  streets?: Street[];
  onStreetChange?: (value: string) => void;
  household?: Household | null;
  householdID?: string;
  households?: Household[];
  onHouseholdChange?: (value: string) => void;
  onSubmit: (event: React.FormEvent<HTMLFormElement>) => void;
}) {
  return (
    <form className="form-grid" onSubmit={props.onSubmit}>
      {props.streets && props.onStreetChange && (
        <StreetSelect streets={props.streets} value={props.streetID ?? ''} onChange={props.onStreetChange} name="street_id" />
      )}
      {props.households && props.onHouseholdChange && (
        <select name="household_id" value={props.householdID ?? ''} onChange={(event) => props.onHouseholdChange?.(event.target.value)}>
          <option value="">Select household</option>
          {props.households.map((item) => <option key={item.id} value={item.id}>{item.house_number} - {item.status}</option>)}
        </select>
      )}
      <input name="house_number" placeholder="House number" required defaultValue={props.household?.house_number ?? ''} />
      <input name="total_numbers" placeholder="Total" type="number" min="0" required defaultValue={props.household?.total_numbers ?? 0} />
      <input name="contacted_numbers" placeholder="Contacted" type="number" min="0" required defaultValue={props.household?.contacted_numbers ?? 0} />
      <input name="voted_numbers" placeholder="Reported voted" type="number" min="0" required defaultValue={props.household?.voted_numbers ?? 0} />
      <select name="status" defaultValue={props.household?.status ?? 'NEW'}>
        {statuses.map((status) => <option key={status} value={status}>{status}</option>)}
      </select>
      <input name="notes" placeholder="Notes" defaultValue={props.household?.notes ?? ''} />
      <button type="submit" disabled={props.busy}>{props.action}</button>
    </form>
  );
}

function text(form: FormData, name: string): string {
  return String(form.get(name) ?? '').trim();
}

function nullableText(form: FormData, name: string): string | null {
  const value = text(form, name);
  return value || null;
}

function numberValue(form: FormData, name: string): number {
  const raw = text(form, name);
  if (!raw) return 0;
  const value = Number(raw);
  return Number.isFinite(value) ? value : 0;
}

function nullableNumber(form: FormData, name: string): number | null {
  const raw = text(form, name);
  if (!raw) return null;
  const value = Number(raw);
  return Number.isFinite(value) ? value : null;
}

function upsert<T extends { id: string }>(items: T[], item: T): T[] {
  const exists = items.some((existing) => existing.id === item.id);
  if (!exists) return [item, ...items];
  return items.map((existing) => (existing.id === item.id ? item : existing));
}

function userName(users: User[], id: string): string {
  return users.find((user) => user.id === id)?.full_name ?? id;
}

function copyID(id: string, setNotice: (notice: Notice) => void) {
  navigator.clipboard?.writeText(id).catch(() => undefined);
  setNotice({ tone: 'success', text: 'ID copied' });
}

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
