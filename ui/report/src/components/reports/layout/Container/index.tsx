import Children from "../common/Children";
import LayoutPanel from "../common/LayoutPanel";
import React from "react";
import { ContainerDefinition } from "../../../../hooks/useReport";

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
