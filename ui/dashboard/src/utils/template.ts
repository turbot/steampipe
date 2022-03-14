import jq from "jq-web";

const interpolatedStringSplitter = /((?<!\\){(?<!\\){[^}]+(?<!\\)}(?<!\\)})/gm;
const interpolatedMatcher = /(?<!\\){(?<!\\){([^}]+)(?<!\\)}(?<!\\)}/gm;

interface TemplatesMap {
  [key: string]: string;
}

interface DataMap {
  [key: string]: string;
}

export interface RenderResults {
  [key: string]: {
    result?: string;
    error?: string;
  };
}

const renderTemplates = async (
  templates: TemplatesMap,
  data: DataMap[]
): Promise<RenderResults[]> => {
  const filters: TemplatesMap = {};
  for (const [field, template] of Object.entries(templates)) {
    const templateParts = template
      .split(interpolatedStringSplitter)
      .filter((p) => p);
    const newTemplateParts: string[] = [];
    for (const templatePart of templateParts) {
      const interpolatedMatch = templatePart.match(interpolatedMatcher);
      if (!interpolatedMatch) {
        newTemplateParts.push(`"${templatePart}"`);
      } else {
        let newInterpolatedTemplate = templatePart;
        newInterpolatedTemplate = newInterpolatedTemplate.replace(
          /(?<!\\){(?<!\\){/,
          "("
        );
        newInterpolatedTemplate = newInterpolatedTemplate.replace(
          /(?<!\\)}(?<!\\)}/,
          ")"
        );
        const doubleQuotedFilter = (newInterpolatedTemplate || "").replace(
          /(?<!\\)'/gm,
          '"'
        );
        newTemplateParts.push(doubleQuotedFilter);
      }
    }

    filters[field] = `(${newTemplateParts.join(" + ")})`;
  }

  const allFieldsFilter = Object.entries(filters)
    .map(([field, filter]) => `"${field}": ${filter}`)
    .join(", ");

  const finalFilter = `[ .[] | { ${allFieldsFilter} }]`;

  try {
    const results = await jq.json(data, finalFilter);
    return results.map((result) => {
      const mapped = {};
      Object.entries(result).forEach(([field, rendered]) => {
        mapped[field] = {
          result: rendered,
        };
      });
      return mapped;
    });
  } catch (err) {
    console.error(err);
  }

  return [];
};

export { renderTemplates };
