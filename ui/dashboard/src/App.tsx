import Dashboard from "./components/dashboards/layout/Dashboard";
import DashboardErrorModal from "./components/dashboards/DashboardErrorModal";
import DashboardHeader from "./components/DashboardHeader";
import DashboardList from "./components/DashboardList";
import { BreakpointProvider } from "./hooks/useBreakpoint";
import { DashboardProvider } from "./hooks/useDashboard";
import { Route, Routes } from "react-router-dom";

const DashboardApp = () => (
  <DashboardProvider>
    <DashboardHeader />
    <DashboardErrorModal />
    <DashboardList />
    <Dashboard />
  </DashboardProvider>
);

const App = () => (
  <BreakpointProvider>
    <Routes>
      <Route path="/" element={<DashboardApp />} />
      <Route path="/:dashboardName" element={<DashboardApp />} />
    </Routes>
  </BreakpointProvider>
);

export default App;

/**
 * steampipe dashboard <command>
 *
 * 1) Socket and Static file server started - socket server is aware of context/workspace it's launched in
 *    e.g. what workspace and/or specific report path are we interested in?
 * 2) Browser opened at Static file server endpoint e.g. http://localhost:3000
 * 3) Browser connects to the socket server
 * 4) Socket server broadcasts its context to clients on connect
 * 5) Browser will display its current context
 * 6) Socket server will broadcast updates to the connected client(s)
 *    TODO handle multiple clients with diff contexts?
 * 7) Browser will handle updates and handle state of report and its panels
 *
 * **/
