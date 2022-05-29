import Children from "../common/Children";
import LayoutPanel from "../common/LayoutPanel";
import PanelDetail from "../PanelDetail";
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
  <>
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
  const { dashboard, search, selectedDashboard, selectedPanel } =
    useDashboard();

  if (search.value || !dashboard || !selectedDashboard) {
    return null;
  }

  if (selectedPanel) {
    return <PanelDetail definition={selectedPanel} />;
  }

  return (
    <Dashboard definition={dashboard} withPadding={!dashboard.artificial} />
  );
};

export default DashboardWrapper;

export { Dashboard };
