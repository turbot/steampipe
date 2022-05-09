import CodeBlock from "../../../CodeBlock";
import { PanelDetailProps } from "./index";
import { useMemo } from "react";

const PanelDetailDefinition = ({ definition }: PanelDetailProps) => {
  const formattedDefinition = useMemo(() => {
    if (!definition.source_definition) {
      return null;
    }
    const tabsToSpaces = definition.source_definition.replace(/\t/g, "  ");
    const initialSpaces = tabsToSpaces.search(/\S/);
    const spaceString = " ".repeat(initialSpaces);
    return tabsToSpaces
      .replace(new RegExp(`^${spaceString}`), "")
      .replaceAll(`\n${spaceString}`, "\n");
  }, [definition.source_definition]);

  if (!formattedDefinition) {
    return <></>;
  }

  return <CodeBlock language="hcl">{formattedDefinition}</CodeBlock>;
};

export default PanelDetailDefinition;
