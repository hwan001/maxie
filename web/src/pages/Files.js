import React, { useState, useEffect, useCallback, useRef } from "react";
import axios from "axios";
import Navbar from "../components/Navbar";
import "../styles/Dashboard.css";
import { Title, Logo, MenuItems, BASE_URL } from "../constants";

const DRIVE_ICONS = {
	google: "fa-brands fa-google",
	naver: "fa-solid fa-cloud",
	onedrive: "fa-brands fa-microsoft",
	dropbox: "fa-brands fa-dropbox",
	local: "fa-solid fa-hard-drive",
	other: "fa-solid fa-folder",
};

const INTERVAL_OPTIONS = [
	{ label: "5 minutes", value: 5 },
	{ label: "10 minutes", value: 10 },
	{ label: "30 minutes", value: 30 },
	{ label: "1 hour", value: 60 },
	{ label: "3 hours", value: 180 },
	{ label: "6 hours", value: 360 },
	{ label: "24 hours", value: 1440 },
];

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

function basename(path) {
	return path ? path.replace(/\\/g, "/").split("/").pop() : "—";
}

// Returns a human-readable group label for a file based on agent + drive path matching
function getDriveGroupLabel(file, agents) {
	const agent = agents.find(a => a.agent_id === file.agent_id);
	if (!agent) return file.agent_name || file.agent_id || "Unknown";

	const drives = agent.drives ?? [];
	let bestDrive = null;
	let bestLen = 0;
	for (const d of drives) {
		if (file.fullpath.startsWith(d.path) && d.path.length > bestLen) {
			bestDrive = d;
			bestLen = d.path.length;
		}
	}

	const agentLabel = agent.name || agent.agent_id;
	if (bestDrive) return `${agentLabel} / ${bestDrive.label || basename(bestDrive.path)}`;
	return agentLabel;
}

function groupFilesByDrive(files, agents) {
	const groups = {};
	for (const file of files) {
		const label = getDriveGroupLabel(file, agents);
		if (!groups[label]) groups[label] = [];
		groups[label].push(file);
	}
	return groups;
}

function groupDupesByDrive(dupeGroups, agents) {
	const driveMap = {};
	for (const g of dupeGroups) {
		const firstFile = g.files?.[0];
		const label = firstFile ? getDriveGroupLabel(firstFile, agents) : "Unknown";
		if (!driveMap[label]) driveMap[label] = [];
		driveMap[label].push(g);
	}
	return driveMap;
}

// ── All Files Tab ──────────────────────────────────────────────────────────────

