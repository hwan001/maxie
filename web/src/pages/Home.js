import React from "react";
import "../styles/Home.css";

import Navbar from "../components/Navbar.js";
import Footer from "../components/Footer.js";

import { Title, Logo, MenuItems } from "../constants";

function Home() {
	return (
		<>
			<Navbar title={Title} logo={Logo} menuItems={MenuItems["home"]} />

			<section id="home" className="section">
				<h1>Welcome to File Optimizer</h1>
				<p>Scan, analyze, and optimize your files with ease.</p>
				<button className="cta-button">Get Started</button>
			</section>

			<section id="features" className="section">
				<h2>Smart Scanning</h2>
				<p>Automatically detect duplicate and redundant files across your system.</p>
				<button className="cta-button">Learn More</button>
			</section>

			<section id="features2" className="section">
				<h2>Detailed Insights</h2>
				<p>View network activity, system info, and file distribution at a glance.</p>
				<button className="cta-button">Learn More</button>
			</section>

			<Footer />
		</>
	);
}

export default Home;
