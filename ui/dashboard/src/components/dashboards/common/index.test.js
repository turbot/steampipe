import "../../../test/matchMedia";
import {
  adjustMinValue,
  adjustMaxValue,
  buildNodesAndEdges,
  themeColors,
} from "./index";

describe("common.adjustMinValue", () => {
  test("5", () => {
    expect(adjustMinValue(5)).toEqual(0);
  });

  test("-8", () => {
    expect(adjustMinValue(-8)).toEqual(-9);
  });

  test("-13", () => {
    expect(adjustMinValue(-13)).toEqual(-14);
  });

  test("-20", () => {
    expect(adjustMinValue(-20)).toEqual(-25);
  });

  test("-26", () => {
    expect(adjustMinValue(-26)).toEqual(-30);
  });

  test("-35", () => {
    expect(adjustMinValue(-35)).toEqual(-40);
  });

  test("-50", () => {
    expect(adjustMinValue(-50)).toEqual(-60);
  });

  test("-52", () => {
    expect(adjustMinValue(-52)).toEqual(-60);
  });

  test("-180", () => {
    expect(adjustMinValue(-180)).toEqual(-190);
  });

  test("-210", () => {
    expect(adjustMinValue(-200)).toEqual(-250);
  });

  test("-250", () => {
    expect(adjustMinValue(-250)).toEqual(-300);
  });

  test("-362", () => {
    expect(adjustMinValue(-362)).toEqual(-400);
  });

  test("-1000", () => {
    expect(adjustMinValue(-1000)).toEqual(-1100);
  });

  test("-2363", () => {
    expect(adjustMinValue(-2363)).toEqual(-2400);
  });

  test("-7001", () => {
    expect(adjustMinValue(-7001)).toEqual(-7100);
  });

  test("-10000", () => {
    expect(adjustMinValue(-10000)).toEqual(-11000);
  });

  test("-26526", () => {
    expect(adjustMinValue(-26526)).toEqual(-27000);
  });
});

describe("common.adjustMaxValue", () => {
  test("-5", () => {
    expect(adjustMaxValue(-5)).toEqual(0);
  });

  test("8", () => {
    expect(adjustMaxValue(8)).toEqual(9);
  });

  test("13", () => {
    expect(adjustMaxValue(13)).toEqual(14);
  });

  test("20", () => {
    expect(adjustMaxValue(20)).toEqual(25);
  });

  test("26", () => {
    expect(adjustMaxValue(26)).toEqual(30);
  });

  test("35", () => {
    expect(adjustMaxValue(35)).toEqual(40);
  });

  test("50", () => {
    expect(adjustMaxValue(50)).toEqual(60);
  });

  test("52", () => {
    expect(adjustMaxValue(52)).toEqual(60);
  });

  test("180", () => {
    expect(adjustMaxValue(180)).toEqual(190);
  });

  test("210", () => {
    expect(adjustMaxValue(210)).toEqual(250);
  });

  test("250", () => {
    expect(adjustMaxValue(250)).toEqual(300);
  });

  test("362", () => {
    expect(adjustMaxValue(362)).toEqual(400);
  });

  test("1000", () => {
    expect(adjustMaxValue(1000)).toEqual(1100);
  });

  test("2363", () => {
    expect(adjustMaxValue(2363)).toEqual(2400);
  });

  test("7001", () => {
    expect(adjustMaxValue(7001)).toEqual(7100);
  });

  test("10000", () => {
    expect(adjustMaxValue(10000)).toEqual(11000);
  });

  test("26526", () => {
    expect(adjustMaxValue(26526)).toEqual(27000);
  });
});

