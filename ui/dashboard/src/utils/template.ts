import jq from "jq-web";

const interpolatedStringSplitter = /((?<!\\){(?<!\\){[^}]+(?<!\\)}(?<!\\)})/gm;
const interpolatedMatcher = /(?<!\\){(?<!\\){([^}]+)(?<!\\)}(?<!\\)}/gm;

interface TemplatesMap {
  [key: string]: string;
}

interface DataMap {
  [key: string]: any;
}

export interface RowRenderResult {
  [key: string]: {
    result?: string;
    error?: string;
  };
}

const renderInterpolatedTemplates = async (
  templates: TemplatesMap,
  data: DataMap[]
): Promise<RowRenderResult[]> => {
  const filters: TemplatesMap = {};
  // Iterate over all the template fields
  for (const [field, template] of Object.entries(templates)) {
    // First, we want to split the string, but unlike a normal string split where you wouldn't
    // include the split char(s), here we want to split on interpolated expressions and include those
    // in the split array - we can do that with a regex split that has a capturing group
    const templateParts = template
      .split(interpolatedStringSplitter)
      .filter((p) => p);
    const newTemplateParts: string[] = [];
    // Iterate over each template part - we want to distinguish between regular strings and
    // interpolated strings - we'll treat them differently.
    for (const templatePart of templateParts) {
      const interpolatedMatch = templatePart.match(interpolatedMatcher);
      // If it's a plain string, quote it
      if (!interpolatedMatch) {
        newTemplateParts.push(`"${templatePart}"`);
      } else {
        // If it's an interpolated string, replace the double curly braces with single parenthese
        // to frame this particular jq sub-expression
        let newInterpolatedTemplate = templatePart;
        newInterpolatedTemplate = newInterpolatedTemplate.replace(
          /(?<!\\){(?<!\\){/,
          "("
        );
        newInterpolatedTemplate = newInterpolatedTemplate.replace(
          /(?<!\\)}(?<!\\)}/,
          ")"
        );
        // Ensure that unescape single quotes are escaped to jq-compatible double quotes
        const doubleQuotedFilter = (newInterpolatedTemplate || "").replace(
          /(?<!\\)'/gm,
          '"'
        );
        newTemplateParts.push(doubleQuotedFilter);
      }
    }

    // Join all field parts with + and then quote the overall filter
    filters[field] = `(${newTemplateParts.join(" + ")})`;
  }

  // Now we want to include all fields with their filter - we're going to map the output to
  // an object with a key per field and the value is the rendered template
  const allFieldsFilter = Object.entries(filters)
    .map(([field, filter]) => `"${field}": ${filter}`)
    .join(", ");

  // Finally, build the overall filter - which will iterate over all passed in rows of data, then turn
  // the result into an array of the object returned by the combined field filters
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
    const interpolatedMatcher = /\{\{([^}]+)}}/gm;
    const testRow = data[0];
    const fieldResults: RowRenderResult = {};
    for (const [field, template] of Object.entries(templates)) {
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
          const rendered = await jq.json(testRow, doubleQuotedString);

          updatedTemplate = updatedTemplate.replace(match[0], rendered);
        }
      } catch (err) {
        // @ts-ignore
        fieldResults[field] = { error: err.stack };
      }
    }
    const errorResult: RowRenderResult[] = [];
    for (let i = 0; i < data.length; i++) {
      errorResult.push(fieldResults);
    }
    return errorResult;
  }
};

export { renderInterpolatedTemplates };