function AllFilesTab({ agents }) {
	const [files, setFiles] = useState([]);
	const [total, setTotal] = useState(0);
	const [page, setPage] = useState(1);
	const [limit] = useState(50);
	const [search, setSearch] = useState("");
	const [draftSearch, setDraftSearch] = useState("");
	const [agentFilter, setAgentFilter] = useState("");
	const [driveFilter, setDriveFilter] = useState("");
	const [sortBy, setSortBy] = useState("synced_at");
	const [sortDir, setSortDir] = useState("desc");
	const [loading, setLoading] = useState(false);
	const [fetchError, setFetchError] = useState(null);
	const [deleting, setDeleting] = useState(null);
	const [groupView, setGroupView] = useState(false);
	const [collapsedGroups, setCollapsedGroups] = useState({});

	const tableWrapRef = useRef(null);
	const dragRef = useRef({ dragging: false, startY: 0, startH: 0 });

	const onDragStart = (e) => {
		const el = tableWrapRef.current;
		if (!el) return;
		dragRef.current = { dragging: true, startY: e.clientY, startH: el.offsetHeight };
		document.addEventListener("mousemove", onDragMove);
		document.addEventListener("mouseup", onDragEnd);
		e.preventDefault();
	};
	const onDragMove = (e) => {
		if (!dragRef.current.dragging) return;
		const delta = e.clientY - dragRef.current.startY;
		const newH = Math.max(160, dragRef.current.startH + delta);
		tableWrapRef.current.style.height = newH + "px";
	};
	const onDragEnd = () => {
		dragRef.current.dragging = false;
		document.removeEventListener("mousemove", onDragMove);
		document.removeEventListener("mouseup", onDragEnd);
	};

	const totalPages = Math.max(1, Math.ceil(total / limit));

	const fetchFiles = useCallback(() => {
		setLoading(true);
		const params = { page, limit, sort_by: sortBy, sort_dir: sortDir };
		if (search) params.search = search;
		if (agentFilter) params.agent_id = agentFilter;
		if (driveFilter) params.drive_type = driveFilter;

		setFetchError(null);
		axios.get(`${BASE_URL}/protected/files`, { params, withCredentials: true })
			.then(r => {
				setFiles(r.data?.files ?? []);
				setTotal(r.data?.total ?? 0);
			})
			.catch(err => {
				const msg = err?.response?.data?.error || err?.message || "Failed to load files";
				setFetchError(msg);
			})
			.finally(() => setLoading(false));
	}, [page, limit, search, agentFilter, driveFilter, sortBy, sortDir]);

	useEffect(() => { fetchFiles(); }, [fetchFiles]);

	const handleSearch = (e) => {
		e.preventDefault();
		setSearch(draftSearch);
		setPage(1);
	};

	const handleSort = (col) => {
		if (sortBy === col) {
			setSortDir(d => d === "asc" ? "desc" : "asc");
		} else {
			setSortBy(col);
			setSortDir("desc");
		}
		setPage(1);
	};

	const handleDelete = async (file) => {
		if (!window.confirm(`Delete ${basename(file.fullpath)}?\n\nThis will queue a delete action on the agent.`)) return;
		setDeleting(file.fullpath);
		try {
			await axios.delete(`${BASE_URL}/protected/files`, { data: { agent_id: file.agent_id, fullpath: file.fullpath }, withCredentials: true });
			fetchFiles();
		} catch {
			alert("Failed to queue delete action.");
		} finally {
			setDeleting(null);
		}
	};

	const toggleGroup = (label) => {
		setCollapsedGroups(prev => ({ ...prev, [label]: !prev[label] }));
	};

	const SortIcon = ({ col }) => {
		if (sortBy !== col) return <i className="fa-solid fa-sort" style={{ opacity: 0.3 }} />;
		return <i className={`fa-solid fa-sort-${sortDir === "asc" ? "up" : "down"}`} />;
	};

	const FileRow = ({ f, showAgent = true }) => (
		<tr>
			<td className="files-name-cell">
				<i className={`${DRIVE_ICONS[f.drive_type] ?? "fa-solid fa-file"} files-row-icon`} />
				<span className="files-name" title={f.fullpath}>{basename(f.fullpath)}</span>
				<span className="files-path" title={f.fullpath}>{f.fullpath}</span>
			</td>
			<td className="files-size">{fmtBytes(f.size)}</td>
			<td className="files-date">{fmtTime(f.modified_at)}</td>
			<td><span className="port-badge">{f.drive_type}</span></td>
			{showAgent && <td className="files-agent">{f.agent_name || f.agent_id}</td>}
			<td>
				<button
					className="files-delete-btn"
					onClick={() => handleDelete(f)}
					disabled={deleting === f.fullpath}
					title="Delete file"
				>
					{deleting === f.fullpath
						? <i className="fa-solid fa-spinner fa-spin" />
						: <i className="fa-solid fa-trash" />}
				</button>
			</td>
		</tr>
	);

	const renderGroupedView = () => {
		const grouped = groupFilesByDrive(files, agents);
		return Object.entries(grouped).map(([label, groupFiles]) => (
			<div key={label} className="files-drive-group">
				<div className="files-drive-group-header" onClick={() => toggleGroup(label)}>
					<i className={`fa-solid fa-chevron-${collapsedGroups[label] ? "right" : "down"} files-drive-chevron`} />
					<i className="fa-solid fa-hard-drive" style={{ color: "var(--color-accent)", fontSize: "0.8rem" }} />
					<span className="files-drive-group-label">{label}</span>
					<span className="files-drive-count">{groupFiles.length} files</span>
				</div>
				{!collapsedGroups[label] && (
					<div className="files-table-wrap">
						<table className="data-table files-table">
							<thead>
								<tr>
									<th>Name</th>
									<th>Size</th>
									<th>Modified</th>
									<th>Type</th>
									<th></th>
								</tr>
							</thead>
							<tbody>
								{groupFiles.map((f, i) => (
									<FileRow key={i} f={f} showAgent={false} />
								))}
							</tbody>
						</table>
					</div>
				)}
			</div>
		));
	};

	return (
		<div className="files-panel">
			<div className="files-toolbar">
				<form onSubmit={handleSearch} className="files-search-form">
					<i className="fa-solid fa-magnifying-glass" />
					<input
						value={draftSearch}
						onChange={e => setDraftSearch(e.target.value)}
						placeholder="Search by path…"
						className="files-search-input"
					/>
					<button type="submit" className="files-search-btn">Search</button>
				</form>
				<div className="files-filters">
					<select value={agentFilter} onChange={e => { setAgentFilter(e.target.value); setPage(1); }} className="files-select">
						<option value="">All Agents</option>
						{agents.map(a => <option key={a.agent_id} value={a.agent_id}>{a.name || a.agent_id}</option>)}
					</select>
					<select value={driveFilter} onChange={e => { setDriveFilter(e.target.value); setPage(1); }} className="files-select">
						<option value="">All Drive Types</option>
						{["local", "google", "naver", "onedrive", "dropbox", "other"].map(t => (
							<option key={t} value={t}>{t}</option>
						))}
					</select>
				</div>
				<button
					className={`files-view-btn${groupView ? " active" : ""}`}
					onClick={() => setGroupView(g => !g)}
					title={groupView ? "Switch to flat view" : "Group by drive"}
				>
					<i className={`fa-solid ${groupView ? "fa-list" : "fa-layer-group"}`} />
					{groupView ? "Flat" : "Group by Drive"}
				</button>
				<span className="files-count">{total.toLocaleString()} files</span>
			</div>

			{loading ? (
				<div className="loading-row"><span className="spinner" />Loading…</div>
			) : fetchError ? (
				<div className="empty-state" style={{ color: "var(--color-danger, #ef4444)" }}>
					<i className="fa-solid fa-circle-exclamation" />
					<span>Could not load files: <strong>{fetchError}</strong></span>
				</div>
			) : files.length === 0 ? (
				<div className="empty-state">
					<i className="fa-solid fa-folder-open" />
					{agents.length === 0
						? <span>No agents linked to your account.<br />Open the Maxie agent → Settings → <strong>Link User Code</strong> and paste your User ID.</span>
						: "No files found"}
				</div>
			) : groupView ? (
				<>
					<div className="files-table-wrap files-table-wrap--resizable" ref={tableWrapRef}>
						<div className="files-drive-groups">
							{renderGroupedView()}
						</div>
					</div>
					<div className="files-resize-handle" onMouseDown={onDragStart} title="Drag to resize">
						<i className="fa-solid fa-grip-lines" />
					</div>
					<div className="files-pagination">
						<button className="page-btn" onClick={() => setPage(1)} disabled={page === 1}>
							<i className="fa-solid fa-angles-left" />
						</button>
						<button className="page-btn" onClick={() => setPage(p => p - 1)} disabled={page === 1}>
							<i className="fa-solid fa-angle-left" />
						</button>
						<span className="page-info">Page {page} of {totalPages}</span>
						<button className="page-btn" onClick={() => setPage(p => p + 1)} disabled={page >= totalPages}>
							<i className="fa-solid fa-angle-right" />
						</button>
						<button className="page-btn" onClick={() => setPage(totalPages)} disabled={page >= totalPages}>
							<i className="fa-solid fa-angles-right" />
						</button>
					</div>
				</>
			) : (
				<>
					<div className="files-table-wrap files-table-wrap--resizable" ref={tableWrapRef}>
						<table className="data-table files-table">
							<thead>
								<tr>
									<th onClick={() => handleSort("name")} className="sortable">
										Name <SortIcon col="name" />
									</th>
									<th onClick={() => handleSort("size")} className="sortable">
										Size <SortIcon col="size" />
									</th>
									<th onClick={() => handleSort("modified_at")} className="sortable">
										Modified <SortIcon col="modified_at" />
									</th>
									<th>Type</th>
									<th>Agent</th>
									<th></th>
								</tr>
							</thead>
							<tbody>
								{files.map((f, i) => (
									<FileRow key={i} f={f} showAgent={true} />
								))}
							</tbody>
						</table>
					</div>
					<div className="files-resize-handle" onMouseDown={onDragStart} title="Drag to resize">
						<i className="fa-solid fa-grip-lines" />
					</div>
					<div className="files-pagination">
						<button className="page-btn" onClick={() => setPage(1)} disabled={page === 1}>
							<i className="fa-solid fa-angles-left" />
						</button>
						<button className="page-btn" onClick={() => setPage(p => p - 1)} disabled={page === 1}>
							<i className="fa-solid fa-angle-left" />
						</button>
						<span className="page-info">Page {page} of {totalPages}</span>
						<button className="page-btn" onClick={() => setPage(p => p + 1)} disabled={page >= totalPages}>
							<i className="fa-solid fa-angle-right" />
						</button>
						<button className="page-btn" onClick={() => setPage(totalPages)} disabled={page >= totalPages}>
							<i className="fa-solid fa-angles-right" />
						</button>
					</div>
				</>
			)}
		</div>
	);
}

