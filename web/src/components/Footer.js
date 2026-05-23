import React from "react";
import { Link } from "react-router-dom";

import { Title, Logo } from "../constants.js";
import "../styles/Footer.css";

function Footer() {
	return (
		<footer className="footer-container">
			<Link to="/" className="footer-logo-link">
				{Title}
				<i className={Logo}></i>
			</Link>
			<small className="footer-rights">
				{Title} &copy; {new Date().getFullYear()}
			</small>
		</footer>
	);
}

export default Footer;
