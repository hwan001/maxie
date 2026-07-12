import React, { useState, useEffect, useCallback } from "react";
import axios from "axios";
import Navbar from "../components/Navbar";
import "../styles/Dashboard.css";
import "../styles/Admin.css";
import { Title, Logo, MenuItems, BASE_URL } from "../constants";

const MIN_PASSWORD_LEN = 8;

const OS_ICONS = {
	darwin: "fa-brands fa-apple",
	windows: "fa-brands fa-windows",
	linux: "fa-brands fa-linux",
};

function errorMessage(err, fallback) {
	return err?.response?.data?.error || fallback;
}

// First-time setup: no admin password exists yet, so anyone can set one.
function AdminSetupCard({ onDone }) {
	const [password, setPassword] = useState("");
	const [confirm, setConfirm] = useState("");
	const [error, setError] = useState("");
	const [saving, setSaving] = useState(false);

	const canSubmit = password.length >= MIN_PASSWORD_LEN && password === confirm && !saving;

	const handleSubmit = async (e) => {
		e.preventDefault();
		if (password.length < MIN_PASSWORD_LEN) {
			setError(`Password must be at least ${MIN_PASSWORD_LEN} characters.`);
			return;
		}
		if (password !== confirm) {
			setError("Passwords do not match.");
			return;
		}
		setError("");
		setSaving(true);
		try {
			await axios.post(`${BASE_URL}/admin/password`, { password }, { withCredentials: true });
			onDone();
		} catch (err) {
			setError(errorMessage(err, "Could not save the password. Please try again."));
		} finally {
			setSaving(false);
		}
	};

	return (
		<section className="admin-card" aria-labelledby="admin-setup-heading">
			<div className="admin-card-header">
				<i className="fa-solid fa-shield-halved" />
				<div>
					<h2 id="admin-setup-heading">First-time setup</h2>
					<p>Set an admin password to protect this console.</p>
				</div>
				<span className="admin-badge admin-badge--warn">Not configured</span>
			</div>

			<form className="admin-form" onSubmit={handleSubmit}>
				<label className="admin-field">
					<span className="admin-field-label">New password</span>
					<input
						type="password"
						className="admin-input"
						value={password}
						onChange={(e) => setPassword(e.target.value)}
						placeholder={`At least ${MIN_PASSWORD_LEN} characters`}
						autoComplete="new-password"
					/>
				</label>
				<label className="admin-field">
					<span className="admin-field-label">Confirm password</span>
					<input
						type="password"
						className="admin-input"
						value={confirm}
						onChange={(e) => setConfirm(e.target.value)}
						placeholder="Re-enter password"
						autoComplete="new-password"
					/>
				</label>

				{error && (
					<div className="admin-form-note" role="alert">
						<i className="fa-solid fa-circle-info" />
						{error}
					</div>
				)}

				<button type="submit" className="admin-submit" disabled={!canSubmit}>
					{saving
						? <><i className="fa-solid fa-spinner fa-spin" /> Saving…</>
						: <><i className="fa-solid fa-floppy-disk" /> Save password</>}
				</button>
			</form>
		</section>
	);
}

// Console already has a password — prompt for it.
function AdminLoginCard({ onDone }) {
	const [password, setPassword] = useState("");
	const [error, setError] = useState("");
	const [busy, setBusy] = useState(false);

	const handleSubmit = async (e) => {
		e.preventDefault();
		setError("");
		setBusy(true);
		try {
			await axios.post(`${BASE_URL}/admin/login`, { password }, { withCredentials: true });
			onDone();
		} catch (err) {
			setError(errorMessage(err, "Sign in failed. Please try again."));
		} finally {
			setBusy(false);
		}
	};

	return (
		<section className="admin-card" aria-labelledby="admin-login-heading">
			<div className="admin-card-header">
				<i className="fa-solid fa-lock" />
				<div>
					<h2 id="admin-login-heading">Admin sign in</h2>
					<p>Enter the admin password to continue.</p>
				</div>
			</div>

			<form className="admin-form" onSubmit={handleSubmit}>
				<label className="admin-field">
					<span className="admin-field-label">Password</span>
					<input
						type="password"
						className="admin-input"
						value={password}
						onChange={(e) => setPassword(e.target.value)}
						placeholder="Admin password"
						autoComplete="current-password"
						autoFocus
					/>
				</label>

				{error && (
					<div className="admin-form-note" role="alert">
						<i className="fa-solid fa-circle-info" />
						{error}
					</div>
				)}

				<button type="submit" className="admin-submit" disabled={!password || busy}>
					{busy
						? <><i className="fa-solid fa-spinner fa-spin" /> Signing in…</>
						: <><i className="fa-solid fa-right-to-bracket" /> Sign in</>}
				</button>
			</form>
		</section>
	);
}