// ── Duplicates Tab ─────────────────────────────────────────────────────────────

function DuplicatesTab({ agents }) {
	const [groups, setGroups] = useState([]);
	const [loading, setLoading] = useState(false);
	const [agentFilter, setAgentFilter] = useState("");
	const [expanded, setExpanded] = useState({});
	const [deleting, setDeleting] = useState(null);
	const [driveView, setDriveView] = useState(false);
	const [collapsedDrives, setCollapsedDrives] = useState({});

	const fetchDupes = useCallback(() => {
		setLoading(true);
		const params = agentFilter ? { agent_id: agentFilter } : {};
		axios.get(`${BASE_URL}/protected/files/duplicates`, { params, withCredentials: true })
			.then(r => setGroups(r.data?.groups ?? []))
			.catch(() => {})
			.finally(() => setLoading(false));
	}, [agentFilter]);

	useEffect(() => { fetchDupes(); }, [fetchDupes]);

	const toggleGroup = (hash) => setExpanded(prev => ({ ...prev, [hash]: !prev[hash] }));
	const toggleDrive = (label) => setCollapsedDrives(prev => ({ ...prev, [label]: !prev[label] }));

	const handleDelete = async (file) => {
		if (!window.confirm(`Delete ${basename(file.fullpath)}?`)) return;
		setDeleting(file.fullpath);
		try {
			await axios.delete(`${BASE_URL}/protected/files`, { data: { agent_id: file.agent_id, fullpath: file.fullpath }, withCredentials: true });
			fetchDupes();
		} catch {
			alert("Failed to queue delete action.");
		} finally {
			setDeleting(null);
		}
	};

	const totalWasted = groups.reduce((s, g) => s + g.size * (g.count - 1), 0);

	const listRef = useRef(null);
	const dragRef = useRef({ dragging: false, startY: 0, startH: 0 });

	const onDragStart = (e) => {
		const el = listRef.current;
		if (!el) return;
		dragRef.current = { dragging: true, startY: e.clientY, startH: el.offsetHeight };
		document.addEventListener("mousemove", onDragMove);
		document.addEventListener("mouseup", onDragEnd);
		e.preventDefault();
	};
	const onDragMove = (e) => {
		if (!dragRef.current.dragging) return;
		const delta = e.clientY - dragRef.current.startY;
		const newH = Math.max(120, dragRef.current.startH + delta);
		listRef.current.style.height = newH + "px";
	};
	const onDragEnd = () => {
		dragRef.current.dragging = false;
		document.removeEventListener("mousemove", onDragMove);
		document.removeEventListener("mouseup", onDragEnd);
	};

	const DupeGroupContent = ({ g }) => (
		<>
			<div className="dupe-group-header" onClick={() => toggleGroup(g.hash)}>
				<i className={`fa-solid fa-chevron-${expanded[g.hash] ? "down" : "right"} dupe-chevron`} />
				<div className="dupe-group-info">
					<span className="dupe-group-count">
						<i className="fa-solid fa-clone" /> {g.count} copies
					</span>
					<span className="dupe-group-size">{fmtBytes(g.size)} each</span>
					<span className="dupe-wasted">
						<i className="fa-solid fa-triangle-exclamation" />
						{fmtBytes(g.size * (g.count - 1))} wasted
					</span>
				</div>
				<span className="dupe-hash">{g.hash.slice(0, 12)}…</span>
			</div>
			{expanded[g.hash] && (
				<table className="data-table dupe-files-table">
					<thead>
						<tr>
							<th>Path</th>
							<th>Agent</th>
							<th>Modified</th>
							<th></th>
						</tr>
					</thead>
					<tbody>
						{g.files.map((f, i) => (
							<tr key={i} className={i === 0 ? "dupe-newest" : ""}>
								<td className="files-name-cell">
									<span className="files-name" title={f.fullpath}>{basename(f.fullpath)}</span>
									<span className="files-path" title={f.fullpath}>{f.fullpath}</span>
								</td>
								<td className="files-agent">{f.agent_name || f.agent_id}</td>
								<td className="files-date">{fmtTime(f.modified_at)}</td>
								<td style={{ display: "flex", alignItems: "center", gap: "0.5rem" }}>
									<button
										className="files-delete-btn"
										onClick={() => handleDelete(f)}
										disabled={deleting === f.fullpath}
									>
										{deleting === f.fullpath
											? <i className="fa-solid fa-spinner fa-spin" />
											: <i className="fa-solid fa-trash" />}
									</button>
									{i === 0 && <span className="newest-badge">newest</span>}
								</td>
							</tr>
						))}
					</tbody>
				</table>
			)}
		</>
	);

	const renderDriveView = () => {
		const driveMap = groupDupesByDrive(groups, agents);
		return Object.entries(driveMap).map(([driveLabel, driveGroups]) => {
			const driveWasted = driveGroups.reduce((s, g) => s + g.size * (g.count - 1), 0);
			return (
				<div key={driveLabel} className="dupe-drive-section">
					<div className="dupe-drive-section-header" onClick={() => toggleDrive(driveLabel)}>
						<i className={`fa-solid fa-chevron-${collapsedDrives[driveLabel] ? "right" : "down"} dupe-chevron`} />
						<i className="fa-solid fa-hard-drive" style={{ color: "var(--color-accent)", fontSize: "0.8rem" }} />
						<span className="dupe-drive-label">{driveLabel}</span>
						<span className="dupe-group-count" style={{ marginLeft: "auto" }}>
							{driveGroups.length} groups
						</span>
						<span className="dupe-wasted">
							<i className="fa-solid fa-triangle-exclamation" />
							{fmtBytes(driveWasted)} wasted
						</span>
					</div>
					{!collapsedDrives[driveLabel] && (
						<div className="dupe-drive-groups">
							{driveGroups.map(g => (
								<div key={g.hash} className="dupe-group">
									<DupeGroupContent g={g} />
								</div>
							))}
						</div>
					)}
				</div>
			);
		});
	};

	return (
		<div className="files-panel">
			<div className="files-toolbar">
				<select value={agentFilter} onChange={e => setAgentFilter(e.target.value)} className="files-select">
					<option value="">All Agents</option>
					{agents.map(a => <option key={a.agent_id} value={a.agent_id}>{a.name || a.agent_id}</option>)}
				</select>
				{groups.length > 0 && (
					<span className="files-count">
						{groups.length} duplicate groups · {fmtBytes(totalWasted)} wasted
					</span>
				)}
				<button
					className={`files-view-btn${driveView ? " active" : ""}`}
					onClick={() => setDriveView(v => !v)}
					title={driveView ? "Switch to grouped view" : "View by drive"}
				>
					<i className={`fa-solid ${driveView ? "fa-clone" : "fa-hard-drive"}`} />
					{driveView ? "By Hash" : "By Drive"}
				</button>
			</div>

			{loading ? (
				<div className="loading-row"><span className="spinner" />Loading duplicates…</div>
			) : groups.length === 0 ? (
				<div className="empty-state">
					<i className="fa-solid fa-circle-check" style={{ color: "#22c55e" }} />
					No duplicates found
				</div>
			) : driveView ? (
				<>
					<div className="dupe-list" ref={listRef}>
						{renderDriveView()}
					</div>
					<div className="dupe-resize-handle" onMouseDown={onDragStart} title="Drag to resize">
						<i className="fa-solid fa-grip-lines" />
					</div>
				</>
			) : (
				<>
					<div className="dupe-list" ref={listRef}>
						{groups.map(g => (
							<div key={g.hash} className="dupe-group">
								<DupeGroupContent g={g} />
							</div>
						))}
					</div>
					<div className="dupe-resize-handle" onMouseDown={onDragStart} title="Drag to resize">
						<i className="fa-solid fa-grip-lines" />
					</div>
				</>
			)}
		</div>
	);
}

