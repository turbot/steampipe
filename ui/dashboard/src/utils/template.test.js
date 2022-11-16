import { buildJQFilter } from "./template";

describe("buildJQFilter", () => {
  test("empty", () => {
    expect(buildJQFilter("")).toEqual("");
  });

  test("no interpolated expression", () => {
    expect(buildJQFilter("simple string")).toEqual(
      '(["simple string"] | join(""))'
    );
  });

  test("basic interpolated expression", () => {
    expect(
      buildJQFilter("simple string with {{ .embedded }} expression")
    ).toEqual(
      '(["simple string with ", ( .embedded ), " expression"] | join(""))'
    );
  });

  test("multiple interpolated expressions", () => {
    expect(
      buildJQFilter(
        "simple string with two {{ .embedded }} expressions {{ .in }} there"
      )
    ).toEqual(
      '(["simple string with two ", ( .embedded ), " expressions ", ( .in ), " there"] | join(""))'
    );
  });

  test("two interpolated expressions with no spacing", () => {
    expect(
      buildJQFilter(
        "simple string with two {{ .embedded }}{{ .adjacent }} expressions"
      )
    ).toEqual(
      '(["simple string with two ", ( .embedded ), ( .adjacent ), " expressions"] | join(""))'
    );
  });

  test("three interpolated expressions with no spacing", () => {
    expect(
      buildJQFilter(
        "simple string with three {{ .embedded }}{{ .compact }}{{ .adjacent }} expressions"
      )
    ).toEqual(
      '(["simple string with three ", ( .embedded ), ( .compact ), ( .adjacent ), " expressions"] | join(""))'
    );
  });

  test("replace single quotes with double quotes", () => {
    expect(
      buildJQFilter(
        "simple string with {{ .embedded | 'foo' }} expression using single quotes"
      )
    ).toEqual(
      '(["simple string with ", ( .embedded | "foo" ), " expression using single quotes"] | join(""))'
    );
  });

  test("ignore unicode single quote", () => {
    expect(
      buildJQFilter(
        "simple string with {{ .embedded | 'what\\u0027s this?' }} expression using plain + unicode single quotes"
      )
    ).toEqual(
      `(["simple string with ", ( .embedded | "what\\u0027s this?" ), " expression using plain + unicode single quotes"] | join(""))`
    );
  });

  test("ignore unicode single quote", () => {
    expect(
      buildJQFilter(
        "simple string with {{ .embedded | 'what\\u0027s this?' }} expression using plain + unicode single quotes"
      )
    ).toEqual(
      `(["simple string with ", ( .embedded | "what\\u0027s this?" ), " expression using plain + unicode single quotes"] | join(""))`
    );
  });

  test("ignore additional open interpolation braces if in expression", () => {
    expect(buildJQFilter('string with nested {{ "{{" }} braces')).toEqual(
      `(["string with nested ", ( "{{" ), " braces"] | join(""))`
    );
  });
});
