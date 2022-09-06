import { getDashboardIconName } from "./";

describe("common.adjustMinValue", () => {
  test("null returns null", () => {
    expect(getDashboardIconName(null)).toEqual(null);
  });

  test("undefined returns null", () => {
    expect(getDashboardIconName(undefined)).toEqual(null);
  });

  test("existing icon not mapped", () => {
    expect(getDashboardIconName("bell")).toEqual("bell");
  });

  test("existing namespaced icon not mapped", () => {
    expect(getDashboardIconName("heroicons-outline:bell")).toEqual(
      "heroicons-outline:bell"
    );
  });

  test("migrated icon mapped", () => {
    expect(getDashboardIconName("search")).toEqual("magnifying-glass");
  });

  test("prefixed migrated icon mapped", () => {
    expect(getDashboardIconName("heroicons-solid:search")).toEqual(
      "heroicons-solid:magnifying-glass"
    );
  });
});
