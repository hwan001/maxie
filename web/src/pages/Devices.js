import React, { useState, useEffect, useRef } from "react";
import { useSearchParams } from "react-router-dom";
import axios from "axios";
import Navbar from "../components/Navbar";
import "../styles/Dashboard.css";
import { Title, Logo, MenuItems, BASE_URL } from "../constants";

const OS_ICONS = {
	darwin: "fa-brands fa-apple",
	windows: "fa-brands fa-windows",
	linux: "fa-brands fa-linux",
};

const DRIVE_TYPE_OPTIONS = [
	{ value: "local", label: "Local" },
	{ value: "external", label: "External" },
	{ value: "network", label: "Network" },
	{ value: "google", label: "Google Drive" },
	{ value: "onedrive", label: "OneDrive" },
	{ value: "dropbox", label: "Dropbox" },
	{ value: "naver", label: "NAVER Cloud" },
];

const DRIVE_TYPE_ICONS = {
	google: "fa-brands fa-google",
	naver: "fa-solid fa-cloud",
	onedrive: "fa-brands fa-microsoft",
	dropbox: "fa-brands fa-dropbox",
	local: "fa-solid fa-hard-drive",
	external: "fa-solid fa-plug",
	network: "fa-solid fa-network-wired",
};

function fmtBytes(bytes) {
	if (!bytes) return "0 B";
	const units = ["B", "KB", "MB", "GB", "TB"];
	let i = 0, val = bytes;
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

// ── Searchable, height-limited, drag-resizable info table ──────────────────────

function SearchableInfoTable({ rows, columns, defaultHeight = 180, emptyMsg = "No data" }) {
	const [search, setSearch] = useState("");
	const containerRef = useRef(null);
	const dragRef = useRef({ dragging: false, startY: 0, startH: 0 });

	const filtered = rows.filter(row =>
		columns.some(col => String(row[col.key] ?? "").toLowerCase().includes(search.toLowerCase()))
	);

	const onDragStart = (e) => {
		const el = containerRef.current;
		if (!el) return;
		dragRef.current = { dragging: true, startY: e.clientY, startH: el.offsetHeight };
		const onMove = (ev) => {
			if (!dragRef.current.dragging) return;
			const newH = Math.max(80, dragRef.current.startH + ev.clientY - dragRef.current.startY);
			el.style.height = newH + "px";
		};
		const onUp = () => {
			dragRef.current.dragging = false;
			document.removeEventListener("mousemove", onMove);
			document.removeEventListener("mouseup", onUp);
		};
		document.addEventListener("mousemove", onMove);
		document.addEventListener("mouseup", onUp);
		e.preventDefault();
	};

	return (
		<div>
			<div className="info-card-search">
				<i className="fa-solid fa-magnifying-glass" />
				<input
					value={search}
					onChange={e => setSearch(e.target.value)}
					placeholder="Search…"
					className="info-card-search-input"
				/>
			</div>
			<div ref={containerRef} className="info-card-table-wrap" style={{ height: defaultHeight }}>
				{rows.length === 0 ? (
					<div className="empty-state" style={{ padding: "1.25rem", fontSize: "0.8rem" }}>
						{emptyMsg}
					</div>
				) : filtered.length === 0 ? (
					<div className="empty-state" style={{ padding: "1.25rem", fontSize: "0.8rem" }}>
						No matches for "{search}"
					</div>
				) : (
					<table className="data-table">
						<thead>
							<tr>{columns.map(col => <th key={col.key}>{col.label}</th>)}</tr>
						</thead>
						<tbody>
							{filtered.map((row, i) => (
								<tr key={i}>
									{columns.map(col => (
										<td key={col.key}>{col.render ? col.render(row) : row[col.key]}</td>
									))}
								</tr>
							))}
						</tbody>
					</table>
				)}
			</div>
			<div className="info-card-resize-handle" onMouseDown={onDragStart}>
				<i className="fa-solid fa-grip-lines" />
			</div>
		</div>
	);
}

// ── Drives tab ─────────────────────────────────────────────────────────────────

const EMPTY_ADD_FORM = { path: "", label: "", drive_type: "local" };

function DriveExcludeSection({ drive, onChange }) {
	const [dirInput, setDirInput] = useState("");
	const [extInput, setExtInput] = useState("");

	const addDir = (e) => {
		e.preventDefault();
		const v = dirInput.trim();
		if (!v) return;
		if (!(drive.exclude_dirs ?? []).includes(v)) {
			onChange({ ...drive, exclude_dirs: [...(drive.exclude_dirs ?? []), v] });
		}
		setDirInput("");
	};

	const addExt = (e) => {
		e.preventDefault();
		const v = extInput.trim().replace(/^\.+/, "");
		if (!v) return;
		if (!(drive.exclude_exts ?? []).includes(v)) {
			onChange({ ...drive, exclude_exts: [...(drive.exclude_exts ?? []), v] });
		}
		setExtInput("");
	};

	return (
		<div className="drive-item-excludes">
			<div className="drive-exclude-field">
				<div className="drive-exclude-field-label">
					<i className="fa-solid fa-folder-minus" /> Exclude Dirs
				</div>
				<div className="drive-exclude-tags">
					{(drive.exclude_dirs ?? []).map(d => (
						<span key={d} className="exclude-tag">
							{d}
							<button className="exclude-tag-remove" onClick={() =>
								onChange({ ...drive, exclude_dirs: drive.exclude_dirs.filter(x => x !== d) })
							}>×</button>
						</span>
					))}
				</div>
				<form onSubmit={addDir} className="drive-exclude-input-row">
					<input value={dirInput} onChange={e => setDirInput(e.target.value)}
						placeholder="/path/to/exclude" className="drive-exclude-input" />
					<button type="submit" className="drive-exclude-add-btn">Add</button>
				</form>
			</div>

			<div className="drive-exclude-field">
				<div className="drive-exclude-field-label">
					<i className="fa-solid fa-file-slash" /> Exclude Extensions
				</div>
				<div className="drive-exclude-tags">
					{(drive.exclude_exts ?? []).map(e => (
						<span key={e} className="exclude-tag">
							.{e}
							<button className="exclude-tag-remove" onClick={() =>
								onChange({ ...drive, exclude_exts: drive.exclude_exts.filter(x => x !== e) })
							}>×</button>
						</span>
					))}
				</div>
				<form onSubmit={addExt} className="drive-exclude-input-row">
					<input value={extInput} onChange={e => setExtInput(e.target.value)}
						placeholder="e.g. tmp" className="drive-exclude-input" />
					<button type="submit" className="drive-exclude-add-btn">Add</button>
				</form>
			</div>
		</div>
	);
}

function DrivesTab({ agent, onRefresh }) {
	const [drives, setDrives] = useState(agent.drives ?? []);
	const [saving, setSaving] = useState(false);
	const [expandedPath, setExpandedPath] = useState(null);
	const [deleteConfirm, setDeleteConfirm] = useState(null);
	const [addForm, setAddForm] = useState(EMPTY_ADD_FORM);
	const [addOpen, setAddOpen] = useState(false);
	const [errMsg, setErrMsg] = useState("");

	useEffect(() => {
		setDrives(agent.drives ?? []);
	}, [agent.drives]);

	const persistDrives = async (newDrives) => {
		setSaving(true);
		setErrMsg("");
		try {
			await axios.put(`${BASE_URL}/protected/agents/${agent.agent_id}/drives`, { drives: newDrives }, { withCredentials: true });
			setDrives(newDrives);
			if (onRefresh) onRefresh();
		} catch {
			setErrMsg("Failed to save drive settings. Please try again.");
		} finally {
			setSaving(false);
		}
	};

	const saveDriveEdit = (updatedDrive) => {
		const newDrives = drives.map(d => d.path === updatedDrive.path ? updatedDrive : d);
		persistDrives(newDrives);
	};

	const confirmDelete = async () => {
		const newDrives = drives.filter(d => d.path !== deleteConfirm);
		await persistDrives(newDrives);
		setDeleteConfirm(null);
		if (expandedPath === deleteConfirm) setExpandedPath(null);
	};

	const submitAdd = async (e) => {
		e.preventDefault();
		const path = addForm.path.trim();
		if (!path) return;
		if (drives.some(d => d.path === path)) {
			setErrMsg("This path is already monitored.");
			return;
		}
		const newDrive = {
			path,
			label: addForm.label.trim(),
			drive_type: addForm.drive_type || "local",
			exclude_dirs: [],
			exclude_exts: [],
		};
		await persistDrives([...drives, newDrive]);
		setAddForm(EMPTY_ADD_FORM);
		setAddOpen(false);
	};

	return (
		<div className="drives-tab">
			{errMsg && (
				<div className="drives-error-banner" role="alert">
					<i className="fa-solid fa-circle-exclamation" />
					{errMsg}
					<button className="drives-error-dismiss" onClick={() => setErrMsg("")} aria-label="Dismiss">
						<i className="fa-solid fa-xmark" />
					</button>
				</div>
			)}
			{/* Drives list */}
			{drives.length === 0 ? (
				<div className="drives-empty">
					<i className="fa-solid fa-hard-drive" />
					No drives monitored — add one below
				</div>
			) : (
				<div className="drives-list">
					{drives.map(d => {
						const expanded = expandedPath === d.path;
						const icon = DRIVE_TYPE_ICONS[d.drive_type] ?? "fa-solid fa-hard-drive";
						const pendingDelete = deleteConfirm === d.path;
						return (
							<div key={d.path} className={`drive-item${expanded ? " expanded" : ""}`}>
								<div className="drive-item-row">
									<i className={`drive-item-icon ${icon}`} />
									<div className="drive-item-info">
										<span className="drive-item-label">{d.label || d.path.split("/").pop() || d.path}</span>
										<span className="drive-item-path" title={d.path}>{d.path}</span>
									</div>
									<span className="drive-item-type-badge">{d.drive_type || "local"}</span>
									<button
										className="drive-item-btn"
										title={expanded ? "Collapse" : "Edit excludes"}
										onClick={() => setExpandedPath(expanded ? null : d.path)}
									>
										<i className={`fa-solid ${expanded ? "fa-chevron-up" : "fa-sliders"}`} />
									</button>
									{pendingDelete ? (
										<div className="drive-delete-confirm">
											<span>Remove drive?</span>
											<button className="drive-delete-yes" onClick={confirmDelete} disabled={saving}>Yes</button>
											<button className="drive-delete-no" onClick={() => setDeleteConfirm(null)}>No</button>
										</div>
									) : (
										<button
											className="drive-item-btn danger"
											title="Remove drive"
											onClick={() => setDeleteConfirm(d.path)}
										>
											<i className="fa-solid fa-trash" />
										</button>
									)}
								</div>

								{expanded && (
									<DriveExcludeSection
										drive={d}
										onChange={(updated) => {
											// Update local state immediately; save on collapse or explicit save
											setDrives(prev => prev.map(x => x.path === updated.path ? updated : x));
										}}
									/>
								)}

								{expanded && (
									<div className="drive-item-save-row">
										<button
											className="drive-item-save-btn"
											onClick={() => saveDriveEdit(drives.find(x => x.path === d.path))}
											disabled={saving}
										>
											{saving
												? <><i className="fa-solid fa-spinner fa-spin" /> Saving…</>
												: <><i className="fa-solid fa-floppy-disk" /> Save changes</>}
										</button>
									</div>
								)}
							</div>
						);
					})}
				</div>
			)}

			{/* Delete warning note */}
			<p className="drives-note">
				<i className="fa-solid fa-triangle-exclamation" />
				Removing a drive also deletes all its indexed files from the server database.
				The agent config is updated on next heartbeat.
			</p>

			{/* Add drive */}
			{addOpen ? (
				<form className="drives-add-form" onSubmit={submitAdd}>
					<div className="drives-add-form-title">
						<i className="fa-solid fa-plus-circle" /> Add Monitored Drive
					</div>
					<div className="drives-add-row">
						<label className="drives-add-label">Path <span className="required">*</span></label>
						<input
							className="drives-add-input"
							placeholder="/home/user/Documents"
							value={addForm.path}
							onChange={e => setAddForm(f => ({ ...f, path: e.target.value }))}
							required
						/>
					</div>
					<div className="drives-add-row">
						<label className="drives-add-label">Label</label>
						<input
							className="drives-add-input"
							placeholder="My Documents (optional)"
							value={addForm.label}
							onChange={e => setAddForm(f => ({ ...f, label: e.target.value }))}
						/>
					</div>
					<div className="drives-add-row">
						<label className="drives-add-label">Type</label>
						<select
							className="drives-add-select"
							value={addForm.drive_type}
							onChange={e => setAddForm(f => ({ ...f, drive_type: e.target.value }))}
						>
							{DRIVE_TYPE_OPTIONS.map(o => (
								<option key={o.value} value={o.value}>{o.label}</option>
							))}
						</select>
					</div>
					<div className="drives-add-actions">
						<button type="submit" className="drives-add-submit" disabled={saving}>
							{saving ? <><i className="fa-solid fa-spinner fa-spin" /> Adding…</> : "Add Drive"}
						</button>
						<button type="button" className="drives-add-cancel" onClick={() => {
							setAddOpen(false);
							setAddForm(EMPTY_ADD_FORM);
						}}>Cancel</button>
					</div>
				</form>
			) : (
				<button className="drives-add-open-btn" onClick={() => setAddOpen(true)}>
					<i className="fa-solid fa-plus" /> Add Drive
				</button>
			)}
		</div>
	);
}

// ── Device list sidebar item ───────────────────────────────────────────────────

function DeviceListItem({ agent, selected, onClick, onDelete }) {
	const os = agent.client_data?.system_info?.os ?? "unknown";
	const online = isOnline(agent.last_seen);
	const [confirming, setConfirming] = useState(false);

	const handleDeleteClick = (e) => {
		e.stopPropagation();
		setConfirming(true);
	};

	const handleConfirm = (e) => {
		e.stopPropagation();
		onDelete(agent.agent_id);
		setConfirming(false);
	};

	const handleCancel = (e) => {
		e.stopPropagation();
		setConfirming(false);
	};

	return (
		<div
			className={`device-list-item${selected ? " active" : ""}`}
			onClick={() => onClick(agent)}
		>
			<div className={`device-status ${online ? "online" : "offline"}`} />
			<i className={OS_ICONS[os] ?? "fa-solid fa-desktop"} />
			<div className="device-list-info">
				<div className="device-list-name">{agent.name || agent.agent_id}</div>
				<div className="device-list-meta">{os} · {fmtTime(agent.last_seen)}</div>
			</div>
			{confirming ? (
				<div className="device-list-delete-confirm">
					<button className="device-list-confirm-yes" onClick={handleConfirm} title="Confirm delete">
						<i className="fa-solid fa-check" />
					</button>
					<button className="device-list-confirm-no" onClick={handleCancel} title="Cancel">
						<i className="fa-solid fa-xmark" />
					</button>
				</div>
			) : (
				<button className="device-list-delete-btn" onClick={handleDeleteClick} title="Remove agent">
					<i className="fa-solid fa-trash" />
				</button>
			)}
		</div>
	);
}

// ── Device detail panel ────────────────────────────────────────────────────────

function DeviceDetail({ agent, onRefresh }) {
	const [detailTab, setDetailTab] = useState("info");

	if (!agent) {
		return (
			<div className="empty-state">
				<i className="fa-solid fa-desktop" />
				Select a device to view details
			</div>
		);
	}

	const os = agent.client_data?.system_info?.os ?? "unknown";
	const sysInfo = agent.client_data?.system_info ?? {};
	const cpu = sysInfo.cpus?.[0];
	const ifaces = agent.client_data?.network_interfaces ?? [];
	const ports = agent.client_data?.active_ports ?? [];
	const stats = agent.file_stats ?? {};
	const online = isOnline(agent.last_seen);
	const driveCount = agent.drives?.length ?? 0;

	const publicIP = agent.client_data?.public_ip ?? null;
	const geoLoc = agent.client_data?.geo_location ?? null;
	const networkDevices = agent.client_data?.network_devices ?? [];
	const wifiNetworks = agent.client_data?.wifi_networks ?? [];
	const wifiHistory = agent.client_data?.wifi_history ?? [];
	const btDevices = agent.client_data?.bluetooth_devices ?? [];

	const ifaceColumns = [
		{ key: "name", label: "Name" },
		{ key: "ip", label: "IP Address" },
	];

	const portColumns = [
		{
			key: "port", label: "Port",
			render: p => <span className="port-badge">{p.port}</span>
		},
		{ key: "local_address", label: "Address" },
	];

	const neighborColumns = [
		{ key: "ip", label: "IP Address" },
		{ key: "mac", label: "MAC" },
		{ key: "hostname", label: "Hostname" },
	];

	const wifiColumns = [
		{ key: "ssid", label: "SSID" },
		{ key: "bssid", label: "BSSID" },
		{ key: "signal", label: "Signal" },
		{ key: "security", label: "Security" },
		{ key: "channel", label: "CH" },
	];

	const btColumns = [
		{ key: "name", label: "Name" },
		{ key: "address", label: "Address" },
		{
			key: "connected", label: "Connected",
			render: d => (
				<span className="port-badge" style={{
					background: d.connected ? "var(--color-success-light, #d1fae5)" : "var(--color-accent-light)",
					color: d.connected ? "var(--color-success, #059669)" : "var(--color-accent)",
				}}>
					{d.connected ? "Yes" : "No"}
				</span>
			),
		},
	];

	return (
		<div className="device-detail">
			{/* Header */}
			<div className="device-detail-header">
				<div className={`device-status ${online ? "online" : "offline"}`} />
				<i className={`device-os-icon ${OS_ICONS[os] ?? "fa-solid fa-desktop"}`} />
				<div>
					<h2>{agent.name || agent.agent_id}</h2>
					<div className="device-detail-sub">{agent.agent_id}</div>
				</div>
				<span className={`status-badge ${online ? "online" : "offline"}`}>
					{online ? "Online" : "Offline"}
				</span>
			</div>

			{/* Tab bar */}
			<div className="detail-tabs">
				<button
					className={`detail-tab-btn${detailTab === "info" ? " active" : ""}`}
					onClick={() => setDetailTab("info")}
				>
					<i className="fa-solid fa-circle-info" /> Info
				</button>
				<button
					className={`detail-tab-btn${detailTab === "drives" ? " active" : ""}`}
					onClick={() => setDetailTab("drives")}
				>
					<i className="fa-solid fa-hard-drive" /> Drives
					{driveCount > 0 && <span className="detail-tab-count">{driveCount}</span>}
				</button>
			</div>

			{/* Info tab */}
			{detailTab === "info" && (
				<div className="info-grid">
					<div className="info-card">
						<div className="info-card-header">
							<i className="fa-solid fa-microchip" />
							<h3>System Info</h3>
						</div>
						<div className="info-card-body">
							{[
								["OS", sysInfo.os],
								["Platform", sysInfo.platform],
								["Version", sysInfo.platform_version],
								["CPU", cpu?.model_name],
								["Cores", cpu?.cores],
								["Speed", cpu?.speed_ghz ? `${cpu.speed_ghz.toFixed(2)} GHz` : null],
								["Registered", fmtTime(agent.registered_at)],
								["Last Seen", fmtTime(agent.last_seen)],
							].map(([k, v]) => v != null && (
								<div className="kv-row" key={k}>
									<span className="kv-key">{k}</span>
									<span className="kv-val">{v}</span>
								</div>
							))}
						</div>
					</div>

					<div className="info-card">
						<div className="info-card-header">
							<i className="fa-solid fa-chart-bar" />
							<h3>File Statistics</h3>
						</div>
						<div className="info-card-body">
							{[
								["Total Files", stats.total_files?.toLocaleString()],
								["Total Size", fmtBytes(stats.total_size)],
								["Duplicates", stats.duplicate_count],
								["Last Scanned", fmtTime(stats.last_scanned)],
							].map(([k, v]) => v != null && (
								<div className="kv-row" key={k}>
									<span className="kv-key">{k}</span>
									<span className="kv-val">{v}</span>
								</div>
							))}
						</div>
					</div>

					<div className="info-card">
						<div className="info-card-header">
							<i className="fa-solid fa-network-wired" />
							<h3>Network Interfaces</h3>
						</div>
						<SearchableInfoTable
							rows={ifaces}
							columns={ifaceColumns}
							defaultHeight={180}
							emptyMsg="No interfaces"
						/>
					</div>

					<div className="info-card">
						<div className="info-card-header">
							<i className="fa-solid fa-plug" />
							<h3>Active Ports</h3>
						</div>
						<SearchableInfoTable
							rows={ports}
							columns={portColumns}
							defaultHeight={180}
							emptyMsg="No active ports"
						/>
					</div>

					{/* Public IP & Location */}
					{(publicIP || geoLoc) && (
						<div className="info-card">
							<div className="info-card-header">
								<i className="fa-solid fa-earth-asia" />
								<h3>Public IP &amp; Location</h3>
							</div>
							<div className="info-card-body">
								{[
									["Public IP", publicIP],
									["Country", geoLoc?.country],
									["Region", geoLoc?.region],
									["City", geoLoc?.city],
									["Coordinates", geoLoc?.lat != null ? `${geoLoc.lat.toFixed(4)}, ${geoLoc.lon.toFixed(4)}` : null],
									["Timezone", geoLoc?.timezone],
									["Source", geoLoc?.source],
								].map(([k, v]) => v != null && (
									<div className="kv-row" key={k}>
										<span className="kv-key">{k}</span>
										<span className="kv-val">{v}</span>
									</div>
								))}
							</div>
						</div>
					)}

					{/* Nearby Network Devices (ARP) */}
					<div className="info-card">
						<div className="info-card-header">
							<i className="fa-solid fa-sitemap" />
							<h3>Nearby Devices</h3>
						</div>
						<SearchableInfoTable
							rows={networkDevices}
							columns={neighborColumns}
							defaultHeight={180}
							emptyMsg="No nearby devices found"
						/>
					</div>

					{/* Wi-Fi Networks */}
					<div className="info-card">
						<div className="info-card-header">
							<i className="fa-solid fa-wifi" />
							<h3>Wi-Fi Networks</h3>
						</div>
						<SearchableInfoTable
							rows={wifiNetworks}
							columns={wifiColumns}
							defaultHeight={180}
							emptyMsg="No Wi-Fi networks scanned"
						/>
					</div>

					{/* Wi-Fi History */}
					{wifiHistory.length > 0 && (
						<div className="info-card">
							<div className="info-card-header">
								<i className="fa-solid fa-clock-rotate-left" />
								<h3>Wi-Fi History</h3>
							</div>
							<div className="info-card-body" style={{ maxHeight: 200, overflowY: "auto" }}>
								{wifiHistory.map((ssid, i) => (
									<div className="kv-row" key={i}>
										<span className="kv-key"><i className="fa-solid fa-wifi" style={{ fontSize: "0.7rem", marginRight: "0.3rem" }} />{i + 1}</span>
										<span className="kv-val">{ssid}</span>
									</div>
								))}
							</div>
						</div>
					)}

					{/* Bluetooth Devices */}
					<div className="info-card">
						<div className="info-card-header">
							<i className="fa-brands fa-bluetooth-b" />
							<h3>Bluetooth Devices</h3>
						</div>
						<SearchableInfoTable
							rows={btDevices}
							columns={btColumns}
							defaultHeight={180}
							emptyMsg="No Bluetooth devices found"
						/>
					</div>
				</div>
			)}

			{/* Drives tab */}
			{detailTab === "drives" && (
				<DrivesTab agent={agent} onRefresh={onRefresh} />
			)}
		</div>
	);
}

// ── Page ───────────────────────────────────────────────────────────────────────

const Devices = () => {
	const [searchParams, setSearchParams] = useSearchParams();
	const [agents, setAgents] = useState([]);
	const [loading, setLoading] = useState(true);
	const [deleteErr, setDeleteErr] = useState("");

	const selectedId = searchParams.get("id");
	const selected = agents.find(a => a.agent_id === selectedId) ?? (agents.length > 0 ? agents[0] : null);

	const fetchData = () => {
		setLoading(true);
		axios.get(`${BASE_URL}/protected/agents`, { withCredentials: true })
			.then(r => setAgents(r.data?.agents ?? []))
			.catch(() => {})
			.finally(() => setLoading(false));
	};

	useEffect(() => { fetchData(); }, []);

	const selectAgent = (agent) => {
		setSearchParams({ id: agent.agent_id });
	};

	const handleDelete = async (agentId) => {
		setDeleteErr("");
		try {
			await axios.delete(`${BASE_URL}/protected/agents/${agentId}`, { withCredentials: true });
			setAgents(prev => prev.filter(a => a.agent_id !== agentId));
			if (selectedId === agentId) {
				setSearchParams({});
			}
		} catch {
			setDeleteErr("Failed to remove agent. Please try again.");
		}
	};

	return (
		<>
			<Navbar title={Title} logo={Logo} menuItems={MenuItems["app"]} />
			<div className="dashboard-layout">
				<aside className="dash-sidebar">
					<div className="dash-sidebar-header">Devices</div>
					{loading ? (
						<div className="loading-row"><span className="spinner" /></div>
					) : agents.length === 0 ? (
						<div style={{ padding: "1rem", fontSize: "0.8rem", color: "var(--color-text-muted)" }}>
							No devices registered
						</div>
					) : (
						<div className="device-list-sidebar">
							{agents.map(a => (
								<DeviceListItem
									key={a.agent_id}
									agent={a}
									selected={selected?.agent_id === a.agent_id}
									onClick={selectAgent}
									onDelete={handleDelete}
								/>
							))}
						</div>
					)}
				</aside>
				<main className="dash-main">
					{deleteErr && (
						<div className="drives-error-banner" role="alert">
							<i className="fa-solid fa-circle-exclamation" />
							{deleteErr}
							<button className="drives-error-dismiss" onClick={() => setDeleteErr("")} aria-label="Dismiss">
								<i className="fa-solid fa-xmark" />
							</button>
						</div>
					)}
					<div className="dash-header">
						<div>
							<h1>Devices</h1>
							<p>Agent system info and drive management</p>
						</div>
						<button className="dash-refresh-btn" onClick={fetchData}>
							<i className="fa-solid fa-rotate-right" />
							Refresh
						</button>
					</div>
					{loading ? (
						<div className="loading-row"><span className="spinner" />Loading devices…</div>
					) : (
						<DeviceDetail agent={selected} onRefresh={fetchData} />
					)}
				</main>
			</div>
		</>
	);
};

export default Devices;
