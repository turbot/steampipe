import Dashboard from "./components/dashboards/layout/Dashboard";
import DashboardErrorModal from "./components/dashboards/DashboardErrorModal";
import DashboardHeader from "./components/DashboardHeader";
import DashboardList from "./components/DashboardList";
import useAnalytics from "./hooks/useAnalytics";
import { DashboardProvider } from "./hooks/useDashboard";
import { Route, Routes } from "react-router-dom";
import { useBreakpoint } from "./hooks/useBreakpoint";
import { useTheme } from "./hooks/useTheme";

const Dashboards = ({
  analyticsContext,
  breakpointContext,
  header,
  themeContext,
}) => (
  <DashboardProvider
    analyticsContext={analyticsContext}
    breakpointContext={breakpointContext}
    themeContext={themeContext}
  >
    {header}
    <DashboardErrorModal />
    <DashboardList />
    <Dashboard />
  </DashboardProvider>
);

const DashboardApp = ({
  analyticsContext,
  breakpointContext,
  header,
  themeContext,
}) => {
  const dashboards = (
    <Dashboards
      analyticsContext={analyticsContext}
      breakpointContext={breakpointContext}
      header={header}
      themeContext={themeContext}
    />
  );

  return (
    <Routes>
      <Route path="/" element={dashboards} />
      <Route path="/:dashboardName" element={dashboards} />
    </Routes>
  );
};

const App = () => {
  const analyticsContext = useAnalytics();
  const breakpointContext = useBreakpoint();
  const themeContext = useTheme();

  return (
    <DashboardApp
      analyticsContext={analyticsContext}
      breakpointContext={breakpointContext}
      header={<DashboardHeader />}
      themeContext={themeContext}
    />
  );
};

export default App;
