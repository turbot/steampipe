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
  <LayoutPanel definition={definition} withPadding={withPadding}>
    {
      <>
        {definition.title ? (
          <Children
            children={[
              {
                name: `${definition.name}.root.text.title`,
                node_type: "text",
                properties: {
                  type: "markdown",
                  value: `# ${definition.title}`,
                },
              },
            ]}
          />
        ) : null}
      </>
    }
    <Children children={definition.children} />
  </LayoutPanel>
);

const DashboardWrapper = () => {
  const { dashboard, selectedDashboard, selectedPanel } = useDashboard();

  if (!dashboard || !selectedDashboard) {
    return null;
  }

  if (selectedPanel) {
    return <PanelDetail definition={selectedPanel} />;
  }

  return <Dashboard definition={dashboard} withPadding={true} />;
};

export default DashboardWrapper;

export { Dashboard };
