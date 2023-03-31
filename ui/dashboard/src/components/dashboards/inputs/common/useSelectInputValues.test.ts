import useSelectInputValues from "./useSelectInputValues";
import { DashboardRunState } from "../../../../types";
import { LeafNodeData } from "../../common";
import { renderHook } from "@testing-library/react";
import { SelectInputOption, SelectOption } from "../types";

const options = [
  {
    name: "name1",
    label: "label1",
  },
  {
    name: "name2",
    label: "label2",
  },
];

const dataNoTags = {
  columns: [
    {
      name: "label",
      data_type: "TEXT",
    },
    {
      name: "value",
      data_type: "TEXT",
    },
  ],
  rows: [
    {
      label: "label1",
      value: "value1",
    },
    {
      label: "label2",
      value: "value2",
    },
  ],
};

const dataWithTags = {
  columns: [
    {
      name: "label",
      data_type: "TEXT",
    },
    {
      name: "value",
      data_type: "TEXT",
    },
    {
      name: "tags",
      data_type: "JSONB",
    },
  ],
  rows: [
    {
      label: "label1",
      value: "value1",
      tags: { foo: "bar" },
    },
    {
      label: "label2",
      value: "value2",
      tags: { bar: "foo" },
    },
  ],
};

type Test = {
  name: string;
  options: SelectInputOption[] | undefined;
  data: LeafNodeData | undefined;
  status: DashboardRunState;
  expected: SelectOption[];
};

const tests: Test[] = [
  {
    name: "returns expected array for static options",
    options,
    data: undefined,
    status: "complete",
    expected: [
      {
        label: "label1",
        value: "name1",
        tags: {},
      },
      {
        label: "label2",
        value: "name2",
        tags: {},
      },
    ],
  },
  {
    name: "returns expected array for data options with no tags",
    options: undefined,
    data: dataNoTags,
    status: "complete",
    expected: [
      {
        label: "label1",
        value: "value1",
        tags: {},
      },
      {
        label: "label2",
        value: "value2",
        tags: {},
      },
    ],
  },
  {
    name: "returns expected array for data options with tags",
    options: undefined,
    data: dataWithTags,
    status: "complete",
    expected: [
      {
        label: "label1",
        value: "value1",
        tags: { foo: "bar" },
      },
      {
        label: "label2",
        value: "value2",
        tags: { bar: "foo" },
      },
    ],
  },
  {
    name: "returns empty array if undefined options and data",
    options: undefined,
    data: undefined,
    status: "complete",
    expected: [],
  },
  {
    name: "returns empty array if options provided and not complete",
    options: options,
    data: undefined,
    status: "initialized",
    expected: [],
  },
  {
    name: "returns empty array if data provided and not complete",
    options: undefined,
    data: dataNoTags,
    status: "initialized",
    expected: [],
  },
  {
    name: "uses static option name if no label property",
    options: [
      {
        name: "name1",
      },
      {
        name: "name2",
      },
    ],
    data: undefined,
    status: "complete",
    expected: [
      {
        label: "name1",
        value: "name1",
        tags: {},
      },
      {
        label: "name2",
        value: "name2",
        tags: {},
      },
    ],
  },
  {
    name: "returns empty array if label data column not present",
    options: undefined,
    data: {
      columns: [
        {
          name: "value",
          data_type: "TEXT",
        },
      ],
      rows: [
        {
          value: "value1",
        },
        {
          value: "value2",
        },
      ],
    },
    status: "complete",
    expected: [],
  },
  {
    name: "returns empty array if value data column not present",
    options: undefined,
    data: {
      columns: [
        {
          name: "label",
          data_type: "TEXT",
        },
      ],
      rows: [
        {
          label: "label1",
        },
        {
          label: "label2",
        },
      ],
    },
    status: "complete",
    expected: [],
  },
  {
    name: "label is empty string if data label column is not truthy",
    options: undefined,
    data: {
      columns: [
        {
          name: "label",
          data_type: "TEXT",
        },
        {
          name: "value",
          data_type: "TEXT",
        },
      ],
      rows: [
        {
          label: "label1",
          value: "value1",
        },
        {
          label: undefined,
          value: "value2",
        },
      ],
    },
    status: "complete",
    expected: [
      {
        label: "label1",
        value: "value1",
        tags: {},
      },
      {
        label: "",
        value: "value2",
        tags: {},
      },
    ],
  },
  {
    name: "value is null if data value column is not truthy",
    options: undefined,
    data: {
      columns: [
        {
          name: "label",
          data_type: "TEXT",
        },
        {
          name: "value",
          data_type: "TEXT",
        },
      ],
      rows: [
        {
          label: "label1",
          value: undefined,
        },
        {
          label: "label2",
          value: "value2",
        },
      ],
    },
    status: "complete",
    expected: [
      {
        label: "label1",
        value: null,
        tags: {},
      },
      {
        label: "label2",
        value: "value2",
        tags: {},
      },
    ],
  },
];

describe("hooks", () => {
  describe("useSelectInputValues", () => {
    test.each(tests)("$name", ({ options, data, status, expected }) => {
      const { result } = renderHook(() =>
        useSelectInputValues(options, data, status)
      );

      expect(result.current).toEqual(expected);
    });
  });
});
