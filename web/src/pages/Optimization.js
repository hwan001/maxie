import React, { useState, useEffect } from "react";
import axios from "axios";
import Navbar from "../components/Navbar";
import "../styles/Dashboard.css";
import { Title, Logo, MenuItems } from "../constants";

function fmtBytes(bytes) {
	if (!bytes) return "0 B";
	const units = ["B", "KB", "MB", "GB", "TB"];
	let i = 0, val = bytes;
	while (val >= 1024 && i < units.length - 1) { val /= 1024; i++; }
	return `${val.toFixed(1)} ${units[i]}`;
}

function OpportunityCard({ agent, onPlan }) {
	const stats = agent.file_stats ?? {};
	const dupes = stats.duplicate_count ?? 0;
	const totalSize = stats.total_size ?? 0;
	const estimatedSavings = dupes > 0 ? Math.floor(totalSize * (dupes / Math.max(stats.total_files ?? 1, 1))) : 0;

	return (
		<div className="info-card">
			<div className="info-card-header">
				<i className="fa-solid fa-robot" />
				<h3>{agent.name || agent.agent_id}</h3>
				{dupes > 0 ? (
					<span className="dupe-badge" style={{ marginLeft: "auto" }}>
						<i className="fa-solid fa-triangle-exclamation" />
						{dupes} duplicates
					</span>
				) : (
					<span className="clean-badge" style={{ marginLeft: "auto" }}>
						<i className="fa-solid fa-circle-check" />
						Clean
					</span>
				)}
			</div>
			<div className="info-card-body">
				<div className="opt-row">
					<div className="opt-stat">
						<div className="opt-stat-label">Total Files</div>
						<div className="opt-stat-value">{stats.total_files?.toLocaleString() ?? "—"}</div>
					</div>
					<div className="opt-stat">
						<div className="opt-stat-label">Total Size</div>
						<div className="opt-stat-value">{fmtBytes(totalSize)}</div>
					</div>
					<div className="opt-stat">
						<div className="opt-stat-label">Duplicates</div>
						<div className={`opt-stat-value${dupes > 0 ? " warn" : ""}`}>{dupes}</div>
					</div>
					<div className="opt-stat">
						<div className="opt-stat-label">Est. Savings</div>
						<div className={`opt-stat-value${estimatedSavings > 0 ? " accent" : ""}`}>
							{fmtBytes(estimatedSavings)}
						</div>
					</div>
				</div>
				{dupes > 0 && (
					<div style={{ marginTop: "1rem" }}>
						<button className="plan-btn" onClick={() => onPlan(agent)}>
							<i className="fa-solid fa-list-check" />
							View Optimization Plan
						</button>
					</div>
				)}
			</div>
		</div>
	);
}

function PlanModal({ agent, onClose }) {
	const stats = agent.file_stats ?? {};
	const dupes = stats.duplicate_count ?? 0;

	return (
		<div className="modal-overlay" onClick={onClose}>
			<div className="modal-box" onClick={e => e.stopPropagation()}>
				<div className="modal-header">
					<h2>Optimization Plan</h2>
					<div className="modal-sub">{agent.name || agent.agent_id}</div>
					<button className="modal-close" onClick={onClose}>
						<i className="fa-solid fa-xmark" />
					</button>
				</div>
				<div className="modal-body">
					<div className="plan-step">
						<div className="plan-step-num">1</div>
						<div className="plan-step-content">
							<div className="plan-step-title">Scan for duplicates</div>
							<div className="plan-step-desc">
								Agent will hash all files and identify {dupes} duplicate file{dupes !== 1 ? "s" : ""}.
							</div>
						</div>
						<span className="plan-step-status done">Done</span>
					</div>
					<div className="plan-step">
						<div className="plan-step-num">2</div>
						<div className="plan-step-content">
							<div className="plan-step-title">Review duplicates</div>
							<div className="plan-step-desc">
								Review each duplicate group and select which copy to keep.
							</div>
						</div>
						<span className="plan-step-status pending">Pending</span>
					</div>
					<div className="plan-step">
						<div className="plan-step-num">3</div>
						<div className="plan-step-content">
							<div className="plan-step-title">Apply cleanup</div>
							<div className="plan-step-desc">
								Remove selected duplicates. Estimated savings: {fmtBytes(
									Math.floor((stats.total_size ?? 0) * (dupes / Math.max(stats.total_files ?? 1, 1)))
								)}.
							</div>
						</div>
						<span className="plan-step-status pending">Pending</span>
					</div>
				</div>
				<div className="modal-footer">
					<button className="plan-btn" disabled title="Requires agent support">
						<i className="fa-solid fa-play" />
						Apply (coming soon)
					</button>
					<button className="dash-refresh-btn" onClick={onClose}>Cancel</button>
				</div>
			</div>
		</div>
	);
}

