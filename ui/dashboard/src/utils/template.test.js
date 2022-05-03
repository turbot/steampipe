import { buildJQFilter } from "./template";

describe("buildJQFilter", () => {
  test("empty", () => {
    expect(buildJQFilter("")).toEqual("");
  });

  test("no interpolated expression", () => {
    expect(buildJQFilter("simple string")).toEqual('("simple string")');
  });

  test("basic interpolated expression", () => {
    expect(
      buildJQFilter("simple string with {{ .embedded }} expression")
    ).toEqual('("simple string with " + ( .embedded ) + " expression")');
  });

  test("multiple interpolated expressions", () => {
    expect(
      buildJQFilter(
        "simple string with two {{ .embedded }} expressions {{ .in }} there"
      )
    ).toEqual(
      '("simple string with two " + ( .embedded ) + " expressions " + ( .in ) + " there")'
    );
  });

  test("replace single quotes with double quotes", () => {
    expect(
      buildJQFilter(
        "simple string with {{ .embedded | 'foo' }} expression using single quotes"
      )
    ).toEqual(
      '("simple string with " + ( .embedded | "foo" ) + " expression using single quotes")'
    );
  });

  test("ignore escaped single quotes", () => {
    expect(
      buildJQFilter(
        "simple string with {{ .embedded | 'foo' | \\'ignore\\' }} expression using unescaped and escaped single quotes"
      )
    ).toEqual(
      `("simple string with " + ( .embedded | "foo" | \\'ignore\\' ) + " expression using unescaped and escaped single quotes")`
    );
  });
});
