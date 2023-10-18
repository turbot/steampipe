import { validateFilter, validateOrFilter, validateAndFilter } from "./"; // Replace with the actual module path

interface TestCase {
  name: string;
  input: any; // Input data for the test
  expected: boolean; // Expected result
}

const filterTestCases: TestCase[] = [
  {
    name: "valid filter with type and key",
    input: { type: "resource", key: "name" },
    expected: true,
  },
  {
    name: "valid filter with type and value",
    input: { type: "tag", value: "production" },
    expected: true,
  },
  {
    name: "filter missing type",
    input: { key: "name" },
    expected: false,
  },
  {
    name: "filter missing both key and value",
    input: { type: "resource" },
    expected: false,
  },
];

const orFilterTestCases: TestCase[] = [
  {
    name: "valid OR filter with valid filters",
    input: {
      or: [
        { type: "resource", value: "*mybucket*" },
        { type: "tag", key: "environment", value: "production" },
      ],
    },
    expected: true,
  },
  {
    name: "empty OR filter",
    input: { or: [] },
    expected: true,
  },
  {
    name: "OR filter with an invalid filter",
    input: {
      or: [{ type: "resource", key: "name" }, { key: "name" }],
    },
    expected: false,
  },
  {
    name: "OR filter with one invalid and one valid filter",
    input: {
      or: [{ type: "resource", value: "*mybucket*" }, { key: "name" }],
    },
    expected: false,
  },
];

const andFilterTestCases: TestCase[] = [
  {
    name: "valid AND filter with valid filters",
    input: {
      and: [
        {
          or: [
            { type: "resource", value: "*mybucket*" },
            { type: "tag", key: "environment", value: "production" },
          ],
        },
        { type: "dimension", key: "region", value: "us*" },
      ],
    },
    expected: true,
  },
  {
    name: "empty AND filter",
    input: { and: [] },
    expected: true,
  },
  {
    name: "AND filter with an invalid filter",
    input: {
      and: [
        {
          or: [{ type: "resource", key: "name" }, { key: "name" }],
        },
        { type: "dimension", key: "region", value: "us*" },
      ],
    },
    expected: false,
  },
];

function runTestCases(
  testCases: TestCase[],
  validationFunction: (input: any) => boolean,
) {
  testCases.forEach((testCase) => {
    it(`should return ${testCase.expected} for ${testCase.name}`, () => {
      const result = validationFunction(testCase.input);
      expect(result).toBe(testCase.expected);
    });
  });
}

describe("Check Filter Validation", () => {
  describe("validateFilter", () => {
    runTestCases(filterTestCases, validateFilter);
  });

  describe("validateOrFilter", () => {
    runTestCases(orFilterTestCases, validateOrFilter);
  });

  describe("validateAndFilter", () => {
    runTestCases(andFilterTestCases, validateAndFilter);
  });
});
