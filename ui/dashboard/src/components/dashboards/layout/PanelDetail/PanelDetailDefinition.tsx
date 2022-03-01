import CodeBlock from "../../../CodeBlock";
import { PanelDetailProps } from "./index";

const PanelDetailDefinition = ({ definition }: PanelDetailProps) => {
  if (!definition.source_definition) {
    return <></>;
  }

  return <CodeBlock language="hcl">{definition.source_definition}</CodeBlock>;
};

export default PanelDetailDefinition;
