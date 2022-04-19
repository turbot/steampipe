import Children from "../common/Children";
import LayoutPanel from "../common/LayoutPanel";
import { ContainerDefinition } from "../../../../hooks/useDashboard";

interface ContainerProps {
  allowChildPanelExpand?: boolean;
  definition: ContainerDefinition;
  withNarrowVertical?: boolean;
}

const Container = ({
  allowChildPanelExpand = true,
  definition,
  withNarrowVertical,
}: ContainerProps) => {
  return (
    <LayoutPanel
      definition={definition}
      withNarrowVertical={withNarrowVertical}
    >
      <Children
        allowPanelExpand={allowChildPanelExpand}
        children={definition.children}
      />
    </LayoutPanel>
  );
};

export default Container;
