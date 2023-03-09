import {
  KeyValuePairs,
  TemplatesMap,
} from "../components/dashboards/common/types";
import { renderInterpolatedTemplates } from "../utils/template";
import { useCallback, useEffect, useState } from "react";

const useTemplateRender = () => {
  const [jqWeb, setJqWeb] = useState<any | null>(null);

  // Dynamically import jq-web from its own bundle
  useEffect(() => {
    import("jq-web").then((m) => setJqWeb(m));
  }, []);

  const renderTemplates = useCallback(
    async (templates: TemplatesMap, data: KeyValuePairs[]) => {
      if (!jqWeb) {
        return [];
      }
      return renderInterpolatedTemplates(templates, data, jqWeb);
    },
    [jqWeb]
  );

  return {
    renderTemplates,
    ready: !!jqWeb,
  };
};

export default useTemplateRender;