const ADMIN_TABS = [
	{ key: "overview", label: "Overview", icon: "fa-solid fa-gauge-high" },
	{ key: "security", label: "Security", icon: "fa-solid fa-lock" },
	{ key: "users", label: "Users & Roles", icon: "fa-solid fa-users-gear" },
	{ key: "auth", label: "Authentication", icon: "fa-solid fa-key" },
	{ key: "topology", label: "Topology", icon: "fa-solid fa-sitemap" },
];

// Change the admin password (requires the current password). Fully wired.
function ChangePasswordPanel() {
	const [current, setCurrent] = useState("");
	const [next, setNext] = useState("");
	const [confirm, setConfirm] = useState("");
	const [note, setNote] = useState(null); // { type: "ok" | "error", text }
	const [busy, setBusy] = useState(false);

	const canSubmit = current && next.length >= MIN_PASSWORD_LEN && next === confirm && !busy;

	const handleSubmit = async (e) => {
		e.preventDefault();
		if (next.length < MIN_PASSWORD_LEN) {
			setNote({ type: "error", text: `New password must be at least ${MIN_PASSWORD_LEN} characters.` });
			return;
		}
		if (next !== confirm) {
			setNote({ type: "error", text: "New passwords do not match." });
			return;
		}
		setBusy(true);
		setNote(null);
		try {
			await axios.post(
				`${BASE_URL}/admin/password`,
				{ password: next, current_password: current },
				{ withCredentials: true },
			);
			setNote({ type: "ok", text: "Password updated." });
			setCurrent(""); setNext(""); setConfirm("");
		} catch (err) {
			setNote({ type: "error", text: errorMessage(err, "Could not update the password.") });
		} finally {
			setBusy(false);
		}
	};

	return (
		<section className="admin-card" aria-labelledby="admin-security-heading">
			<div className="admin-card-header">
				<i className="fa-solid fa-lock" />
				<div>
					<h2 id="admin-security-heading">Change password</h2>
					<p>Update the password used to sign in to this console.</p>
				</div>
			</div>

			<form className="admin-form" onSubmit={handleSubmit}>
				<label className="admin-field">
					<span className="admin-field-label">Current password</span>
					<input type="password" className="admin-input" value={current}
						onChange={(e) => setCurrent(e.target.value)} autoComplete="current-password" />
				</label>
				<label className="admin-field">
					<span className="admin-field-label">New password</span>
					<input type="password" className="admin-input" value={next}
						onChange={(e) => setNext(e.target.value)}
						placeholder={`At least ${MIN_PASSWORD_LEN} characters`} autoComplete="new-password" />
				</label>
				<label className="admin-field">
					<span className="admin-field-label">Confirm new password</span>
					<input type="password" className="admin-input" value={confirm}
						onChange={(e) => setConfirm(e.target.value)} autoComplete="new-password" />
				</label>

				{note && (
					<div className={`admin-form-note${note.type === "ok" ? " admin-form-note--ok" : ""}`} role="alert">
						<i className={`fa-solid ${note.type === "ok" ? "fa-circle-check" : "fa-circle-info"}`} />
						{note.text}
					</div>
				)}

				<button type="submit" className="admin-submit" disabled={!canSubmit}>
					{busy
						? <><i className="fa-solid fa-spinner fa-spin" /> Updating…</>
						: <><i className="fa-solid fa-floppy-disk" /> Update password</>}
				</button>
			</form>
		</section>
	);
}

