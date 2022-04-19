import Benchmark from "../../check/Benchmark";
import Card from "../../Card";
import Container from "../Container";
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
  withTitle?: boolean;
}

const Children = ({
  children = [],
  allowPanelExpand = true,
  withTitle = true,
}: ChildrenProps) => (
  <>
    {children.map((child) => {
      switch (child.node_type) {
        case "benchmark":
          return (
            <Benchmark {...child} />
            // <Panel
            //   key={child.name}
            //   definition={child}
            //   allowExpand={allowPanelExpand}
            //   withTitle={withTitle}
            // >
            //   <Benchmark {...child.execution_tree} />
            // </Panel>
          );
        case "card":
          return (
            <Panel
              key={child.name}
              definition={child}
              allowExpand={allowPanelExpand}
              withTitle={withTitle}
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
              withTitle={withTitle}
            >
              <Chart {...child} />
            </Panel>
          );
        case "container":
          return <Container key={child.name} definition={child} />;
        // case "control":
        //   return (
        //     <Panel
        //       key={child.name}
        //       definition={child}
        //       allowExpand={allowPanelExpand}
        //       withTitle={withTitle}
        //     >
        //       <Control {...child} />
        //     </Panel>
        //   );
        case "dashboard":
          return <Dashboard key={child.name} definition={child} />;
        case "error":
          return (
            <Panel
              key={child.name}
              definition={child}
              allowExpand={allowPanelExpand}
              withTitle={withTitle}
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
              withTitle={withTitle}
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
              withTitle={withTitle}
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
              withTitle={withTitle}
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
                allowPanelExpand &&
                (child.title || child.properties?.type === "table")
              }
              withTitle={withTitle}
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
              withTitle={withTitle}
            >
              <Table {...child} />
            </Panel>
          );
        case "text":
          return (
            <Panel
              key={child.name}
              definition={child}
              allowExpand={false}
              withTitle={withTitle}
            >
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
