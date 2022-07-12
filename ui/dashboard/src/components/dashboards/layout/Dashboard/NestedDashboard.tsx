import Children from "../common/Children";
import LayoutPanel from "../common/LayoutPanel";
import { DashboardDefinition } from "../../../../hooks/useDashboard";

interface NestedDashboardProps {
  definition: DashboardDefinition;
  withPadding?: boolean;
}

// TODO allow full-screen of a panel
const NestedDashboard = ({
  definition,
  withPadding = false,
}: NestedDashboardProps) => (
  <LayoutPanel
    definition={definition}
    isDashboard={true}
    withPadding={withPadding}
  >
    <Children children={definition.children} />
  </LayoutPanel>
);

export default NestedDashboard;
