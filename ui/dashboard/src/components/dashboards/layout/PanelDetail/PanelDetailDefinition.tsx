import CodeBlock from "../../../CodeBlock";
import { PanelDetailProps } from "./index";

const PanelDetailDefinition = ({ definition }: PanelDetailProps) => {
  if (!definition.source_definition) {
    return <></>;
  }

  return (
    <div className="col-span-12 mt-4">
      <CodeBlock language="hcl">{definition.source_definition}</CodeBlock>
    </div>
  );
};

export default PanelDetailDefinition;
