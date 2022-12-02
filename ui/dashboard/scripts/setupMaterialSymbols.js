const fs = require("fs-extra");
const camelCase = require("lodash/camelCase");
const kebabCase = require("lodash/kebabCase");
const upperFirst = require("lodash/upperFirst");

(async () => {
  const nodeModulesPath = "@material-symbols/svg-200/rounded";
  const dir = await fs.readdir("./node_modules/" + nodeModulesPath);
  let generatedFile = "// @ts-nocheck\n";
  const outlineIcons = {};
  const solidIcons = {};
  for (const file of dir) {
    const fileNameParts = file.split(".");
    let importName = upperFirst(camelCase(fileNameParts[0]));
    if (/^\d/.test(importName)) {
      importName = "_" + importName;
    }
    const nameParts = fileNameParts[0].split("-");
    const nameKebab = kebabCase(nameParts[0]);
    const isFillIcon = nameParts.length === 2 && nameParts[1] === "fill";
    if (isFillIcon) {
      solidIcons[nameKebab] = {
        component: importName,
      };
    } else {
      outlineIcons[nameKebab] = {
        component: importName,
      };
    }
    generatedFile += `import { ReactComponent as ${importName} } from "${nodeModulesPath}/${file}";\n`;
  }
  generatedFile += "\n";
  generatedFile += "const outline = {\n";
  for (const [name, definition] of Object.entries(outlineIcons)) {
    generatedFile += `  "${name}": { Component: ${definition.component} },\n`;
  }
  generatedFile += "}\n\n";
  generatedFile += "const solid = {\n";
  for (const [name, definition] of Object.entries(solidIcons)) {
    generatedFile += `  "${name}": { Component: ${definition.component} },\n`;
  }
  generatedFile += "}\n\n";
  generatedFile += "export {\n";
  generatedFile += "  outline,\n";
  generatedFile += "  solid,\n";
  generatedFile += "}";

  await fs.writeFile("./src/components/Icon/materialSymbols.ts", generatedFile);
})();
