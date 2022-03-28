import App from "./App";
import ErrorBoundary from "./components/ErrorBoundary";
import React from "react";
import ReactDOM from "react-dom";
import { AnalyticsProvider } from "./hooks/useAnalytics";
import { BreakpointProvider } from "./hooks/useBreakpoint";
import { BrowserRouter as Router } from "react-router-dom";
import { ThemeProvider } from "./hooks/useTheme";
import "./styles/index.css";

ReactDOM.render(
  <React.StrictMode>
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
  </React.StrictMode>,
  document.getElementById("root")
);
