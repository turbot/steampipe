import Children from "../common/Children";
import LayoutPanel from "../common/LayoutPanel";
import { ContainerDefinition } from "../../../../hooks/useDashboard";

interface ContainerProps {
  definition: ContainerDefinition;
  withNarrowVertical?: boolean;
}

const Container = ({ definition, withNarrowVertical }: ContainerProps) => {
  return (
    <LayoutPanel
      definition={definition}
      withNarrowVertical={withNarrowVertical}
    >
      <Children children={definition.children} />
    </LayoutPanel>
  );
};

export default Container;
