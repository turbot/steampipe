import { findDimension } from "./index";

test("findDimension with undefined dimensions and key returns undefined", () => {
  expect(findDimension(undefined, undefined)).toBe(undefined);
});

test("findDimension with undefined dimensions returns undefined", () => {
  expect(findDimension(undefined, "some_dimension")).toBe(undefined);
});

test("findDimension with undefined key returns undefined", () => {
  expect(findDimension([], undefined)).toBe(undefined);
});

test("findDimension with missing key returns undefined", () => {
  expect(
    findDimension([{ key: "region", value: "us-east-1" }], "account_id")
  ).toBe(undefined);
});

test("findDimension with present key returns dimension", () => {
  const dimensions = [
    { key: "account_id", value: "123456789012" },
    { key: "region", value: "us-east-1" },
  ];
  expect(findDimension(dimensions, "region")).toBe(dimensions[1]);
});
