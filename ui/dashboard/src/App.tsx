import Dashboard from "./components/dashboards/layout/Dashboard";
import DashboardErrorModal from "./components/dashboards/DashboardErrorModal";
import DashboardHeader from "./components/DashboardHeader";
import DashboardList from "./components/DashboardList";
import useAnalytics from "./hooks/useAnalytics";
import { DashboardProvider } from "./hooks/useDashboard";
import { DashboardProviderNew } from "./hooks/refactor/useDashboard";
import { FullHeightThemeWrapper, useTheme } from "./hooks/useTheme";
import { Route, Routes } from "react-router-dom";
import { useBreakpoint } from "./hooks/useBreakpoint";

const Dashboards = ({ analyticsContext, breakpointContext, themeContext }) => (
  <DashboardProviderNew
    analyticsContext={analyticsContext}
    breakpointContext={breakpointContext}
    themeContext={themeContext}
  >
    <DashboardProvider>
      <DashboardHeader />
      <DashboardErrorModal />
      <DashboardList wrapperClassName="p-4" />
      <Dashboard />
    </DashboardProvider>
  </DashboardProviderNew>
);

const DashboardApp = ({
  analyticsContext,
  breakpointContext,
  themeContext,
}) => {
  const dashboards = (
    <Dashboards
      analyticsContext={analyticsContext}
      breakpointContext={breakpointContext}
      themeContext={themeContext}
    />
  );

  return (
    <Routes>
      <Route path="/" element={dashboards} />
      <Route path="/:dashboard_name" element={dashboards} />
    </Routes>
  );
};

const App = () => {
  const analyticsContext = useAnalytics();
  const breakpointContext = useBreakpoint();
  const themeContext = useTheme();

  return (
    <FullHeightThemeWrapper>
      <DashboardApp
        analyticsContext={analyticsContext}
        breakpointContext={breakpointContext}
        themeContext={themeContext}
      />
    </FullHeightThemeWrapper>
  );
};

export default App;

export { DashboardApp };
