import React from "react";
import ReactDOM from "react-dom/client";
import "./index.css";
import App from "./App";
import reportWebVitals from "./reportWebVitals";

import { GoogleOAuthProvider } from "@react-oauth/google";

// Fall back to a non-empty placeholder so GoogleOAuthProvider does not throw
// "Missing required parameter client_id" at mount time when the env var is
// absent. The Google button in Login.js is disabled whenever this placeholder
// is in use, so no real OAuth call is ever attempted without a real ID.
const googleClientId =
	process.env.REACT_APP_GOOGLE_CLIENT_ID || "google-oauth-not-configured";

const root = ReactDOM.createRoot(document.getElementById("root"));
root.render(
	<GoogleOAuthProvider clientId={googleClientId}>
		<React.StrictMode>
			<App />
		</React.StrictMode>
	</GoogleOAuthProvider>
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
