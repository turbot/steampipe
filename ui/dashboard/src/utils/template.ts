import {
  KeyValuePairs,
  RowRenderResult,
  TemplatesMap,
} from "../components/dashboards/common/types";

const replaceSingleQuotesWithDoubleQuotes = (str) => {
  if (!str) {
    return str;
  }
  return str.replaceAll("'", '"');
};

export const buildJQFilter = (template) => {
  if (!template) {
    return template;
  }

  const templateParts = template.split(/({{.*?}})/gs).filter((p) => p);
  const newTemplateParts: string[] = [];
  // Iterate over each template part - we want to distinguish between regular strings and
  // interpolated strings - we'll treat them differently.
  for (const templatePart of templateParts) {
    const interpolatedMatch = /{{(.*?)}}/gs.exec(templatePart);
    // If it's a plain string, quote it
    if (!interpolatedMatch) {
      newTemplateParts.push(JSON.stringify(templatePart));
    } else {
      // If it's an interpolated string, replace the double curly braces with single parentheses
      // to frame this particular jq sub-expression
      let newInterpolatedTemplate = templatePart;
      newInterpolatedTemplate =
        "(" + newInterpolatedTemplate.substring(interpolatedMatch.index + 2);
      newInterpolatedTemplate =
        newInterpolatedTemplate.substring(0, interpolatedMatch[0].length - 3) +
        ")";

      // Replace any single quotes with jq-compatible double quotes
      const doubleQuotedString = replaceSingleQuotesWithDoubleQuotes(
        newInterpolatedTemplate
      );

      newTemplateParts.push(doubleQuotedString);
    }
  }

  // Join all field parts into an array and then use jq
  // to join then into the final filter.
  // This ensures that types are coerced
  // e.g. (5 + " hello") would result in an error,
  // whereas ([5, " hello"] | join("")) would give "5 hello"
  return `([${newTemplateParts.join(", ")}] | join(""))`;
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
  data: KeyValuePairs[],
  jq: any
): Promise<RowRenderResult[]> => {
  try {
    const finalFilter = buildCombinedJQFilter(templates);
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
    const interpolatedMatcher = /{{(.*?)}}/gs;
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

          // Replace any single quotes with jq-compatible double quotes
          const doubleQuotedString =
            replaceSingleQuotesWithDoubleQuotes(templatePart);

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
