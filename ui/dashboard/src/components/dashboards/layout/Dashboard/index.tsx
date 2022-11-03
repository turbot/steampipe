import Children from "../Children";
import DashboardProgress from "./DashboardProgress";
import Grid from "../Grid";
import PanelDetail from "../PanelDetail";
import SnapshotRenderComplete from "../../../snapshot/SnapshotRenderComplete";
import { classNames } from "../../../../utils/styles";
import { DashboardDataModeLive, DashboardDefinition } from "../../../../types";
import { registerComponent } from "../../index";
import { useDashboard } from "../../../../hooks/useDashboard";

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
}: DashboardProps) => {
  const title =
    isRoot && definition.title ? (
      <h1 className={classNames("col-span-12")}>{definition.title}</h1>
    ) : null;
  const grid = (
    <Grid name={definition.name} width={isRoot ? 12 : definition.width}>
      {title}
      <Children
        allowPanelExpand={allowPanelExpand}
        children={definition.children}
      />
    </Grid>
  );
  return (
    <>
      {isRoot ? <DashboardProgress /> : <></>}
      {isRoot ? <div className="h-full overflow-y-auto p-4">{grid}</div> : grid}
    </>
  );
};

const DashboardWrapper = ({
  allowPanelExpand = true,
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
        allowPanelExpand={allowPanelExpand}
        definition={dashboard}
        withPadding={true}
      />
      <SnapshotRenderComplete />
    </>
  );
};

registerComponent("dashboard", Dashboard);

export default DashboardWrapper;

export { Dashboard };
