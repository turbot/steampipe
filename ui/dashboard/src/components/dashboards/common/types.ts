import { TableColumnDisplay, TableColumnWrap } from "../Table";

export interface ChartTooltipFormatter {
  format(params: Object | any[]): string;
}

export interface CategoryFields {
  [name: string]: CategoryField;
}

export interface CategoryField {
  name: string;
  href?: string | null;
  display?: TableColumnDisplay;
  wrap?: TableColumnWrap;
}

export interface KeyValuePairs {
  [key: string]: any;
}