// ── Scan Schedule Content ─────────────────────────────────────────────────────

function ScanScheduleContent({ agents, onAgentsChange }) {
	const [intervals, setIntervals] = useState({});
	const [saving, setSaving] = useState(null);
	const [saved, setSaved] = useState({});

	useEffect(() => {
		const init = {};
		for (const a of agents) {
			init[a.agent_id] = a.scan_interval_minutes > 0 ? a.scan_interval_minutes : 10;
		}
		setIntervals(init);
	}, [agents]);

	const saveInterval = async (agentId) => {
		setSaving(agentId);
		try {
			await axios.put(`${BASE_URL}/protected/agents/${agentId}/config`, {
				scan_interval_minutes: intervals[agentId],
			}, { withCredentials: true });
			setSaved(prev => ({ ...prev, [agentId]: true }));
			setTimeout(() => setSaved(prev => ({ ...prev, [agentId]: false })), 2000);
			if (onAgentsChange) onAgentsChange();
		} catch {
			alert("Failed to save scan interval.");
		} finally {
			setSaving(null);
		}
	};

	if (agents.length === 0) {
		return (
			<div className="empty-state">
				<i className="fa-solid fa-clock" />
				No agents registered
			</div>
		);
	}

	return (
		<div className="scan-schedule-list">
			{agents.map(agent => (
				<div key={agent.agent_id} className="scan-schedule-row">
					<div className="scan-schedule-agent">
						<i className="fa-solid fa-computer" />
						<div>
							<div className="scan-schedule-name">{agent.name || agent.agent_id}</div>
							<div className="scan-schedule-id">{agent.agent_id}</div>
						</div>
					</div>
					<select
						className="files-select"
						value={intervals[agent.agent_id] ?? 10}
						onChange={e => setIntervals(prev => ({ ...prev, [agent.agent_id]: Number(e.target.value) }))}
					>
						{INTERVAL_OPTIONS.map(o => (
							<option key={o.value} value={o.value}>{o.label}</option>
						))}
					</select>
					<button
						className="drive-settings-save-btn"
						onClick={() => saveInterval(agent.agent_id)}
						disabled={saving === agent.agent_id}
					>
						{saving === agent.agent_id
							? <><i className="fa-solid fa-spinner fa-spin" /> Saving…</>
							: saved[agent.agent_id]
								? <><i className="fa-solid fa-check" /> Saved</>
								: <><i className="fa-solid fa-floppy-disk" /> Save</>}
					</button>
				</div>
			))}
		</div>
	);
}

