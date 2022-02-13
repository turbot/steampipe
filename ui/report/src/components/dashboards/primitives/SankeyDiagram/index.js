import Chart from "../../charts/Chart";
import Primitive from "../../Primitive";
import { ColorGenerator } from "../../../../utils/color";

const buildOptions = (data) => {
  if (!data) {
    return null;
  }
  const colorGenerator = new ColorGenerator(24, 4);
  const builtData = [];
  const categories = {};
  const usedIds = {};
  const objectData = data.slice(1).map((dataRow) => {
    const row = {};
    for (let i = 0; i < data[0].length; i++) {
      const value = data[0][i];
      row[value] = dataRow[i];
    }

    if (row.category && !categories[row.category]) {
      categories[row.category] = { color: colorGenerator.nextColor().hex };
    }

    if (!usedIds[row.id]) {
      builtData.push({
        ...row,
        itemStyle: {
          color: categories[row.category].color,
        },
      });
      usedIds[row.id] = true;
    }
    return row;
  });
  const edges = [];
  const edgeValues = {};
  for (const d of objectData) {
    // TODO remove <null> after Kai fixes base64 issue and removes col string conversion
    if (d.parent === null || d.parent === "<null>") {
      d.parent = null;
      continue;
    }
    edges.push({ source: d.parent, target: d.id, value: 0.01 });
    edgeValues[d.parent] = (edgeValues[d.parent] || 0) + 0.01;
  }
  for (const e of edges) {
    var v = 0;
    if (edgeValues[e.target]) {
      for (const e2 of edges) {
        if (e.target === e2.source) {
          v += edgeValues[e2.target] || 0.01;
        }
      }
      e.value = v;
    }
  }
  const options = {
    //tooltip: {
    //    trigger: 'item'
    //},
    series: {
      type: "sankey",
      layout: "none",
      draggable: true,
      label: { formatter: "{b}" },
      emphasis: {
        focus: "adjacency",
        blurScope: "coordinateSystem",
      },
      //data: objectData.map(o => ),
      data: builtData,
      links: edges,
      // categories: Object.entries(categories).map(([category, info]) => ({
      //   name: category,
      //   symbol: "rect",
      //   symbolSize: [160, 40],
      //   itemStyle: { color: info.color },
      // })),
    },
  };
  // console.log(options);
  return options;
};

const SankeyDiagram = ({ data, error }) => {
  const options = buildOptions(data);
  return (
    <Primitive error={error} ready={!!data}>
      {options && <Chart options={options} />}
    </Primitive>
  );
};

export default {
  type: "sankey",
  component: SankeyDiagram,
};
