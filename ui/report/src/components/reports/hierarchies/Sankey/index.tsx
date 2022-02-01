import Hierarchy from "../Hierarchy";
import React from "react";
import { HierarchyProps, IHierarchy } from "../index";

const Sankey = (props: HierarchyProps) => (
  <Hierarchy
    data={props.data}
    inputs={{ type: "sankey", ...props.properties }}
  />
);

const definition: IHierarchy = {
  type: "sankey",
  component: Sankey,
};

export default definition;