// ── Drive Settings Content ─────────────────────────────────────────────────────

function DriveSettingsContent({ agents, onAgentsChange }) {
	const [saving, setSaving] = useState(null);
	const [localAgents, setLocalAgents] = useState([]);

	useEffect(() => {
		setLocalAgents(agents.map(a => ({
			...a,
			drives: (a.drives ?? []).map(d => ({
				...d,
				exclude_dirs: d.exclude_dirs ?? [],
				exclude_exts: d.exclude_exts ?? [],
			})),
		})));
	}, [agents]);

	const updateDrive = (agentIdx, driveIdx, field, value) => {
		setLocalAgents(prev => prev.map((a, ai) => {
			if (ai !== agentIdx) return a;
			return {
				...a,
				drives: a.drives.map((d, di) => di === driveIdx ? { ...d, [field]: value } : d),
			};
		}));
	};

	const addItem = (agentIdx, driveIdx, field, value) => {
		const trimmed = value.trim();
		if (!trimmed) return;
		const drive = localAgents[agentIdx].drives[driveIdx];
		if ((drive[field] ?? []).includes(trimmed)) return;
		updateDrive(agentIdx, driveIdx, field, [...(drive[field] ?? []), trimmed]);
	};

	const removeItem = (agentIdx, driveIdx, field, item) => {
		const drive = localAgents[agentIdx].drives[driveIdx];
		updateDrive(agentIdx, driveIdx, field, (drive[field] ?? []).filter(x => x !== item));
	};

	const saveDrives = async (agentIdx) => {
		const agent = localAgents[agentIdx];
		setSaving(agent.agent_id);
		try {
			await axios.put(`${BASE_URL}/protected/agents/${agent.agent_id}/drives`, { drives: agent.drives }, { withCredentials: true });
			if (onAgentsChange) onAgentsChange();
		} catch {
			alert("Failed to save drive settings.");
		} finally {
			setSaving(null);
		}
	};

	if (localAgents.length === 0) {
		return (
			<div className="empty-state">
				<i className="fa-solid fa-gear" />
				No agents registered
			</div>
		);
	}

	return (
		<>
			{localAgents.map((agent, ai) => (
				<div key={agent.agent_id} className="drive-settings-agent">
					<div className="drive-settings-agent-header">
						<i className="fa-solid fa-computer" />
						<span className="drive-settings-agent-name">{agent.name || agent.agent_id}</span>
						<button
							className="drive-settings-save-btn"
							onClick={() => saveDrives(ai)}
							disabled={saving === agent.agent_id}
						>
							{saving === agent.agent_id
								? <><i className="fa-solid fa-spinner fa-spin" /> Saving…</>
								: <><i className="fa-solid fa-floppy-disk" /> Save</>}
						</button>
					</div>
					{(agent.drives ?? []).length === 0 ? (
						<div className="drive-settings-empty">No drives configured for this agent</div>
					) : (
						agent.drives.map((drive, di) => (
							<DriveExcludeEditor
								key={drive.path}
								drive={drive}
								onAdd={(field, val) => addItem(ai, di, field, val)}
								onRemove={(field, item) => removeItem(ai, di, field, item)}
							/>
						))
					)}
				</div>
			))}
		</>
	);
}

