import Children from "../common/Children";
import DashboardProgress from "./DashboardProgress";
import LayoutPanel from "../common/LayoutPanel";
import PanelDetail from "../PanelDetail";
import {
  DashboardDefinition,
  DashboardRunState,
  useDashboard,
} from "../../../../hooks/useDashboard";
import { registerComponent } from "../../index";

interface DashboardProps {
  allowPanelExpand?: boolean;
  definition: DashboardDefinition;
  isRoot?: boolean;
  progress?: number;
  state?: DashboardRunState;
  withPadding?: boolean;
}

interface DashboardWrapperProps {
  allowPanelExpand?: boolean;
}

// TODO allow full-screen of a panel
const Dashboard = ({
  allowPanelExpand = true,
  definition,
  progress = 0,
  isRoot = true,
  state = "ready",
  withPadding = false,
}: DashboardProps) => (
  <>
    {isRoot ? <DashboardProgress state={state} progress={progress} /> : <></>}
    <LayoutPanel
      className={isRoot ? "h-full overflow-y-auto" : undefined}
      definition={definition}
      isDashboard={true}
      withPadding={withPadding}
    >
      <Children
        allowPanelExpand={allowPanelExpand}
        children={definition.children}
      />
    </LayoutPanel>
  </>
);

const DashboardWrapper = ({
  allowPanelExpand = true,
}: DashboardWrapperProps) => {
  const {
    dashboard,
    dataMode,
    progress,
    search,
    selectedDashboard,
    selectedPanel,
    state,
  } = useDashboard();

  if (
    search.value ||
    !dashboard ||
    (!selectedDashboard && dataMode === "live")
  ) {
    return null;
  }

  if (selectedPanel) {
    return <PanelDetail definition={selectedPanel} />;
  }

  return (
    <Dashboard
      allowPanelExpand={allowPanelExpand}
      definition={dashboard}
      progress={progress}
      withPadding={true}
      state={state}
    />
  );
};

registerComponent("dashboard", Dashboard);

export default DashboardWrapper;

export { Dashboard };
