import jq from "jq-web";

const interpolatedStringSplitter = /({{.*?}})/gm;
const interpolatedMatcher = /{{(.*?)}}/gm;
const singleQuoteMatcher = /(?:^|[^\\])(')/gm;

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

export const buildJQFilter = (template) => {
  if (!template) {
    return template;
  }

  const templateParts = template
    .split(interpolatedStringSplitter)
    .filter((p) => p);
  const newTemplateParts: string[] = [];
  // Iterate over each template part - we want to distinguish between regular strings and
  // interpolated strings - we'll treat them differently.
  for (const templatePart of templateParts) {
    const interpolatedMatch = interpolatedMatcher.exec(templatePart);
    // If it's a plain string, quote it
    if (!interpolatedMatch) {
      newTemplateParts.push(`"${templatePart}"`);
    } else {
      // If it's an interpolated string, replace the double curly braces with single parentheses
      // to frame this particular jq sub-expression
      let newInterpolatedTemplate = templatePart;
      newInterpolatedTemplate =
        "(" + newInterpolatedTemplate.substring(interpolatedMatch.index + 2);
      newInterpolatedTemplate =
        newInterpolatedTemplate.substring(0, interpolatedMatch[0].length - 3) +
        ")";

      // Replace any unescaped single quotes with jq-compatible double quotes
      const singleQuoteMatches =
        newInterpolatedTemplate.matchAll(singleQuoteMatcher);
      for (const singleQuoteMatch of singleQuoteMatches) {
        const matchPrefix = singleQuoteMatch[0]
          .split(singleQuoteMatch[1])
          .join("");
        newInterpolatedTemplate =
          newInterpolatedTemplate.substring(0, singleQuoteMatch.index) +
          `${matchPrefix}"` +
          newInterpolatedTemplate.substring(
            singleQuoteMatch.index + singleQuoteMatch[0].length
          );
      }
      newTemplateParts.push(newInterpolatedTemplate);
    }
  }

  // Join all field parts with + and then quote the overall filter
  return `(${newTemplateParts.join(" + ")})`;
};

const buildCombinedJQFilter = (templates: TemplatesMap) => {
  const filters: TemplatesMap = {};
  // Iterate over all the template fields
  for (const [field, template] of Object.entries(templates)) {
    filters[field] = buildJQFilter(template);
  }

  // Now we want to include all fields with their filter - we're going to map the output to
  // an object with a key per field and the value is the rendered template
  const allFieldsFilter = Object.entries(filters)
    .map(([field, filter]) => `"${field}": ${filter}`)
    .join(", ");

  // Finally, build the overall filter - which will iterate over all passed in rows of data, then turn
  // the result into an array of the object returned by the combined field filters
  return `[ .[] | { ${allFieldsFilter} }]`;
};

const renderInterpolatedTemplates = async (
  templates: TemplatesMap,
  data: DataMap[]
): Promise<RowRenderResult[]> => {
  const finalFilter = buildCombinedJQFilter(templates);
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

          const singleQuoteMatches = templatePart.matchAll(singleQuoteMatcher);
          let doubleQuotedString = templatePart;

          for (const singleQuoteMatch of singleQuoteMatches) {
            const matchPrefix = singleQuoteMatch[0]
              .split(singleQuoteMatch[1])
              .join("");
            doubleQuotedString =
              doubleQuotedString.substring(0, singleQuoteMatch.index) +
              `${matchPrefix}"` +
              doubleQuotedString.substring(
                singleQuoteMatch.index + singleQuoteMatch[0].length
              );
          }

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