function DriveExcludeEditor({ drive, onAdd, onRemove }) {
	const [dirInput, setDirInput] = useState("");
	const [extInput, setExtInput] = useState("");

	const submitDir = (e) => {
		e.preventDefault();
		onAdd("exclude_dirs", dirInput);
		setDirInput("");
	};

	const submitExt = (e) => {
		e.preventDefault();
		onAdd("exclude_exts", extInput);
		setExtInput("");
	};

	return (
		<div className="drive-exclude-editor">
			<div className="drive-exclude-header">
				<i className="fa-solid fa-hard-drive drive-exclude-icon" />
				<div className="drive-exclude-info">
					<span className="drive-exclude-label">{drive.label || drive.path}</span>
					<span className="drive-exclude-path" title={drive.path}>{drive.path}</span>
				</div>
				<span className="port-badge">{drive.drive_type}</span>
			</div>
			<div className="drive-exclude-fields">
				<div className="drive-exclude-field">
					<div className="drive-exclude-field-label">
						<i className="fa-solid fa-folder-minus" /> Exclude Directories
					</div>
					<div className="drive-exclude-tags">
						{(drive.exclude_dirs ?? []).map(d => (
							<span key={d} className="exclude-tag">
								{d}
								<button onClick={() => onRemove("exclude_dirs", d)} className="exclude-tag-remove">
									<i className="fa-solid fa-xmark" />
								</button>
							</span>
						))}
					</div>
					<form onSubmit={submitDir} className="drive-exclude-input-row">
						<input
							value={dirInput}
							onChange={e => setDirInput(e.target.value)}
							placeholder="e.g. build, dist, tmp"
							className="drive-exclude-input"
						/>
						<button type="submit" className="drive-exclude-add-btn">Add</button>
					</form>
				</div>
				<div className="drive-exclude-field">
					<div className="drive-exclude-field-label">
						<i className="fa-solid fa-file-circle-minus" /> Exclude Extensions
					</div>
					<div className="drive-exclude-tags">
						{(drive.exclude_exts ?? []).map(e => (
							<span key={e} className="exclude-tag">
								{e}
								<button onClick={() => onRemove("exclude_exts", e)} className="exclude-tag-remove">
									<i className="fa-solid fa-xmark" />
								</button>
							</span>
						))}
					</div>
					<form onSubmit={submitExt} className="drive-exclude-input-row">
						<input
							value={extInput}
							onChange={e => setExtInput(e.target.value)}
							placeholder="e.g. .log, .tmp, .class"
							className="drive-exclude-input"
						/>
						<button type="submit" className="drive-exclude-add-btn">Add</button>
					</form>
				</div>
			</div>
		</div>
	);
}