function OverviewPanel() {
	return (
		<section className="admin-card">
			<div className="admin-card-header">
				<i className="fa-solid fa-gauge-high" />
				<div>
					<h2>Overview</h2>
					<p>Console status and quick facts.</p>
				</div>
				<span className="admin-badge admin-badge--ok">Protected</span>
			</div>
			<ul className="admin-roadmap">
				<li><i className="fa-solid fa-lock" /> Admin password is set and this session is authenticated.</li>
				<li><i className="fa-solid fa-users-gear" /> Manage users and roles from the Users &amp; Roles tab.</li>
				<li><i className="fa-solid fa-key" /> Connect an identity provider from the Authentication tab.</li>
				<li><i className="fa-solid fa-sitemap" /> Inspect the user–device map from the Topology tab.</li>
			</ul>
		</section>
	);
}

// Placeholder for panels whose backend is not built yet. No fake actions.
function PlaceholderPanel({ icon, title, description, items }) {
	return (
		<section className="admin-card admin-card--muted">
			<div className="admin-card-header">
				<i className={icon} />
				<div>
					<h2>{title}</h2>
					<p>{description}</p>
				</div>
				<span className="admin-badge">Coming soon</span>
			</div>
			<ul className="admin-roadmap">
				{items.map((it) => (
					<li key={it}><i className="fa-solid fa-circle-dot" /> {it}</li>
				))}
			</ul>
		</section>
	);
}

// One user with their owned devices, rendered as a small tree row.
function TopoUser({ user }) {
	return (
		<div className="topo-user">
			<div className="topo-user-head">
				<i className="fa-solid fa-user topo-user-icon" />
				<div className="topo-user-info">
					<span className="topo-user-name">
						{user.name || user.user_id}
						{user.is_guest && <span className="topo-guest">guest</span>}
					</span>
					{user.email && <span className="topo-user-sub">{user.email}</span>}
				</div>
				<span className="topo-device-count">
					{user.device_count} {user.device_count === 1 ? "device" : "devices"}
				</span>
			</div>
			{user.devices.length > 0 && (
				<div className="topo-devices">
					{user.devices.map((d) => (
						<div key={d.agent_id} className="topo-device">
							<i className={`topo-device-os ${OS_ICONS[d.os] ?? "fa-solid fa-desktop"}`} />
							<span className="topo-device-name">{d.name || d.agent_id}</span>
							<span className={`device-status-badge ${d.online ? "online" : "offline"}`}>
								<span className={`device-status ${d.online ? "online" : "offline"}`} />
								{d.online ? "Online" : "Offline"}
							</span>
							<span className="topo-device-files">{(d.files ?? 0).toLocaleString()} files</span>
						</div>
					))}
				</div>
			)}
		</div>
	);
}

// Topology: live user–device map built from server data.
function TopologyPanel() {
	const [data, setData] = useState(null);
	const [error, setError] = useState("");

	useEffect(() => {
		axios.get(`${BASE_URL}/admin/topology`, { withCredentials: true })
			.then((r) => { setData(r.data); setError(""); })
			.catch((err) => setError(errorMessage(err, "Could not load topology.")));
	}, []);

	if (error) {
		return (
			<section className="admin-card admin-status-msg" role="alert">
				<i className="fa-solid fa-triangle-exclamation" />
				<span>{error}</span>
			</section>
		);
	}
	if (!data) {
		return <div className="loading-row"><span className="spinner" /> Loading topology…</div>;
	}

	const users = data.users ?? [];
	const unassigned = data.unassigned ?? [];
	const totalDevices = users.reduce((s, u) => s + u.device_count, 0) + unassigned.length;

	return (
		<section className="admin-card">
			<div className="admin-card-header">
				<i className="fa-solid fa-sitemap" />
				<div>
					<h2>Topology</h2>
					<p>{users.length} {users.length === 1 ? "user" : "users"} · {totalDevices} {totalDevices === 1 ? "device" : "devices"}</p>
				</div>
			</div>

			{users.length === 0 && unassigned.length === 0 ? (
				<div className="admin-empty">No users or devices yet.</div>
			) : (
				<div className="topo-tree">
					{users.map((u) => <TopoUser key={u.user_id} user={u} />)}
					{unassigned.length > 0 && (
						<TopoUser
							user={{
								user_id: "__unassigned",
								name: "Unassigned",
								email: "legacy agents with no linked user",
								is_guest: false,
								device_count: unassigned.length,
								devices: unassigned,
							}}
						/>
					)}
				</div>
			)}
		</section>
	);
}

