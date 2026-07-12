import React, { useState, useEffect } from "react";
import axios from "axios";
import { BASE_URL } from "../constants";
import { errorMessage } from "../lib/http";

function fmtDate(iso) {
	if (!iso) return "—";
	const d = new Date(iso);
	if (isNaN(d) || d.getFullYear() < 2000) return "—";
	return d.toLocaleDateString();
}

// Read-only list of all accounts loaded from the server. Role management is
// intentionally out of scope here — this is just the listing.
function UsersPanel() {
	const [users, setUsers] = useState(null);
	const [error, setError] = useState("");

	useEffect(() => {
		axios.get(`${BASE_URL}/admin/users`, { withCredentials: true })
			.then((r) => { setUsers(r.data?.users ?? []); setError(""); })
			.catch((err) => setError(errorMessage(err, "Could not load users.")));
	}, []);

	if (error) {
		return (
			<section className="admin-card admin-status-msg" role="alert">
				<i className="fa-solid fa-triangle-exclamation" />
				<span>{error}</span>
			</section>
		);
	}
	if (!users) {
		return <div className="loading-row"><span className="spinner" /> Loading users…</div>;
	}

	return (
		<section className="admin-card">
			<div className="admin-card-header">
				<i className="fa-solid fa-users-gear" />
				<div>
					<h2>Users &amp; Roles</h2>
					<p>{users.length} {users.length === 1 ? "account" : "accounts"}</p>
				</div>
			</div>

			{users.length === 0 ? (
				<div className="admin-empty">No users yet.</div>
			) : (
				<div className="admin-table-wrap">
					<table className="data-table admin-users-table">
						<thead>
							<tr>
								<th>User</th>
								<th>Type</th>
								<th>Devices</th>
								<th>Created</th>
							</tr>
						</thead>
						<tbody>
							{users.map((u) => (
								<tr key={u.user_id}>
									<td>
										<div className="admin-user-cell">
											<span className="admin-user-name">{u.name || u.user_id}</span>
											<span
												className={`admin-user-sub${u.email ? "" : " topo-user-id"}`}
												title={u.user_id}
											>
												{u.email || u.user_id}
											</span>
										</div>
									</td>
									<td>
										<span className={`admin-badge${u.is_guest ? "" : " admin-badge--ok"}`}>
											{u.is_guest ? "Guest" : "User"}
										</span>
									</td>
									<td>{u.device_count}</td>
									<td>{fmtDate(u.created_at)}</td>
								</tr>
							))}
						</tbody>
					</table>
				</div>
			)}
		</section>
	);
}

export default UsersPanel;
