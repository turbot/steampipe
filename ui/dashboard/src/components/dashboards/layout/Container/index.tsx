import Children from "../Children";
import ContainerTitle from "../../titles/ContainerTitle";
import Grid from "../Grid";
import { ContainerDefinition } from "../../../../types";
import { registerComponent } from "../../index";
import { useDashboard } from "../../../../hooks/useDashboard";

interface ContainerProps {
  layoutDefinition?: ContainerDefinition;
  definition?: ContainerDefinition;
}

const Container = ({ definition, layoutDefinition }: ContainerProps) => {
  const { panelsMap } = useDashboard();

  if (!definition && !layoutDefinition) {
    return null;
  }

  const panelDefinition = definition
    ? definition
    : layoutDefinition && panelsMap[layoutDefinition.name]
    ? panelsMap[layoutDefinition.name]
    : layoutDefinition;

  if (!panelDefinition) {
    return null;
  }

  return (
    <Grid name={panelDefinition.name} width={panelDefinition.width}>
      <ContainerTitle title={panelDefinition.title} />
      <Children
        children={
          definition
            ? definition.children
            : layoutDefinition
            ? layoutDefinition.children
            : []
        }
      />
    </Grid>
  );
};

registerComponent("container", Container);

export default Container;
