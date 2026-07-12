import React, { useState, useEffect } from "react";
import axios from "axios";
import { BASE_URL } from "../constants";
import { errorMessage } from "../lib/http";

const OS_ICONS = {
	darwin: "fa-brands fa-apple",
	windows: "fa-brands fa-windows",
	linux: "fa-brands fa-linux",
};

function truncate(s, n) {
	return s && s.length > n ? s.slice(0, n - 1) + "…" : s;
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
					<span
						className={`topo-user-sub${user.email ? "" : " topo-user-id"}`}
						title={user.user_id}
					>
						{user.email || user.user_id}
					</span>
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

// ── Graph view ───────────────────────────────────────────────────────────────

const GRAPH = { rootX: 30, userX: 210, deviceX: 430, viewW: 680, row: 44, top: 28, gap: 16 };

// Lay out a Server → users → devices tree into node/edge coordinates.
function buildGraphLayout(users, unassigned) {
	const groups = users.map((u) => ({ ...u }));
	if (unassigned.length) {
		groups.push({ user_id: "__unassigned", name: "Unassigned", is_guest: false, devices: unassigned });
	}

	const { userX, deviceX, row, top, gap } = GRAPH;
	let y = top;
	const userNodes = [];
	const deviceNodes = [];
	const edges = [];

	groups.forEach((g) => {
		const devs = g.devices ?? [];
		const sub = g.is_guest ? g.user_id : (g.email || g.user_id);
		if (devs.length === 0) {
			userNodes.push({ id: g.user_id, x: userX, y, label: truncate(g.name || g.user_id, 22), guest: g.is_guest, sub });
			y += row + gap;
			return;
		}
		const startY = y;
		devs.forEach((d) => {
			deviceNodes.push({ id: g.user_id + "|" + d.agent_id, x: deviceX, y, label: truncate(d.name || d.agent_id, 24), online: d.online });
			y += row;
		});
		const uy = (startY + (y - row)) / 2;
		userNodes.push({ id: g.user_id, x: userX, y: uy, label: truncate(g.name || g.user_id, 22), guest: g.is_guest, sub });
		devs.forEach((d, i) => edges.push({ x1: userX, y1: uy, x2: deviceX, y2: startY + i * row, kind: "ud" }));
		y += gap;
	});

	const height = Math.max(y, top + row) + 10;
	const rootY = height / 2;
	userNodes.forEach((u) => edges.push({ x1: GRAPH.rootX, y1: rootY, x2: userX, y2: u.y, kind: "ru" }));
	return { userNodes, deviceNodes, edges, height, rootY };
}

function GraphView({ users, unassigned }) {
	const { userNodes, deviceNodes, edges, height, rootY } = buildGraphLayout(users, unassigned);
	return (
		<div className="topo-graph-wrap">
			<svg
				className="topo-graph"
				viewBox={`0 0 ${GRAPH.viewW} ${height}`}
				width="100%"
				height={height}
				role="img"
				aria-label="User to device topology graph"
			>
				{edges.map((e, i) => (
					<line key={i} x1={e.x1} y1={e.y1} x2={e.x2} y2={e.y2} className={`topo-edge topo-edge--${e.kind}`} />
				))}

				<g className="topo-gnode topo-gnode--root">
					<circle cx={GRAPH.rootX} cy={rootY} r="7" />
					<text x={GRAPH.rootX} y={rootY - 13} textAnchor="middle">Server</text>
				</g>

				{userNodes.map((n) => (
					<g key={n.id} className={`topo-gnode topo-gnode--user${n.guest ? " guest" : ""}`}>
						<circle cx={n.x} cy={n.y} r="6" />
						<text x={n.x + 13} y={n.y - 2}>{n.label}</text>
						{n.sub && <text x={n.x + 13} y={n.y + 11} className="topo-gnode-sub">{truncate(n.sub, 30)}</text>}
					</g>
				))}

				{deviceNodes.map((n) => (
					<g key={n.id} className={`topo-gnode topo-gnode--device ${n.online ? "online" : "offline"}`}>
						<circle cx={n.x} cy={n.y} r="5" />
						<text x={n.x + 13} y={n.y + 3}>{n.label}</text>
					</g>
				))}
			</svg>
		</div>
	);
}

// ── Panel ────────────────────────────────────────────────────────────────────

// Topology: live user–device map built from server data, as a list or a graph.
function TopologyPanel() {
	const [data, setData] = useState(null);
	const [error, setError] = useState("");
	const [view, setView] = useState("list");

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
	const isEmpty = users.length === 0 && unassigned.length === 0;

	return (
		<section className="admin-card">
			<div className="admin-card-header">
				<i className="fa-solid fa-sitemap" />
				<div>
					<h2>Topology</h2>
					<p>{users.length} {users.length === 1 ? "user" : "users"} · {totalDevices} {totalDevices === 1 ? "device" : "devices"}</p>
				</div>
				<div className="topo-toggle" role="group" aria-label="Topology view">
					<button
						className={`topo-toggle-btn${view === "list" ? " active" : ""}`}
						onClick={() => setView("list")}
						aria-pressed={view === "list"}
					>
						<i className="fa-solid fa-list-ul" /> List
					</button>
					<button
						className={`topo-toggle-btn${view === "graph" ? " active" : ""}`}
						onClick={() => setView("graph")}
						aria-pressed={view === "graph"}
					>
						<i className="fa-solid fa-diagram-project" /> Graph
					</button>
				</div>
			</div>

			{isEmpty ? (
				<div className="admin-empty">No users or devices yet.</div>
			) : view === "graph" ? (
				<GraphView users={users} unassigned={unassigned} />
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

export default TopologyPanel;
