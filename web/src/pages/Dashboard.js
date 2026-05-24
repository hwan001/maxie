import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import axios from "axios";
import Navbar from "../components/Navbar";
import "../styles/Dashboard.css";
import { Title, Logo, MenuItems, BASE_URL } from "../constants";

const OS_ICONS = {
	darwin: "fa-brands fa-apple",
	windows: "fa-brands fa-windows",
	linux: "fa-brands fa-linux",
};

const DRIVE_ICONS = {
	google: "fa-brands fa-google",
	naver: "fa-solid fa-cloud",
	onedrive: "fa-brands fa-microsoft",
	dropbox: "fa-brands fa-dropbox",
	local: "fa-solid fa-hard-drive",
	other: "fa-solid fa-folder",
};

function fmtBytes(bytes) {
	if (!bytes) return "0 B";
	const units = ["B", "KB", "MB", "GB", "TB"];
	let i = 0;
	let val = bytes;
	while (val >= 1024 && i < units.length - 1) { val /= 1024; i++; }
	return `${val.toFixed(1)} ${units[i]}`;
}

function fmtTime(iso) {
	if (!iso) return "—";
	const d = new Date(iso);
	if (isNaN(d) || d.getFullYear() < 2000) return "—";
	return d.toLocaleString();
}

function isOnline(lastSeen) {
	if (!lastSeen) return false;
	return Date.now() - new Date(lastSeen).getTime() < 5 * 60 * 1000;
}

function FilterSidebar({ agents, osFilter, driveFilter, onOsFilter, onDriveFilter }) {
	const osList = [...new Set(agents.map(a => a.client_data?.system_info?.os).filter(Boolean))];
	const driveTypes = [...new Set(agents.flatMap(a => (a.drives ?? []).map(d => d.drive_type)).filter(Boolean))];

	return (
		<aside className="dash-sidebar">
			<div className="dash-sidebar-header">Filters</div>
			{osList.length > 0 && (
				<div className="filter-section">
					<div className="filter-label">OS</div>
					{osList.map(os => (
						<label key={os} className="filter-item">
							<input type="checkbox" checked={osFilter.includes(os)} onChange={() => onOsFilter(os)} />
							<i className={OS_ICONS[os] ?? "fa-solid fa-desktop"} />
							{os}
						</label>
					))}
				</div>
			)}
			{driveTypes.length > 0 && (
				<div className="filter-section">
					<div className="filter-label">Drive Type</div>
					{driveTypes.map(dt => (
						<label key={dt} className="filter-item">
							<input type="checkbox" checked={driveFilter.includes(dt)} onChange={() => onDriveFilter(dt)} />
							<i className={DRIVE_ICONS[dt] ?? "fa-solid fa-folder"} />
							{dt}
						</label>
					))}
				</div>
			)}
		</aside>
	);
}

function DeviceCard({ agent, onClick }) {
	const os = agent.client_data?.system_info?.os ?? "unknown";
	const online = isOnline(agent.last_seen);
	const drives = agent.drives ?? [];
	const stats = agent.file_stats ?? {};

	return (
		<div className="device-card" onClick={() => onClick(agent)}>
			<div className="device-card-header">
				<div className={`device-status ${online ? "online" : "offline"}`} title={online ? "Online" : "Offline"} />
				<i className={`device-os-icon ${OS_ICONS[os] ?? "fa-solid fa-desktop"}`} />
				<div className="device-name">{agent.name || agent.agent_id}</div>
				<span className="device-os-tag">{os}</span>
			</div>
			<div className="device-card-body">
				<div className="device-stat">
					<span className="device-stat-label">Files</span>
					<span className="device-stat-value">{stats.total_files?.toLocaleString() ?? "—"}</span>
				</div>
				<div className="device-stat">
					<span className="device-stat-label">Size</span>
					<span className="device-stat-value">{fmtBytes(stats.total_size)}</span>
				</div>
				<div className="device-stat">
					<span className="device-stat-label">Duplicates</span>
					<span className={`device-stat-value${stats.duplicate_count > 0 ? " warn" : ""}`}>
						{stats.duplicate_count ?? "—"}
					</span>
				</div>
				<div className="device-stat">
					<span className="device-stat-label">Last seen</span>
					<span className="device-stat-value small">{fmtTime(agent.last_seen)}</span>
				</div>
			</div>
			{drives.length > 0 && (
				<div className="device-card-drives">
					{drives.map((d, i) => (
						<span key={i} className="drive-pill">
							<i className={DRIVE_ICONS[d.drive_type] ?? "fa-solid fa-folder"} />
							{d.label || d.drive_type}
						</span>
					))}
				</div>
			)}
		</div>
	);
}

