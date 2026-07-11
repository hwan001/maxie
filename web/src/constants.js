export const DEBUG = process.env.NODE_ENV !== "production";
export const BASE_URL = process.env.REACT_APP_BASE_URL
  ?? (process.env.NODE_ENV === 'development' ? 'http://localhost:50000' : '');
export const AGENT_RELEASE_BASE_URL = process.env.REACT_APP_AGENT_RELEASE_BASE_URL ||
	'https://github.com/hwan001/maxie/releases/latest/download';

export const Title = "Maxie";
export const Logo = "fab fa-typo3";

export const MenuItems = {
	app: [
		{ i: "fa-solid fa-table-columns", label: "Dashboard", path: "/dashboard" },
		{ i: "fa-solid fa-desktop", label: "Devices", path: "/devices" },
		{ i: "fa-solid fa-file-lines", label: "Files", path: "/files" },
		{ i: "fa-solid fa-wand-magic-sparkles", label: "Optimize", path: "/optimization" },
	],
	dashboard: [{ i: "fa-solid fa-house", label: "Home", path: "/" }],
	home: [
		{ i: "fa-solid fa-table-columns", label: "Dashboard", path: "/dashboard" },
	],
	login: [],
};
