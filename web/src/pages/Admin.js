import React, { useState } from "react";
import Navbar from "../components/Navbar";
import "../styles/Dashboard.css";
import "../styles/Admin.css";
import { Title, Logo, MenuItems } from "../constants";

const MIN_PASSWORD_LEN = 8;

// First-time admin setup screen (MAX-10).
//
// This is the initial scaffold: it renders the setup form and validates input
// locally. Persisting the password to the server (POST → app_config
// admin_password_hash) and the access gate are wired in the follow-up step.
function AdminSetupCard() {
	const [password, setPassword] = useState("");
	const [confirm, setConfirm] = useState("");
	const [error, setError] = useState("");

	const canSubmit = password.length >= MIN_PASSWORD_LEN && password === confirm;

	const handleSubmit = (e) => {
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
		// Backend wiring (bcrypt hash → app_config) lands in the next step.
		setError("Saving isn't available yet.");
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
					<i className="fa-solid fa-floppy-disk" /> Save password
				</button>
			</form>
		</section>
	);
}

const Admin = () => {
	return (
		<>
			<Navbar title={Title} logo={Logo} menuItems={MenuItems["app"]} />
			<div className="dashboard-layout">
				<main className="dash-main">
					<div className="dash-header">
						<div>
							<h1>Admin</h1>
							<p>Console settings, users, and roles</p>
						</div>
					</div>

					<AdminSetupCard />

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
			</div>
		</>
	);
};

export default Admin;