function StatCard({ icon, label, value, sub }) {
	return (
		<div className="stat-card">
			<div className="stat-icon"><i className={icon} /></div>
			<div className="stat-label">{label}</div>
			<div className="stat-value">{value ?? "—"}</div>
			{sub && <div className="stat-sub">{sub}</div>}
		</div>
	);
}

const Dashboard = () => {
	const navigate = useNavigate();
	const [agents, setAgents] = useState([]);
	const [loading, setLoading] = useState(true);
	const [osFilter, setOsFilter] = useState([]);
	const [driveFilter, setDriveFilter] = useState([]);

	const fetchData = () => {
		setLoading(true);
		axios.get(`${BASE_URL}/protected/agents`, { withCredentials: true })
			.then(r => setAgents(r.data?.agents ?? []))
			.catch(() => {})
			.finally(() => setLoading(false));
	};

	useEffect(() => { fetchData(); }, []);

	const toggleOs = (os) => setOsFilter(prev => prev.includes(os) ? prev.filter(x => x !== os) : [...prev, os]);
	const toggleDrive = (dt) => setDriveFilter(prev => prev.includes(dt) ? prev.filter(x => x !== dt) : [...prev, dt]);

	const filtered = agents.filter(a => {
		const os = a.client_data?.system_info?.os;
		if (osFilter.length > 0 && !osFilter.includes(os)) return false;
		if (driveFilter.length > 0) {
			const drives = (a.drives ?? []).map(d => d.drive_type);
			if (!driveFilter.some(dt => drives.includes(dt))) return false;
		}
		return true;
	});

	const totalFiles = agents.reduce((s, a) => s + (a.file_stats?.total_files ?? 0), 0);
	const totalSize = agents.reduce((s, a) => s + (a.file_stats?.total_size ?? 0), 0);
	const totalDupes = agents.reduce((s, a) => s + (a.file_stats?.duplicate_count ?? 0), 0);

	return (
		<>
			<Navbar title={Title} logo={Logo} menuItems={MenuItems["app"]} />
			<div className="dashboard-layout">
				<FilterSidebar
					agents={agents}
					osFilter={osFilter}
					driveFilter={driveFilter}
					onOsFilter={toggleOs}
					onDriveFilter={toggleDrive}
				/>
				<main className="dash-main">
					<div className="dash-header">
						<div>
							<h1>Dashboard</h1>
							<p>All registered agents and their file statistics</p>
						</div>
						<button className="dash-refresh-btn" onClick={fetchData}>
							<i className="fa-solid fa-rotate-right" />
							Refresh
						</button>
					</div>

					<div className="stat-grid">
						<StatCard icon="fa-solid fa-robot" label="Agents" value={agents.length} sub="registered" />
						<StatCard icon="fa-solid fa-file" label="Total Files" value={totalFiles.toLocaleString()} sub="across all drives" />
						<StatCard icon="fa-solid fa-hard-drive" label="Total Size" value={fmtBytes(totalSize)} sub="indexed" />
						<StatCard icon="fa-solid fa-clone" label="Duplicates" value={totalDupes} sub="files found" />
					</div>

					{loading ? (
						<div className="loading-row"><span className="spinner" />Loading devices…</div>
					) : filtered.length === 0 ? (
						<div className="empty-state">
							<i className="fa-solid fa-satellite-dish" />
							{agents.length === 0 ? "No agents registered yet" : "No agents match the current filters"}
						</div>
					) : (
						<div className="device-grid">
							{filtered.map(a => (
								<DeviceCard
									key={a.agent_id}
									agent={a}
									onClick={(agent) => navigate(`/devices?id=${agent.agent_id}`)}
								/>
							))}
						</div>
					)}
				</main>
			</div>
		</>
	);
};

export default Dashboard;
