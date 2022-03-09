// // @ts-ignore
import jq from "jq-web";
//
// // console.log(foo);
// console.log(foo.json());

// const interpolatedMatcher = /(?<=\{\{).*?(?=}})/g;

// @ts-ignore
// import("../jq.wasm").then(({ json }) => console.log(json));

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

export { getInterpolatedTemplateValue };
