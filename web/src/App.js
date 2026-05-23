import React from "react";
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";

import Home from "./pages/Home";
import AuthPage from "./pages/Login";
import Dashboard from "./pages/Dashboard";
import Devices from "./pages/Devices";
import Files from "./pages/Files";
import Optimization from "./pages/Optimization";

function App() {
	return (
		<Router>
			<Routes>
				<Route path="/" element={<Home />} />
				<Route path="/login" element={<AuthPage />} />
				<Route path="/signup" element={<AuthPage />} />
				<Route path="/dashboard" element={<Dashboard />} />
				<Route path="/devices" element={<Devices />} />
				<Route path="/files" element={<Files />} />
				<Route path="/optimization" element={<Optimization />} />
			</Routes>
		</Router>
	);
}

export default App;