// ── Drive Settings Tab (with sub-tabs) ───────────────────────────────────────

function DriveSettingsTab({ agents, onAgentsChange }) {
	const [subTab, setSubTab] = useState("drives");

	return (
		<div className="files-panel">
			<div className="settings-subtabs">
				<button
					className={`settings-subtab-btn${subTab === "drives" ? " active" : ""}`}
					onClick={() => setSubTab("drives")}
				>
					<i className="fa-solid fa-hard-drive" /> Drive Settings
				</button>
				<button
					className={`settings-subtab-btn${subTab === "schedule" ? " active" : ""}`}
					onClick={() => setSubTab("schedule")}
				>
					<i className="fa-solid fa-clock" /> Scan Schedule
				</button>
			</div>

			{subTab === "drives" && (
				<DriveSettingsContent agents={agents} onAgentsChange={onAgentsChange} />
			)}
			{subTab === "schedule" && (
				<ScanScheduleContent agents={agents} onAgentsChange={onAgentsChange} />
			)}
		</div>
	);
}

// ── Main Page ─────────────────────────────────────────────────────────────────

const Files = () => {
	const [agents, setAgents] = useState([]);
	const [tab, setTab] = useState("files");

	const fetchAgents = () => {
		axios.get(`${BASE_URL}/protected/agents`, { withCredentials: true })
			.then(r => setAgents(r.data?.agents ?? []))
			.catch(() => {});
	};

	useEffect(() => { fetchAgents(); }, []);

	return (
		<>
			<Navbar title={Title} logo={Logo} menuItems={MenuItems["app"]} />
			<div className="dashboard-layout">
				<main className="dash-main">
					<div className="dash-header">
						<div>
							<h1>Files</h1>
							<p>Browse and manage files indexed across all agents</p>
						</div>
					</div>

					<div className="files-tabs">
						<button
							className={`files-tab-btn${tab === "files" ? " active" : ""}`}
							onClick={() => setTab("files")}
						>
							<i className="fa-solid fa-list" /> All Files
						</button>
						<button
							className={`files-tab-btn${tab === "dupes" ? " active" : ""}`}
							onClick={() => setTab("dupes")}
						>
							<i className="fa-solid fa-clone" /> Duplicates
						</button>
						<button
							className={`files-tab-btn${tab === "settings" ? " active" : ""}`}
							onClick={() => setTab("settings")}
						>
							<i className="fa-solid fa-gear" /> Settings
						</button>
					</div>

					{tab === "files" && <AllFilesTab agents={agents} />}
					{tab === "dupes" && <DuplicatesTab agents={agents} />}
					{tab === "settings" && <DriveSettingsTab agents={agents} onAgentsChange={fetchAgents} />}
				</main>
			</div>
		</>
	);
};

export default Files;
