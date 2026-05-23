import React, { useState, useEffect } from "react";
import { Link } from "react-router-dom";
import { Button } from "./Button";

import Cookies from "js-cookie";
import "../styles/Navbar.css";
import "../styles/DropdownItem.css";

import DropdownItem from "../components/DropdownItem.js";

function Navbar({ title, logo, menuItems }) {
	const [click, setClick] = useState(false);
	const handleClick = () => setClick(!click);
	const closeMobileMenu = () => setClick(false);

	const [session, setSession] = useState(false);
	const getJwt = () => {
		/* 백엔드 미들웨어로 인증 받아서 처리 */
		const jwt = Cookies.get("jwt");
		if (jwt) {
			console.log(jwt);
			setSession(true);
		} else {
			setSession(false);
		}
	};
	useEffect(() => {
		getJwt();
	}, []);

	return (
		<>
			<nav className="navbar">
				<div className="navbar-container">
					<Link to="/" className="navbar-logo" onClick={closeMobileMenu}>
						{title || "LOGO"}
						<i className={logo} />
					</Link>

					<div className="menu-icon" onClick={handleClick}>
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
								{item.submenu ? (
									<div className="dropdown-content">
										{item.submenu.map((submenuItem, subIndex) => (
											<DropdownItem
												key={subIndex}
												icon={submenuItem.i}
												label={submenuItem.label}
												path={submenuItem.path}
											/>
										))}
									</div>
								) : (
									<></>
								)}
							</li>
						))}
					</ul>

					<div className="nav-btns">
						{!session && (
							<Button linkTo="/login" buttonStyle="btn--outline">
								Log in
							</Button>
						)}
						{!session && (
							<Button linkTo="/signup" buttonStyle="btn--outline">
								Sign up
							</Button>
						)}
					</div>
				</div>
			</nav>
		</>
	);
}

export default Navbar;
