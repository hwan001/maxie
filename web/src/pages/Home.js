import React from "react";
import "../styles/Common.css";

import Navbar from "../components/Navbar.js";

function Home() {
  const homeMenuItems = [
    { 
        label: "Downloads", 
        submenu: [
          { img: "https://cdn.prod.website-files.com/63ed4bc7a4b189da942a6b8c/65ce30961d407eff31c44dd4_BookOpenText.svg", label: "Windows", path: "/download/windows" },
          { img: "https://cdn.prod.website-files.com/63ed4bc7a4b189da942a6b8c/65ce30961d407eff31c44dd4_BookOpenText.svg", label: "linux", path: "/download/linux" },
          { img: "https://cdn.prod.website-files.com/63ed4bc7a4b189da942a6b8c/65ce30961d407eff31c44dd4_BookOpenText.svg", label: "mac", path: "/download/mac" }
        ]
    },
    { label: "Blog", path: "https://hwan001.co.kr"}
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