import Benchmark from "../../check/Benchmark";
import Children from "../common/Children";
import LayoutPanel from "../common/LayoutPanel";
import Panel from "../Panel";
import PanelDetail from "../PanelDetail";
import React from "react";
import {
  DashboardDefinition,
  useDashboard,
} from "../../../../hooks/useDashboard";

interface DashboardProps {
  definition: DashboardDefinition;
  withPadding?: boolean;
}

// TODO allow full-screen of a panel
const Dashboard = ({ definition, withPadding = false }: DashboardProps) => (
  <LayoutPanel
    definition={definition}
    isDashboard={true}
    withPadding={withPadding}
  >
    <>
      {definition.node_type === "benchmark" && (
        <Panel definition={definition} allowExpand={true} withTitle={false}>
          {/*@ts-ignore*/}
          <Benchmark {...definition} />
        </Panel>
      )}
    </>
    <>
      {definition.node_type === "dashboard" && (
        <Children children={definition.children} />
      )}
    </>
  </LayoutPanel>
);

const DashboardWrapper = () => {
  const { dashboard, search, selectedDashboard, selectedPanel } =
    useDashboard();

  if (search.value || !dashboard || !selectedDashboard) {
    return null;
  }

  if (selectedPanel) {
    return <PanelDetail definition={selectedPanel} />;
  }

  return <Dashboard definition={dashboard} withPadding={true} />;
};

export default DashboardWrapper;

export { Dashboard };
