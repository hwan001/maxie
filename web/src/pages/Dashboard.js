import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import axios from "axios";
import { MapContainer, TileLayer, Marker, Popup } from "react-leaflet";
import L from "leaflet";
import "leaflet/dist/leaflet.css";
import Navbar from "../components/Navbar";
import "../styles/Dashboard.css";
import { Title, Logo, MenuItems, BASE_URL, AGENT_RELEASE_BASE_URL } from "../constants";

// Fix default Leaflet marker icons broken by webpack
delete L.Icon.Default.prototype._getIconUrl;
L.Icon.Default.mergeOptions({
	iconRetinaUrl: require("leaflet/dist/images/marker-icon-2x.png"),
	iconUrl: require("leaflet/dist/images/marker-icon.png"),
	shadowUrl: require("leaflet/dist/images/marker-shadow.png"),
});

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
				<span className={`device-status-badge ${online ? "online" : "offline"}`}>
					<span className={`device-status ${online ? "online" : "offline"}`} />
					{online ? "Online" : "Offline"}
				</span>
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

const AGENT_DOWNLOADS = [
	{
		key: "darwin",
		icon: "fa-brands fa-apple",
		label: "macOS",
		desc: "Universal Binary (Apple Silicon + Intel)",
		file: "maxie-agent-darwin",
	},
	{
		key: "linux-amd64",
		icon: "fa-brands fa-linux",
		label: "Linux x86_64",
		desc: "amd64 (Ubuntu, Debian, CentOS…)",
		file: "maxie-agent-linux-amd64",
	},
	{
		key: "linux-arm64",
		icon: "fa-brands fa-linux",
		label: "Linux ARM64",
		desc: "arm64 (Raspberry Pi, AWS Graviton…)",
		file: "maxie-agent-linux-arm64",
	},
	{
		key: "windows",
		icon: "fa-brands fa-windows",
		label: "Windows",
		desc: "x86_64 (.exe)",
		file: "maxie-agent-windows-amd64.exe",
	},
];

function AgentDownloadSection() {
	return (
		<div className="agent-download-section">
			<div className="agent-download-header">
				<i className="fa-solid fa-download" />
				<h2>Download Agent</h2>
				<span className="agent-download-badge">latest</span>
			</div>
			<p className="agent-download-desc">
				Install the agent on each device you want to monitor.
				Run it once — it registers automatically to your account.
			</p>
			<div className="agent-download-grid">
				{AGENT_DOWNLOADS.map((d) => (
					<a
						key={d.key}
						href={`${AGENT_RELEASE_BASE_URL}/${d.file}`}
						target="_blank"
						rel="noreferrer"
						className="agent-download-card"
					>
						<i className={`agent-dl-os-icon ${d.icon}`} />
						<div className="agent-dl-info">
							<div className="agent-dl-label">{d.label}</div>
							<div className="agent-dl-desc">{d.desc}</div>
						</div>
						<i className="fa-solid fa-arrow-down agent-dl-arrow" />
					</a>
				))}
			</div>
		</div>
	);
}

function AgentMap({ agents }) {
	const geoAgents = agents.filter(a => {
		const geo = a.client_data?.geo_location;
		return geo?.lat != null && geo?.lon != null;
	});

	const center = geoAgents.length > 0 ? [
		geoAgents.reduce((s, a) => s + a.client_data.geo_location.lat, 0) / geoAgents.length,
		geoAgents.reduce((s, a) => s + a.client_data.geo_location.lon, 0) / geoAgents.length,
	] : [20, 0];

	return (
		<div className="agent-map-section">
			<div className="agent-map-header">
				<i className="fa-solid fa-earth-asia" />
				<h2>Agent Locations</h2>
				<span className="agent-map-count">
					{geoAgents.length > 0 ? `${geoAgents.length} located` : "pending"}
				</span>
			</div>
			{geoAgents.length === 0 ? (
				<div className="agent-map-empty">
					<i className="fa-solid fa-location-dot" />
					<span>Location data not yet available — will appear after the next agent sync</span>
				</div>
			) : (
				<MapContainer
					center={center}
					zoom={geoAgents.length === 1 ? 6 : 3}
					className="agent-map"
					scrollWheelZoom={false}
				>
					<TileLayer
						attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
						url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
					/>
					{geoAgents.map(a => {
						const geo = a.client_data.geo_location;
						const online = isOnline(a.last_seen);
						return (
							<Marker key={a.agent_id} position={[geo.lat, geo.lon]}>
								<Popup>
									<div className="map-popup">
										<div className="map-popup-name">
											<span className={`map-popup-dot ${online ? "online" : "offline"}`} />
											{a.name || a.agent_id}
										</div>
										{geo.city && <div className="map-popup-loc">{geo.city}{geo.country ? `, ${geo.country}` : ""}</div>}
										<div className="map-popup-coords">{geo.lat.toFixed(4)}, {geo.lon.toFixed(4)}</div>
									</div>
								</Popup>
							</Marker>
						);
					})}
				</MapContainer>
			)}
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

					<AgentDownloadSection />

					<AgentMap agents={agents} />

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
