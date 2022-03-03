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
  const interpolatedMatcher = /\{\{([^}]+)}}/g;
  let updatedTemplate = template;
  try {
    let match;
    while ((match = interpolatedMatcher.exec(template)) !== null) {
      // This is necessary to avoid infinite loops with zero-width matches
      if (match.index === interpolatedMatcher.lastIndex) {
        interpolatedMatcher.lastIndex++;
      }

      const templatePart = match[1];
      // console.log("Rendering", templatePart, context);
      const rendered = await jq.json(context, templatePart);

      updatedTemplate = updatedTemplate.replace(match[0], rendered);

      // // The result can be accessed through the `m`-variable.
      // match.forEach((match, groupIndex) => {
      //   // console.log(`Found match, group ${groupIndex}: ${match}`);
      // });
    }
  } catch (err) {
    console.log(err);
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
