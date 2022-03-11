import Dashboard from "./components/dashboards/layout/Dashboard";
import DashboardErrorModal from "./components/dashboards/DashboardErrorModal";
import DashboardHeader from "./components/DashboardHeader";
import DashboardList from "./components/DashboardList";
import { AnalyticsProvider } from "./hooks/useAnalytics";
import { BreakpointProvider } from "./hooks/useBreakpoint";
import { DashboardProvider } from "./hooks/useDashboard";
import { Route, Routes } from "react-router-dom";

const DashboardApp = () => (
  <DashboardProvider>
    <AnalyticsProvider>
      <DashboardHeader />
      <DashboardErrorModal />
      <DashboardList />
      <Dashboard />
    </AnalyticsProvider>
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
