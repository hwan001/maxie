import React from "react";

import "../styles/Common.css";
import "../styles/Home.css";

import Navbar from "../components/Navbar.js";

function Home() {
  const homeMenuItems = [
    { i:"fa-solid fa-table-columns", label: "Dashboard", path:"/dashboard"},
    { 
        i: "fa-solid fa-download",
        label: "Downloads",
        submenu: [
          { i: "windows", label: "windows", path: "/api/download/client/windows" },
          { i: "linux", label: "linux", path: "/api/download/client/linux" },
          { i: "apple", label: "mac", path: "/api/download/client/mac" }
        ]
    },
    { i:"fa-solid fa-rss", label: "Blog", path: "https://hwan001.co.kr"},
    { i:"fa-brands fa-github", label: "GitHub", path: "https://github.com/hwan001"},
  ];

  return (
    <>
      <Navbar menuItems={homeMenuItems} />

      <section id="home" class="section">
        <h1>Welcome to Our Site</h1>
        <p>Introduction content goes here...</p>
      </section>

      <section id="features" class="section">
        <h2>Features</h2>
        <p>Describe your features here...</p>
        <button class="cta-button">Learn More</button>
      </section>

      <section id="contact" class="section">
        <h2>Contact Us</h2>
        <form>
            <input type="text" placeholder="Your Name" required/>
            <input type="email" placeholder="Your Email" required/>
            <textarea placeholder="Your Message" required></textarea>
            <button type="submit">Send</button>
        </form>
      </section>

      <footer>
        <p>Footer content here...</p>
      </footer>
    </>
  );
}

export default Home;