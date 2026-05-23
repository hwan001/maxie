export const DEBUG = true;
// export const BASE_URL = "https://goserver.666lab.org/api";
export const BASE_URL='http://localhost:3000';

export const Title = "File Optimizer";
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
		// { i: "fa-solid fa-rss", label: "Blog", path: "https://hwan001.co.kr" },
		// {
		// 	i: "fa-brands fa-github",
		// 	label: "GitHub",
		// 	path: "https://github.com/hwan001",
		// },
	],
	login: [],
};
