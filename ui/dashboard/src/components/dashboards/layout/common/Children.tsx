import Benchmark, { BenchmarkTree } from "../../check/Benchmark";
import Card from "../../Card";
import Container from "../Container";
import ErrorPanel from "../../Error";
import Image from "../../Image";
import Panel from "../Panel";
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
}: ChildrenProps) => {
  return (
    <>
      {children.map((child) => {
        switch (child.node_type) {
          case "benchmark":
            return (
              <Benchmark
                key={child.name}
                {...(child as PanelDefinition)}
                withTitle={withTitle}
              />
            );
          case "benchmark_tree":
            return <BenchmarkTree key={child.name} {...child} />;
          case "card":
            return (
              <Panel
                key={child.name}
                layoutDefinition={child}
                allowExpand={allowPanelExpand}
                withTitle={withTitle}
              >
                {/*@ts-ignore*/}
                {(definition) => <Card {...definition} />}
              </Panel>
            );
          case "chart":
            return (
              <Panel
                key={child.name}
                layoutDefinition={child}
                ready={(definition) => !!definition.data}
                allowExpand={allowPanelExpand}
                withTitle={withTitle}
              >
                {/*@ts-ignore*/}
                {(definition) => <Chart {...definition} />}
              </Panel>
            );
          case "container":
            return (
              <Container
                key={child.name}
                allowChildPanelExpand={child.allow_child_panel_expand}
                definition={child}
                expandDefinition={child}
                withTitle={withTitle}
              />
            );
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
                layoutDefinition={child}
                allowExpand={allowPanelExpand}
                withTitle={withTitle}
              >
                {(definition) => <ErrorPanel {...definition} />}
              </Panel>
            );
          case "flow":
            return (
              <Panel
                key={child.name}
                layoutDefinition={child}
                ready={(definition) => !!definition.data}
                allowExpand={allowPanelExpand}
                withTitle={withTitle}
              >
                {/*@ts-ignore*/}
                {(definition) => <Flow {...definition} />}
              </Panel>
            );
          case "hierarchy":
            return (
              <Panel
                key={child.name}
                layoutDefinition={child}
                ready={(definition) => !!definition.data}
                allowExpand={allowPanelExpand}
                withTitle={withTitle}
              >
                {/*@ts-ignore*/}
                {(definition) => <Hierarchy {...definition} />}
              </Panel>
            );
          case "image":
            return (
              <Panel
                key={child.name}
                layoutDefinition={child}
                ready={(definition) =>
                  definition.sql
                    ? !!definition.data
                    : !!definition.properties.src
                }
                allowExpand={allowPanelExpand}
                withTitle={withTitle}
              >
                {/*@ts-ignore*/}
                {(definition) => <Image {...definition} />}
              </Panel>
            );
          case "input":
            return (
              <Panel
                key={child.name}
                layoutDefinition={child}
                allowExpand={
                  allowPanelExpand &&
                  (child.title || child.properties?.type === "table")
                }
                withTitle={withTitle}
              >
                {/*@ts-ignore*/}
                {(definition) => <Input {...definition} />}
              </Panel>
            );
          case "table":
            return (
              <Panel
                key={child.name}
                layoutDefinition={child}
                ready={(definition) => !!definition.data}
                allowExpand={allowPanelExpand}
                withTitle={withTitle}
              >
                {/*@ts-ignore*/}
                {(definition) => <Table {...definition} />}
              </Panel>
            );
          case "text":
            return (
              <Panel
                key={child.name}
                layoutDefinition={child}
                allowExpand={false}
                withTitle={withTitle}
              >
                {/*@ts-ignore*/}
                {(definition) => <Text {...definition} />}
              </Panel>
            );
          default:
            return null;
        }
      })}
    </>
  );
};

export default Children;
