import React, { useState, useEffect, useCallback } from "react";
import axios from "axios";
import Navbar from "../components/Navbar";
import "../styles/Dashboard.css";
import "../styles/Admin.css";
import { Title, Logo, MenuItems, BASE_URL } from "../constants";

const MIN_PASSWORD_LEN = 8;

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

// Authenticated console view.
function AdminConsole({ onLogout }) {
	return (
		<main className="dash-main">
			<div className="dash-header">
				<div>
					<h1>Admin</h1>
					<p>Console settings, users, and roles</p>
				</div>
				<button className="dash-refresh-btn" onClick={onLogout}>
					<i className="fa-solid fa-right-from-bracket" />
					Sign out
				</button>
			</div>

			<section className="admin-card admin-card--muted">
				<div className="admin-card-header">
					<i className="fa-solid fa-diagram-project" />
					<div>
						<h2>Users &amp; topology</h2>
						<p>OIDC, roles, and the user–device map arrive in later steps.</p>
					</div>
					<span className="admin-badge">Planned</span>
				</div>
				<ul className="admin-roadmap">
					<li><i className="fa-solid fa-key" /> OIDC provider connection &amp; role management</li>
					<li><i className="fa-solid fa-sitemap" /> User–device list &amp; topology view</li>
				</ul>
			</section>
		</main>
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
