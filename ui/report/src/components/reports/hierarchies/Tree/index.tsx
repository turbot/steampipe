import Hierarchy from "../Hierarchy";
import React from "react";
import { HierarchyProps, IHierarchy } from "../index";

const Tree = (props: HierarchyProps) => (
  <Hierarchy data={props.data} inputs={{ type: "tree", ...props.properties }} />
);

const definition: IHierarchy = {
  type: "tree",
  component: Tree,
};

export default definition;
