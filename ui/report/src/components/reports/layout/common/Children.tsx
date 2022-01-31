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
                  type="benchmark"
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
                  type="chart"
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
                  type="control"
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
                  type="counter"
                >
                  <Counter {...child} />
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
                <Panel
                  definition={child}
                  showExpand={showPanelExpand}
                  type="error"
                >
                  <ErrorPanel error={`Unknown resource type: ${child.name}`} />
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
                  type="image"
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
              type="input"
            >
              <Input {...child} />
            </Panel>
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
                  type="table"
                >
                  <Table {...child} />
                </Panel>
              )}
            />
          );
        case "text":
          return (
            <ChildWithTitle
              key={child.name}
              child={child}
              level="panel"
              renderChild={() => (
                <Panel
                  definition={child}
                  showExpand={showPanelExpand}
                  type="text"
                >
                  <Text {...child} />
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
