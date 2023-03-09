import * as echarts from "echarts/core";
import {
  BarChart,
  GraphChart,
  LineChart,
  PieChart,
  SankeyChart,
  TreeChart,
} from "echarts/charts";
import { CanvasRenderer } from "echarts/renderers";
import {
  DatasetComponent,
  GridComponent,
  LegendComponent,
  TitleComponent,
  TooltipComponent,
} from "echarts/components";
import { LabelLayout } from "echarts/features";

echarts.use([
  BarChart,
  CanvasRenderer,
  DatasetComponent,
  GraphChart,
  GridComponent,
  LabelLayout,
  LegendComponent,
  LineChart,
  PieChart,
  SankeyChart,
  TitleComponent,
  TooltipComponent,
  TreeChart,
]);

export { echarts };
