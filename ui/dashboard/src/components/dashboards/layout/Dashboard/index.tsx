import Children from "../common/Children";
import DashboardProgress from "./DashboardProgress";
import LayoutPanel from "../common/LayoutPanel";
import PanelDetail from "../PanelDetail";
import {
  DashboardDefinition,
  useDashboard,
} from "../../../../hooks/useDashboard";
import { registerComponent } from "../../index";

interface DashboardProps {
  allowPanelExpand?: boolean;
  definition: DashboardDefinition;
  isRoot?: boolean;
  withPadding?: boolean;
}

interface DashboardWrapperProps {
  allowPanelExpand?: boolean;
}

// TODO allow full-screen of a panel
const Dashboard = ({
  allowPanelExpand = true,
  definition,
  isRoot = true,
  withPadding = false,
}: DashboardProps) => (
  <>
    {isRoot ? <DashboardProgress /> : <></>}
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
  const { dashboard, dataMode, search, selectedDashboard, selectedPanel } =
    useDashboard();

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
      withPadding={true}
    />
  );
};

registerComponent("dashboard", Dashboard);

export default DashboardWrapper;

export { Dashboard };
