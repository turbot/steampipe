import Children from "../common/Children";
import DashboardProgress from "./DashboardProgress";
import LayoutPanel from "../common/LayoutPanel";
import PanelDetail from "../PanelDetail";
import {
  DashboardDefinition,
  DashboardRunState,
  useDashboard,
} from "../../../../hooks/useDashboard";

interface DashboardProps {
  definition: DashboardDefinition;
  isRoot?: boolean;
  progress?: number;
  state?: DashboardRunState;
  withPadding?: boolean;
}

// TODO allow full-screen of a panel
const Dashboard = ({
  definition,
  progress = 0,
  state = "running",
  withPadding = false,
}: DashboardProps) => (
  <>
    <DashboardProgress state={state} progress={progress} />
    <LayoutPanel
      definition={definition}
      isDashboard={true}
      withPadding={withPadding}
    >
      <Children children={definition.children} />
    </LayoutPanel>
  </>
);

const DashboardWrapper = () => {
  const {
    dashboard,
    progress,
    search,
    selectedDashboard,
    selectedPanel,
    state,
  } = useDashboard();

  if (search.value || !dashboard || !selectedDashboard) {
    return null;
  }

  if (selectedPanel) {
    return <PanelDetail definition={selectedPanel} />;
  }

  return (
    <Dashboard
      definition={dashboard}
      progress={progress}
      withPadding={true}
      state={state}
    />
  );
};

export default DashboardWrapper;

export { Dashboard };
