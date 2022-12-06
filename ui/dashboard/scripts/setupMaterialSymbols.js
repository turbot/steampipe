const camelCase = require("lodash/camelCase");
const fs = require("fs-extra");
const upperFirst = require("lodash/upperFirst");

(async () => {
  const nodeModulesPath = "@material-symbols/svg-300/rounded";
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
    const nameAndStyleParts = fileNameParts[0].split("-");
    const nameKebab = nameAndStyleParts[0].replaceAll("_", "-");
    const isFillIcon =
      nameAndStyleParts.length === 2 && nameAndStyleParts[1] === "fill";

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
  generatedFile += "const icons = {\n";
  for (const [name, definition] of Object.entries(outlineIcons)) {
    generatedFile += `  "${name}": ${definition.component},\n`;
    generatedFile += `  "materialsymbols-outline:${name}": ${definition.component},\n`;
  }

  for (const [name, definition] of Object.entries(solidIcons)) {
    generatedFile += `  "materialsymbols-solid:${name}": ${definition.component},\n`;
  }
  generatedFile += "}\n\n";
  generatedFile += "export {\n";
  generatedFile += "  icons,\n";
  generatedFile += "}";

  await fs.writeFile("./src/icons/materialSymbols.ts", generatedFile);
})();