const Optimization = () => {
	const [agents, setAgents] = useState([]);
	const [loading, setLoading] = useState(true);
	const [planAgent, setPlanAgent] = useState(null);

	const fetchData = () => {
		setLoading(true);
		axios.get("/api/agents")
			.then(r => setAgents(r.data?.agents ?? []))
			.catch(() => {})
			.finally(() => setLoading(false));
	};

	useEffect(() => { fetchData(); }, []);

	const withDupes = agents.filter(a => (a.file_stats?.duplicate_count ?? 0) > 0);
	const clean = agents.filter(a => (a.file_stats?.duplicate_count ?? 0) === 0);
	const totalSavings = agents.reduce((s, a) => {
		const stats = a.file_stats ?? {};
		const dupes = stats.duplicate_count ?? 0;
		return s + Math.floor((stats.total_size ?? 0) * (dupes / Math.max(stats.total_files ?? 1, 1)));
	}, 0);

	return (
		<>
			<Navbar title={Title} logo={Logo} menuItems={MenuItems["app"]} />
			<div className="dashboard-layout">
				<main className="dash-main" style={{ maxWidth: "none" }}>
					<div className="dash-header">
						<div>
							<h1>Optimization</h1>
							<p>Identify and remove duplicate files across all agents</p>
						</div>
						<button className="dash-refresh-btn" onClick={fetchData}>
							<i className="fa-solid fa-rotate-right" />
							Refresh
						</button>
					</div>

					<div className="stat-grid">
						<div className="stat-card">
							<div className="stat-icon"><i className="fa-solid fa-triangle-exclamation" /></div>
							<div className="stat-label">Agents with Dupes</div>
							<div className="stat-value" style={{ color: withDupes.length > 0 ? "var(--color-warn)" : undefined }}>
								{withDupes.length}
							</div>
							<div className="stat-sub">of {agents.length} total</div>
						</div>
						<div className="stat-card">
							<div className="stat-icon"><i className="fa-solid fa-circle-check" /></div>
							<div className="stat-label">Clean Agents</div>
							<div className="stat-value">{clean.length}</div>
							<div className="stat-sub">no duplicates</div>
						</div>
						<div className="stat-card">
							<div className="stat-icon"><i className="fa-solid fa-floppy-disk" /></div>
							<div className="stat-label">Est. Savings</div>
							<div className="stat-value" style={{ color: totalSavings > 0 ? "var(--color-accent)" : undefined }}>
								{fmtBytes(totalSavings)}
							</div>
							<div className="stat-sub">recoverable</div>
						</div>
					</div>

					{loading ? (
						<div className="loading-row"><span className="spinner" />Loading agents…</div>
					) : agents.length === 0 ? (
						<div className="empty-state">
							<i className="fa-solid fa-satellite-dish" />
							No agents registered yet
						</div>
					) : (
						<>
							{withDupes.length > 0 && (
								<div style={{ display: "flex", flexDirection: "column", gap: "1rem" }}>
									{withDupes.map(a => (
										<OpportunityCard key={a.agent_id} agent={a} onPlan={setPlanAgent} />
									))}
								</div>
							)}
							{clean.length > 0 && withDupes.length > 0 && (
								<div style={{ fontSize: "0.8rem", color: "var(--color-text-muted)", marginTop: "0.5rem" }}>
									<i className="fa-solid fa-circle-check" style={{ marginRight: "0.4rem", color: "#22c55e" }} />
									{clean.length} agent{clean.length !== 1 ? "s are" : " is"} already clean.
								</div>
							)}
							{withDupes.length === 0 && (
								<div className="empty-state">
									<i className="fa-solid fa-circle-check" style={{ color: "#22c55e" }} />
									All agents are clean — no duplicates found.
								</div>
							)}
						</>
					)}
				</main>
			</div>

			{planAgent && <PlanModal agent={planAgent} onClose={() => setPlanAgent(null)} />}
		</>
	);
};

export default Optimization;