function AdminPanel({ tab }) {
	switch (tab) {
		case "security":
			return <ChangePasswordPanel />;
		case "users":
			return (
				<PlaceholderPanel
					icon="fa-solid fa-users-gear"
					title="Users & Roles"
					description="View accounts and assign admin or user roles."
					items={["List users linked to the server", "Grant or revoke the admin role", "Search and filter accounts"]}
				/>
			);
		case "auth":
			return (
				<PlaceholderPanel
					icon="fa-solid fa-key"
					title="Authentication"
					description="Connect and manage an identity provider."
					items={["Configure an OIDC provider", "Test the connection", "Enable provider-based sign in"]}
				/>
			);
		case "topology":
			return <TopologyPanel />;
		default:
			return <OverviewPanel />;
	}
}

// Authenticated console: sub-tab sidebar + active panel.
function AdminConsole({ onLogout }) {
	const [tab, setTab] = useState("overview");
	const active = ADMIN_TABS.find((t) => t.key === tab) ?? ADMIN_TABS[0];

	return (
		<>
			<aside className="dash-sidebar">
				<div className="dash-sidebar-header">Admin</div>
				<nav className="admin-nav">
					{ADMIN_TABS.map((t) => (
						<button
							key={t.key}
							className={`admin-nav-item${t.key === tab ? " active" : ""}`}
							onClick={() => setTab(t.key)}
						>
							<i className={t.icon} />
							{t.label}
						</button>
					))}
				</nav>
			</aside>
			<main className="dash-main">
				<div className="dash-header">
					<div>
						<h1>{active.label}</h1>
						<p>Admin console</p>
					</div>
					<button className="dash-refresh-btn" onClick={onLogout}>
						<i className="fa-solid fa-right-from-bracket" />
						Sign out
					</button>
				</div>
				<AdminPanel tab={tab} />
			</main>
		</>
	);
}

const Admin = () => {
	const [status, setStatus] = useState(null); // { configured, authenticated }
	const [loadError, setLoadError] = useState("");

	const loadStatus = useCallback(async () => {
		try {
			const res = await axios.get(`${BASE_URL}/admin/status`, { withCredentials: true });
			setStatus(res.data);
			setLoadError("");
		} catch (err) {
			setLoadError(errorMessage(err, "Could not reach the server."));
		}
	}, []);

	useEffect(() => { loadStatus(); }, [loadStatus]);

	const handleLogout = async () => {
		try {
			await axios.post(`${BASE_URL}/admin/logout`, {}, { withCredentials: true });
		} catch { /* clearing client state below regardless */ }
		loadStatus();
	};

	const renderBody = () => {
		if (loadError) {
			return (
				<div className="admin-center">
					<div className="admin-card admin-status-msg" role="alert">
						<i className="fa-solid fa-triangle-exclamation" />
						<span>{loadError}</span>
					</div>
				</div>
			);
		}
		if (!status) {
			return (
				<div className="admin-center">
					<div className="loading-row"><span className="spinner" /></div>
				</div>
			);
		}
		if (status.authenticated) {
			return <AdminConsole onLogout={handleLogout} />;
		}
		return (
			<div className="admin-center">
				{status.configured
					? <AdminLoginCard onDone={loadStatus} />
					: <AdminSetupCard onDone={loadStatus} />}
			</div>
		);
	};

	return (
		<>
			<Navbar title={Title} logo={Logo} menuItems={MenuItems["app"]} />
			<div className="dashboard-layout">
				{renderBody()}
			</div>
		</>
	);
};

export default Admin;
