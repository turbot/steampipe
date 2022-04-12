import App from "./App";
import ErrorBoundary from "./components/ErrorBoundary";
import React from "react";
import { AnalyticsProvider } from "./hooks/useAnalytics";
import { BreakpointProvider } from "./hooks/useBreakpoint";
import { BrowserRouter as Router } from "react-router-dom";
import { createRoot } from "react-dom/client";
import { ThemeProvider } from "./hooks/useTheme";
import "./styles/index.css";

const container = document.getElementById("root");
// @ts-ignore
const root = createRoot(container);

root.render(
  <Router>
    <ThemeProvider>
      <ErrorBoundary>
        <BreakpointProvider>
          <AnalyticsProvider>
            <App />
          </AnalyticsProvider>
        </BreakpointProvider>
      </ErrorBoundary>
    </ThemeProvider>
  </Router>
);
