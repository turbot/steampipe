import jq from "jq-web";

const getInterpolatedTemplateValue = (raw, context) => {
  console.log(raw);
  return raw;
  // const templateParts = {};
  // return jq.json(context) raw;
};

export { getInterpolatedTemplateValue };
