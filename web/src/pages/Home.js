import React from "react";
import "../styles/Home.css";

import Navbar from "../components/Navbar.js";
import Footer from "../components/Footer.js";

import {Title, Logo, MenuItems} from '../constants';

function Home() {
  return (
    <>
      <Navbar title={Title} logo={Logo} menuItems={MenuItems['home']} />

      <section id="home" class="section">
        <h1>Welcome to Our Site</h1>
        <p>Introduction content goes here...</p>
      </section>

      <section id="features" class="section">
        <h2>Features</h2>
        <p>Describe your features here...</p>
        <button class="cta-button">Learn More</button>
      </section>

      <Footer/>
    </>
  );
}

export default Home;