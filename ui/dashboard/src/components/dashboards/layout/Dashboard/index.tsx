import Children from "../Children";
import DashboardProgress from "./DashboardProgress";
import DashboardTitle from "../../titles/DashboardTitle";
import Grid from "../Grid";
import PanelDetail from "../PanelDetail";
import SnapshotRenderComplete from "../../../snapshot/SnapshotRenderComplete";
import { DashboardDataModeLive, DashboardDefinition } from "../../../../types";
import { registerComponent } from "../../index";
import { useDashboard } from "../../../../hooks/useDashboard";

interface DashboardProps {
  definition: DashboardDefinition;
  isRoot?: boolean;
  showPanelControls?: boolean;
  withPadding?: boolean;
}

interface DashboardWrapperProps {
  showPanelControls?: boolean;
}

// TODO allow full-screen of a panel
const Dashboard = ({
  definition,
  isRoot = true,
  showPanelControls = true,
}: DashboardProps) => {
  const grid = (
    <Grid name={definition.name} width={isRoot ? 12 : definition.width}>
      {isRoot && <DashboardTitle title={definition.title} />}
      <Children
        showPanelControls={showPanelControls}
        children={definition.children}
      />
    </Grid>
  );
  return (
    <>
      {isRoot ? <DashboardProgress /> : null}
      {isRoot ? <div className="h-full overflow-y-auto p-4">{grid}</div> : grid}
    </>
  );
};

const DashboardWrapper = ({
  showPanelControls = true,
}: DashboardWrapperProps) => {
  const { dashboard, dataMode, search, selectedDashboard, selectedPanel } =
    useDashboard();

  if (
    search.value ||
    !dashboard ||
    (!selectedDashboard && dataMode === DashboardDataModeLive)
  ) {
    return null;
  }

  if (selectedPanel) {
    return <PanelDetail definition={selectedPanel} />;
  }

  return (
    <>
      <Dashboard
        definition={dashboard}
        showPanelControls={showPanelControls}
        withPadding={true}
      />
      <SnapshotRenderComplete />
    </>
  );
};

registerComponent("dashboard", Dashboard);

export default DashboardWrapper;

export { Dashboard };
