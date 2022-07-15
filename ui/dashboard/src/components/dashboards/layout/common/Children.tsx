import {
  ContainerDefinition,
  PanelDefinition,
} from "../../../../hooks/useDashboard";
import { getComponent } from "../../index";

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
  const Panel = getComponent("panel");
  return (
    <>
      {children.map((child) => {
        switch (child.panel_type) {
          case "benchmark":
            const Benchmark = getComponent("benchmark");
            return (
              <Benchmark
                key={child.name}
                {...(child as PanelDefinition)}
                withTitle={withTitle}
              />
            );
          case "benchmark_tree":
            const BenchmarkTree = getComponent("benchmark_tree");
            return <BenchmarkTree key={child.name} {...child} />;
          case "card":
            const Card = getComponent("card");
            return (
              <Panel
                key={child.name}
                layoutDefinition={child}
                allowExpand={allowPanelExpand}
                withTitle={withTitle}
              >
                {(definition) => <Card {...definition} />}
              </Panel>
            );
          case "chart":
            const Chart = getComponent("chart");
            return (
              <Panel
                key={child.name}
                layoutDefinition={child}
                ready={(definition) => !!definition.data}
                allowExpand={allowPanelExpand}
                withTitle={withTitle}
              >
                {(definition) => <Chart {...definition} />}
              </Panel>
            );
          case "container":
            const Container = getComponent("container");
            return (
              <Container
                key={child.name}
                allowChildPanelExpand={child.allow_child_panel_expand}
                expandDefinition={child}
                layoutDefinition={child}
                withTitle={withTitle}
              />
            );
          case "dashboard":
            const Dashboard = getComponent("dashboard");
            return (
              <Dashboard key={child.name} definition={child} isRoot={false} />
            );
          case "error":
            const ErrorPanel = getComponent("error");
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
            const Flow = getComponent("flow");
            return (
              <Panel
                key={child.name}
                layoutDefinition={child}
                ready={(definition) => !!definition.data}
                allowExpand={allowPanelExpand}
                withTitle={withTitle}
              >
                {(definition) => <Flow {...definition} />}
              </Panel>
            );
          case "hierarchy":
            const Hierarchy = getComponent("hierarchy");
            return (
              <Panel
                key={child.name}
                layoutDefinition={child}
                ready={(definition) => !!definition.data}
                allowExpand={allowPanelExpand}
                withTitle={withTitle}
              >
                {(definition) => <Hierarchy {...definition} />}
              </Panel>
            );
          case "image":
            const Image = getComponent("image");
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
                {(definition) => <Image {...definition} />}
              </Panel>
            );
          case "input":
            const Input = getComponent("input");
            return (
              <Panel
                key={child.name}
                layoutDefinition={child}
                allowExpand={
                  allowPanelExpand &&
                  (child.title || child.display_type === "table")
                }
                withTitle={withTitle}
              >
                {(definition) => <Input {...definition} />}
              </Panel>
            );
          case "table":
            const Table = getComponent("table");
            return (
              <Panel
                key={child.name}
                layoutDefinition={child}
                ready={(definition) => !!definition.data}
                allowExpand={allowPanelExpand}
                withTitle={withTitle}
              >
                {(definition) => <Table {...definition} />}
              </Panel>
            );
          case "text":
            const Text = getComponent("text");
            return (
              <Panel
                key={child.name}
                layoutDefinition={child}
                allowExpand={false}
                withTitle={withTitle}
              >
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
