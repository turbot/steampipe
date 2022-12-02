// import { readdirSync } from "fs";

console.log("Setting up fonts");
// const reqSvgs = require.context(
//   "@material-symbols/svg-200/rounded",
//   false,
//   /\.svg$/
// );
const reqSvgs = require.context(
  "!@svgr/webpack!@material-symbols/svg-200/rounded",
  // "!svg-react-loader!@material-symbols/svg-200/rounded",
  // "!svg-inline-loader!@material-symbols/svg-200/rounded",
  true,
  /\.svg$/
);
// console.log(readdirSync("@material-symbols/svg-200/rounded"));
const symbolsMap = reqSvgs.keys().reduce((images, path) => {
  const key = path.substring(path.lastIndexOf("/") + 1, path.lastIndexOf("."));
  images[key] = reqSvgs(path);
  return images;
}, {});
// const svgs = reqSvgs.keys().map((path) => ({ path, file: reqSvgs(path) }));
// const svgs = reqSvgs.keys();
// import * as icons from "@material-symbols/svg-200/rounded";
export { symbolsMap as icons };