describe("common.buildNodesAndEdges", () => {
  test("single node", () => {
    const rawData = {
      columns: [{ name: "id", data_type: "text" }],
      rows: [{ id: "node" }],
    };
    expect(buildNodesAndEdges(rawData)).toEqual({
      categories: {},
      edges: [],
      metadata: { contains_duplicate_edges: false, has_multiple_roots: false },
      next_color_index: 0,
      nodes: [{ id: "node", title: null, category: null, depth: null }],
      root_nodes: {
        node: {
          id: "node",
          title: null,
          category: null,
          depth: null,
        },
      },
    });
  });

  test("single node with from_id", () => {
    const rawData = {
      columns: [
        { name: "id", data_type: "text" },
        { name: "from_id", data_type: "text" },
      ],
      rows: [{ id: "node", from_id: "from_node" }],
    };
    expect(buildNodesAndEdges(rawData)).toEqual({
      categories: {},
      edges: [
        {
          id: "from_node:node",
          from_id: "from_node",
          to_id: "node",
          title: null,
          category: null,
        },
      ],
      metadata: { contains_duplicate_edges: false, has_multiple_roots: false },
      next_color_index: 0,
      nodes: [
        { id: "node", title: null, category: null, depth: null },
        {
          id: "from_node",
          title: null,
          category: null,
          depth: null,
        },
      ],
      root_nodes: {
        from_node: {
          id: "from_node",
          title: null,
          category: null,
          depth: null,
        },
      },
    });
  });

  test("single node with to_id", () => {
    const rawData = {
      columns: [
        { name: "id", data_type: "text" },
        { name: "to_id", data_type: "text" },
      ],
      rows: [{ id: "node", to_id: "to_node" }],
    };
    expect(buildNodesAndEdges(rawData)).toEqual({
      categories: {},
      edges: [
        {
          id: "node:to_node",
          from_id: "node",
          to_id: "to_node",
          title: null,
          category: null,
        },
      ],
      metadata: { contains_duplicate_edges: false, has_multiple_roots: false },
      next_color_index: 0,
      nodes: [
        { id: "node", title: null, category: null, depth: null },
        {
          id: "to_node",
          title: null,
          category: null,
          depth: null,
        },
      ],
      root_nodes: {
        node: {
          id: "node",
          title: null,
          category: null,
          depth: null,
        },
      },
    });
  });

  test("single node with from_id and to_id", () => {
    const rawData = {
      columns: [
        { name: "id", data_type: "text" },
        { name: "from_id", data_type: "text" },
        { name: "to_id", data_type: "text" },
      ],
      rows: [{ id: "node", from_id: "from_node", to_id: "to_node" }],
    };
    expect(buildNodesAndEdges(rawData)).toEqual({
      categories: {},
      edges: [
        {
          id: "from_node:to_node",
          from_id: "from_node",
          to_id: "to_node",
          title: null,
          category: null,
        },
      ],
      metadata: { contains_duplicate_edges: false, has_multiple_roots: true },
      next_color_index: 0,
      nodes: [
        { id: "node", title: null, category: null, depth: null },
        {
          id: "from_node",
          title: null,
          category: null,
          depth: null,
        },
        {
          id: "to_node",
          title: null,
          category: null,
          depth: null,
        },
      ],
      root_nodes: {
        node: {
          id: "node",
          title: null,
          category: null,
          depth: null,
        },
        from_node: {
          id: "from_node",
          title: null,
          category: null,
          depth: null,
        },
      },
    });
  });

  test("two node with separate edge declaration", () => {
    const rawData = {
      columns: [
        { name: "id", data_type: "text" },
        { name: "from_id", data_type: "text" },
        { name: "to_id", data_type: "text" },
      ],
      rows: [
        { id: "from_node" },
        { id: "to_node" },
        { from_id: "from_node", to_id: "to_node" },
      ],
    };
    expect(buildNodesAndEdges(rawData)).toEqual({
      categories: {},
      edges: [
        {
          id: "from_node:to_node",
          from_id: "from_node",
          to_id: "to_node",
          title: null,
          category: null,
        },
      ],
      metadata: { contains_duplicate_edges: false, has_multiple_roots: false },
      next_color_index: 0,
      nodes: [
        { id: "from_node", title: null, category: null, depth: null },
        { id: "to_node", title: null, category: null, depth: null },
      ],
      root_nodes: {
        from_node: {
          id: "from_node",
          title: null,
          category: null,
          depth: null,
        },
      },
    });
  });

  // test("single node with title", () => {
  //   const rawData = {
  //     columns: [
  //       { name: "id", data_type: "text" },
  //       { name: "title", data_type: "text" },
  //     ],
  //     rows: [{ id: "a_node", title: "A Node Title" }],
  //   };
  //   expect(buildNodesAndEdges(rawData)).toEqual({
  //     categories: {},
  //     edges: [],
  //     metadata: { contains_duplicate_edges: false, has_multiple_roots: false },
  //     next_color_index: 0,
  //     nodes: [
  //       { category: null, depth: null, id: "a_node", title: "A Node Title" },
  //     ],
  //     root_nodes: {
  //       a_node: {
  //         category: null,
  //         depth: null,
  //         id: "a_node",
  //         title: "A Node Title",
  //       },
  //     },
  //   });
  // });
  //
  // test("single node with category", () => {
  //   const rawData = {
  //     columns: [
  //       { name: "id", data_type: "text" },
  //       { name: "category", data_type: "text" },
  //     ],
  //     rows: [{ id: "a_node", category: "a_category" }],
  //   };
  //   expect(buildNodesAndEdges(rawData)).toEqual({
  //     categories: {},
  //     edges: [],
  //     metadata: { contains_duplicate_edges: false, has_multiple_roots: false },
  //     next_color_index: 0,
  //     nodes: [
  //       { category: "a_category", depth: null, id: "a_node", title: null },
  //     ],
  //     root_nodes: {
  //       a_node: {
  //         category: "a_category",
  //         depth: null,
  //         id: "a_node",
  //         title: null,
  //       },
  //     },
  //   });
  // });
});
