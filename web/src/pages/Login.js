import React, { useState, useEffect } from "react";
import { Link } from "react-router-dom";
import axios from "axios";
import { useGoogleLogin } from "@react-oauth/google";
import "../styles/Login.css";
import IMG_PROFILE from "../images/profile.png";

import { BASE_URL, Title, Logo } from "../constants";

const GoogleIcon = () => (
	<svg width="18" height="18" viewBox="0 0 18 18">
		<path fill="#4285F4" d="M17.64 9.2c0-.637-.057-1.251-.164-1.84H9v3.481h4.844a4.14 4.14 0 01-1.796 2.716v2.259h2.908c1.702-1.567 2.684-3.875 2.684-6.615z"/>
		<path fill="#34A853" d="M9 18c2.43 0 4.467-.806 5.956-2.184l-2.908-2.259c-.806.54-1.837.86-3.048.86-2.344 0-4.328-1.584-5.036-3.711H.957v2.332A8.997 8.997 0 009 18z"/>
		<path fill="#FBBC05" d="M3.964 10.706A5.41 5.41 0 013.682 9c0-.593.102-1.17.282-1.706V4.962H.957A8.996 8.996 0 000 9c0 1.452.348 2.827.957 4.038l3.007-2.332z"/>
		<path fill="#EA4335" d="M9 3.58c1.321 0 2.508.454 3.44 1.345l2.582-2.58C13.463.891 11.426 0 9 0A8.997 8.997 0 00.957 4.962L3.964 6.294C4.672 4.167 6.656 3.58 9 3.58z"/>
	</svg>
);

const GuestIcon = () => (
	<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
		<path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
		<circle cx="12" cy="7" r="4"/>
	</svg>
);

const googleOAuthEnabled =
	!!process.env.REACT_APP_GOOGLE_CLIENT_ID &&
	process.env.REACT_APP_GOOGLE_CLIENT_ID !== "google-oauth-not-configured";

function AuthPage() {
	const [profile, setProfile] = useState(null);
	const [copied, setCopied] = useState(false);
	const [guestLoading, setGuestLoading] = useState(false);

	const login = useGoogleLogin({
		flow: "auth-code",
		onSuccess: async (codeResponse) => {
			try {
				const response = await axios.post(
					`${BASE_URL}/auth/google`,
					{ code: codeResponse.code },
					{ withCredentials: true }
				);
				if (response.data.message === "Logged in") {
					setProfile(response.data.profile);
				}
			} catch {
				// auth error — user stays on page
			}
		},
		onError: () => {},
	});

	const loginAsGuest = async () => {
		setGuestLoading(true);
		try {
			const response = await axios.post(
				`${BASE_URL}/auth/guest`,
				{},
				{ withCredentials: true }
			);
			if (response.data.profile) {
				setProfile(response.data.profile);
			}
		} catch {
			// guest creation failed
		} finally {
			setGuestLoading(false);
		}
	};

	const copyUserCode = () => {
		if (!profile?.id) return;
		navigator.clipboard.writeText(profile.id).then(() => {
			setCopied(true);
			setTimeout(() => setCopied(false), 2000);
		});
	};

	const fetchProfile = async () => {
		try {
			const response = await axios.get(`${BASE_URL}/protected/profile`, {
				withCredentials: true,
			});
			setProfile(response.data.profile ?? response.data);
		} catch {
			setProfile(null);
		}
	};

	const logout = () => {
		fetch(`${BASE_URL}/protected/logout`, {
			method: "POST",
			credentials: "include",
		})
			.then((r) => r.json())
			.then(() => setProfile(null))
			.catch(() => {});
	};

	useEffect(() => {
		fetchProfile();
	}, []);

	return (
		<div className="auth-page">
			<aside className="auth-brand">
				<div className="auth-brand-logo">
					{Title}
					<i className={Logo} />
				</div>
				<h1>Optimize your files smarter</h1>
				<p>
					Scan, analyze, and manage files across all your devices from one
					unified dashboard.
				</p>
				<div className="auth-brand-features">
					<div className="auth-feature-item">
						<i className="fa-solid fa-magnifying-glass" />
						<span>Deep file scanning &amp; deduplication</span>
					</div>
					<div className="auth-feature-item">
						<i className="fa-solid fa-network-wired" />
						<span>Real-time network monitoring</span>
					</div>
					<div className="auth-feature-item">
						<i className="fa-solid fa-shield-halved" />
						<span>Secure — your data stays yours</span>
					</div>
				</div>
			</aside>

			<main className="auth-panel">
				<div className="auth-card">
					{profile ? (
						<>
							<div className="auth-card-header">
								<h2>Welcome{profile.is_guest ? ", Guest" : " back"}</h2>
								<p>
									{profile.is_guest
										? "Your temporary session is active."
										: `Signed in as ${profile.email}`}
								</p>
							</div>
							<div className="profile-logged">
								{!profile.is_guest && (
									<img
										src={profile.picture || IMG_PROFILE}
										alt="profile"
										className="profile-avathar"
									/>
								)}
								<div className="profile-name">{profile.name || "Guest"}</div>
								{!profile.is_guest && (
									<div className="profile-email">{profile.email}</div>
								)}

								{profile.is_guest && (
									<div className="guest-code-box">
										<div className="guest-code-label">
											<i className="fa-solid fa-key" /> Your User Code
										</div>
										<div className="guest-code-value">{profile.id}</div>
										<button className="guest-code-copy" onClick={copyUserCode}>
											{copied ? (
												<><i className="fa-solid fa-check" /> Copied!</>
											) : (
												<><i className="fa-regular fa-copy" /> Copy</>
											)}
										</button>
										{profile.expires_at && (
											<div className="guest-code-expiry">
												<i className="fa-regular fa-clock" />{" "}
												Session expires{" "}
												{new Date(profile.expires_at).toLocaleString()}
											</div>
										)}
										<p className="guest-code-hint">
											Paste this code when registering the desktop agent to link
											your devices.
										</p>
									</div>
								)}

								<div className="profile-actions">
									<Link
										to="/dashboard"
										className="btn-full btn-accent"
										style={{ textDecoration: "none", textAlign: "center" }}
									>
										Go to Dashboard
									</Link>
									<button className="btn-full btn-ghost" onClick={logout}>
										Sign out
									</button>
								</div>
							</div>
						</>
					) : (
						<>
							<div className="auth-card-header">
								<h2>Welcome</h2>
								<p>Sign in to access your dashboard.</p>
							</div>
							<button
								className="google-btn"
								onClick={googleOAuthEnabled ? login : undefined}
								disabled={!googleOAuthEnabled}
								title={
									googleOAuthEnabled
										? undefined
										: "Set REACT_APP_GOOGLE_CLIENT_ID in web/.env to enable Google login"
								}
							>
								<GoogleIcon />
								{googleOAuthEnabled
									? "Continue with Google"
									: "Google login not configured"}
							</button>
							<div className="auth-divider">
								<span>or</span>
							</div>
							<button
								className="guest-btn"
								onClick={loginAsGuest}
								disabled={guestLoading}
							>
								<GuestIcon />
								{guestLoading ? "Creating session…" : "Continue as Guest"}
							</button>
							<div className="auth-terms">
								Guest sessions expire after 24 hours. Sign in with Google for
								persistent access.
							</div>
						</>
					)}
				</div>
			</main>
		</div>
	);
}

export default AuthPage;
