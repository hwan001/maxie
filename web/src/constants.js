export const DEBUG=true;
export const BASE_URL='https://goserver.666lab.org/api';
//export const BASE_URL='http://localhost:8080/api';

export const Title="Fillite";
export const Logo="fab fa-typo3";

export const MenuItems = {
    "dashboard":[ 
        { i:"fa-solid fa-house", label: "Home", path:"/"},
    ],
    "home":[
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
    ],
    "login":[],
};