import Children from "../common/Children";
import LayoutPanel from "../common/LayoutPanel";
import { ContainerDefinition } from "../../../../hooks/useDashboard";

interface ContainerProps {
  definition: ContainerDefinition;
}

const Container = ({ definition }: ContainerProps) => {
  return (
    <LayoutPanel definition={definition}>
      <Children children={definition.children} />
    </LayoutPanel>
  );
};

export default Container;
