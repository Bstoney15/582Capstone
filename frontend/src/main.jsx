// main.jsx – application entry point; mounts the React root into the DOM.
// Author: Ben Stonestreet, Connor Williamson
// Created: 2026-02-02

import { createRoot } from "react-dom/client";
import "./index.css";
import App from "./App.jsx";

// Mount the React application into the #root element defined in index.html.
createRoot(document.getElementById("root")).render(<App />);
