import Benchmark from "../../check/Benchmark";
import Container from "../Container";
import Control from "../../check/Control";
import Counter from "../../Counter";
import ErrorPanel from "../../Error";
import Image from "../../Image";
import Panel from "../Panel";
import React from "react";
import Table from "../../Table";
import Text from "../../Text";
import TitleWrapper from "../../TitleWrapper";
import {
  ContainerDefinition,
  PanelDefinition,
} from "../../../../hooks/useReport";
import { RenderChart as Chart } from "../../charts/Chart";
import { RenderInput as Input } from "../../inputs/Input";

const ChildWithTitle = ({ child, level, renderChild }) => {
  return (
    <TitleWrapper definition={child} level={level} title={child.title}>
      {child.title ? null : renderChild()}
    </TitleWrapper>
  );
};

interface ChildrenProps {
  children: ContainerDefinition[] | PanelDefinition[] | undefined;
  showPanelExpand?: boolean;
}

const Children = ({ children = [], showPanelExpand = true }: ChildrenProps) => (
  <>
    {children.map((child) => {
      switch (child.node_type) {
        case "container":
          return (
            <ChildWithTitle
              key={child.name}
              child={child}
              level="container"
              renderChild={() => <Container definition={child} />}
            />
          );
        case "benchmark":
          return (
            <ChildWithTitle
              key={child.name}
              child={child}
              level="container"
              renderChild={() => (
                <Panel
                  definition={child}
                  ready={true}
                  showExpand={showPanelExpand}
                >
                  <Benchmark {...child} />
                </Panel>
              )}
            />
          );
        case "chart":
          return (
            <ChildWithTitle
              key={child.name}
              child={child}
              level="panel"
              renderChild={() => (
                <Panel
                  definition={child}
                  ready={!!child.data}
                  showExpand={showPanelExpand}
                >
                  <Chart {...child} />
                </Panel>
              )}
            />
          );
        case "control":
          return (
            <ChildWithTitle
              key={child.name}
              child={child}
              level="panel"
              renderChild={() => (
                <Panel
                  definition={child}
                  ready={true}
                  showExpand={showPanelExpand}
                >
                  <Control {...child} />
                </Panel>
              )}
            />
          );
        case "counter":
          return (
            <ChildWithTitle
              key={child.name}
              child={child}
              level="panel"
              renderChild={() => (
                <Panel
                  definition={child}
                  ready={true}
                  showExpand={showPanelExpand}
                >
                  <Counter {...child} />
                </Panel>
              )}
            />
          );
        case "image":
          return (
            <ChildWithTitle
              key={child.name}
              child={child}
              level="panel"
              renderChild={() => (
                <Panel
                  definition={child}
                  ready={child.sql ? !!child.data : !!child.properties.src}
                  showExpand={showPanelExpand}
                >
                  <Image {...child} />
                </Panel>
              )}
            />
          );
        case "input":
          return (
            <Panel
              key={child.name}
              definition={child}
              ready={true}
              showExpand={showPanelExpand}
            >
              <Input {...child} />
            </Panel>
          );
        case "text":
          return (
            <ChildWithTitle
              key={child.name}
              child={child}
              level="panel"
              renderChild={() => (
                <Panel definition={child} showExpand={showPanelExpand}>
                  <Text {...child} />
                </Panel>
              )}
            />
          );
        case "table":
          return (
            <ChildWithTitle
              key={child.name}
              child={child}
              level="panel"
              renderChild={() => (
                <Panel
                  definition={child}
                  ready={!!child.data}
                  showExpand={showPanelExpand}
                >
                  <Table {...child} />
                </Panel>
              )}
            />
          );
        case "error":
          return (
            <ChildWithTitle
              key={child.name}
              child={child}
              level="panel"
              renderChild={() => (
                <Panel definition={child} showExpand={showPanelExpand}>
                  <ErrorPanel error={`Unknown resource type: ${child.name}`} />
                </Panel>
              )}
            />
          );
        default:
          return null;
      }
    })}
  </>
);

export default Children;
