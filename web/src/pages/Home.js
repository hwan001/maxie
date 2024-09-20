import React from "react";
import "../styles/Common.css";

import Navbar from "../components/Navbar.js";

function Home() {
  const homeMenuItems = [
    { 
        label: "Downloads", 
        submenu: [
          { img: "", label: "Windows", path: "/api/download/client/windows" },
          { img: "", label: "linux", path: "/api/download/client/linux" },
          { img: "", label: "mac", path: "/api/download/client/mac" }
        ]
    },
    { label: "Blog", path: "https://hwan001.co.kr"},
    { label: "GitHub", path: "https://github.com/hwan001"},
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