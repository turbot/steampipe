import CodeBlock from "../../../CodeBlock";
import { format } from "@supabase/sql-formatter";
import { PanelDetailProps } from "./index";
import { useMemo } from "react";

const beautify = (query) => {
  if (!query) {
    return null;
  }
  return format(query || "", {
    language: "postgresql",
  });
};

const PanelDetailQuery = ({ definition }: PanelDetailProps) => {
  const formattedQuery = useMemo(
    () => beautify(definition.sql),
    [definition.sql]
  );

  if (!formattedQuery) {
    return <></>;
  }

  return <CodeBlock language="sql">{formattedQuery}</CodeBlock>;
};

export default PanelDetailQuery;
