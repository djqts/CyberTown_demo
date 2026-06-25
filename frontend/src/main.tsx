import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import TownApp from "./app/TownApp";
import "./styles/globals.css";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <TownApp />
  </StrictMode>
);
