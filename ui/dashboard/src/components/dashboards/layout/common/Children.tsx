import Benchmark from "../../check/Benchmark";
import Card from "../../Card";
import Container from "../Container";
import Control from "../../check/Control";
import ErrorPanel from "../../Error";
import Image from "../../Image";
import Panel from "../Panel";
import React from "react";
import Table from "../../Table";
import Text from "../../Text";
import {
  ContainerDefinition,
  PanelDefinition,
} from "../../../../hooks/useDashboard";
import { Dashboard } from "../Dashboard";
import { RenderChart as Chart } from "../../charts/Chart";
import { RenderFlow as Flow } from "../../flows/Flow";
import { RenderHierarchy as Hierarchy } from "../../hierarchies/Hierarchy";
import { RenderInput as Input } from "../../inputs/Input";

interface ChildrenProps {
  children: ContainerDefinition[] | PanelDefinition[] | undefined;
  allowPanelExpand?: boolean;
}

const Children = ({
  children = [],
  allowPanelExpand = true,
}: ChildrenProps) => (
  <>
    {children.map((child) => {
      switch (child.node_type) {
        case "benchmark":
          return (
            <Panel
              key={child.name}
              definition={child}
              allowExpand={allowPanelExpand}
            >
              <Benchmark {...child} />
            </Panel>
          );
        case "card":
          return (
            <Panel
              key={child.name}
              definition={child}
              allowExpand={allowPanelExpand}
            >
              <Card {...child} />
            </Panel>
          );
        case "chart":
          return (
            <Panel
              key={child.name}
              definition={child}
              ready={!!child.data}
              allowExpand={allowPanelExpand}
            >
              <Chart {...child} />
            </Panel>
          );
        case "container":
          return <Container key={child.name} definition={child} />;
        case "control":
          return (
            <Panel
              key={child.name}
              definition={child}
              allowExpand={allowPanelExpand}
            >
              <Control {...child} />
            </Panel>
          );
        case "dashboard":
          return <Dashboard key={child.name} definition={child} />;
        case "error":
          return (
            <Panel
              key={child.name}
              definition={child}
              allowExpand={allowPanelExpand}
            >
              <ErrorPanel error={child.error} />
            </Panel>
          );
        case "flow":
          return (
            <Panel
              key={child.name}
              definition={child}
              ready={!!child.data}
              allowExpand={allowPanelExpand}
            >
              <Flow {...child} />
            </Panel>
          );
        case "hierarchy":
          return (
            <Panel
              key={child.name}
              definition={child}
              ready={!!child.data}
              allowExpand={allowPanelExpand}
            >
              <Hierarchy {...child} />
            </Panel>
          );
        case "image":
          return (
            <Panel
              key={child.name}
              definition={child}
              ready={child.sql ? !!child.data : !!child.properties.src}
              allowExpand={allowPanelExpand}
            >
              <Image {...child} />
            </Panel>
          );
        case "input":
          return (
            <Panel
              key={child.name}
              definition={child}
              allowExpand={
                allowPanelExpand && child.properties?.type === "table"
              }
            >
              <Input {...child} />
            </Panel>
          );
        case "table":
          return (
            <Panel
              key={child.name}
              definition={child}
              ready={!!child.data}
              allowExpand={allowPanelExpand}
            >
              <Table {...child} />
            </Panel>
          );
        case "text":
          return (
            <Panel key={child.name} definition={child} allowExpand={false}>
              <Text {...child} />
            </Panel>
          );
        default:
          return null;
      }
    })}
  </>
);

export default Children;
