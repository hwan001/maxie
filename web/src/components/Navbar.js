import React, { useState, useEffect, useRef, useCallback } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Button } from "./Button";
import axios from "axios";

import "../styles/Navbar.css";
import "../styles/DropdownItem.css";

import DropdownItem from "../components/DropdownItem.js";
import { BASE_URL } from "../constants";
import IMG_PROFILE from "../images/profile.png";

function ProfileDropdown({ profile, onLogout, onClose }) {
	const [copied, setCopied] = useState(false);
	const ref = useRef(null);

	useEffect(() => {
		const handler = (e) => {
			if (ref.current && !ref.current.contains(e.target)) onClose();
		};
		document.addEventListener("mousedown", handler);
		return () => document.removeEventListener("mousedown", handler);
	}, [onClose]);

	const copyCode = () => {
		if (!profile?.id) return;
		navigator.clipboard.writeText(profile.id).then(() => {
			setCopied(true);
			setTimeout(() => setCopied(false), 2000);
		});
	};

	return (
		<div className="profile-dropdown" ref={ref}>
			<div className="profile-dropdown-header">
				{profile.is_guest ? (
					<div className="profile-dropdown-avatar guest-avatar">
						<i className="fa-solid fa-user" />
					</div>
				) : (
					<img
						src={profile.picture || IMG_PROFILE}
						alt="avatar"
						className="profile-dropdown-avatar"
					/>
				)}
				<div className="profile-dropdown-info">
					<span className="profile-dropdown-name">
						{profile.name || "Guest"}
					</span>
					{profile.is_guest ? (
						<span className="profile-dropdown-sub">Guest session</span>
					) : (
						<span className="profile-dropdown-sub">{profile.email}</span>
					)}
				</div>
			</div>

			{profile.is_guest && (
				<div className="profile-dropdown-code">
					<div className="profile-dropdown-code-label">
						<i className="fa-solid fa-key" /> User Code
					</div>
					<div className="profile-dropdown-code-row">
						<span className="profile-dropdown-code-value">{profile.id}</span>
						<button className="profile-dropdown-code-copy" onClick={copyCode}>
							{copied ? (
								<i className="fa-solid fa-check" />
							) : (
								<i className="fa-regular fa-copy" />
							)}
						</button>
					</div>
					{profile.expires_at && (
						<div className="profile-dropdown-expiry">
							<i className="fa-regular fa-clock" />{" "}
							Expires {new Date(profile.expires_at).toLocaleString()}
						</div>
					)}
					<p className="profile-dropdown-hint">
						Paste this code when registering the desktop agent.
					</p>
				</div>
			)}

			<div className="profile-dropdown-actions">
				<Link
					to="/dashboard"
					className="profile-dropdown-btn profile-dropdown-btn--accent"
					onClick={onClose}
				>
					<i className="fa-solid fa-table-columns" /> Dashboard
				</Link>
				<button
					className="profile-dropdown-btn profile-dropdown-btn--ghost"
					onClick={onLogout}
				>
					<i className="fa-solid fa-right-from-bracket" /> Sign out
				</button>
			</div>
		</div>
	);
}

function Navbar({ title, logo, menuItems }) {
	const [click, setClick] = useState(false);
	const [profile, setProfile] = useState(null);
	const [dropdownOpen, setDropdownOpen] = useState(false);
	const navigate = useNavigate();

	const fetchProfile = useCallback(async () => {
		try {
			const res = await axios.get(`${BASE_URL}/protected/profile`, {
				withCredentials: true,
			});
			setProfile(res.data.profile ?? res.data);
		} catch {
			setProfile(null);
		}
	}, []);

	useEffect(() => {
		fetchProfile();
	}, [fetchProfile]);

	const handleLogout = async () => {
		try {
			await fetch(`${BASE_URL}/protected/logout`, {
				method: "POST",
				credentials: "include",
			});
		} catch {}
		setProfile(null);
		setDropdownOpen(false);
		navigate("/login");
	};

	const closeMobileMenu = () => setClick(false);

	return (
		<>
			<nav className="navbar">
				<div className="navbar-container">
					<Link to="/" className="navbar-logo" onClick={closeMobileMenu}>
						{title || "LOGO"}
						<i className={logo} />
					</Link>

					<div className="menu-icon" onClick={() => setClick(!click)}>
						<i className={click ? "fas fa-times" : "fas fa-bars"} />
					</div>

					<ul className={click ? "nav-menu active" : "nav-menu"}>
						{menuItems.map((item, index) => (
							<li className="nav-item" key={index}>
								<Link
									to={item.path}
									className="nav-links"
									onClick={closeMobileMenu}
								>
									{item.label}
									{item.i ? <i className={`${item.i}`} /> : <></>}
								</Link>
								{item.submenu && (
									<div className="dropdown-content">
										{item.submenu.map((sub, si) => (
											<DropdownItem
												key={si}
												icon={sub.i}
												label={sub.label}
												path={sub.path}
											/>
										))}
									</div>
								)}
							</li>
						))}
					</ul>

					<div className="nav-btns">
						{profile ? (
							<div className="nav-profile-wrap">
								<button
									className="nav-profile-btn"
									onClick={() => setDropdownOpen((v) => !v)}
								>
									{profile.is_guest ? (
										<div className="nav-profile-avatar nav-profile-avatar--guest">
											<i className="fa-solid fa-user" />
										</div>
									) : (
										<img
											src={profile.picture || IMG_PROFILE}
											alt="avatar"
											className="nav-profile-avatar"
										/>
									)}
									<span className="nav-profile-name">
										{profile.name || "Guest"}
									</span>
									<i
										className={`fa-solid fa-chevron-down nav-profile-chevron${dropdownOpen ? " open" : ""}`}
									/>
								</button>

								{dropdownOpen && (
									<ProfileDropdown
										profile={profile}
										onLogout={handleLogout}
										onClose={() => setDropdownOpen(false)}
									/>
								)}
							</div>
						) : (
							<Button linkTo="/login" buttonStyle="btn--outline">
								Log in
							</Button>
						)}
					</div>
				</div>
			</nav>
		</>
	);
}

export default Navbar;
