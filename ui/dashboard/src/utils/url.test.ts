import { isRelativeUrl } from "./url";

describe("isRelativeUrl", () => {
  test("null", () => {
    expect(isRelativeUrl(null)).toEqual(true);
  });

  test("undefined", () => {
    expect(isRelativeUrl(undefined)).toEqual(true);
  });

  test("empty", () => {
    expect(isRelativeUrl("")).toEqual(true);
  });

  test("relative", () => {
    expect(isRelativeUrl("/foo/bar")).toEqual(true);
  });

  test("relative with query string", () => {
    expect(isRelativeUrl("/foo/bar?bar=foo")).toEqual(true);
  });

  test("absolute", () => {
    expect(isRelativeUrl("https://foo.bar")).toEqual(false);
  });
});
