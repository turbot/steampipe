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

const getInterpolatedTemplateValue = async (
  template,
  context
): Promise<string | null> => {
  const interpolatedMatcher = /\{\{([^}]+)}}/gm;
  let updatedTemplate = template;
  try {
    let match;
    while ((match = interpolatedMatcher.exec(template)) !== null) {
      // This is necessary to avoid infinite loops with zero-width matches
      if (match.index === interpolatedMatcher.lastIndex) {
        interpolatedMatcher.lastIndex++;
      }

      const templatePart = match[1];

      const doubleQuotedString = (templatePart || "").replace(
        /(?<!\\)'/gm,
        '"'
      );
      const rendered = await jq.json(context, doubleQuotedString);

      // If we get a null result, we don't want to continue
      if (rendered === null) {
        return null;
      }

      updatedTemplate = updatedTemplate.replace(match[0], rendered);
    }
  } catch (err) {
    console.log("Error rendering column template", err);
    return null;
  }

  return updatedTemplate;

  //
  // console.log(raw);
  // if (!raw) {
  //   return raw;
  // }
  // const interpolatedParts = interpolatedMatcher.exec(raw);
  // console.log(interpolatedParts);
  // return raw;
  // // const templateParts = {};
  // // return jq.json(context) raw;
};

export { getInterpolatedTemplateValue, renderTemplates };
