import CodeBlock from "../../../CodeBlock";
import CopyToClipboard from "../../../CopyToClipboard";
import { PanelDetailProps } from "./index";
import { useMemo } from "react";

const PanelDetailDefinition = ({ definition }: PanelDetailProps) => {
  const formattedDefinition = useMemo(() => {
    if (!definition.source_definition) {
      return null;
    }
    const tabsToSpaces = definition.source_definition.replace(
      /\t/g,
      "&nbsp;&nbsp;"
    );
    const initialSpaces = tabsToSpaces.search(/\S/);
    const spaceString = " ".repeat(initialSpaces);
    return tabsToSpaces
      .replace(new RegExp(`^${spaceString}`), "")
      .replaceAll(`\n${spaceString}`, "\n");
  }, [definition.source_definition]);

  if (!formattedDefinition) {
    return <></>;
  }

  return (
    <div className="space-y-4">
      <CopyToClipboard data={formattedDefinition} />
      <CodeBlock language="hcl">{formattedDefinition}</CodeBlock>
    </div>
  );
};

export default PanelDetailDefinition;
