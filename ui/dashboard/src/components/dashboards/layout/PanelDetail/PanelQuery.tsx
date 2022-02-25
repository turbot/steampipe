import CodeBlock from "../../../CodeBlock";
import { format } from "@supabase/sql-formatter";
import { useMemo } from "react";

const beautify = (query) => {
  return format(query || "", {
    language: "postgresql",
  });
};

interface PanelQueryProps {
  query: string;
}

const PanelQuery = ({ query }: PanelQueryProps) => {
  const formattedQuery = useMemo(() => beautify(query), [query]);

  return <CodeBlock language="sql">{formattedQuery}</CodeBlock>;
};

export default PanelQuery;
